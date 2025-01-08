package services

import (
	"fmt"
	"irmin-sdk/client"
	"irmin-sdk/models"
	"net/http"
	"net/url"
	"strings"
)

// QueryService handles query-related API calls
type QueryService struct {
	client *client.Client
}

// NewQueryService creates a new QueryService
func NewQueryService(client *client.Client) *QueryService {
	return &QueryService{
		client: client,
	}
}

// ExecuteScript executes a script (e.g., Irmin SQL query or Compute Sandbox script)
func (s *QueryService) ExecuteScript(scriptType, content string) (*models.QueryExecutionResult, *client.IrminAPIResponse, error) {
	form := url.Values{}
	form.Set("type", scriptType)
	form.Set("content", content)

	var result models.QueryExecutionResult
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    "/v1/queries/execute",
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(form.Encode()),
	}, &result)
	if err != nil {
		return nil, nil, fmt.Errorf("execute script error: %w", err)
	}
	return &result, apiResp, nil
}

// CreateQuery creates a new query
func (s *QueryService) CreateQuery(
	scriptType, content, name, description string,
	stored, run bool,
) (*models.Query, *client.IrminAPIResponse, error) {
	form := url.Values{}
	form.Set("type", scriptType)
	form.Set("content", content)
	if name != "" {
		form.Set("name", name)
	}
	if description != "" {
		form.Set("description", description)
	}
	form.Set("stored", fmt.Sprintf("%t", stored))
	form.Set("run", fmt.Sprintf("%t", run))

	var query models.Query
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    "/v1/queries",
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(form.Encode()),
	}, &query)
	if err != nil {
		return nil, nil, fmt.Errorf("create query error: %w", err)
	}
	return &query, apiResp, nil
}

// GetQueries retrieves all queries in the workspace
func (s *QueryService) GetQueries() ([]models.Query, *client.IrminAPIResponse, error) {
	endpoint := "/v1/queries"
	var queries []models.Query

	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}, &queries)
	if err != nil {
		return nil, nil, fmt.Errorf("get queries error: %w", err)
	}
	return queries, apiResp, nil
}

// GetQuery retrieves a single query by ID
func (s *QueryService) GetQuery(queryID string) (*models.Query, *client.IrminAPIResponse, error) {
	endpoint := fmt.Sprintf("/v1/queries/%s", queryID)
	var query models.Query

	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}, &query)
	if err != nil {
		return nil, nil, fmt.Errorf("get query error: %w", err)
	}
	return &query, apiResp, nil
}

// DeleteQuery deletes a query by ID
func (s *QueryService) DeleteQuery(queryID string) (*client.IrminAPIResponse, error) {
	form := url.Values{}
	form.Set("_method", "DELETE")

	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/queries/%s", queryID),
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(form.Encode()),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("delete query error: %w", err)
	}
	return apiResp, nil
}

// UpdateQuery updates a query by ID
func (s *QueryService) UpdateQuery(
	queryID, scriptType, content, name, description string,
	stored bool,
) (*models.Query, *client.IrminAPIResponse, error) {
	form := url.Values{}
	form.Set("_method", "PATCH")
	form.Set("type", scriptType)
	form.Set("content", content)
	form.Set("name", name)
	form.Set("description", description)
	form.Set("stored", fmt.Sprintf("%t", stored))

	var query models.Query
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/queries/%s", queryID),
		ContentType: "application/x-www-form-urlencoded",
		Body:        strings.NewReader(form.Encode()),
	}, &query)
	if err != nil {
		return nil, nil, fmt.Errorf("update query error: %w", err)
	}
	return &query, apiResp, nil
}

// ExecuteQuery executes a query by ID
func (s *QueryService) ExecuteQuery(queryID string) (*client.IrminAPIResponse, error) {
	endpoint := fmt.Sprintf("/v1/queries/%s/execute", queryID)
	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("execute query error: %w", err)
	}
	return apiResp, nil
}

// GetQueryResults retrieves the results of a query, paginated
func (s *QueryService) GetQueryResults(queryID string, page int) (*models.QueryExecutionResult, *client.IrminAPIResponse, error) {
	endpoint := fmt.Sprintf("/v1/queries/%s/results?page=%d", queryID, page)
	var result models.QueryExecutionResult

	apiResp, err := s.client.FetchAPI(client.RequestOptions{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}, &result)
	if err != nil {
		return nil, nil, fmt.Errorf("get query results error: %w", err)
	}
	return &result, apiResp, nil
}
