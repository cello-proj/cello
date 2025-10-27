//go:generate moq -out ../../test/testhelpers/dbClientMock.go -pkg testhelpers . Client:DBClientMock

package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cello-proj/cello/internal/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type ProjectEntry struct {
	ProjectID  string `db:"project"`
	Repository string `db:"repository"`
}

type TokenEntry struct {
	CreatedAt string `db:"created_at"`
	ExpiresAt string `db:"expires_at"`
	ProjectID string `db:"project"`
	TokenID   string `db:"token_id"`
}

// IsEmpty returns whether a struct is empty.
func (t TokenEntry) IsEmpty() bool {
	return t == (TokenEntry{})
}

// Client allows for db crud operations
type Client interface {
	CreateProjectEntry(ctx context.Context, pe ProjectEntry) error
	DeleteProjectEntry(ctx context.Context, project string) error
	ReadProjectEntry(ctx context.Context, project string) (ProjectEntry, error)
	CreateTokenEntry(ctx context.Context, token types.Token) error
	DeleteTokenEntry(ctx context.Context, token string) error
	// This only exists for dynamodb, as the token ID is the sort key which also requires the project ID as the primary key.
	DeleteTokenEntryByProject(ctx context.Context, project, token string) error
	ReadTokenEntry(ctx context.Context, token string) (TokenEntry, error)
	// This only exists for dynamodb, as the token ID is the sort key which also requires the project ID as the primary key.
	ReadTokenEntryByProject(ctx context.Context, project, token string) (TokenEntry, error)
	ListTokenEntries(ctx context.Context, project string) ([]TokenEntry, error)
	Health(ctx context.Context) error
}

// Verify interface implementations at compile time
var (
	_ Client = (*DynamoDBClient)(nil)
)

const (
	ProjectEntryDB = "projects"
	TokenEntryDB   = "tokens"
)

// DynamoDBSvc defines the interface for DynamoDB operations
type DynamoDBSvc interface {
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDBClient allows for db crud operations using dynamodb
type DynamoDBClient struct {
	svc       DynamoDBSvc
	tableName string
}

const (
	primaryKey = "pk"
	sortKey    = "sk"

	projectPKFmt = "PROJECT#%s"
	metadataSK   = "METADATA"
	tokenSKFmt   = "TOKEN#%s"
)

var (
	ErrProjectNotFound = fmt.Errorf("project not found")
	ErrTokenNotFound   = fmt.Errorf("token not found")
)

func NewDynamoDBClient(tableName string, endpointURL string, assumeRoleARN string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Setup assume role if provided
	if assumeRoleARN != "" {
		stsClient := sts.NewFromConfig(cfg)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithCredentialsProvider(stscreds.NewAssumeRoleProvider(
				stsClient,
				assumeRoleARN,
				func(options *stscreds.AssumeRoleOptions) {
					options.RoleSessionName = "cello-session"
				}),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load assume role config: %w", err)
		}
	}

	svc := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpointURL != "" {
			o.BaseEndpoint = aws.String(endpointURL)
		}
	})

	return &DynamoDBClient{
		svc:       svc,
		tableName: tableName,
	}, nil
}

func (d *DynamoDBClient) Health(ctx context.Context) error {
	// No-op implementation to avoid unnecessary API calls
	return nil
}

func (d *DynamoDBClient) CreateProjectEntry(ctx context.Context, pe ProjectEntry) error {
	item := map[string]ddbtypes.AttributeValue{
		primaryKey:   &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(projectPKFmt, pe.ProjectID)},
		sortKey:      &ddbtypes.AttributeValueMemberS{Value: metadataSK},
		"repository": &ddbtypes.AttributeValueMemberS{Value: pe.Repository},
	}

	_, err := d.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	if err != nil {
		var ccf *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("project %s already exists", pe.ProjectID)
		}
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

func (d *DynamoDBClient) ReadProjectEntry(ctx context.Context, project string) (ProjectEntry, error) {
	result, err := d.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			primaryKey: &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(projectPKFmt, project)},
			sortKey:    &ddbtypes.AttributeValueMemberS{Value: metadataSK},
		},
	})
	if err != nil {
		return ProjectEntry{}, err
	}

	if result.Item == nil {
		return ProjectEntry{}, ErrProjectNotFound
	}

	repo, ok := result.Item["repository"].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return ProjectEntry{}, fmt.Errorf("invalid repository attribute")
	}

	return ProjectEntry{
		ProjectID:  project,
		Repository: repo.Value,
	}, nil
}

func (d *DynamoDBClient) DeleteProjectEntry(ctx context.Context, project string) error {
	projectPK := fmt.Sprintf(projectPKFmt, project)

	// First, query all items associated with this project (metadata, tokens, etc.)
	var allItems []map[string]ddbtypes.AttributeValue

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk": &ddbtypes.AttributeValueMemberS{Value: projectPK},
		},
	}

	// Handle pagination to get all items
	for {
		result, err := d.svc.Query(ctx, queryInput)
		if err != nil {
			return fmt.Errorf("failed to query project items: %w", err)
		}

		allItems = append(allItems, result.Items...)

		// Check if there are more pages
		if result.LastEvaluatedKey == nil {
			break
		}

		// Set the exclusive start key for the next page
		queryInput.ExclusiveStartKey = result.LastEvaluatedKey
	}

	if len(allItems) == 0 {
		// No items to delete, project doesn't exist
		return nil
	}

	// Prepare batch delete requests
	var writeRequests []ddbtypes.WriteRequest
	for _, item := range allItems {
		sk, ok := item[sortKey].(*ddbtypes.AttributeValueMemberS)
		if !ok {
			continue
		}

		writeRequests = append(writeRequests, ddbtypes.WriteRequest{
			DeleteRequest: &ddbtypes.DeleteRequest{
				Key: map[string]ddbtypes.AttributeValue{
					primaryKey: &ddbtypes.AttributeValueMemberS{Value: projectPK},
					sortKey:    &ddbtypes.AttributeValueMemberS{Value: sk.Value},
				},
			},
		})
	}

	// Delete items in batches of 25 (DynamoDB limit)
	const batchSize = 25
	const delay = 100 * time.Millisecond
	const maxRetries = 5

	for i := 0; i < len(writeRequests); i += batchSize {
		end := min(i+batchSize, len(writeRequests))

		batch := writeRequests[i:end]
		batchInput := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]ddbtypes.WriteRequest{
				d.tableName: batch,
			},
		}

		// Delete this batch of items and retry if there are unprocessed items
		retryCount := 0
		for retryCount < maxRetries {
			result, err := d.svc.BatchWriteItem(ctx, batchInput)
			if err != nil {
				return fmt.Errorf("failed to batch delete project items: %w", err)
			}

			// If there are no unprocessed items, we're done
			if len(result.UnprocessedItems) == 0 {
				break
			}

			retryCount++
			if retryCount >= maxRetries {
				return fmt.Errorf("failed to delete all project items after %d retries", maxRetries)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}

			batchInput.RequestItems = result.UnprocessedItems
		}
	}

	return nil
}

func (d *DynamoDBClient) CreateTokenEntry(ctx context.Context, token types.Token) error {
	item := map[string]ddbtypes.AttributeValue{
		primaryKey:   &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(projectPKFmt, token.ProjectID)},
		sortKey:      &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(tokenSKFmt, token.ProjectToken.ID)},
		"created_at": &ddbtypes.AttributeValueMemberS{Value: token.CreatedAt},
		"expires_at": &ddbtypes.AttributeValueMemberS{Value: token.ExpiresAt},
	}

	_, err := d.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(sk)"),
	})
	if err != nil {
		var ccf *ddbtypes.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return fmt.Errorf("token %s already exists for project %s", token.ProjectToken.ID, token.ProjectID)
		}
		return fmt.Errorf("failed to create token: %w", err)
	}
	return nil
}

func (d *DynamoDBClient) ReadTokenEntryByProject(ctx context.Context, project, token string) (TokenEntry, error) {
	result, err := d.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			primaryKey: &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(projectPKFmt, project)},
			sortKey:    &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(tokenSKFmt, token)},
		},
	})
	if err != nil {
		return TokenEntry{}, fmt.Errorf("failed to get token: %w", err)
	}

	if result.Item == nil {
		return TokenEntry{}, ErrTokenNotFound
	}

	return d.parseTokenFromItem(result.Item, project)
}

// ReadTokenEntry should not be used. Use ReadTokenEntryByProject instead.
// This is due to a difference in how the data is stored in postgres vs dynamodb.
// In postgres, the token ID is the primary key, while in dynamodb, the token ID is
// the sort key which requires the project ID as the primary key.
func (d *DynamoDBClient) ReadTokenEntry(ctx context.Context, token string) (TokenEntry, error) {
	return TokenEntry{}, fmt.Errorf("ReadTokenEntry should not be used. Use ReadTokenEntryByProject instead")
}

// DeleteTokenEntry should not be used. Use DeleteTokenEntryByProject instead.
// This is due to a difference in how the data is stored in postgres vs dynamodb.
// In postgres, the token ID is the primary key, while in dynamodb, the token ID is
// the sort key which requires the project ID as the primary key.
func (d *DynamoDBClient) DeleteTokenEntry(ctx context.Context, token string) error {
	return fmt.Errorf("DeleteTokenEntry should not be used. Use DeleteTokenEntryByProject instead")
}

func (d *DynamoDBClient) DeleteTokenEntryByProject(ctx context.Context, project, token string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]ddbtypes.AttributeValue{
			primaryKey: &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(projectPKFmt, project)},
			sortKey:    &ddbtypes.AttributeValueMemberS{Value: fmt.Sprintf(tokenSKFmt, token)},
		},
	}

	if _, err := d.svc.DeleteItem(ctx, input); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

func (d *DynamoDBClient) ListTokenEntries(ctx context.Context, project string) ([]TokenEntry, error) {
	projectPK := fmt.Sprintf(projectPKFmt, project)
	tokenSKPrefix := fmt.Sprintf(tokenSKFmt, "")

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(d.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk_prefix)"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk":        &ddbtypes.AttributeValueMemberS{Value: projectPK},
			":sk_prefix": &ddbtypes.AttributeValueMemberS{Value: tokenSKPrefix},
		},
	}

	result, err := d.svc.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}

	tokens := []TokenEntry{}
	for _, item := range result.Items {
		token, err := d.parseTokenFromItem(item, project)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// parseTokenFromItem converts a DynamoDB item to a TokenEntry
func (d *DynamoDBClient) parseTokenFromItem(item map[string]ddbtypes.AttributeValue, project string) (TokenEntry, error) {
	tokenSKPrefix := fmt.Sprintf(tokenSKFmt, "")

	// Extract the actual token ID from the sort key
	sk, ok := item[sortKey].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return TokenEntry{}, fmt.Errorf("invalid sort key attribute")
	}

	tokenID := strings.TrimPrefix(sk.Value, tokenSKPrefix)

	createdAt, ok := item["created_at"].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return TokenEntry{}, fmt.Errorf("invalid created_at attribute")
	}

	expiresAt, ok := item["expires_at"].(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return TokenEntry{}, fmt.Errorf("invalid expires_at attribute")
	}

	return TokenEntry{
		CreatedAt: createdAt.Value,
		ExpiresAt: expiresAt.Value,
		ProjectID: project,
		TokenID:   tokenID,
	}, nil
}
