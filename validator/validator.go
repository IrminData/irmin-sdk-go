package irminsdkvalidator

import (
	"net/url"
	"reflect"
	"strings"
	"time"

	irminsqids "github.com/IrminData/irmin-sdk-go/sqids"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
	"github.com/teambition/rrule-go"
)

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

	// Register custom pipeline stage validation
	err = v.RegisterValidation("validpipelinestage", validator.validatePipelineStage)
	if err != nil {
		panic(err)
	}

	// Register custom workflowable validation
	err = v.RegisterValidation("validworkflowable", validator.validateWorkflowable)
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

	// Register custom pipeline stage validation
	err = v.RegisterValidation("validpipelinestage", validator.validatePipelineStage)
	if err != nil {
		panic(err)
	}

	// Register custom workflowable validation
	err = v.RegisterValidation("validworkflowable", validator.validateWorkflowable)
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

	// Get all the relevant fields from the parent struct
	executableField := parentStruct.FieldByName("Executable")
	connectionIDField := parentStruct.FieldByName("ConnectionID")
	repositoryField := parentStruct.FieldByName("Repository")

	switch stageType {
	case "action":
		// For action stages, Executable is required
		if !executableField.IsValid() || executableField.IsNil() {
			return false
		}
		executableValue := executableField.Elem().String()
		if executableValue == "" {
			return false
		}

	case "connection":
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

	case "repository":
		// For repository stages, Repository is required
		if !repositoryField.IsValid() || repositoryField.IsNil() {
			return false
		}
		repositoryValue := repositoryField.Elem().String()
		if repositoryValue == "" {
			return false
		}

	default:
		// Unknown stage type
		return false
	}

	return true
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

	// Get all the relevant fields from the parent struct
	fieldMappingsField := parentStruct.FieldByName("FieldMappings")
	connectionIDField := parentStruct.FieldByName("ConnectionID")
	repositoryField := parentStruct.FieldByName("Repository")
	repositoryBranchField := parentStruct.FieldByName("RepositoryBranch")
	importFromConnectionPathsField := parentStruct.FieldByName("ImportFromConnectionPaths")
	importToRepositoryPathField := parentStruct.FieldByName("ImportToRepositoryPath")
	exportFromRepositoryPathsField := parentStruct.FieldByName("ExportFromRepositoryPaths")
	exportToConnectionPathField := parentStruct.FieldByName("ExportToConnectionPath")
	stagesField := parentStruct.FieldByName("Stages")
	executableField := parentStruct.FieldByName("Executable")

	switch workflowableType {
	case "import":
		// For import workflowables, these fields are required:
		// FieldMappings, ConnectionID, Repository, RepositoryBranch, ImportFromConnectionPaths, ImportToRepositoryPath

		// Check FieldMappings
		if !fieldMappingsField.IsValid() || fieldMappingsField.Len() == 0 {
			return false
		}

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

		// Check Repository
		if !repositoryField.IsValid() || repositoryField.String() == "" {
			return false
		}

		// Check RepositoryBranch
		if !repositoryBranchField.IsValid() || repositoryBranchField.String() == "" {
			return false
		}

		// Check ImportFromConnectionPaths
		if !importFromConnectionPathsField.IsValid() || importFromConnectionPathsField.Len() == 0 {
			return false
		}

		// Check ImportToRepositoryPath
		if !importToRepositoryPathField.IsValid() || importToRepositoryPathField.String() == "" {
			return false
		}

	case "export":
		// For export workflowables, these fields are required:
		// FieldMappings, ConnectionID, Repository, RepositoryBranch, ExportFromRepositoryPaths, ExportToConnectionPath

		// Check FieldMappings
		if !fieldMappingsField.IsValid() || fieldMappingsField.Len() == 0 {
			return false
		}

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

		// Check Repository
		if !repositoryField.IsValid() || repositoryField.String() == "" {
			return false
		}

		// Check RepositoryBranch
		if !repositoryBranchField.IsValid() || repositoryBranchField.String() == "" {
			return false
		}

		// Check ExportFromRepositoryPaths
		if !exportFromRepositoryPathsField.IsValid() || exportFromRepositoryPathsField.Len() == 0 {
			return false
		}

		// Check ExportToConnectionPath
		if !exportToConnectionPathField.IsValid() || exportToConnectionPathField.String() == "" {
			return false
		}

	case "pipeline":
		// For pipeline workflowables, Stages is required
		if !stagesField.IsValid() || stagesField.Len() == 0 {
			return false
		}

	case "action":
		// For action workflowables, Executable is required
		if !executableField.IsValid() || executableField.String() == "" {
			return false
		}

	default:
		// Unknown workflowable type
		return false
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
