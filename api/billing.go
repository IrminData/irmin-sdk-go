package irmincore

import (
	"context"
	"fmt"
	"net/http"

	irminmodels "github.com/IrminData/irmin-sdk-go/models"
)

// GetSubscription retrieves the current billing subscription for a workspace.
func (c *Client) GetSubscription(
	ctx context.Context,
	workspaceSlug string,
) (*irminmodels.PlanInfo, *irminmodels.IrminAPIResponse, error) {
	var plan irminmodels.PlanInfo
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/billing/subscription", workspaceSlug),
	}, &plan)
	if err != nil {
		return nil, nil, fmt.Errorf("get subscription error: %w", err)
	}
	return &plan, apiResp, nil
}

// GetUsage retrieves current period usage for a workspace.
func (c *Client) GetUsage(
	ctx context.Context,
	workspaceSlug string,
) ([]irminmodels.UsageDimensionInfo, *irminmodels.IrminAPIResponse, error) {
	var usage []irminmodels.UsageDimensionInfo
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/billing/usage", workspaceSlug),
	}, &usage)
	if err != nil {
		return nil, nil, fmt.Errorf("get usage error: %w", err)
	}
	return usage, apiResp, nil
}

// GetUsageHistory retrieves usage history for a workspace over multiple periods.
func (c *Client) GetUsageHistory(
	ctx context.Context,
	workspaceSlug string,
	periods int,
) ([]irminmodels.UsageDimensionInfo, *irminmodels.IrminAPIResponse, error) {
	var usage []irminmodels.UsageDimensionInfo
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/billing/usage/history?periods=%d", workspaceSlug, periods),
	}, &usage)
	if err != nil {
		return nil, nil, fmt.Errorf("get usage history error: %w", err)
	}
	return usage, apiResp, nil
}

// CreateCheckout creates a Polar checkout session for a workspace.
func (c *Client) CreateCheckout(
	ctx context.Context,
	workspaceSlug string,
	req irminmodels.CheckoutRequest,
) (*irminmodels.CheckoutResponse, *irminmodels.IrminAPIResponse, error) {
	var checkout irminmodels.CheckoutResponse
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/workspaces/%s/billing/checkout", workspaceSlug),
		ContentType: "application/json",
		Body:        req,
	}, &checkout)
	if err != nil {
		return nil, nil, fmt.Errorf("create checkout error: %w", err)
	}
	return &checkout, apiResp, nil
}

// ChangePlan changes the billing plan for a workspace.
func (c *Client) ChangePlan(
	ctx context.Context,
	workspaceSlug string,
	req irminmodels.CheckoutRequest,
) (*irminmodels.CheckoutResponse, *irminmodels.IrminAPIResponse, error) {
	var checkout irminmodels.CheckoutResponse
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    fmt.Sprintf("/v1/workspaces/%s/billing/change-plan", workspaceSlug),
		ContentType: "application/json",
		Body:        req,
	}, &checkout)
	if err != nil {
		return nil, nil, fmt.Errorf("change plan error: %w", err)
	}
	return &checkout, apiResp, nil
}

// GetPortalURL retrieves the Polar customer portal URL for a workspace.
func (c *Client) GetPortalURL(
	ctx context.Context,
	workspaceSlug string,
) (*irminmodels.PortalResponse, *irminmodels.IrminAPIResponse, error) {
	var portal irminmodels.PortalResponse
	apiResp, err := c.FetchAPI(ctx, RequestOptions{
		Method:   http.MethodPost,
		Endpoint: fmt.Sprintf("/v1/workspaces/%s/billing/portal", workspaceSlug),
	}, &portal)
	if err != nil {
		return nil, nil, fmt.Errorf("get portal URL error: %w", err)
	}
	return &portal, apiResp, nil
}
