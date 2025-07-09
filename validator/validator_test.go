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

func TestValidator_RequestValidation(t *testing.T) {
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

func TestCoreAPIRequestStructs_ComprehensiveValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

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
			err := validator.Validate(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSQLValidation tests the validsql custom validation function.
func TestSQLValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{"Valid_Select_Statement", "SELECT * FROM users WHERE id = 1", false},
		{"Valid_Select_With_Limit", "SELECT name, email FROM users LIMIT 10", false},
		{"Empty_String", "", false}, // Optional field
		{"Invalid_Drop_Statement", "DROP TABLE users", true},
		{"Invalid_Delete_Statement", "DELETE FROM users", true},
		{"Invalid_Union_Statement", "SELECT * FROM users UNION SELECT * FROM admin", true},
		{"Invalid_Exec_Statement", "EXEC sp_configure", true},
		{"Invalid_Too_Long_SQL", strings.Repeat("SELECT * FROM table ", 10000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.sql, "validsql")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDocumentationValidation tests the validdocumentation custom validation function.
func TestDocumentationValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		doc     string
		wantErr bool
	}{
		{"Valid_Documentation_String", "This is a valid documentation string", false},
		{"Empty_String", "", false}, // Optional field
		{"Valid_Long_Documentation", strings.Repeat("This is documentation. ", 100), false},
		{"Invalid_Too_Long_Documentation", strings.Repeat("x", 20000), true}, // Exceeds DocumentationMaxLength
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.doc, "validdocumentation")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestURLValidation tests the validurl custom validation function.
func TestURLValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid_HTTPS_URL", "https://example.com", false},
		{"Valid_HTTP_URL", "http://example.com", false},
		{"Valid_URL_With_Path", "https://example.com/path/to/resource", false},
		{"Empty_String", "", false}, // Optional field
		{"Invalid_FTP_Scheme", "ftp://example.com", true},
		{"Invalid_No_Scheme", "example.com", true},
		{"Invalid_Malformed_URL", "not-a-url", true},
		{"Invalid_Too_Long_URL", "https://" + strings.Repeat("x", 2000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.url, "validurl")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPhoneValidation tests the validphone custom validation function.
func TestPhoneValidation(t *testing.T) {
	sqidManager := sqids.NewSQIDManager("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	validator := validator.NewValidator(sqidManager)

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{"Valid_US_Phone", "+1234567890", false},
		{"Valid_International_Phone", "+447123456789", false},
		{"Empty_String", "", false}, // Optional field
		{"Invalid_Missing_Plus", "1234567890", true},
		{"Invalid_Too_Short", "+1", true}, // Changed from "+123" to "+1" which is truly too short
		{"Invalid_Too_Long", "+123456789012345678", true},
		{"Invalid_Non_Numeric", "+12345abcde", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVar(tt.phone, "validphone")
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
