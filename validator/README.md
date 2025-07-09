# Validation Utility

Simple validation utility for Irmin SDK models using [go-playground/validator](https://github.com/go-playground/validator).

## Usage

```go
package main

import (
    "fmt"
    irminsdkvalidator "github.com/IrminData/irmin-sdk-go/validator"
    "github.com/IrminData/irmin-sdk-go/models"
)

func main() {
    // Create validator instance
    validator := irminsdkvalidator.NewValidator()

    // Validate a model
    user := models.User{
        ID:        "user-123",
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john@example.com",
        Phone:     "+1234567890",
        Roles:     []models.Role{{ID: "role-1", Role: "admin"}},
    }

    if err := validator.Validate(user); err != nil {
        fmt.Printf("Validation failed: %v\n", err)
    }

    // Validate individual fields
    if err := validator.ValidateVar("test@example.com", "email"); err != nil {
        fmt.Printf("Email validation failed: %v\n", err)
    }
}
```

## API

### `NewValidator(sqidManager *SQIDManager) *Validator`

Creates a new validator instance with SQID validation support.

### `NewClientValidator() *Validator`

Creates a new validator instance for client-side use (skips SQID validation).

### `Validate(s any) error`

Validates a struct and returns validation errors from go-playground/validator.

### `ValidateVar(field any, tag string) error`

Validates a single variable against a validation tag.

## Custom Validation Rules

The validator includes several custom validation rules beyond the standard go-playground/validator rules:

### Core Validators

- `validtoken` - Validates API tokens (must start with "cred_", be at least 64 chars, alphanumeric + underscore)
- `validslug` - Validates slugs/branch names (1-100 chars, alphanumeric + underscore + hyphen)
- `validsqid` - Validates SQIDs using the provided SQID manager (server-side only)

### Schedule Validators

- `validrrule` - Validates RRule expressions for recurring schedules
- `validcron` - Validates cron expressions for scheduled tasks
- `validschedule` - Validates schedule trigger configuration (cross-field validation)

### Enhanced Content Validators

- `validsql` - Validates SQL queries with length limits and basic security checks
- `validdocumentation` - Validates documentation fields with appropriate length limits (max 10,000 chars)
- `validurl` - Enhanced URL validation with scheme restrictions (http/https only, max 2,000 chars)
- `validphone` - Enhanced phone validation with E.164 format and length checks

## Validation Constants

The validator uses the following constants for validation limits:

```go
const (
    TokenPrefix            = "cred_"
    TokenLength            = 64
    SlugMinLength          = 1
    SlugMaxLength          = 100
    DocumentationMaxLength = 10000
    SQLMaxLength          = 50000
    URLMaxLength          = 2000
)
```

## Example Models

All models in the `models` package include comprehensive validation tags. Common patterns:

### Basic Validation
- `validate:"required"` - Field is required
- `validate:"email"` - Valid email format
- `validate:"min=1,max=100"` - String length constraints
- `validate:"min=0,max=10"` - Numeric range constraints

### Custom Validation
- `validate:"validsqid=users"` - Valid SQID for users table
- `validate:"validslug"` - Valid slug format
- `validate:"validurl"` - Valid HTTP/HTTPS URL
- `validate:"validsql"` - Safe SQL query
- `validate:"validdocumentation"` - Documentation with length limits

### Enum Validation
- `validate:"oneof=value1 value2 value3"` - Must be one of the specified values

### Cross-field Validation
- `validate:"required_if=Field value"` - Required if another field has specific value
- `validate:"required_with=Field"` - Required if another field is present

## Security Features

### SQL Injection Prevention
The `validsql` validator includes basic SQL injection prevention by blocking dangerous patterns:
- DDL operations: `DROP`, `CREATE`, `ALTER`, `TRUNCATE`
- DML operations: `DELETE`, `INSERT`, `UPDATE`
- System procedures: `EXEC`, `EXECUTE`, `sp_`, `xp_`
- Advanced techniques: `UNION`, `/*!`

### URL Security
The `validurl` validator restricts URLs to safe schemes:
- Allowed: `http://`, `https://`
- Blocked: `ftp://`, `file://`, `javascript:`, etc.

## Error Handling

Error handling is up to the consuming application. The validator returns detailed error messages that can be parsed and displayed to users:

```go
err := validator.Validate(model)
if err != nil {
    // Handle validation errors
    fmt.Printf("Validation failed: %v\n", err)
}
```

## Client vs Server Validation

The validator supports both client-side and server-side scenarios:

- **Server-side**: Use `NewValidator(sqidManager)` - includes full SQID validation
- **Client-side**: Use `NewClientValidator()` - skips SQID validation since clients don't have access to the SQID alphabet

This ensures the same validation rules can be applied on both sides while accommodating the different capabilities of each environment.
