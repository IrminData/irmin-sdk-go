package services

import (
	"encoding/json"
	"fmt"
	"irmin-sdk/client"
	"irmin-sdk/models"
	"net/http"
	"net/url"
)

// ConnectorService handles operations related to connectors
type ConnectorService struct {
	client *client.Client
}

// NewConnectorService creates a new instance of ConnectorService
func NewConnectorService(client *client.Client) *ConnectorService {
	return &ConnectorService{
		client: client,
	}
}

// FetchAllConnectors retrieves all available connectors
func (s *ConnectorService) FetchAllConnectors() ([]models.Connector, *client.IrminAPIResponse, error) {
	var connectors []models.Connector
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: "/v1/connectors",
	}, &connectors)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch connectors error: %w", err)
	}
	return connectors, apiResp, nil
}

// FetchConnector retrieves a connector by its ID
func (s *ConnectorService) FetchConnector(connectorID string) (*models.Connector, *client.IrminAPIResponse, error) {
	var connector models.Connector
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/connectors/%s", connectorID),
	}, &connector)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch connector error: %w", err)
	}
	return &connector, apiResp, nil
}

// FetchConnectorConfigurationFields retrieves configuration fields for a connector
func (s *ConnectorService) FetchConnectorConfigurationFields(
	connectorID, configType string,
	currentDetails map[string]string,
	currentSettings map[string]string,
) (map[string]interface{}, *client.IrminAPIResponse, error) {
	form := map[string]string{}
	for key, value := range currentDetails {
		form[fmt.Sprintf("details[%s]", key)] = value
	}
	for key, value := range currentSettings {
		form[fmt.Sprintf("settings[%s]", key)] = value
	}

	var fields map[string]interface{}
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/connectors/%s/%s", connectorID, configType),
		ContentType: "application/x-www-form-urlencoded",
		FormFields:  form,
	}, &fields)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch connector configuration fields error: %w", err)
	}
	return fields, apiResp, nil
}

// ValidateConnectorConfiguration validates the configuration for a connector
func (s *ConnectorService) ValidateConnectorConfiguration(
	connectorID string,
	details map[string]string,
	settings map[string]string,
) (*models.ConnectorConfigurationValidationResult, *client.IrminAPIResponse, error) {
	form := map[string]string{}
	for key, value := range details {
		form[fmt.Sprintf("details[%s]", key)] = value
	}
	for key, value := range settings {
		form[fmt.Sprintf("settings[%s]", key)] = value
	}

	var validationResult models.ConnectorConfigurationValidationResult
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/connectors/%s/validate", connectorID),
		ContentType: "application/x-www-form-urlencoded",
		FormFields:  form,
	}, &validationResult)
	if err != nil {
		return nil, nil, fmt.Errorf("validate connector configuration error: %w", err)
	}
	return &validationResult, apiResp, nil
}

// FetchConnectorSchema retrieves the object schema for a connector
func (s *ConnectorService) FetchConnectorSchema(
	connectorID, operation string,
	details map[string]string,
	settings map[string]string,
) (*models.ObjectSchema, *client.IrminAPIResponse, error) {
	form := map[string]string{}
	for key, value := range details {
		form[fmt.Sprintf("details[%s]", key)] = value
	}
	for key, value := range settings {
		form[fmt.Sprintf("settings[%s]", key)] = value
	}

	var schema models.ObjectSchema
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/connectors/%s/schema/%s", connectorID, operation),
		ContentType: "application/x-www-form-urlencoded",
		FormFields:  form,
	}, &schema)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch connector schema error: %w", err)
	}
	return &schema, apiResp, nil
}

// ValidateConnectorData validates data against a connector schema
func (s *ConnectorService) ValidateConnectorData(
	connectorID, operation string,
	data map[string]interface{},
	details map[string]string,
	settings map[string]string,
) (*models.ConnectorSchemaValidationResult, *client.IrminAPIResponse, error) {
	form := url.Values{}

	config := make(map[string]interface{})
	for key, value := range details {
		config[fmt.Sprintf("details[%s]", key)] = value
	}
	for key, value := range settings {
		config[fmt.Sprintf("settings[%s]", key)] = value
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal config error: %w", err)
	}
	form.Set("configuration", string(configJSON))

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal data error: %w", err)
	}
	form.Set("data", string(dataJSON))

	var validationResult models.ConnectorSchemaValidationResult
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/connectors/%s/schema/%s/validate", connectorID, operation),
		ContentType: "application/x-www-form-urlencoded",
		Body:        []byte(form.Encode()),
	}, &validationResult)
	if err != nil {
		return nil, nil, fmt.Errorf("validate connector data error: %w", err)
	}
	return &validationResult, apiResp, nil
}

// RegisterNewConnector registers a new connector with the system. Requests to this endpoint must be authenticated with a system token.
func (s *ConnectionService) RegisterNewConnector(baseURL, systemToken string) (*models.Connector, *client.IrminAPIResponse, error) {
	var connector models.Connector
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    "/v1/connectors",
		ContentType: "application/x-www-form-urlencoded",
		FormFields: map[string]string{
			"url":          baseURL,
			"system_token": systemToken,
		},
	}, &connector)
	if err != nil {
		return nil, nil, fmt.Errorf("register new connector error: %w", err)
	}
	return &connector, apiResp, nil
}

// UpdateRegisteredConnector updates the details of a registered connector. Requests to this endpoint must be authenticated with a system token.
func (s *ConnectionService) UpdateRegisteredConnector(connectorID, baseURL, systemToken string) (*models.Connector, *client.IrminAPIResponse, error) {
	var connector models.Connector
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/connectors/%s", connectorID),
		ContentType: "application/x-www-form-urlencoded",
		FormFields: map[string]string{
			"_method":      "PATCH",
			"url":          baseURL,
			"system_token": systemToken,
		},
	}, &connector)
	if err != nil {
		return nil, nil, fmt.Errorf("update registered connector error: %w", err)
	}
	return &connector, apiResp, nil
}
