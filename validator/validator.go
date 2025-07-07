package irminsdkvalidator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation functionality for Irmin models
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("validtoken", validateToken)
	v.RegisterValidation("validbranchname", validateBranchName)

	// Use JSON field names in error messages
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

// validateTokenPrefix is a custom validation function for API token prefixes
// Token prefixes must:
// - Start with "cred_"
// - Be at least 64 characters total
// - Contain only alphanumeric characters and underscores after the prefix
func validateToken(fl validator.FieldLevel) bool {
	token := fl.Field().String()

	// Must start with "cred_"
	if !strings.HasPrefix(token, "cred_") {
		return false
	}

	// Must be at least 64 characters (this is also checked by min=64)
	if len(token) < 64 {
		return false
	}

	// Check that the rest of the token contains only alphanumeric and underscores
	suffix := token[5:] // Remove "cred_" prefix
	for _, char := range suffix {
		isAlphaNumeric := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9')
		isUnderscore := char == '_'

		if !isAlphaNumeric && !isUnderscore {
			return false
		}
	}

	return true
}

// validateBranchName is a custom validation function for branch names
// Branch names must:
// - Be at least 1 character
// - Be at most 100 characters
// - Contain only alphanumeric characters, underscores and hyphens
func validateBranchName(fl validator.FieldLevel) bool {
	branchName := fl.Field().String()

	// Must be at least 1 character
	if len(branchName) < 1 {
		return false
	}

	// Must be at most 100 characters
	if len(branchName) > 100 {
		return false
	}

	// Check that the branch name contains only alphanumeric characters and underscores
	for _, char := range branchName {
		isAlphaNumeric := (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9')
		isUnderscore := char == '_'
		isHyphen := char == '-'

		if !isAlphaNumeric && !isUnderscore && !isHyphen {
			return false
		}
	}

	return true
}

// Validate validates a struct and returns validation errors
func (v *Validator) Validate(s any) error {
	return v.validate.Struct(s)
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}
