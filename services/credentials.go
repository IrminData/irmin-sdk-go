package services

import (
	"bytes"
	"fmt"
	"irmin-sdk/client"
	"irmin-sdk/models"
	"mime/multipart"
	"net/http"
)

// CredentialService handles operations related to system tokens
type CredentialService struct {
	client *client.Client
}

// NewCredentialService creates a new instance of CredentialService
func NewCredentialService(client *client.Client) *CredentialService {
	return &CredentialService{
		client: client,
	}
}

// GetSystemTokens retrieves the user's system tokens
func (s *CredentialService) GetSystemTokens() ([]models.SystemToken, *client.IrminAPIResponse, error) {
	var tokens []models.SystemToken
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: "/v1/credentials",
	}, &tokens)
	if err != nil {
		return nil, nil, fmt.Errorf("get system tokens error: %w", err)
	}
	return tokens, apiResp, nil
}

// CreateSystemToken creates a new system token
func (s *CredentialService) CreateSystemToken(name string, expiry int) (*models.SystemToken, *client.IrminAPIResponse, error) {
	endpoint := "/v1/credentials"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("name", name); err != nil {
		return nil, nil, fmt.Errorf("write name field error: %w", err)
	}
	if err := writer.WriteField("expiry", fmt.Sprintf("%d", expiry)); err != nil {
		return nil, nil, fmt.Errorf("write expiry field error: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, nil, fmt.Errorf("close writer error: %w", err)
	}

	var token models.SystemToken
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    endpoint,
		ContentType: writer.FormDataContentType(),
		Body:        body,
	}, &token)
	if err != nil {
		return nil, nil, fmt.Errorf("create system token error: %w", err)
	}
	return &token, apiResp, nil
}

// RevokeSystemToken revokes a system token
func (s *CredentialService) RevokeSystemToken(tokenID string) (*client.IrminAPIResponse, error) {
	endpoint := fmt.Sprintf("/v1/credentials/%s", tokenID)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("_method", "DELETE"); err != nil {
		return nil, fmt.Errorf("write _method field error: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer error: %w", err)
	}

	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    endpoint,
		ContentType: writer.FormDataContentType(),
		Body:        body,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("revoke system token error: %w", err)
	}
	return apiResp, nil
}
