//go:generate moq -out ../../test/testhelpers/credsProviderMock.go -pkg testhelpers . Provider:CredsProviderMock

package credentials

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cello-proj/cello/internal/responses"
	"github.com/cello-proj/cello/internal/types"
	"github.com/cello-proj/cello/internal/validations"
	"github.com/cello-proj/cello/service/internal/env"

	vault "github.com/hashicorp/vault/api"
)

const (
	authorizationKeyAdmin = "admin"
)

// Provider defines the interface required by providers.
type Provider interface {
	CreateProject(string) (string, string, string, error)
	CreateTarget(string, types.Target) error
	CreateToken(string) (string, string, string, error)
	UpdateTarget(string, types.Target) error
	DeleteProject(string) error
	DeleteTarget(string, string) error
	GetProject(string) (responses.GetProject, error)
	GetTarget(string, string) (types.Target, error)
	GetToken() (string, error)
	DeleteProjectToken(string, string) error
	GetProjectToken(string, string) (types.ProjectToken, error)
	ListTargets(string) ([]string, error)
	ProjectExists(string) (bool, error)
	TargetExists(string, string) (bool, error)
}

type vaultLogical interface {
	Delete(path string) (*vault.Secret, error)
	List(path string) (*vault.Secret, error)
	Read(path string) (*vault.Secret, error)
	Write(path string, data map[string]interface{}) (*vault.Secret, error)
}

type vaultSys interface {
	DeletePolicy(name string) error
	PutPolicy(name, rules string) error
}

// Vault
const (
	vaultAppRolePrefix = "auth/approle/role"
	vaultProjectPrefix = "argo-cloudops-projects"
)

var (
	// ErrNotFound conveys that the item was not found.
	ErrNotFound = errors.New("item not found")
	// ErrTargetNotFound conveys that the target was not round.
	ErrTargetNotFound = errors.New("target not found")
)

type VaultProvider struct {
	roleID          string
	secretID        string
	vaultLogicalSvc vaultLogical
	vaultSysSvc     vaultSys
}

// NewVaultProvider returns a new VaultProvider
func NewVaultProvider(a Authorization, env env.Vars, h http.Header, vaultConfigFn VaultConfigFn, vaultSvcFn VaultSvcFn) (Provider, error) {
	config := vaultConfigFn(&vault.Config{Address: env.VaultAddress}, env.VaultRole, env.VaultSecret)
	svc, err := vaultSvcFn(*config, h)
	if err != nil {
		return nil, err
	}
	return &VaultProvider{
		vaultLogicalSvc: vaultLogical(svc.Logical()),
		vaultSysSvc:     vaultSys(svc.Sys()),
		roleID:          a.Key,
		secretID:        a.Secret,
	}, nil
}

type VaultConfig struct {
	config *vault.Config
	role   string
	secret string
}

type VaultConfigFn func(config *vault.Config, role, secret string) *VaultConfig

// NewVaultConfig returns a new VaultConfig.
func NewVaultConfig(config *vault.Config, role, secret string) *VaultConfig {
	return &VaultConfig{
		config: config,
		role:   role,
		secret: secret,
	}
}

type VaultSvcFn func(c VaultConfig, h http.Header) (svc *vault.Client, err error)

// NewVaultSvc returns a new vault.Client.
// TODO before open sourcing we should provide the token instead of generating it
// TODO rename to client?
func NewVaultSvc(c VaultConfig, h http.Header) (*vault.Client, error) {
	vaultSvc, err := vault.NewClient(c.config)
	if err != nil {
		return nil, err
	}

	vaultSvc.SetHeaders(h)

	options := map[string]interface{}{
		"role_id":   c.role,
		"secret_id": c.secret,
	}

	sec, err := vaultSvc.Logical().Write("auth/approle/login", options)
	if err != nil {
		return nil, err
	}

	vaultSvc.SetToken(sec.Auth.ClientToken)
	return vaultSvc, nil
}

// Authorization represents a user's authorization token.
type Authorization struct {
	Provider string `valid:"required"`
	Key      string `valid:"required"`
	Secret   string `valid:"required"`
}

func (a Authorization) Validate(optionalValidations ...func() error) error {
	v := []func() error{
		func() error {
			if a.Provider != "vault" {
				return errors.New("provider must be vault")
			}
			return nil
		},
		func() error { return validations.ValidateStruct(a) },
	}

	v = append(v, optionalValidations...)

	return validations.Validate(v...)
}

// ValidateAuthorizedAdmin determines if the Authorization is valid and an admin.
// TODO See if this can be removed when refactoring auth.
// Optional validation should be passed as parameter to Validate().
func (a Authorization) ValidateAuthorizedAdmin(adminSecret string) func() error {
	return func() error {
		if a.Key != "admin" {
			return fmt.Errorf("must be an authorized admin")
		}

		if a.Secret != adminSecret {
			return fmt.Errorf("must be an authorized admin, invalid admin secret")
		}

		return nil
	}
}

// NewAuthorization provides an Authorization from a header.
// This is separate from admin functions which use the admin env var
func NewAuthorization(authorizationHeader string) (*Authorization, error) {
	var a Authorization
	auth := strings.SplitN(authorizationHeader, ":", 3)
	if len(auth) < 3 {
		return nil, fmt.Errorf("invalid authorization header")
	}
	a.Provider = auth[0]
	a.Key = auth[1]
	a.Secret = auth[2]
	return &a, nil
}

func (v VaultProvider) createPolicyState(name, policy string) error {
	return v.vaultSysSvc.PutPolicy(fmt.Sprintf("%s-%s", vaultProjectPrefix, name), policy)
}

func genProjectAppRole(name string) string {
	return fmt.Sprintf("%s/%s-%s", vaultAppRolePrefix, vaultProjectPrefix, name)
}

func (v VaultProvider) CreateToken(name string) (string, string, string, error) {
	if !v.isAdmin() {
		return "", "", "", errors.New("admin credentials must be used to create token")
	}

	secretID, secretAccessor, err := v.generateSecrets(name)
	if err != nil {
		return "", "", "", err
	}

	roleID, err := v.readRoleID(name)
	if err != nil {
		return "", "", "", err
	}

	return roleID, secretID, secretAccessor, nil
}

func (v VaultProvider) CreateProject(name string) (string, string, string, error) {
	if !v.isAdmin() {
		return "", "", "", errors.New("admin credentials must be used to create project")
	}

	policy := defaultVaultReadonlyPolicyAWS(name)
	err := v.createPolicyState(name, policy)
	if err != nil {
		return "", "", "", err
	}

	if err := v.writeProjectState(name); err != nil {
		return "", "", "", err
	}

	return v.CreateToken(name)
}

// CreateTarget creates a target for the project.
// TODO validate policy and other information is correct in target
// TODO Validate role exists (if possible, etc)
func (v VaultProvider) CreateTarget(projectName string, target types.Target) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to create target")
	}

	options := map[string]interface{}{
		"credential_type": target.Properties.CredentialType,
		"policy_arns":     target.Properties.PolicyArns,
		"policy_document": target.Properties.PolicyDocument,
		"role_arns":       target.Properties.RoleArn,
	}

	path := fmt.Sprintf("aws/roles/%s-%s-target-%s", vaultProjectPrefix, projectName, target.Name)
	_, err := v.vaultLogicalSvc.Write(path, options)
	return err
}

func defaultVaultReadonlyPolicyAWS(projectName string) string {
	return fmt.Sprintf(
		"path \"aws/sts/argo-cloudops-projects-%s-target-*\" { capabilities = [\"read\"] }",
		projectName,
	)
}

func (v VaultProvider) deletePolicyState(name string) error {
	return v.vaultSysSvc.DeletePolicy(fmt.Sprintf("%s-%s", vaultProjectPrefix, name))
}

func (v VaultProvider) DeleteProject(name string) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to delete project")
	}

	err := v.deletePolicyState(name)
	if err != nil {
		return fmt.Errorf("vault delete project error: %w", err)
	}

	if _, err = v.vaultLogicalSvc.Delete(genProjectAppRole(name)); err != nil {
		return fmt.Errorf("vault delete project error: %w", err)
	}
	return nil
}

func (v VaultProvider) DeleteTarget(projectName string, targetName string) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to delete target")
	}

	path := fmt.Sprintf("aws/roles/%s-%s-target-%s", vaultProjectPrefix, projectName, targetName)
	_, err := v.vaultLogicalSvc.Delete(path)
	return err
}

const (
	vaultSecretTTL   = "8776h" // 1 year
	vaultTokenMaxTTL = "10m"
	// When set to 1 with the cli or api, it will not return the creds as it
	// says it's hit the limit of uses.
	vaultTokenNumUses = 3
)

func (v VaultProvider) GetProject(projectName string) (responses.GetProject, error) {
	sec, err := v.vaultLogicalSvc.Read(genProjectAppRole(projectName))
	if err != nil {
		return responses.GetProject{}, fmt.Errorf("vault get project error: %w", err)
	}
	if sec == nil {
		return responses.GetProject{}, ErrNotFound
	}

	return responses.GetProject{Name: projectName}, nil
}

func (v VaultProvider) GetTarget(projectName, targetName string) (types.Target, error) {
	if !v.isAdmin() {
		return types.Target{}, errors.New("admin credentials must be used to get target information")
	}

	sec, err := v.vaultLogicalSvc.Read(fmt.Sprintf("aws/roles/argo-cloudops-projects-%s-target-%s", projectName, targetName))
	if err != nil {
		return types.Target{}, fmt.Errorf("vault get target error: %w", err)
	}

	if sec == nil {
		return types.Target{}, ErrTargetNotFound
	}

	// These should always exist.
	roleArn := sec.Data["role_arns"].([]interface{})[0].(string)
	credentialType := sec.Data["credential_type"].(string)

	// Optional.
	policies := []string{}
	if val, ok := sec.Data["policy_arns"]; ok {
		for _, v := range val.([]interface{}) {
			policies = append(policies, v.(string))
		}
	}

	// Optional.
	var policyDocument string
	if val, ok := sec.Data["policy_document"]; ok {
		policyDocument = val.(string)
	}

	return types.Target{
		Name: targetName,
		// target 'Type' always 'aws_account', currently not stored in Vault
		Type: "aws_account",
		Properties: types.TargetProperties{
			CredentialType: credentialType,
			PolicyArns:     policies,
			PolicyDocument: policyDocument,
			RoleArn:        roleArn,
		},
	}, nil
}

func (v VaultProvider) DeleteProjectToken(projectName, tokenID string) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to delete tokens")
	}

	data := map[string]interface{}{
		"secret_id_accessor": tokenID,
	}

	path := fmt.Sprintf("%s/secret-id-accessor/destroy", genProjectAppRole(projectName))
	_, err := v.vaultLogicalSvc.Write(path, data)
	if err != nil {
		return err
	}

	return nil
}

func (v VaultProvider) GetProjectToken(projectName, tokenID string) (types.ProjectToken, error) {
	token := types.ProjectToken{}

	if !v.isAdmin() {
		return token, errors.New("admin credentials must be used to delete tokens")
	}

	data := map[string]interface{}{
		"secret_id_accessor": tokenID,
	}

	path := fmt.Sprintf("%s/secret-id-accessor/lookup", genProjectAppRole(projectName))
	projectToken, err := v.vaultLogicalSvc.Write(path, data)
	if err != nil {
		return token, fmt.Errorf("vault get secret ID accessor error: %w", err)
	}

	if projectToken == nil {
		return token, nil
	}

	return types.ProjectToken{
		ID: projectToken.Data["secret_id_accessor"].(string),
	}, nil
}

func (v VaultProvider) GetToken() (string, error) {
	if v.isAdmin() {
		return "", errors.New("admin credentials cannot be used to get tokens")
	}

	options := map[string]interface{}{
		"role_id":   v.roleID,
		"secret_id": v.secretID,
	}

	sec, err := v.vaultLogicalSvc.Write("auth/approle/login", options)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return sec.Auth.ClientToken, nil
}

// TODO See if this can be removed when refactoring auth.
func (v VaultProvider) isAdmin() bool {
	return v.roleID == authorizationKeyAdmin
}

func (v VaultProvider) ListTargets(project string) ([]string, error) {
	if !v.isAdmin() {
		return nil, errors.New("admin credentials must be used to list targets")
	}

	sec, err := v.vaultLogicalSvc.List("aws/roles/")
	if err != nil {
		return nil, fmt.Errorf("vault list error: %w", err)
	}

	// allow empty array to render json as []
	list := make([]string, 0)
	if sec != nil {
		for _, target := range sec.Data["keys"].([]interface{}) {
			value := target.(string)
			prefix := fmt.Sprintf("argo-cloudops-projects-%s-target-", project)
			if strings.HasPrefix(value, prefix) {
				list = append(list, strings.Replace(value, prefix, "", 1))
			}
		}
	}

	return list, nil
}

func (v VaultProvider) ProjectExists(name string) (bool, error) {
	p, err := v.GetProject(name)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return p.Name != "", nil
}

func (v VaultProvider) readRoleID(appRoleName string) (string, error) {
	secret, err := v.vaultLogicalSvc.Read(fmt.Sprintf("%s/role-id", genProjectAppRole(appRoleName)))
	if err != nil {
		return "", err
	}
	return secret.Data["role_id"].(string), nil
}

func (v VaultProvider) generateSecrets(appRoleName string) (string, string, error) {
	options := map[string]interface{}{
		"force": true,
	}
	secret, err := v.vaultLogicalSvc.Write(fmt.Sprintf("%s/secret-id", genProjectAppRole(appRoleName)), options)
	if err != nil {
		return "", "", err
	}
	return secret.Data["secret_id"].(string), secret.Data["secret_id_accessor"].(string), nil
}

func (v VaultProvider) TargetExists(projectName, targetName string) (bool, error) {
	_, err := v.GetTarget(projectName, targetName)
	return !errors.Is(err, ErrTargetNotFound), nil
}

// UpdateTarget updates a targets policies for the project.
func (v VaultProvider) UpdateTarget(projectName string, target types.Target) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to update target")
	}

	options := map[string]interface{}{
		"credential_type": target.Properties.CredentialType,
		"policy_arns":     target.Properties.PolicyArns,
		"policy_document": target.Properties.PolicyDocument,
		"role_arns":       target.Properties.RoleArn,
	}

	path := fmt.Sprintf("aws/roles/%s-%s-target-%s", vaultProjectPrefix, projectName, target.Name)
	_, err := v.vaultLogicalSvc.Write(path, options)
	return err
}

func (v VaultProvider) writeProjectState(name string) error {
	options := map[string]interface{}{
		"secret_id_ttl":           vaultSecretTTL,
		"token_max_ttl":           vaultTokenMaxTTL,
		"token_no_default_policy": "true",
		"token_num_uses":          vaultTokenNumUses,
		"token_policies":          fmt.Sprintf("%s-%s", vaultProjectPrefix, name),
	}

	_, err := v.vaultLogicalSvc.Write(genProjectAppRole(name), options)
	if err != nil {
		return err
	}
	return nil
}
