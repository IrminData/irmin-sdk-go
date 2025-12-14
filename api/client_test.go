package irmincore_test

import (
	"net/url"
	"testing"

	irmincore "github.com/IrminData/irmin-sdk-go/api"
)

func TestClient_ValidationIntegration(t *testing.T) {
	// Create a client with validation enabled
	client := irmincore.NewClient("https://api.example.com", "fake-token", "en")

	t.Run("ValidateRequest method works", func(t *testing.T) {
		// Test valid request
		validReq := irmincore.CreateConnectionRequest{
			Name:      "Test Connection",
			Connector: "postgres",
		}

		err := client.ValidateRequest(validReq)
		if err != nil {
			t.Errorf("Expected valid request to pass validation, got error: %v", err)
		}

		// Test invalid request
		invalidReq := irmincore.CreateConnectionRequest{
			// Missing required fields
		}

		err = client.ValidateRequest(invalidReq)
		if err == nil {
			t.Error("Expected invalid request to fail validation")
		}
	})

	t.Run("ValidateVar method works", func(t *testing.T) {
		// Test valid email
		err := client.ValidateVar("test@example.com", "email")
		if err != nil {
			t.Errorf("Expected valid email to pass validation, got error: %v", err)
		}

		// Test invalid email
		err = client.ValidateVar("not-an-email", "email")
		if err == nil {
			t.Error("Expected invalid email to fail validation")
		}
	})

	t.Run("Client has validator initialized", func(t *testing.T) {
		if client.Validator == nil {
			t.Error("Expected client to have validator initialized")
		}
	})

	t.Run("SQID validation is skipped on client-side", func(t *testing.T) {
		// This should pass even with an "invalid" SQID because client-side validation skips SQID checks
		req := irmincore.TransferConnectionOwnershipRequest{
			NewOwnerID: "definitely_not_a_valid_sqid",
		}

		err := client.ValidateRequest(req)
		if err != nil {
			t.Errorf("Client-side validator should skip SQID validation, got error: %v", err)
		}
	})
}

func TestAddLimitResponseParam(t *testing.T) {
	t.Run("simple endpoint without query params", func(t *testing.T) {
		result := irmincore.AddLimitResponseParam("/api/users")
		assertURLHasParam(t, result, "limit-response", "true")
		assertPath(t, result, "/api/users")
	})

	t.Run("endpoint with existing query params", func(t *testing.T) {
		result := irmincore.AddLimitResponseParam("/api/users?name=john&age=30")
		assertURLHasParam(t, result, "limit-response", "true")
		assertURLHasParam(t, result, "name", "john")
		assertURLHasParam(t, result, "age", "30")
	})

	t.Run("endpoint with trailing separator", func(t *testing.T) {
		result := irmincore.AddLimitResponseParam("/api/users?name=john&")
		assertURLHasParam(t, result, "limit-response", "true")
		assertURLHasParam(t, result, "name", "john")
	})

	t.Run("endpoint with fragment", func(t *testing.T) {
		result := irmincore.AddLimitResponseParam("/api/users?name=john#section")
		assertURLHasParam(t, result, "limit-response", "true")
		assertURLHasParam(t, result, "name", "john")
		assertFragment(t, result, "section")
	})

	t.Run("endpoint with trailing separator and fragment", func(t *testing.T) {
		result := irmincore.AddLimitResponseParam("/api/users?name=john&#section")
		assertURLHasParam(t, result, "limit-response", "true")
		assertURLHasParam(t, result, "name", "john")
		assertFragment(t, result, "section")
	})
}

// Helper functions for assertions

func assertURLHasParam(t *testing.T, urlStr, key, expectedValue string) {
	t.Helper()
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", urlStr, err)
	}

	value := u.Query().Get(key)
	if value != expectedValue {
		t.Errorf("Expected parameter %q to be %q, got %q", key, expectedValue, value)
	}
}

func assertPath(t *testing.T, urlStr, expectedPath string) {
	t.Helper()
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", urlStr, err)
	}

	if u.Path != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, u.Path)
	}
}

func assertFragment(t *testing.T, urlStr, expectedFragment string) {
	t.Helper()
	u, err := url.Parse(urlStr)
	if err != nil {
		t.Fatalf("Failed to parse URL %q: %v", urlStr, err)
	}

	if u.Fragment != expectedFragment {
		t.Errorf("Expected fragment %q, got %q", expectedFragment, u.Fragment)
	}
}
