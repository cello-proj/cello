package credentials

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argo-cloudops/internal/requests"
	"github.com/argoproj-labs/argo-cloudops/internal/responses"
	"github.com/argoproj-labs/argo-cloudops/service/internal/validations"

	vault "github.com/hashicorp/vault/api"
)

const (
	authorizationKeyAdmin = "admin"
)

// Provider defines the interface required by providers.
type Provider interface {
	CreateProject(string) (string, string, error)
	CreateTarget(string, requests.CreateTargetRequest) error
	DeleteProject(string) error
	DeleteTarget(string, string) error
	GetProject(string) (responses.GetProject, error)
	GetTarget(string, string) (requests.TargetProperties, error)
	GetToken() (string, error)
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
func NewVaultProvider(a Authorization, svc *vault.Client) (Provider, error) {
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

// NewVaultConfig returns a new VaultConfig.
func NewVaultConfig(config *vault.Config, role, secret string) *VaultConfig {
	return &VaultConfig{
		config: config,
		role:   role,
		secret: secret,
	}
}

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
	Provider string
	Key      string
	Secret   string
}

// NewAuthorization provides an Authorization from a header.
// This is separate from admin functions which use the admin env var
func NewAuthorization(authorizationHeader string) (*Authorization, error) {
	var a Authorization
	if err := validations.InitValidator().Var(authorizationHeader, "valid_auth_header"); err != nil {
		return nil, fmt.Errorf("invalid authorization header format")
	}
	auth := strings.SplitN(authorizationHeader, ":", 3)
	a.Provider = auth[0]
	a.Key = auth[1]
	a.Secret = auth[2]
	return &a, nil
}

// IsAdmin determines if the Authorization is an admin.
// TODO See if this can be removed when refactoring auth.
// AuthorizedAdmin determines if the Authorization is valid and an Admin.
func (a Authorization) ValidateAuthorizedAdmin(adminSecret string) error {
	if err := validations.InitValidator().Var(a.Key, "eq=admin"); err != nil {
		return fmt.Errorf("must be an authorized admin, %s", validations.VarValidationErrors("user", err))
	}
	if err := validations.InitValidator().Var(a.Secret, fmt.Sprintf("eq=%s", adminSecret)); err != nil {
		return fmt.Errorf("must be an authorized admin, invalid admin secret")
	}
	return nil
}

func (v VaultProvider) createPolicyState(name, policy string) error {
	return v.vaultSysSvc.PutPolicy(fmt.Sprintf("%s-%s", vaultProjectPrefix, name), policy)
}

func genProjectAppRole(name string) string {
	return fmt.Sprintf("%s/%s-%s", vaultAppRolePrefix, vaultProjectPrefix, name)
}

func (v VaultProvider) CreateProject(name string) (string, string, error) {
	if !v.isAdmin() {
		return "", "", errors.New("admin credentials must be used to create project")
	}

	policy := defaultVaultReadonlyPolicyAWS(name)
	err := v.createPolicyState(name, policy)
	if err != nil {
		return "", "", err
	}

	if err := v.writeProjectState(name); err != nil {
		return "", "", err
	}

	secretID, err := v.readSecretID(name)
	if err != nil {
		return "", "", err
	}

	roleID, err := v.readRoleID(name)
	if err != nil {
		return "", "", err
	}

	return roleID, secretID, nil
}

// CreateTarget creates a target for the project.
// TODO validate policy and other information is correct in target
// TODO Validate role exists (if possible, etc)
func (v VaultProvider) CreateTarget(projectName string, ctr requests.CreateTargetRequest) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to create target")
	}

	targetName := ctr.Name
	credentialType := ctr.Properties.CredentialType
	policyArns := ctr.Properties.PolicyArns
	roleArn := ctr.Properties.RoleArn

	options := map[string]interface{}{
		"role_arns":       roleArn,
		"credential_type": credentialType,
		"policy_arns":     policyArns,
	}

	path := fmt.Sprintf("aws/roles/%s-%s-target-%s", vaultProjectPrefix, projectName, targetName)
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

func (v VaultProvider) GetTarget(projectName, targetName string) (requests.TargetProperties, error) {
	if !v.isAdmin() {
		return requests.TargetProperties{}, errors.New("admin credentials must be used to get target information")
	}

	sec, err := v.vaultLogicalSvc.Read(fmt.Sprintf("aws/roles/argo-cloudops-projects-%s-target-%s", projectName, targetName))
	if err != nil {
		return requests.TargetProperties{}, fmt.Errorf("vault get target error: %w", err)
	}

	if sec == nil {
		return requests.TargetProperties{}, ErrTargetNotFound
	}

	roleArn := sec.Data["role_arns"].([]interface{})[0].(string)
	policyArns := sec.Data["policy_arns"].([]interface{})
	credentialType := sec.Data["credential_type"].(string)

	var policies []string
	for _, v := range policyArns {
		policies = append(policies, v.(string))
	}

	return requests.TargetProperties{CredentialType: credentialType, RoleArn: roleArn, PolicyArns: policies}, nil
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

func (v VaultProvider) readSecretID(appRoleName string) (string, error) {
	options := map[string]interface{}{
		"force": true,
	}
	secret, err := v.vaultLogicalSvc.Write(fmt.Sprintf("%s/secret-id", genProjectAppRole(appRoleName)), options)
	if err != nil {
		return "", err
	}
	return secret.Data["secret_id"].(string), nil
}

func (v VaultProvider) TargetExists(projectName, targetName string) (bool, error) {
	_, err := v.GetTarget(projectName, targetName)
	return err != ErrTargetNotFound, nil
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
