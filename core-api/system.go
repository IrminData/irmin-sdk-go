package irmincore

import (
	"fmt"
	"net/http"

	irminmodels "github.com/IrminData/irmin-sdk-go/models"
)

// CallSystemWebhook calls the system webhook endpoint.
// The body is expected to be an that will be marshaled to JSON.
//
// Usable only with a system token.
func (c *Client) CallSystemWebhook(
	queryParams map[string]string,
	headers map[string]string,
	body any,
) (*irminmodels.IrminAPIResponse, error) {
	apiResp, err := c.FetchAPI(RequestOptions{
		Method:      http.MethodPost,
		Endpoint:    "/v1/system/webhook",
		ContentType: "application/json",
		Body:        body,
		Headers:     headers,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call system webhook error: %w", err)
	}
	return apiResp, nil
}
