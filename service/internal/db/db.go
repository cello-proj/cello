//go:generate moq -out ../../test/testhelpers/dbClientMock.go -pkg testhelpers . Client:DBClientMock

package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/cello-proj/cello/internal/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
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
	ReadTokenEntry(ctx context.Context, token string) (TokenEntry, error)
	ListTokenEntries(ctx context.Context, project string) ([]TokenEntry, error)
	Health(ctx context.Context) error
}

// Verify interface implementations at compile time
var (
	_ Client = SQLClient{}
	_ Client = (*DynamoDBClient)(nil)
)

// SQLClient allows for db crud operations using postgres db
type SQLClient struct {
	host     string
	database string
	user     string
	password string
	options  map[string]string
}

const (
	ProjectEntryDB = "projects"
	TokenEntryDB   = "tokens"
)

func NewSQLClient(host, database, user, password string, options map[string]string) (SQLClient, error) {
	return SQLClient{
		host:     host,
		database: database,
		user:     user,
		password: password,
		options:  options,
	}, nil
}

func (d SQLClient) createSession() (db.Session, error) {
	settings := postgresql.ConnectionURL{
		Host:     d.host,
		Database: d.database,
		User:     d.user,
		Password: d.password,
		Options:  d.options,
	}

	return postgresql.Open(settings)
}

func (d SQLClient) Health(ctx context.Context) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.WithContext(ctx).Ping()
}

func (d SQLClient) CreateProjectEntry(ctx context.Context, pe ProjectEntry) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.WithContext(ctx).Tx(func(sess db.Session) error {
		if err := sess.Collection(ProjectEntryDB).Find("project", pe.ProjectID).Delete(); err != nil {
			return err
		}

		if _, err = sess.Collection(ProjectEntryDB).Insert(pe); err != nil {
			return err
		}

		return nil
	})
}

func (d SQLClient) ReadProjectEntry(ctx context.Context, project string) (ProjectEntry, error) {
	res := ProjectEntry{}

	sess, err := d.createSession()
	if err != nil {
		return res, err
	}
	defer sess.Close()

	err = sess.WithContext(ctx).Collection(ProjectEntryDB).Find("project", project).One(&res)
	return res, err
}

func (d SQLClient) DeleteProjectEntry(ctx context.Context, project string) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.WithContext(ctx).Collection(ProjectEntryDB).Find("project", project).Delete()
}

func (d SQLClient) CreateTokenEntry(ctx context.Context, token types.Token) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	err = sess.WithContext(ctx).Tx(func(sess db.Session) error {
		res := TokenEntry{
			CreatedAt: token.CreatedAt,
			ExpiresAt: token.ExpiresAt,
			ProjectID: token.ProjectID,
			TokenID:   token.ProjectToken.ID,
		}

		if _, err = sess.Collection(TokenEntryDB).Insert(res); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (d SQLClient) DeleteTokenEntry(ctx context.Context, token string) error {
	sess, err := d.createSession()
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.WithContext(ctx).Collection(TokenEntryDB).Find("token_id", token).Delete()
}

func (d SQLClient) ReadTokenEntry(ctx context.Context, token string) (TokenEntry, error) {
	res := TokenEntry{}
	sess, err := d.createSession()
	if err != nil {
		return res, err
	}
	defer sess.Close()

	err = sess.WithContext(ctx).Collection(TokenEntryDB).Find("token_id", token).One(&res)
	return res, err
}

func (d SQLClient) ListTokenEntries(ctx context.Context, project string) ([]TokenEntry, error) {
	res := []TokenEntry{}

	sess, err := d.createSession()
	if err != nil {
		return res, err
	}
	defer sess.Close()

	err = sess.WithContext(ctx).Collection(TokenEntryDB).Find("project", project).OrderBy("-created_at").All(&res)
	return res, err
}

// DynamoDBSvc defines the interface for DynamoDB operations
type DynamoDBSvc interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
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

func NewDynamoDBClient(tableName string, endpointURL string, sqlClient Client) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
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
		return ProjectEntry{}, fmt.Errorf("project not found")
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
	// No-op implementation
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

func (d *DynamoDBClient) ReadTokenEntry(ctx context.Context, token string) (TokenEntry, error) {
	// No-op implementation
	return TokenEntry{}, nil
}

func (d *DynamoDBClient) DeleteTokenEntry(ctx context.Context, token string) error {
	// No-op implementation
	return nil
}

func (d *DynamoDBClient) ListTokenEntries(ctx context.Context, project string) ([]TokenEntry, error) {
	// No-op implementation
	return []TokenEntry{}, nil
}
