# Conditional Field Validation in Go Models

## Overview

This document explains the implementation of conditional field validation in the Irmin Go SDK. Conditional validation allows certain fields to be required only when specific conditions are met, such as when another field has a particular value.

## Problem Statement

In Go, unlike TypeScript which has union types and discriminated unions, there's no clean way to express that certain fields are only required based on the value of other fields. For example:

- In pipeline stages: `Executable` should only be required when `Type` is "action"
- In workflowables: `ConnectionID` should only be required when `Type` is "import" or "export"
- In schedule triggers: `RRule` or `Cron` should only be validated when `Type` is "time"

## Solution Approach

We implemented conditional validation using custom validator functions that examine the parent struct and apply validation logic based on field relationships.

### 1. Pipeline Stage Validation

**Model**: `PipelineStage` in `models/workflow.go`

**Implementation**: Custom validator `validpipelinestage` applied to the `Type` field.

**Validation Rules**:
- **Action stages** (`Type: "action"`): `Executable` field is required and must be non-empty
- **Connection stages** (`Type: "connection"`): `ConnectionID` field is required, must be non-empty, and must be a valid SQID
- **Repository stages** (`Type: "repository"`): `Repository` field is required and must be non-empty

**Example**:
```go
type PipelineStage struct {
    Type         PipelineStageType `json:"type" validate:"required,oneof=action connection repository,validpipelinestage"`
    Executable   *string          `json:"executable,omitempty"`
    ConnectionID *string          `json:"connection_id,omitempty"`
    Repository   *string          `json:"repository,omitempty"`
    // ... other fields
}
```

### 2. Workflowable Validation

**Model**: `Workflowable` in `models/workflow.go`

**Implementation**: Custom validator `validworkflowable` applied to the `Type` field.

**Validation Rules**:
- **Import workflowables** (`Type: "import"`): Requires `FieldMappings`, `ConnectionID`, `Repository`, `RepositoryBranch`, `ImportFromConnectionPaths`, `ImportToRepositoryPath`
- **Export workflowables** (`Type: "export"`): Requires `FieldMappings`, `ConnectionID`, `Repository`, `RepositoryBranch`, `ExportFromRepositoryPaths`, `ExportToConnectionPath`
- **Pipeline workflowables** (`Type: "pipeline"`): Requires `Stages` (non-empty slice)
- **Action workflowables** (`Type: "action"`): Requires `Executable`

### 3. Schedule Trigger Validation

**Model**: `ScheduleTrigger` in `models/schedule.go`

**Implementation**: Custom validator `validschedule` applied to the `Type` field.

**Validation Rules**:
- **Time triggers** (`Type: "time"`): At least one of `RRule` or `Cron` must be provided and valid
- **Repository event triggers** (`Type: "repository-event"`): `RepositoryEvent` is required
- **Workflow run event triggers** (`Type: "workflow-run-event"`): `WorkflowRunEvent` is required

## Implementation Details

### Custom Validator Functions

Custom validators are implemented as methods on the `Validator` struct:

```go
func (v *Validator) validatePipelineStage(fl validator.FieldLevel) bool {
    // Get parent struct and examine field values
    parentStruct := fl.Parent()
    stageType := fl.Field().String()
    
    // Apply conditional validation based on type
    switch stageType {
    case "action":
        // Validate that Executable is present and non-empty
    case "connection":
        // Validate that ConnectionID is present, non-empty, and valid SQID
    case "repository":
        // Validate that Repository is present and non-empty
    }
    
    return true
}
```

### Registration

Custom validators are registered in both `NewValidator` and `NewClientValidator` functions:

```go
err = v.RegisterValidation("validpipelinestage", validator.validatePipelineStage)
if err != nil {
    panic(err)
}
```

### Model Tags

The validation tag is applied to the discriminator field (typically `Type`):

```go
Type PipelineStageType `json:"type" validate:"required,oneof=action connection repository,validpipelinestage"`
```

## Validation Tag Strategy

For conditional fields, we use a simplified tagging approach:

1. **Discriminator field**: Gets the custom validator (`validpipelinestage`, `validworkflowable`, etc.)
2. **Conditional fields**: Use `omitempty` for optional validation but no `required_if` tags
3. **Always optional fields**: Use `omitempty,min=1` for basic validation when present

### Why Not Use `required_if`?

The `go-playground/validator` package's `required_if` tag doesn't work well with `omitempty` for complex conditional scenarios. When `omitempty` is used, the field validation is skipped entirely, preventing `required_if` from being evaluated. Removing `omitempty` causes validation errors on nil pointer fields.

Our custom validator approach provides:
- **Cleaner logic**: Complex conditional rules in readable Go code
- **Better error handling**: Custom error messages and logic
- **SQID validation**: Integration with our SQID validation system
- **Flexibility**: Easy to extend for complex business rules

## Testing

Each conditional validation scenario has comprehensive test coverage:

### Valid Cases
- Each type with its required fields properly set
- Verification that non-required fields for the type don't cause validation errors

### Invalid Cases
- Each type missing its required fields
- Verification that validation fails appropriately

### Test Examples

```go
// Valid action stage
stage := models.PipelineStage{
    Type:       models.PipelineStageTypeAction,
    Executable: &executable, // Required for action stages
    // ConnectionID and Repository are nil - should not cause errors
}

// Invalid action stage
stage := models.PipelineStage{
    Type: models.PipelineStageTypeAction,
    // Missing Executable - should fail validation
}
```

## Adding New Conditional Validation

To add conditional validation to a new model:

1. **Identify the discriminator field** (usually `Type`)
2. **Create a custom validator function**:
   ```go
   func (v *Validator) validateMyModel(fl validator.FieldLevel) bool {
       parentStruct := fl.Parent()
       modelType := fl.Field().String()
       
       switch modelType {
       case "type1":
           // Validate required fields for type1
       case "type2":
           // Validate required fields for type2
       }
       
       return true
   }
   ```
3. **Register the validator** in both `NewValidator` and `NewClientValidator`
4. **Update the model tags**:
   ```go
   Type MyModelType `json:"type" validate:"required,oneof=type1 type2,validmymodel"`
   ```
5. **Add comprehensive tests** for all type combinations

## Best Practices

1. **Keep custom validators focused**: Each validator should handle one model type
2. **Validate SQID fields**: Use the validator's `sqidManager` when available
3. **Handle nil pointers safely**: Always check `IsValid()` and `IsNil()` for pointer fields
4. **Test thoroughly**: Cover all type combinations and edge cases
5. **Document validation rules**: Clearly specify what fields are required for each type
6. **Use `omitempty`**: For truly optional fields that should only be validated when present

## Migration Notes

When updating existing models with conditional validation:

1. **Review existing `required_if` tags**: These may need to be replaced with custom validators
2. **Add `omitempty` to optional fields**: Prevents validation errors on nil/empty fields
3. **Update tests**: Ensure test coverage for all conditional scenarios
4. **Verify backward compatibility**: Existing valid data should continue to validate

## Example Migration

**Before** (problematic):
```go
type MyModel struct {
    Type  string  `json:"type" validate:"required,oneof=a b"`
    FieldA *string `json:"field_a,omitempty" validate:"required_if=Type a"`
    FieldB *string `json:"field_b,omitempty" validate:"required_if=Type b"`
}
```

**After** (working):
```go
type MyModel struct {
    Type   string  `json:"type" validate:"required,oneof=a b,validmymodel"`
    FieldA *string `json:"field_a,omitempty"`
    FieldB *string `json:"field_b,omitempty"`
}
```

With custom validator:
```go
func (v *Validator) validateMyModel(fl validator.FieldLevel) bool {
    parentStruct := fl.Parent()
    modelType := fl.Field().String()
    
    fieldAField := parentStruct.FieldByName("FieldA")
    fieldBField := parentStruct.FieldByName("FieldB")
    
    switch modelType {
    case "a":
        return !fieldAField.IsNil() && fieldAField.Elem().String() != ""
    case "b":
        return !fieldBField.IsNil() && fieldBField.Elem().String() != ""
    }
    
    return false
}
```

## Conclusion

The custom validator approach provides a robust, maintainable solution for conditional field validation in Go. It addresses the limitations of the `required_if` tag while providing clear, testable validation logic that integrates well with the existing validation framework.