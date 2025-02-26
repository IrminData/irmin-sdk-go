package services

import (
	"fmt"
	"net/http"

	"github.com/IrminData/irmin-sdk-go/models"
)

// RoleService handles Role-related API calls
type RoleService struct {
	client *Client
}

// NewRoleService creates a new RoleService
func NewRoleService(client *Client) *RoleService {
	return &RoleService{
		client: client,
	}
}

// FetchRoles retrieves all available roles
func (s *RoleService) FetchRoles() ([]models.IrminRole, *models.IrminAPIResponse, error) {
	endpoint := "/v1/roles"
	var roles []models.IrminRole

	apiResp, err := s.client.FetchAPI(RequestOptions{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}, &roles)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch roles error: %w", err)
	}
	return roles, apiResp, nil
}
