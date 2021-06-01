package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

func (v vaultCredentialsProvider) newCredentialsProvider(authorization Authorization) (credentialsProvider, error) {
	if authorization.Provider == "vault" {
		return &vaultCredentialsProvider{
			VaultSvc: v.VaultSvc,
			RoleID:   authorization.Key,
			SecretID: authorization.Secret,
		}, nil
	}

	return nil, errors.New("Unknown provider")
}

type credentialsProvider interface {
	createProject(string) (string, string, error)
	createTarget(string, createTargetRequest) error
	deleteProject(string) error
	deleteTarget(string, string) error
	getProject(string) (string, error)
	getTarget(string, string) (targetProperties, error)
	getToken() (string, error)
	listTargets(string) ([]string, error)
	projectExists(string) (bool, error)
	targetExists(name string) (bool, error)
}

// Vault
const (
	vaultAppRolePrefix = "auth/approle/role"
	vaultProjectPrefix = "argo-cloudops-projects"
)

var (
	ErrNotFound = errors.New("item not found")
)

type vaultCredentialsProvider struct {
	VaultSvc *vault.Client
	RoleID   string
	SecretID string
}

type vaultConfig struct {
	config *vault.Config
	role   string
	secret string
}

// TODO before open sourcing we should provide the token instead of generating it
func newVaultSvc(c vaultConfig, h http.Header) (*vault.Client, error) {
	vaultSvc, err := vault.NewClient(c.config)
	if err != nil {
		return nil, err
	}

	options := map[string]interface{}{
		"role_id":   c.role,
		"secret_id": c.secret,
	}

	sec, err := vaultSvc.Logical().Write("auth/approle/login", options)
	if err != nil {
		return nil, err
	}

	vaultSvc.SetToken(sec.Auth.ClientToken)
	vaultSvc.SetHeaders(h)
	return vaultSvc, nil
}

type targetProperties struct {
	CredentialType string   `json:"credential_type"`
	PolicyArns     []string `json:"policy_arns"`
	RoleArn        string   `json:"role_arn"`
}

type createTargetRequest struct {
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Properties targetProperties `json:"properties"`
}

type createProjectRequest struct {
	Name string `json:"name"`
}

func genProjectAppRole(name string) string {
	return fmt.Sprintf("%s/%s-%s", vaultAppRolePrefix, vaultProjectPrefix, name)
}

func (v vaultCredentialsProvider) projectExists(name string) (bool, error) {
	p, err := v.getProject(name)
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return p != "", nil
}

func (v vaultCredentialsProvider) targetExists(name string) (bool, error) {
	// TODO: Implement targetExists call
	return false, nil
}

func (v vaultCredentialsProvider) isAdmin() bool {
	return v.RoleID == "admin"
}

func (v vaultCredentialsProvider) getToken() (string, error) {
	if v.isAdmin() {
		return "", errors.New("admin credentials cannot be used to get tokens")
	}

	options := map[string]interface{}{
		"role_id":   v.RoleID,
		"secret_id": v.SecretID,
	}

	sec, err := v.VaultSvc.Logical().Write("auth/approle/login", options)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	return sec.Auth.ClientToken, nil
}

func (v vaultCredentialsProvider) createPolicyState(name, policy string) error {
	return v.VaultSvc.Sys().PutPolicy(fmt.Sprintf("%s-%s", vaultProjectPrefix, name), policy)
}

func (v vaultCredentialsProvider) deletePolicyState(name string) error {
	return v.VaultSvc.Sys().DeletePolicy(fmt.Sprintf("%s-%s", vaultProjectPrefix, name))
}

// TODO validate policy and other information is correct in target
// Validate role exists (if possible, etc)
func (v vaultCredentialsProvider) createTarget(projectName string, ctr createTargetRequest) error {
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
	_, err := v.VaultSvc.Logical().Write(path, options)
	return err
}

func (v vaultCredentialsProvider) deleteTarget(projectName string, targetName string) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to delete target")
	}

	path := fmt.Sprintf("aws/roles/%s-%s-target-%s", vaultProjectPrefix, projectName, targetName)
	_, err := v.VaultSvc.Logical().Delete(path)
	return err
}

const (
	vaultSecretTTL   = "8776h" // 1 year
	vaultTokenMaxTTL = "10m"
	// When set to 1 with the cli or api, it will not return the creds as it
	// says it's hit the limit of uses.
	vaultTokenNumUses = 3
)

func (v vaultCredentialsProvider) writeProjectState(name string) error {
	options := map[string]interface{}{
		"secret_id_ttl":           vaultSecretTTL,
		"token_max_ttl":           vaultTokenMaxTTL,
		"token_no_default_policy": "true",
		"token_num_uses":          vaultTokenNumUses,
		"token_policies":          fmt.Sprintf("%s-%s", vaultProjectPrefix, name),
	}

	_, err := v.VaultSvc.Logical().Write(genProjectAppRole(name), options)
	if err != nil {
		return err
	}
	return nil
}

func (v vaultCredentialsProvider) readSecretID(appRoleName string) (string, error) {
	options := map[string]interface{}{
		"force": true,
	}
	secret, err := v.VaultSvc.Logical().Write(fmt.Sprintf("%s/secret-id", genProjectAppRole(appRoleName)), options)
	if err != nil {
		return "", err
	}
	return secret.Data["secret_id"].(string), nil
}

func (v vaultCredentialsProvider) readRoleID(appRoleName string) (string, error) {
	secret, err := v.VaultSvc.Logical().Read(fmt.Sprintf("%s/role-id", genProjectAppRole(appRoleName)))
	if err != nil {
		return "", err
	}
	return secret.Data["role_id"].(string), nil
}

func defaultVaultReadonlyPolicyAWS(projectName string) string {
	return fmt.Sprintf(
		"path \"aws/sts/argo-cloudops-projects-%s-target-*\" { capabilities = [\"read\"] }",
		projectName,
	)
}

func (v vaultCredentialsProvider) createProject(name string) (string, string, error) {
	if !v.isAdmin() {
		return "", "", errors.New("admin credentials must be used to create project")
	}

	policy := defaultVaultReadonlyPolicyAWS(name)
	err := v.createPolicyState(name, policy)
	if err != nil {
		return "", "", err
	}

	err = v.writeProjectState(name)
	if err != nil {
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

func (v vaultCredentialsProvider) listTargets(project string) ([]string, error) {
	if !v.isAdmin() {
		return nil, errors.New("admin credentials must be used to list targets")
	}

	sec, err := v.VaultSvc.Logical().List("aws/roles/")
	if err != nil {
		return nil, fmt.Errorf("vault list error: %v", err)
	}

	// allow empty array to render json as []
	list := make([]string, 0)
	if sec != nil {
		for _, target := range sec.Data["keys"].([]interface{}) {
			value := target.(string)
			if strings.HasPrefix(value, fmt.Sprintf("argo-cloudops-projects-%s-target-", project)) {
				list = append(list, value)
			}
		}
	}

	return list, nil
}

func (v vaultCredentialsProvider) getProject(projectName string) (string, error) {
	sec, err := v.VaultSvc.Logical().Read(genProjectAppRole(projectName))
	if err != nil {
		return "", fmt.Errorf("vault get project error: %v", err)
	}
	if sec == nil {
		return "", ErrNotFound
	}
	return fmt.Sprintf(`{"name":"%s"}`, projectName), nil
}

func (v vaultCredentialsProvider) deleteProject(name string) error {
	if !v.isAdmin() {
		return errors.New("admin credentials must be used to delete project")
	}

	err := v.deletePolicyState(name)
	if err != nil {
		return fmt.Errorf("vault delete project error: %v", err)
	}

	_, err = v.VaultSvc.Logical().Delete(genProjectAppRole(name))
	if err != nil {
		return fmt.Errorf("vault delete project error: %v", err)
	}
	return nil
}

func (v vaultCredentialsProvider) getTarget(projectName, targetName string) (targetProperties, error) {
	if !v.isAdmin() {
		return targetProperties{}, errors.New("admin credentials must be used to get target information")
	}

	sec, err := v.VaultSvc.Logical().Read(fmt.Sprintf("aws/roles/argo-cloudops-projects-%s-target-%s", projectName, targetName))
	if err != nil {
		return targetProperties{}, fmt.Errorf("vault get target error: %v", err)
	}

	if sec == nil {
		return targetProperties{}, fmt.Errorf("target not found")
	}

	roleArn := sec.Data["role_arns"].([]interface{})[0].(string)
	policyArns := sec.Data["policy_arns"].([]interface{})
	credentialType := sec.Data["credential_type"].(string)

	var policies []string
	for _, v := range policyArns {
		policies = append(policies, v.(string))
	}

	return targetProperties{CredentialType: credentialType, RoleArn: roleArn, PolicyArns: policies}, nil
}
