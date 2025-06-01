package irmincore

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	irminmodels "github.com/IrminData/irmin-sdk-go/models"
)

// PolicyCreateParams represents the parameters required to create a new policy.
type PolicyCreateParams struct {
	// Required fields
	Effect    irminmodels.PolicyEffect    `form:"effect"`    // The effect of the policy (allow/deny)
	Action    irminmodels.PolicyAction    `form:"action"`    // The action being controlled (read/write/etc)
	Resource  irminmodels.PolicyResource  `form:"resource"`  // The resource type being controlled
	Principal irminmodels.PolicyPrincipal `form:"principal"` // The principal type (role/user)

	// Optional fields
	ResourceID      *string `form:"resource_id,omitempty"`       // ID of the specific resource
	RoleID          *string `form:"role_id,omitempty"`           // ID of the role if principal is role
	WorkspaceUserID *string `form:"workspace_user_id,omitempty"` // ID of the user if principal is user
}

// PolicyUpdateParams represents the parameters that can be updated for a policy.
type PolicyUpdateParams struct {
	// All fields are optional for updates
	Effect          *irminmodels.PolicyEffect    `form:"effect,omitempty"`            // The effect of the policy (allow/deny)
	Action          *irminmodels.PolicyAction    `form:"action,omitempty"`            // The action being controlled (read/write/etc)
	Resource        *irminmodels.PolicyResource  `form:"resource,omitempty"`          // The resource type being controlled
	Principal       *irminmodels.PolicyPrincipal `form:"principal,omitempty"`         // The principal type (role/user)
	ResourceID      *string                      `form:"resource_id,omitempty"`       // ID of the specific resource
	RoleID          *string                      `form:"role_id,omitempty"`           // ID of the role if principal is role
	WorkspaceUserID *string                      `form:"workspace_user_id,omitempty"` // ID of the user if principal is user
}

// toFormFields converts a struct to a map[string]string using form tags.
func toFormFields(v interface{}) map[string]string {
	fields := make(map[string]string)
	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" || tag == "-" {
			continue
		}

		// Remove omitempty from tag
		if idx := field.Tag.Get("form"); idx != "" {
			tag = idx[:len(idx)-9] // remove ",omitempty"
		}

		value := val.Field(i)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}

		// Convert value to string
		var strValue string
		switch value.Kind() {
		case reflect.String:
			strValue = value.String()
		case reflect.Bool:
			strValue = strconv.FormatBool(value.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strValue = strconv.FormatInt(value.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strValue = strconv.FormatUint(value.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			strValue = strconv.FormatFloat(value.Float(), 'f', -1, 64)
		default:
			continue
		}

		fields[tag] = strValue
	}

	return fields
}

// ListPolicies returns a list of all policies for a workspace.
func (c *Client) ListPolicies(workspace string) ([]irminmodels.Policy, *irminmodels.IrminAPIResponse, error) {
	var policies []irminmodels.Policy
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/policies", workspace),
	}, &policies)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch policies error: %w", err)
	}
	return policies, apiResp, nil
}

// GetPolicy returns a single policy.
func (c *Client) GetPolicy(workspace, policyID string) (*irminmodels.Policy, *irminmodels.IrminAPIResponse, error) {
	var policy irminmodels.Policy
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/policies/%s", workspace, policyID),
	}, &policy)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch policy error: %w", err)
	}
	return &policy, apiResp, nil
}

// CreatePolicy creates a new policy for a workspace.
func (c *Client) CreatePolicy(
	workspace string,
	params PolicyCreateParams,
) (*irminmodels.Policy, *irminmodels.IrminAPIResponse, error) {
	var policy irminmodels.Policy
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/workspaces/%s/policies", workspace),
		ContentType: "application/x-www-form-urlencoded",
		FormFields:  toFormFields(params),
	}, &policy)
	if err != nil {
		return nil, nil, fmt.Errorf("create policy error: %w", err)
	}
	return &policy, apiResp, nil
}

// UpdatePolicy updates an existing policy.
func (c *Client) UpdatePolicy(
	workspace, policyID string,
	params PolicyUpdateParams,
) (*irminmodels.Policy, *irminmodels.IrminAPIResponse, error) {
	var policy irminmodels.Policy
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:      http.MethodPatch,
		Endpoint:    fmt.Sprintf("/v1/workspaces/%s/policies/%s", workspace, policyID),
		ContentType: "application/x-www-form-urlencoded",
		FormFields:  toFormFields(params),
	}, &policy)
	if err != nil {
		return nil, nil, fmt.Errorf("update policy error: %w", err)
	}
	return &policy, apiResp, nil
}

// DeletePolicy deletes a policy.
func (c *Client) DeletePolicy(workspace, policyID string) (*irminmodels.IrminAPIResponse, error) {
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/policies/%s", workspace, policyID),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("delete policy error: %w", err)
	}
	return apiResp, nil
}
