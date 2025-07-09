package irminsdkvalidator_test

import (
	"strings"
	"testing"
	"time"

	coreapi "github.com/IrminData/irmin-sdk-go/core-api"
	models "github.com/IrminData/irmin-sdk-go/models"
	sqids "github.com/IrminData/irmin-sdk-go/sqids"
	validator "github.com/IrminData/irmin-sdk-go/validator"
)

func TestValidator_ValidateUser(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	t.Run("valid user", func(t *testing.T) {
		userID, _ := sqidManager.Encode("users", 123)
		roleID, _ := sqidManager.Encode("roles", 123)
		user := models.User{
			ID:             userID,
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john.doe@example.com",
			Phone:          "+1234567890",
			Company:        "Example Inc.",
			ProfilePicture: "https://example.com/profile.jpg",
			Roles: []models.Role{
				{ID: roleID, Role: "admin"},
			},
		}

		err := validator.Validate(user)
		if err != nil {
			t.Errorf("Expected valid user, got error: %v", err)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		user := models.User{
			ID:        "", // Missing required field
			FirstName: "", // Too short
			Email:     "invalid-email",
			Roles:     []models.Role{}, // Required but empty
		}

		err := validator.Validate(user)
		if err == nil {
			t.Error("Expected validation error for invalid user")
		}
	})
}

func TestValidator_ValidateAPIToken(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	t.Run("valid API token", func(t *testing.T) {
		tokenID, _ := sqidManager.Encode("api_tokens", 123)
		token := models.APIToken{
			ID:        tokenID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      "My Token",
			Token:     "cred_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := validator.Validate(token)
		if err != nil {
			t.Errorf("Expected valid token, got error: %v", err)
		}
	})

	t.Run("invalid API token", func(t *testing.T) {
		token := models.APIToken{
			ID:        "token-123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      "My Token",
			Token:     "wrong_prefix_short", // Wrong prefix and too short
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := validator.Validate(token)
		if err == nil {
			t.Error("Expected validation error for invalid token")
		}
	})
}

func TestValidator_ValidateSchedule(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	// valid schedule with rrule
	t.Run("valid rrule time trigger", func(t *testing.T) {
		rrule := "FREQ=DAILY;COUNT=5"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:  models.TimeTriggerType,
					RRule: &rrule,
				},
			},
			MaxRetries: 5,
		}
		err := validator.Validate(schedule)
		if err != nil {
			t.Errorf("Expected valid schedule, got error: %v", err)
		}
	})

	// valid schedule with cron
	t.Run("valid cron time trigger", func(t *testing.T) {
		cron := "0 0 * * *"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type: models.TimeTriggerType,
					Cron: &cron,
				},
			},
		}
		err := validator.Validate(schedule)
		if err != nil {
			t.Errorf("Expected valid schedule, got error: %v", err)
		}
	})

	// valid schedule with repository event
	t.Run("valid repository event trigger", func(t *testing.T) {
		repoEvent := models.PostMerge
		repoSlug := "my-repo"
		repoRef := "main"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:            models.RepositoryTriggerType,
					RepositoryEvent: &repoEvent,
					Repository:      &repoSlug,
					RepositoryRef:   &repoRef,
				},
			},
		}
		err := validator.Validate(schedule)
		if err != nil {
			t.Errorf("Expected valid schedule, got error: %v", err)
		}
	})

	// valid schedule with workflow run event
	t.Run("valid workflow run event trigger", func(t *testing.T) {
		workflowEvent := models.PostWorkflowRun
		workflowSqid, _ := sqidManager.Encode("workflows", 1)
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:             models.WorkflowRunTriggerType,
					WorkflowRunEvent: &workflowEvent,
					WorkflowID:       &workflowSqid,
				},
			},
		}
		err := validator.Validate(schedule)
		if err != nil {
			t.Errorf("Expected valid schedule, got error: %v", err)
		}
	})

	// invalid time trigger - neither rrule nor cron
	t.Run("invalid time trigger - neither rrule nor cron", func(t *testing.T) {
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type: models.TimeTriggerType,
				},
			},
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for time trigger with no rrule or cron")
		}
	})

	// invalid repository event - missing repository
	t.Run("invalid repository event - missing repository", func(t *testing.T) {
		repoEvent := models.PostMerge
		repoRef := "main"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:            models.RepositoryTriggerType,
					RepositoryEvent: &repoEvent,
					RepositoryRef:   &repoRef,
					// Missing Repository
				},
			},
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for repository event with missing repository")
		}
	})

	// invalid repository event - missing repository ref
	t.Run("invalid repository event - missing repository ref", func(t *testing.T) {
		repoEvent := models.PostMerge
		repoSlug := "my-repo"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:            models.RepositoryTriggerType,
					RepositoryEvent: &repoEvent,
					Repository:      &repoSlug,
					// Missing RepositoryRef
				},
			},
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for repository event with missing repository ref")
		}
	})

	// invalid workflow run event - missing workflow id
	t.Run("invalid workflow run event - missing workflow id", func(t *testing.T) {
		workflowEvent := models.PostWorkflowRun
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type:             models.WorkflowRunTriggerType,
					WorkflowRunEvent: &workflowEvent,
					// Missing WorkflowID
				},
			},
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for workflow run event with missing workflow id")
		}
	})

	// invalid schedule - maxretries too high
	t.Run("invalid schedule - maxretries too high", func(t *testing.T) {
		cron := "0 0 * * *"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type: models.TimeTriggerType,
					Cron: &cron,
				},
			},
			MaxRetries: 11,
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for max retries > 10")
		}
	})

	// invalid trigger type
	t.Run("invalid trigger type", func(t *testing.T) {
		cron := "0 0 * * *"
		schedule := models.Schedule{
			Triggers: []models.ScheduleTrigger{
				{
					Type: "invalid-trigger-type",
					Cron: &cron,
				},
			},
		}
		err := validator.Validate(schedule)
		if err == nil {
			t.Error("Expected validation error for invalid trigger type")
		}
	})
}

func TestValidator_ValidateVar(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{"valid email", "test@example.com", "email", false},
		{"invalid email", "not-an-email", "email", true},
		{"valid required string", "hello", "required", false},
		{"invalid required string", "", "required", true},
		{"valid min length", "hello", "min=3", false},
		{"invalid min length", "hi", "min=3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStartsWithValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		value   string
		tag     string
		wantErr bool
	}{
		{"valid prefix", "cred_abc123", "startswith=cred_", false},
		{"invalid prefix", "wrong_abc123", "startswith=cred_", true},
		{"empty string", "", "startswith=cred_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.value, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientValidator_RequestValidation(t *testing.T) {
	// Test client-side validator (no SQID manager)
	clientValidator := validator.NewClientValidator()

	t.Run("CreateConnectionRequest - Valid", func(t *testing.T) {
		req := coreapi.CreateConnectionRequest{
			Name:        "My Database Connection",
			Connector:   "postgres",
			Description: "Connection to production database",
			Details: map[string]any{
				"host":     "localhost",
				"port":     5432,
				"database": "myapp",
			},
			Settings: map[string]any{
				"ssl_mode": "require",
			},
		}

		err := clientValidator.Validate(req)
		if err != nil {
			t.Errorf("Expected valid connection request, got error: %v", err)
		}
	})

	t.Run("CreateConnectionRequest - Invalid (missing required fields)", func(t *testing.T) {
		req := coreapi.CreateConnectionRequest{
			// Missing required Name and Connector fields
			Description: "This request is missing required fields",
		}

		err := clientValidator.Validate(req)
		if err == nil {
			t.Error("Expected validation error for missing required fields")
		}
	})

	t.Run("CreateWorkspaceRequest - Valid", func(t *testing.T) {
		req := coreapi.CreateWorkspaceRequest{
			Name:        "My New Workspace",
			Description: "A workspace for data analysis",
		}

		err := clientValidator.Validate(req)
		if err != nil {
			t.Errorf("Expected valid workspace request, got error: %v", err)
		}
	})

	t.Run("CreateWorkspaceRequest - Invalid (missing required name)", func(t *testing.T) {
		req := coreapi.CreateWorkspaceRequest{
			// Missing required Name field
			Description: "A workspace without a name",
		}

		err := clientValidator.Validate(req)
		if err == nil {
			t.Error("Expected validation error for missing required name")
		}
	})

	t.Run("TransferConnectionOwnershipRequest - Valid", func(t *testing.T) {
		req := coreapi.TransferConnectionOwnershipRequest{
			NewOwnerID: "user_123", // SQID validation will be skipped on client-side
		}

		err := clientValidator.Validate(req)
		if err != nil {
			t.Errorf("Expected valid transfer request, got error: %v", err)
		}
	})

	t.Run("TransferConnectionOwnershipRequest - Invalid (missing required field)", func(t *testing.T) {
		req := coreapi.TransferConnectionOwnershipRequest{
			// Missing required NewOwnerID field
		}

		err := clientValidator.Validate(req)
		if err == nil {
			t.Error("Expected validation error for missing required NewOwnerID")
		}
	})
}

func TestServerValidator_RequestValidation(t *testing.T) {
	// Test server-side validator (with SQID manager)
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	serverValidator := validator.NewValidator(sqidManager)

	t.Run("User model with valid SQID", func(t *testing.T) {
		// Generate a valid user SQID
		userSQID, _ := sqidManager.Encode("users", 123)

		user := models.User{
			ID:             userSQID, // Valid SQID
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@example.com",
			Phone:          "+1234567890", // Valid E164 format
			Company:        "Example Inc",
			ProfilePicture: "https://example.com/profile.jpg", // Valid URL
			Roles:          []models.Role{},                   // Empty roles slice
		}

		err := serverValidator.Validate(user)
		if err != nil {
			t.Errorf("Expected valid user with valid SQID, got error: %v", err)
		}
	})

	t.Run("User model with invalid SQID", func(t *testing.T) {
		user := models.User{
			ID:             "invalid_user_sqid", // Invalid SQID
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@example.com",
			Phone:          "+1234567890", // Valid E164 format
			Company:        "Example Inc",
			ProfilePicture: "https://example.com/profile.jpg", // Valid URL
			Roles:          []models.Role{},                   // Empty roles slice
		}

		err := serverValidator.Validate(user)
		if err == nil {
			t.Error("Expected validation error for invalid SQID on server-side")
		}
	})

	t.Run("TransferConnectionOwnershipRequest - With SQID validation in core-api", func(t *testing.T) {
		// Core-api request structs DO have SQID validation tags
		req := coreapi.TransferConnectionOwnershipRequest{
			NewOwnerID: "invalid_user_sqid", // This WILL trigger SQID validation
		}

		err := serverValidator.Validate(req)
		if err == nil {
			t.Error("Expected SQID validation error for invalid user SQID")
		}
	})
}

func TestClientVsServerValidator_SQIDHandling(t *testing.T) {
	// Setup both validators
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	clientValidator := validator.NewClientValidator()
	serverValidator := validator.NewValidator(sqidManager)

	// Create a user model with an invalid SQID (models have SQID validation tags)
	user := models.User{
		ID:             "definitely_not_a_valid_sqid",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Phone:          "+1234567890", // Valid E164 format
		Company:        "Example Inc",
		ProfilePicture: "https://example.com/profile.jpg", // Valid URL
		Roles:          []models.Role{},                   // Empty roles slice
	}

	t.Run("Client validator skips SQID validation", func(t *testing.T) {
		err := clientValidator.Validate(user)
		if err != nil {
			t.Errorf("Client validator should skip SQID validation, got error: %v", err)
		}
	})

	t.Run("Server validator enforces SQID validation", func(t *testing.T) {
		err := serverValidator.Validate(user)
		if err == nil {
			t.Error("Server validator should enforce SQID validation and fail")
		}
	})
}

func TestCoreAPIRequestStructs_ComprehensiveValidation(t *testing.T) {
	clientValidator := validator.NewClientValidator()

	tests := []struct {
		name    string
		request any
		wantErr bool
	}{
		{
			name: "CreateCredentialRequest - Valid",
			request: coreapi.CreateCredentialRequest{
				Name:   "My API Token",
				Expiry: 3600, // Required field: seconds until expiry
			},
			wantErr: false,
		},
		{
			name:    "CreateCredentialRequest - Invalid (missing required fields)",
			request: coreapi.CreateCredentialRequest{
				// Missing required Name and Expiry fields
			},
			wantErr: true,
		},
		{
			name: "SendInviteRequest - Valid",
			request: coreapi.SendInviteRequest{
				Email: "user@example.com",
				Role:  "viewer",
			},
			wantErr: false,
		},
		{
			name:    "SendInviteRequest - Invalid (missing required fields)",
			request: coreapi.SendInviteRequest{
				// Missing required Email and Role fields
			},
			wantErr: true,
		},
		{
			name: "CreateRepositoryRequest - Valid",
			request: coreapi.CreateRepositoryRequest{
				Name:          "my-repo",
				DefaultBranch: "main",
				Description:   "My data repository",
			},
			wantErr: false,
		},
		{
			name: "CreateRepositoryRequest - Invalid (missing required fields)",
			request: coreapi.CreateRepositoryRequest{
				// Missing required Name field
				Description: "Repository without required fields",
			},
			wantErr: true,
		},
		{
			name: "CreateQueryRequest - Valid",
			request: coreapi.CreateQueryRequest{
				Name:        "My Query",
				SQL:         "SELECT * FROM table",
				Description: "A simple query",
			},
			wantErr: false,
		},
		{
			name: "CreateQueryRequest - Invalid (missing required fields)",
			request: coreapi.CreateQueryRequest{
				// Name field is required in CreateQueryRequest
				Description: "Query missing required name field",
			},
			wantErr: true, // Name is required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clientValidator.Validate(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewCustomValidators tests all the new custom validation functions.
func TestNewCustomValidators(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	t.Run("SQL Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			sql     string
			wantErr bool
		}{
			{"valid select", "SELECT * FROM users WHERE id = 1", false},
			{"valid with limit", "SELECT name, email FROM users LIMIT 10", false},
			{"empty string", "", false}, // Optional field
			{"dangerous drop", "DROP TABLE users", true},
			{"dangerous delete", "DELETE FROM users", true},
			{"dangerous union", "SELECT * FROM users UNION SELECT * FROM admin", true},
			{"dangerous exec", "EXEC sp_configure", true},
			{"too long sql", strings.Repeat("SELECT * FROM table ", 10000), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateVar(tt.sql, "validsql")
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("Documentation Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			doc     string
			wantErr bool
		}{
			{"valid documentation", "This is a valid documentation string", false},
			{"empty string", "", false}, // Optional field
			{"long valid doc", strings.Repeat("This is documentation. ", 100), false},
			{"too long doc", strings.Repeat("x", 20000), true}, // Exceeds DocumentationMaxLength
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateVar(tt.doc, "validdocumentation")
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("URL Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			url     string
			wantErr bool
		}{
			{"valid https url", "https://example.com", false},
			{"valid http url", "http://example.com", false},
			{"valid with path", "https://example.com/path/to/resource", false},
			{"empty string", "", false}, // Optional field
			{"invalid scheme", "ftp://example.com", true},
			{"no scheme", "example.com", true},
			{"malformed url", "not-a-url", true},
			{"too long url", "https://" + strings.Repeat("x", 2000), true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateVar(tt.url, "validurl")
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("Phone Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			phone   string
			wantErr bool
		}{
			{"valid US phone", "+1234567890", false},
			{"valid international", "+447123456789", false},
			{"empty string", "", false}, // Optional field
			{"missing plus", "1234567890", true},
			{"too short", "+1", true}, // Changed from "+123" to "+1" which is truly too short
			{"too long", "+123456789012345678", true},
			{"non-numeric", "+12345abcde", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validator.ValidateVar(tt.phone, "validphone")
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

// TestEnhancedModelValidation tests models with improved validation.
func TestEnhancedModelValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	t.Run("User with enhanced validations", func(t *testing.T) {
		userID, _ := sqidManager.Encode("users", 123)
		user := models.User{
			ID:             userID,
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john.doe@example.com",
			Phone:          "+1234567890",
			Company:        "Example Inc.",
			ProfilePicture: "https://example.com/profile.jpg",
			Roles:          []models.Role{},
		}

		err := validator.Validate(user)
		if err != nil {
			t.Errorf("Expected valid user with enhanced validation, got error: %v", err)
		}
	})

	t.Run("User with invalid phone", func(t *testing.T) {
		userID, _ := sqidManager.Encode("users", 123)
		user := models.User{
			ID:             userID,
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john.doe@example.com",
			Phone:          "invalid-phone", // Invalid phone format
			Company:        "Example Inc.",
			ProfilePicture: "https://example.com/profile.jpg",
			Roles:          []models.Role{},
		}

		err := validator.Validate(user)
		if err == nil {
			t.Error("Expected validation error for invalid phone format")
		}
	})

	t.Run("StoredQuery with SQL validation", func(t *testing.T) {
		queryID, _ := sqidManager.Encode("queries", 123)
		userID, _ := sqidManager.Encode("users", 123)

		query := models.StoredQuery{
			ID:          queryID,
			Name:        "Test Query",
			Description: "A test query",
			SQL:         "SELECT * FROM users WHERE id = 1",
			Owner: models.User{
				ID:        userID,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Roles:     []models.Role{},
			},
			Tags:      []models.Tag{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := validator.Validate(query)
		if err != nil {
			t.Errorf("Expected valid query with SQL validation, got error: %v", err)
		}
	})

	t.Run("StoredQuery with dangerous SQL", func(t *testing.T) {
		queryID, _ := sqidManager.Encode("queries", 123)
		userID, _ := sqidManager.Encode("users", 123)

		query := models.StoredQuery{
			ID:          queryID,
			Name:        "Dangerous Query",
			Description: "A dangerous query",
			SQL:         "DROP TABLE users", // Dangerous SQL
			Owner: models.User{
				ID:        userID,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Roles:     []models.Role{},
			},
			Tags:      []models.Tag{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := validator.Validate(query)
		if err == nil {
			t.Error("Expected validation error for dangerous SQL")
		}
	})
}

func TestEnhancedValidation_ValidationResult(t *testing.T) {
	clientValidator := validator.NewClientValidator()

	t.Run("ValidateEnhanced - Valid request", func(t *testing.T) {
		req := coreapi.CreateConnectionRequest{
			Name:        "My Database Connection",
			Connector:   "postgres",
			Description: "Connection to production database",
			Details: map[string]any{
				"host":     "localhost",
				"port":     5432,
				"database": "myapp",
			},
		}

		result := clientValidator.ValidateEnhanced(req)
		
		if !result.IsValid {
			t.Errorf("Expected valid request, got invalid")
		}
		
		if result.HasErrors() {
			t.Errorf("Expected no errors, but HasErrors() returned true")
		}
		
		if result.GetUserMessage() != "" {
			t.Errorf("Expected empty user message for valid request, got: %s", result.GetUserMessage())
		}
		
		if len(result.GetFieldErrors()) != 0 {
			t.Errorf("Expected no field errors for valid request, got: %v", result.GetFieldErrors())
		}
		
		if result.GetRawErrors() != nil {
			t.Errorf("Expected no raw errors for valid request, got: %v", result.GetRawErrors())
		}
	})

	t.Run("ValidateEnhanced - Invalid request with missing required fields", func(t *testing.T) {
		req := coreapi.CreateConnectionRequest{
			// Missing required Name and Connector fields
			Description: "This request is missing required fields",
		}

		result := clientValidator.ValidateEnhanced(req)
		
		if result.IsValid {
			t.Errorf("Expected invalid request, got valid")
		}
		
		if !result.HasErrors() {
			t.Errorf("Expected errors, but HasErrors() returned false")
		}
		
		userMessage := result.GetUserMessage()
		if userMessage == "" {
			t.Errorf("Expected non-empty user message for invalid request")
		}
		
		fieldErrors := result.GetFieldErrors()
		if len(fieldErrors) == 0 {
			t.Errorf("Expected field errors for invalid request")
		}
		
		// Check that we have errors for the missing required fields
		if _, exists := fieldErrors["name"]; !exists {
			t.Errorf("Expected field error for 'name', but not found in: %v", fieldErrors)
		}
		
		if _, exists := fieldErrors["connector"]; !exists {
			t.Errorf("Expected field error for 'connector', but not found in: %v", fieldErrors)
		}
		
		if result.GetRawErrors() == nil {
			t.Errorf("Expected raw errors for invalid request")
		}
		
		// Test the Error() method for backward compatibility
		if result.Error() == "" {
			t.Errorf("Expected non-empty error string from Error() method")
		}
	})

	t.Run("ValidateVarEnhanced - Valid email", func(t *testing.T) {
		result := clientValidator.ValidateVarEnhanced("test@example.com", "email")
		
		if !result.IsValid {
			t.Errorf("Expected valid email, got invalid")
		}
		
		if result.HasErrors() {
			t.Errorf("Expected no errors for valid email")
		}
	})

	t.Run("ValidateVarEnhanced - Invalid email", func(t *testing.T) {
		result := clientValidator.ValidateVarEnhanced("not-an-email", "email")
		
		if result.IsValid {
			t.Errorf("Expected invalid email, got valid")
		}
		
		if !result.HasErrors() {
			t.Errorf("Expected errors for invalid email")
		}
		
		userMessage := result.GetUserMessage()
		if !strings.Contains(userMessage, "email") {
			t.Errorf("Expected user message to mention email validation, got: %s", userMessage)
		}
	})

	t.Run("ValidateEnhanced - Multiple field errors", func(t *testing.T) {
		// Create a struct with multiple validation errors
		req := coreapi.CreateWorkspaceRequest{
			// Missing required Name field
			Description: strings.Repeat("x", 1001), // Assuming there's a max length validation
		}

		result := clientValidator.ValidateEnhanced(req)
		
		if result.IsValid {
			t.Errorf("Expected invalid request with multiple errors, got valid")
		}
		
		userMessage := result.GetUserMessage()
		fieldErrors := result.GetFieldErrors()
		
		// Should have a generic message for multiple errors
		if len(fieldErrors) > 1 && !strings.Contains(userMessage, "Multiple validation errors") {
			t.Errorf("Expected generic message for multiple errors, got: %s", userMessage)
		}
		
		// Should have field-specific errors
		if len(fieldErrors) == 0 {
			t.Errorf("Expected field errors for invalid request")
		}
	})
}

func TestEnhancedValidation_CustomValidators(t *testing.T) {
	clientValidator := validator.NewClientValidator()

	t.Run("ValidateEnhanced - Custom token validation", func(t *testing.T) {
		// Test valid token
		validToken := "cred_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		result := clientValidator.ValidateVarEnhanced(validToken, "validtoken")
		
		if !result.IsValid {
			t.Errorf("Expected valid token, got invalid: %s", result.GetUserMessage())
		}

		// Test invalid token
		invalidToken := "invalid_token"
		result = clientValidator.ValidateVarEnhanced(invalidToken, "validtoken")
		
		if result.IsValid {
			t.Errorf("Expected invalid token, got valid")
		}
		
		userMessage := result.GetUserMessage()
		if !strings.Contains(userMessage, "API token") {
			t.Errorf("Expected user message to mention API token validation, got: %s", userMessage)
		}
	})

	t.Run("ValidateEnhanced - Custom slug validation", func(t *testing.T) {
		// Test valid slug
		result := clientValidator.ValidateVarEnhanced("valid-slug_123", "validslug")
		
		if !result.IsValid {
			t.Errorf("Expected valid slug, got invalid: %s", result.GetUserMessage())
		}

		// Test invalid slug (contains invalid characters)
		result = clientValidator.ValidateVarEnhanced("invalid slug with spaces", "validslug")
		
		if result.IsValid {
			t.Errorf("Expected invalid slug, got valid")
		}
		
		userMessage := result.GetUserMessage()
		if !strings.Contains(userMessage, "slug") {
			t.Errorf("Expected user message to mention slug validation, got: %s", userMessage)
		}
	})

	t.Run("ValidateEnhanced - Custom URL validation", func(t *testing.T) {
		// Test valid URL
		result := clientValidator.ValidateVarEnhanced("https://example.com", "validurl")
		
		if !result.IsValid {
			t.Errorf("Expected valid URL, got invalid: %s", result.GetUserMessage())
		}

		// Test invalid URL
		result = clientValidator.ValidateVarEnhanced("not-a-url", "validurl")
		
		if result.IsValid {
			t.Errorf("Expected invalid URL, got valid")
		}
		
		userMessage := result.GetUserMessage()
		if !strings.Contains(userMessage, "URL") {
			t.Errorf("Expected user message to mention URL validation, got: %s", userMessage)
		}
	})

	t.Run("ValidateEnhanced - Custom phone validation", func(t *testing.T) {
		// Test valid phone number
		result := clientValidator.ValidateVarEnhanced("+1234567890", "validphone")
		
		if !result.IsValid {
			t.Errorf("Expected valid phone number, got invalid: %s", result.GetUserMessage())
		}

		// Test invalid phone number
		result = clientValidator.ValidateVarEnhanced("123-456-7890", "validphone")
		
		if result.IsValid {
			t.Errorf("Expected invalid phone number, got valid")
		}
		
		userMessage := result.GetUserMessage()
		if !strings.Contains(userMessage, "phone number") {
			t.Errorf("Expected user message to mention phone number validation, got: %s", userMessage)
		}
	})
}

func TestEnhancedValidation_ErrorMessages(t *testing.T) {
	clientValidator := validator.NewClientValidator()

	testCases := []struct {
		name          string
		value         any
		tag           string
		expectedInMsg string
	}{
		{"required validation", "", "required", "required"},
		{"email validation", "invalid", "email", "email"},
		{"min length validation", "ab", "min=5", "at least"},
		{"max length validation", "toolongstring", "max=5", "at most"},
		{"numeric validation", "abc", "numeric", "number"},
		{"alpha validation", "abc123", "alpha", "letters"},
		{"alphanum validation", "abc-123", "alphanum", "letters and numbers"},
		{"url validation", "invalid-url", "url", "url"},
		{"startswith validation", "wrongprefix", "startswith=test", "start with"},
		{"endswith validation", "wrongsuffix", "endswith=test", "end with"},
		{"contains validation", "nomatch", "contains=test", "contain"},
		{"oneof validation", "invalid", "oneof=valid option", "one of"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := clientValidator.ValidateVarEnhanced(tc.value, tc.tag)
			
			if result.IsValid {
				t.Errorf("Expected validation to fail for %s", tc.name)
				return
			}
			
			userMessage := result.GetUserMessage()
			if !strings.Contains(strings.ToLower(userMessage), tc.expectedInMsg) {
				t.Errorf("Expected user message to contain '%s', got: %s", tc.expectedInMsg, userMessage)
			}
		})
	}
}
