package irminsdkvalidator

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	irminsqids "github.com/IrminData/irmin-sdk-go/sqids"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
	"github.com/teambition/rrule-go"
)

// ValidationResult contains validation results in multiple formats for different use cases.
type ValidationResult struct {
	// IsValid indicates whether the validation passed
	IsValid bool

	// UserMessage provides a single, generic error message suitable for end users
	UserMessage string

	// FieldErrors maps field names to user-friendly error messages
	FieldErrors map[string]string

	// RawErrors contains the original validation errors from the underlying library
	RawErrors error
}

// Error implements the error interface for backward compatibility.
func (vr *ValidationResult) Error() string {
	if vr.IsValid {
		return ""
	}
	if vr.UserMessage != "" {
		return vr.UserMessage
	}
	if vr.RawErrors != nil {
		return vr.RawErrors.Error()
	}
	return "validation failed"
}

// HasErrors returns true if there are any validation errors.
func (vr *ValidationResult) HasErrors() bool {
	return !vr.IsValid
}

// GetUserMessage returns a single user-friendly error message.
func (vr *ValidationResult) GetUserMessage() string {
	return vr.UserMessage
}

// GetFieldErrors returns a map of field-specific error messages.
func (vr *ValidationResult) GetFieldErrors() map[string]string {
	return vr.FieldErrors
}

// GetRawErrors returns the original validation errors.
func (vr *ValidationResult) GetRawErrors() error {
	return vr.RawErrors
}

// Validator provides validation functionality for Irmin models.
type Validator struct {
	validate    *validator.Validate
	sqidManager *irminsqids.SQIDManager
}

// Constants are now defined in constants.go

// NewValidator creates a new validator instance.
func NewValidator(sqidManager *irminsqids.SQIDManager) *Validator {
	v := validator.New()

	validator := &Validator{
		validate:    v,
		sqidManager: sqidManager,
	}

	// Register custom validation functions
	err := v.RegisterValidation("validtoken", validateToken)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validslug", validateSlug)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validsqid", validator.validateSQID)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validrrule", validateRRule)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validcron", validateCron)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validschedule", validateScheduleTrigger)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validsql", validateSQL)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validdocumentation", validateDocumentation)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validurl", validateURL)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validphone", validatePhone)
	if err != nil {
		panic(err)
	}

	// // Use JSON field names in error messages
	// v.RegisterTagNameFunc(func(fld reflect.StructField) string {
	// 	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	// 	if name == "-" {
	// 		return ""
	// 	}
	// 	return name
	// })

	return validator
}

// NewClientValidator creates a new validator instance for client-side use.
// This validator skips SQID validation since clients don't have access to the SQID alphabet.
func NewClientValidator() *Validator {
	v := validator.New()

	validator := &Validator{
		validate:    v,
		sqidManager: nil, // No SQID manager for client-side validation
	}

	// Register custom validation functions (excluding SQID validation)
	err := v.RegisterValidation("validtoken", validateToken)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validslug", validateSlug)
	if err != nil {
		panic(err)
	}
	// Register SQID validation but it will be skipped when sqidManager is nil
	err = v.RegisterValidation("validsqid", validator.validateSQID)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validrrule", validateRRule)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validcron", validateCron)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validschedule", validateScheduleTrigger)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validsql", validateSQL)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validdocumentation", validateDocumentation)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validurl", validateURL)
	if err != nil {
		panic(err)
	}
	err = v.RegisterValidation("validphone", validatePhone)
	if err != nil {
		panic(err)
	}

	return validator
}

// validateTokenPrefix is a custom validation function for API token prefixes
// Token prefixes must:
// - Start with "cred_"
// - Be at least 64 characters total
// - Contain only alphanumeric characters and underscores after the prefix.
func validateToken(fl validator.FieldLevel) bool {
	token := fl.Field().String()

	// Must start with "cred_"
	if !strings.HasPrefix(token, TokenPrefix) {
		return false
	}

	// Must be at least 64 characters (this is also checked by min=64)
	if len(token) < TokenLength {
		return false
	}

	// Check that the rest of the token contains only alphanumeric and underscores
	suffix := token[len(TokenPrefix):]
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

func validateRRule(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var rruleValue string
	if field.Kind() == reflect.Ptr {
		rruleValue = field.Elem().String()
	} else {
		rruleValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if rruleValue == "" {
		return true
	}

	// Prepare the RRule string following the same logic as orchestrator
	ruleStr := rruleValue
	ruleStr = strings.TrimPrefix(ruleStr, "RRULE:")
	ruleStr = strings.TrimSpace(ruleStr)
	ruleStr = strings.TrimSuffix(ruleStr, ";")

	// If the RRule string doesn't contain DTSTART, add it
	if !strings.Contains(ruleStr, "DTSTART") {
		// Format with newlines between components
		now := time.Now()
		ruleStr = "DTSTART:" + now.UTC().Format("20060102T150405Z") + "\n" + ruleStr
	} else {
		// Replace semicolons with newlines for existing DTSTART
		ruleStr = strings.ReplaceAll(ruleStr, ";", "\n")
	}

	// Try to parse the RRule string
	_, err := rrule.StrToRRule(ruleStr)
	return err == nil
}

func validateCron(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var cronValue string
	if field.Kind() == reflect.Ptr {
		cronValue = field.Elem().String()
	} else {
		cronValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if cronValue == "" {
		return true
	}

	// Prepare the cron expression following the same logic as orchestrator
	cronStr := cronValue
	cronStr = strings.TrimPrefix(cronStr, "CRON:")
	cronStr = strings.TrimSpace(cronStr)

	// Try to parse the cron expression
	_, err := cron.ParseStandard(cronStr)
	return err == nil
}

// validateBranchName is a custom validation function for branch names
// Branch names must:
// - Be at least 1 character
// - Be at most 100 characters
// - Contain only alphanumeric characters, underscores and hyphens.
func validateSlug(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var branchName string
	if field.Kind() == reflect.Ptr {
		branchName = field.Elem().String()
	} else {
		branchName = field.String()
	}

	// Must be at least 1 character
	if len(branchName) < SlugMinLength {
		return false
	}

	// Must be at most 100 characters
	if len(branchName) > SlugMaxLength {
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

func (v *Validator) validateSQID(fl validator.FieldLevel) bool {
	// Skip SQID validation if no SQID manager is available (client-side scenario)
	if v.sqidManager == nil {
		return true
	}

	// Get the value of the field
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var sqidValue string
	if field.Kind() == reflect.Ptr {
		sqidValue = field.Elem().String()
	} else {
		sqidValue = field.String()
	}

	typeParam := fl.Param()

	// Get the type of the sqid
	typeParam = strings.TrimSpace(typeParam)

	// Get the sqid manager
	sqidManager := v.sqidManager

	// Decode the sqid
	decoded, err := sqidManager.Decode(typeParam, sqidValue)
	if err != nil {
		return false
	}

	// Check if the decoded value is a valid uint64
	if decoded == 0 {
		return false
	}

	return true
}

func validateScheduleTrigger(fl validator.FieldLevel) bool {
	// Get the parent struct (ScheduleTrigger) from the Type field
	parentStruct := fl.Parent()

	// Make sure we're working with a struct
	if parentStruct.Kind() != reflect.Struct {
		return true // Let other validators handle non-struct cases
	}

	// Get the Type field value (this is the current field being validated)
	triggerType := fl.Field().String()

	// Only validate time triggers with this function
	if triggerType != "time" {
		return true
	}

	return validateTimeTrigger(parentStruct)
}

// validateTimeTrigger validates that time triggers have proper RRule or Cron configuration.
func validateTimeTrigger(parentStruct reflect.Value) bool {
	rruleField := parentStruct.FieldByName("RRule")
	cronField := parentStruct.FieldByName("Cron")

	rruleIsEmpty := isFieldEmpty(rruleField)
	cronIsEmpty := isFieldEmpty(cronField)

	// At least one must be provided
	if rruleIsEmpty && cronIsEmpty {
		return false
	}

	// Validate RRule if present
	if !rruleIsEmpty {
		rruleValue := rruleField.Elem().String()
		if !isValidRRule(rruleValue) {
			return false
		}
	}

	// Validate Cron if present
	if !cronIsEmpty {
		cronValue := cronField.Elem().String()
		if !isValidCron(cronValue) {
			return false
		}
	}

	return true
}

// isFieldEmpty checks if a pointer field is nil or contains an empty string.
func isFieldEmpty(field reflect.Value) bool {
	return !field.IsValid() || field.IsNil() ||
		(field.Elem().IsValid() && field.Elem().String() == "")
}

// Helper function to validate RRule.
func isValidRRule(rruleValue string) bool {
	if rruleValue == "" {
		return true
	}

	// Prepare the RRule string following the same logic as orchestrator
	ruleStr := rruleValue
	ruleStr = strings.TrimPrefix(ruleStr, "RRULE:")
	ruleStr = strings.TrimSpace(ruleStr)
	ruleStr = strings.TrimSuffix(ruleStr, ";")

	// If the RRule string doesn't contain DTSTART, add it
	if !strings.Contains(ruleStr, "DTSTART") {
		// Format with newlines between components
		now := time.Now()
		ruleStr = "DTSTART:" + now.UTC().Format("20060102T150405Z") + "\n" + ruleStr
	} else {
		// Replace semicolons with newlines for existing DTSTART
		ruleStr = strings.ReplaceAll(ruleStr, ";", "\n")
	}

	// Try to parse the RRule string
	_, err := rrule.StrToRRule(ruleStr)
	return err == nil
}

// Helper function to validate Cron.
func isValidCron(cronValue string) bool {
	if cronValue == "" {
		return true
	}

	// Prepare the cron expression following the same logic as orchestrator
	cronStr := cronValue
	cronStr = strings.TrimPrefix(cronStr, "CRON:")
	cronStr = strings.TrimSpace(cronStr)

	// Try to parse the cron expression
	_, err := cron.ParseStandard(cronStr)
	return err == nil
}

// validateSQL validates SQL queries to ensure they are safe and properly formatted.
func validateSQL(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var sqlValue string
	if field.Kind() == reflect.Ptr {
		sqlValue = field.Elem().String()
	} else {
		sqlValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if sqlValue == "" {
		return true
	}

	// Check maximum length
	if len(sqlValue) > SQLMaxLength {
		return false
	}

	// Basic SQL injection prevention - check for dangerous patterns
	sqlLower := strings.ToLower(strings.TrimSpace(sqlValue))

	// Allow common SQL operations but block potentially dangerous ones
	dangerousPatterns := []string{
		"drop ", "delete ", "truncate ", "alter ", "create ",
		"insert ", "update ", "exec ", "execute ", "sp_",
		"xp_", "union ", "/*!",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(sqlLower, pattern) {
			return false
		}
	}

	return true
}

// validateDocumentation validates documentation fields with appropriate length limits.
func validateDocumentation(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var docValue string
	if field.Kind() == reflect.Ptr {
		docValue = field.Elem().String()
	} else {
		docValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if docValue == "" {
		return true
	}

	// Check maximum length
	if len(docValue) > DocumentationMaxLength {
		return false
	}

	return true
}

// validateURL validates URLs with additional checks beyond the standard url validator.
func validateURL(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var urlValue string
	if field.Kind() == reflect.Ptr {
		urlValue = field.Elem().String()
	} else {
		urlValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if urlValue == "" {
		return true
	}

	// Check maximum length
	if len(urlValue) > URLMaxLength {
		return false
	}

	// Parse the URL to validate format
	parsedURL, err := url.Parse(urlValue)
	if err != nil {
		return false
	}

	// Check that scheme and host are present
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// Additional checks for allowed schemes
	allowedSchemes := []string{"http", "https"}
	hasValidScheme := false
	for _, scheme := range allowedSchemes {
		if strings.ToLower(parsedURL.Scheme) == scheme {
			hasValidScheme = true
			break
		}
	}

	return hasValidScheme
}

// validatePhone validates phone numbers with E.164 format and additional checks.
func validatePhone(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var phoneValue string
	if field.Kind() == reflect.Ptr {
		phoneValue = field.Elem().String()
	} else {
		phoneValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if phoneValue == "" {
		return true
	}

	// Basic E.164 validation
	// Must start with + and be followed by 1-15 digits
	if !strings.HasPrefix(phoneValue, "+") {
		return false
	}

	// Remove the + and check if the rest are digits
	digits := phoneValue[1:]
	if len(digits) < 3 || len(digits) > 15 {
		return false
	}

	// Check that all remaining characters are digits
	for _, char := range digits {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// Validate validates a struct and returns validation errors.
func (v *Validator) Validate(s any) error {
	return v.validate.Struct(s)
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field any, tag string) error {
	return v.validate.Var(field, tag)
}

// ValidateEnhanced validates a struct and returns a detailed ValidationResult.
// This provides multiple error formats for different use cases.
func (v *Validator) ValidateEnhanced(s any) *ValidationResult {
	err := v.validate.Struct(s)
	if err == nil {
		return &ValidationResult{
			IsValid:     true,
			UserMessage: "",
			FieldErrors: make(map[string]string),
			RawErrors:   nil,
		}
	}

	return v.buildValidationResult(err)
}

// ValidateVarEnhanced validates a single variable and returns a detailed ValidationResult.
func (v *Validator) ValidateVarEnhanced(field any, tag string) *ValidationResult {
	err := v.validate.Var(field, tag)
	if err == nil {
		return &ValidationResult{
			IsValid:     true,
			UserMessage: "",
			FieldErrors: make(map[string]string),
			RawErrors:   nil,
		}
	}

	return v.buildValidationResult(err)
}

// buildValidationResult converts validation errors into a structured ValidationResult.
func (v *Validator) buildValidationResult(err error) *ValidationResult {
	result := &ValidationResult{
		IsValid:     false,
		FieldErrors: make(map[string]string),
		RawErrors:   err,
	}

	// Check if it's a ValidationErrors type from go-playground/validator
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var userMessages []string

		for _, fieldError := range validationErrors {
			fieldName := v.getFieldName(fieldError)
			fieldMessage := v.getFieldErrorMessage(fieldError)

			result.FieldErrors[fieldName] = fieldMessage
			userMessages = append(userMessages, fieldMessage)
		}

		// Create a generic user message
		if len(userMessages) == 1 {
			result.UserMessage = userMessages[0]
		} else if len(userMessages) > 1 {
			result.UserMessage = "Multiple validation errors occurred. Please check the field errors for details."
		} else {
			result.UserMessage = "Validation failed"
		}
	} else {
		// Handle other types of errors
		result.UserMessage = "Validation failed: " + err.Error()
	}

	return result
}

// getFieldName extracts a user-friendly field name from a validation error.
func (v *Validator) getFieldName(fieldError validator.FieldError) string {
	// Use JSON tag name if available, otherwise use the struct field name
	field := fieldError.Field()

	// For single variable validation, field might be empty, use "field" as default
	if field == "" {
		return "field"
	}

	// For nested fields, we want to show the full path but make it user-friendly
	if strings.Contains(field, ".") {
		// Convert something like "User.Address.Street" to "user.address.street"
		parts := strings.Split(field, ".")
		for i, part := range parts {
			parts[i] = strings.ToLower(part)
		}
		return strings.Join(parts, ".")
	}

	return strings.ToLower(field)
}

// getFieldErrorMessage creates a user-friendly error message for a specific field error.
func (v *Validator) getFieldErrorMessage(fieldError validator.FieldError) string {
	field := v.getFieldName(fieldError)
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters long", field, param)
	case "numeric":
		return fmt.Sprintf("Field '%s' must be a number", field)
	case "alpha":
		return fmt.Sprintf("Field '%s' must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("Field '%s' must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, param)
	case "startswith":
		return fmt.Sprintf("Field '%s' must start with '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("Field '%s' must end with '%s'", field, param)
	case "contains":
		return fmt.Sprintf("Field '%s' must contain '%s'", field, param)
	case "validtoken":
		return fmt.Sprintf("Field '%s' must be a valid API token", field)
	case "validslug":
		return fmt.Sprintf("Field '%s' must be a valid slug (letters, numbers, hyphens, and underscores only)", field)
	case "validsqid":
		return fmt.Sprintf("Field '%s' must be a valid SQID", field)
	case "validrrule":
		return fmt.Sprintf("Field '%s' must be a valid recurrence rule", field)
	case "validcron":
		return fmt.Sprintf("Field '%s' must be a valid cron expression", field)
	case "validschedule":
		return fmt.Sprintf("Field '%s' must be a valid schedule trigger", field)
	case "validsql":
		return fmt.Sprintf("Field '%s' must be a valid SQL query", field)
	case "validdocumentation":
		return fmt.Sprintf("Field '%s' must be valid documentation", field)
	case "validurl":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "validphone":
		return fmt.Sprintf("Field '%s' must be a valid phone number in E.164 format", field)
	default:
		// Generic message for unknown validation tags
		return fmt.Sprintf("Field '%s' failed validation: %s", field, tag)
	}
}
