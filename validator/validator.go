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

// ValidationResultError contains validation results in multiple formats for different use cases.
type ValidationResultError struct {
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
func (vr *ValidationResultError) Error() string {
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
func (vr *ValidationResultError) HasErrors() bool {
	return !vr.IsValid
}

// GetUserMessage returns a single user-friendly error message.
func (vr *ValidationResultError) GetUserMessage() string {
	return vr.UserMessage
}

// GetFieldErrors returns a map of field-specific error messages.
func (vr *ValidationResultError) GetFieldErrors() map[string]string {
	return vr.FieldErrors
}

// GetRawErrors returns the original validation errors.
func (vr *ValidationResultError) GetRawErrors() error {
	return vr.RawErrors
}

// Validator provides validation functionality for Irmin models.
type Validator struct {
	validate    *validator.Validate
	sqidManager *irminsqids.SQIDManager
}

// NewValidator creates a new validator instance.
func NewValidator(sqidManager *irminsqids.SQIDManager) *Validator {
	v := validator.New()

	validator := &Validator{
		validate:    v,
		sqidManager: sqidManager,
	}

	registerValidations(v, validator)
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

	registerValidations(v, validator)
	return validator
}

// registerValidations registers all custom validation functions with the validator.
func registerValidations(v *validator.Validate, validator *Validator) {
	// Register custom validation functions
	if err := v.RegisterValidation("validtoken", validateToken); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validslug", validateSlug); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validsqid", validator.validateSQID); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validrrule", validateRRule); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validcron", validateCron); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validschedule", validateScheduleTrigger); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validsql", validateSQL); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validdocumentation", validateDocumentation); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validurl", validateURL); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("validphone", validatePhone); err != nil {
		panic(err)
	}

	// Register custom pipeline stage validation
	if err := v.RegisterValidation("validpipelinestage", validator.validatePipelineStage); err != nil {
		panic(err)
	}

	// Register custom workflowable validation
	if err := v.RegisterValidation("validworkflowable", validator.validateWorkflowable); err != nil {
		panic(err)
	}

	// Register image URL validation
	if err := v.RegisterValidation("validimageurl", validateImageURL); err != nil {
		panic(err)
	}
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

	// Block only dangerous DDL and system operations
	dangerousPatterns := []string{
		"drop ", "truncate ", "alter ",
		"exec ", "execute ", "sp_", "xp_",
		"/*!",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(sqlLower, pattern) {
			return false
		}
	}

	return true
}

// validateDocumentation validates documentation fields to ensure they contain valid markdown
// while being quite permissive. Prevents injection attacks through documentation.
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

	// Basic security checks to prevent malicious content
	// Block potentially dangerous HTML/JS that could be injected through markdown
	docLower := strings.ToLower(docValue)

	// Block script tags and javascript
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "<iframe", "</iframe>",
		"<object", "</object>", "<embed", "</embed>", "<form", "</form>",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(docLower, pattern) {
			return false
		}
	}

	// Basic markdown structure validation - be permissive but check for balanced brackets
	// Count square brackets for links [text](url) and images ![alt](url)
	openSquare := strings.Count(docValue, "[")
	closeSquare := strings.Count(docValue, "]")
	openParen := strings.Count(docValue, "(")
	closeParen := strings.Count(docValue, ")")

	// Allow substantial tolerance for unbalanced brackets (markdown can be flexible)
	// Only fail if extremely unbalanced (indicating potential malformed content)
	if openSquare > 0 && closeSquare > 0 {
		bracketDiff := openSquare - closeSquare
		if bracketDiff > 20 || bracketDiff < -20 {
			return false
		}
	}

	if openParen > 0 && closeParen > 0 {
		parenDiff := openParen - closeParen
		if parenDiff > 20 || parenDiff < -20 {
			return false
		}
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

// validateImageURL validates image URLs with additional checks for image formats.
func validateImageURL(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Handle nil pointers
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return true
	}

	// Get the actual string value
	var imageURLValue string
	if field.Kind() == reflect.Ptr {
		imageURLValue = field.Elem().String()
	} else {
		imageURLValue = field.String()
	}

	// Empty string is considered valid (optional field)
	if imageURLValue == "" {
		return true
	}

	// First, validate as a regular URL
	if !validateURLInternal(imageURLValue) {
		return false
	}

	// Additional checks for image URLs
	imageURLLower := strings.ToLower(imageURLValue)

	// Check for non-image extensions that would indicate this is not an image URL
	nonImageExtensions := []string{
		".txt", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".zip", ".tar", ".gz",
		".mp3", ".mp4", ".avi", ".mov", ".html", ".htm", ".css", ".js", ".json", ".xml",
	}

	for _, ext := range nonImageExtensions {
		if strings.Contains(imageURLLower, ext) {
			return false
		}
	}

	// Allow URLs without specific extensions (could be dynamic/API generated images)
	return true
}

// validateURLInternal is a helper function that validates URL format without the field level interface.
func validateURLInternal(urlValue string) bool {
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
	for _, scheme := range allowedSchemes {
		if strings.ToLower(parsedURL.Scheme) == scheme {
			return true
		}
	}

	return false
}

// validatePipelineStage validates that pipeline stages have the correct required fields based on their type.
func (v *Validator) validatePipelineStage(fl validator.FieldLevel) bool {
	// This validator should be applied to the Type field of a PipelineStage
	// Get the parent struct (PipelineStage)
	parentStruct := fl.Parent()

	// Make sure we're working with a struct
	if parentStruct.Kind() != reflect.Struct {
		return true // Let other validators handle non-struct cases
	}

	// Get the Type field value (this is the current field being validated)
	stageType := fl.Field().String()

	switch stageType {
	case "action":
		return v.validateActionPipelineStage(parentStruct)
	case "connection":
		return v.validateConnectionPipelineStage(parentStruct)
	case "repository":
		return v.validateRepositoryPipelineStage(parentStruct)
	default:
		// Unknown stage type
		return false
	}
}

// validateActionPipelineStage validates action pipeline stages.
func (v *Validator) validateActionPipelineStage(parentStruct reflect.Value) bool {
	executableField := parentStruct.FieldByName("Executable")

	// For action stages, Executable is required
	if !executableField.IsValid() || executableField.IsNil() {
		return false
	}
	executableValue := executableField.Elem().String()
	return executableValue != ""
}

// validateConnectionPipelineStage validates connection pipeline stages.
func (v *Validator) validateConnectionPipelineStage(parentStruct reflect.Value) bool {
	connectionIDField := parentStruct.FieldByName("ConnectionID")

	// For connection stages, ConnectionID is required
	if !connectionIDField.IsValid() || connectionIDField.IsNil() {
		return false
	}
	connectionIDValue := connectionIDField.Elem().String()
	if connectionIDValue == "" {
		return false
	}

	// Validate SQID if we have a SQID manager
	if v.sqidManager != nil {
		decoded, err := v.sqidManager.Decode("connections", connectionIDValue)
		if err != nil || decoded == 0 {
			return false
		}
	}

	return true
}

// validateRepositoryPipelineStage validates repository pipeline stages.
func (v *Validator) validateRepositoryPipelineStage(parentStruct reflect.Value) bool {
	repositoryField := parentStruct.FieldByName("Repository")

	// For repository stages, Repository is required
	if !repositoryField.IsValid() || repositoryField.IsNil() {
		return false
	}
	repositoryValue := repositoryField.Elem().String()
	return repositoryValue != ""
}

// validateWorkflowable validates that workflowables have the correct required fields based on their type.
func (v *Validator) validateWorkflowable(fl validator.FieldLevel) bool {
	// This validator should be applied to the Type field of a Workflowable
	// Get the parent struct (Workflowable)
	parentStruct := fl.Parent()

	// Make sure we're working with a struct
	if parentStruct.Kind() != reflect.Struct {
		return true // Let other validators handle non-struct cases
	}

	// Get the Type field value (this is the current field being validated)
	workflowableType := fl.Field().String()

	switch workflowableType {
	case "import":
		return v.validateImportWorkflowable(parentStruct)
	case "export":
		return v.validateExportWorkflowable(parentStruct)
	case "pipeline":
		return v.validatePipelineWorkflowable(parentStruct)
	case "action":
		return v.validateActionWorkflowable(parentStruct)
	default:
		// Unknown workflowable type
		return false
	}
}

// validateImportWorkflowable validates import workflowables.
func (v *Validator) validateImportWorkflowable(parentStruct reflect.Value) bool {
	// Validate common workflowable fields
	if !v.validateCommonWorkflowableFields(parentStruct) {
		return false
	}

	// Check ImportFromConnectionPaths
	importFromConnectionPathsField := parentStruct.FieldByName("ImportFromConnectionPaths")
	if !importFromConnectionPathsField.IsValid() || importFromConnectionPathsField.Len() == 0 {
		return false
	}

	// Check ImportToRepositoryPath
	importToRepositoryPathField := parentStruct.FieldByName("ImportToRepositoryPath")
	if !importToRepositoryPathField.IsValid() || importToRepositoryPathField.String() == "" {
		return false
	}

	return true
}

// validateExportWorkflowable validates export workflowables.
func (v *Validator) validateExportWorkflowable(parentStruct reflect.Value) bool {
	// Validate common workflowable fields
	if !v.validateCommonWorkflowableFields(parentStruct) {
		return false
	}

	// Check ExportFromRepositoryPaths
	exportFromRepositoryPathsField := parentStruct.FieldByName("ExportFromRepositoryPaths")
	if !exportFromRepositoryPathsField.IsValid() || exportFromRepositoryPathsField.Len() == 0 {
		return false
	}

	// Check ExportToConnectionPath
	exportToConnectionPathField := parentStruct.FieldByName("ExportToConnectionPath")
	if !exportToConnectionPathField.IsValid() || exportToConnectionPathField.String() == "" {
		return false
	}

	return true
}

// validateCommonWorkflowableFields validates fields common to import and export workflowables.
func (v *Validator) validateCommonWorkflowableFields(parentStruct reflect.Value) bool {
	// Check FieldMappings
	fieldMappingsField := parentStruct.FieldByName("FieldMappings")
	if !fieldMappingsField.IsValid() || fieldMappingsField.Len() == 0 {
		return false
	}

	// Check ConnectionID and validate SQID
	connectionIDField := parentStruct.FieldByName("ConnectionID")
	if !v.validateConnectionID(connectionIDField) {
		return false
	}

	// Check Repository
	repositoryField := parentStruct.FieldByName("Repository")
	if !repositoryField.IsValid() || repositoryField.String() == "" {
		return false
	}

	// Check RepositoryBranch
	repositoryBranchField := parentStruct.FieldByName("RepositoryBranch")
	if !repositoryBranchField.IsValid() || repositoryBranchField.String() == "" {
		return false
	}

	return true
}

// validatePipelineWorkflowable validates pipeline workflowables.
func (v *Validator) validatePipelineWorkflowable(parentStruct reflect.Value) bool {
	// For pipeline workflowables, Stages is required
	stagesField := parentStruct.FieldByName("Stages")
	return stagesField.IsValid() && stagesField.Len() > 0
}

// validateActionWorkflowable validates action workflowables.
func (v *Validator) validateActionWorkflowable(parentStruct reflect.Value) bool {
	// For action workflowables, Executable is required
	executableField := parentStruct.FieldByName("Executable")
	return executableField.IsValid() && executableField.String() != ""
}

// validateConnectionID validates a connection ID field and its SQID if available.
func (v *Validator) validateConnectionID(connectionIDField reflect.Value) bool {
	// Check ConnectionID
	if !connectionIDField.IsValid() || connectionIDField.String() == "" {
		return false
	}

	// Validate SQID if we have a SQID manager
	if v.sqidManager != nil {
		decoded, err := v.sqidManager.Decode("connections", connectionIDField.String())
		if err != nil || decoded == 0 {
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

// ValidateEnhanced validates a struct and returns a detailed ValidationResultError.
// This provides multiple error formats for different use cases.
func (v *Validator) ValidateEnhanced(s any) *ValidationResultError {
	err := v.validate.Struct(s)
	if err == nil {
		return &ValidationResultError{
			IsValid:     true,
			UserMessage: "",
			FieldErrors: make(map[string]string),
			RawErrors:   nil,
		}
	}

	return v.buildValidationResult(err)
}

// ValidateVarEnhanced validates a single variable and returns a detailed ValidationResultError.
func (v *Validator) ValidateVarEnhanced(field any, tag string) *ValidationResultError {
	err := v.validate.Var(field, tag)
	if err == nil {
		return &ValidationResultError{
			IsValid:     true,
			UserMessage: "",
			FieldErrors: make(map[string]string),
			RawErrors:   nil,
		}
	}

	return v.buildValidationResult(err)
}

// buildValidationResult converts validation errors into a structured ValidationResultError.
func (v *Validator) buildValidationResult(err error) *ValidationResultError {
	result := &ValidationResultError{
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
		switch {
		case len(userMessages) == 1:
			result.UserMessage = userMessages[0]
		case len(userMessages) > 1:
			result.UserMessage = "Multiple validation errors occurred. Please check the field errors for details."
		default:
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

// getStandardValidationMessage returns error messages for standard validation tags.
func (v *Validator) getStandardValidationMessage(field, tag, param string) (string, bool) {
	switch tag {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field), true
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field), true
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, param), true
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, param), true
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters long", field, param), true
	case "numeric":
		return fmt.Sprintf("Field '%s' must be a number", field), true
	case "alpha":
		return fmt.Sprintf("Field '%s' must contain only letters", field), true
	case "alphanum":
		return fmt.Sprintf("Field '%s' must contain only letters and numbers", field), true
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field), true
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field), true
	default:
		return "", false
	}
}

// getStringValidationMessage returns error messages for string-related validation tags.
func (v *Validator) getStringValidationMessage(field, tag, param string) (string, bool) {
	switch tag {
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, param), true
	case "startswith":
		return fmt.Sprintf("Field '%s' must start with '%s'", field, param), true
	case "endswith":
		return fmt.Sprintf("Field '%s' must end with '%s'", field, param), true
	case "contains":
		return fmt.Sprintf("Field '%s' must contain '%s'", field, param), true
	default:
		return "", false
	}
}

// getCustomValidationMessage returns error messages for custom validation tags.
func (v *Validator) getCustomValidationMessage(field, tag string) (string, bool) {
	switch tag {
	case "validtoken":
		return fmt.Sprintf("Field '%s' must be a valid API token", field), true
	case "validslug":
		return fmt.Sprintf(
			"Field '%s' must be a valid slug (letters, numbers, hyphens, and underscores only)",
			field,
		), true
	case "validsqid":
		return fmt.Sprintf("Field '%s' must be a valid SQID", field), true
	case "validrrule":
		return fmt.Sprintf("Field '%s' must be a valid recurrence rule", field), true
	case "validcron":
		return fmt.Sprintf("Field '%s' must be a valid cron expression", field), true
	case "validschedule":
		return fmt.Sprintf("Field '%s' must be a valid schedule trigger", field), true
	case "validsql":
		return fmt.Sprintf("Field '%s' must be a valid SQL query", field), true
	case "validdocumentation":
		return fmt.Sprintf("Field '%s' must be valid documentation", field), true
	case "validurl":
		return fmt.Sprintf("Field '%s' must be a valid URL", field), true
	case "validphone":
		return fmt.Sprintf("Field '%s' must be a valid phone number in E.164 format", field), true
	case "validpipelinestage":
		return fmt.Sprintf("Field '%s' must be a valid pipeline stage type with required fields", field), true
	case "validworkflowable":
		return fmt.Sprintf("Field '%s' must be a valid workflowable type with required fields", field), true
	case "validimageurl":
		return fmt.Sprintf("Field '%s' must be a valid image URL", field), true
	default:
		return "", false
	}
}

// getFieldErrorMessage creates a user-friendly error message for a specific field error.
func (v *Validator) getFieldErrorMessage(fieldError validator.FieldError) string {
	field := v.getFieldName(fieldError)
	tag := fieldError.Tag()
	param := fieldError.Param()

	// Try standard validation messages
	if msg, found := v.getStandardValidationMessage(field, tag, param); found {
		return msg
	}

	// Try string validation messages
	if msg, found := v.getStringValidationMessage(field, tag, param); found {
		return msg
	}

	// Try custom validation messages
	if msg, found := v.getCustomValidationMessage(field, tag); found {
		return msg
	}

	// Generic message for unknown validation tags
	return fmt.Sprintf("Field '%s' failed validation: %s", field, tag)
}
