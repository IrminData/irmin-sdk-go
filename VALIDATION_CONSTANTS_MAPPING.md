# Validation Constants Mapping

This document maps the validation constants defined in `constants.go` and `validator/constants.go` to their expected usage in struct validation tags.

## Overview

Since Go struct tags must be string literals at compile time, we cannot directly use constants in validation tags. However, we maintain constants for consistency and use them in custom validation functions. This document ensures that hardcoded values in struct tags match their corresponding constants.

## Field Length Constants Mapping

### Name Fields

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `NameMinLength` | 1 | Minimum for most name fields | `min=1` |
| `ShortNameMaxLength` | 50 | First names, last names, roles | `max=50` |
| `NameMaxLength` | 100 | Standard name fields (entities, branches, etc.) | `max=100` |
| `LongNameMaxLength` | 255 | Long names (object names, schemas) | `max=255` |

### Description Fields

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `ShortDescriptionMaxLength` | 200 | Tags, roles, help text | `max=200` |
| `DescriptionMaxLength` | 500 | Standard descriptions | `max=500` |
| `LongDescriptionMaxLength` | 1000 | Long descriptions, messages | `max=1000` |

### Technical Fields

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `VersionMaxLength` | 20 | Version strings, languages | `max=20` |
| `ContentTypeMaxLength` | 100 | MIME types, system tokens, companies | `max=100` |
| `HashMaxLength` | 64 | Git hashes, commit hashes | `max=64` |
| `RefMaxLength` | 100 | Git references, slugs | `max=100` |

### User Input Fields

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `KeyMaxLength` | 50 | Configuration keys | `max=50` |
| `ValueMaxLength` | 100 | Configuration values, examples | `max=100` |

### Locale Fields

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `LocaleMinLength` | 2 | Locale identifiers | `min=2` |
| `LocaleMaxLength` | 5 | Locale identifiers | `max=5` |

### Retention and Limits

| Constant | Value | Usage | Validation Tag |
|----------|-------|-------|----------------|
| `RetentionMinDays` | 1 | Minimum retention period | `min=1` |
| `RetentionMaxDays` | 3650 | Maximum retention period | `max=3650` |
| `MaxRetriesMin` | 0 | Minimum retries | `min=0` |
| `MaxRetriesMax` | 10 | Maximum retries | `max=10` |
| `SearchLimitMin` | 1 | Minimum search/page limits | `min=1` |
| `SearchLimitMax` | 100 | Maximum search limits | `max=100` |
| `PerPageMax` | 1000 | Maximum items per page | `max=1000` |

## Validation Rules for Update Requests

All `Update*Request` structs should follow these rules:

1. **No required fields**: Update requests should never have `required` validation since only provided fields should be updated
2. **Use `omitempty` in JSON tags**: All fields should have `omitempty` to exclude empty values from JSON
3. **Apply appropriate length limits**: Use the same validation limits as the corresponding model fields
4. **Use consistent validation**: Phone fields use `validphone`, URL fields use `validurl`

## Custom Validation Functions

The following custom validation functions use the constants defined above:

- `validurl`: Validates URLs (uses `URLMaxLength`)
- `validphone`: Validates phone numbers  
- `validdocumentation`: Validates documentation fields (uses `DocumentationMaxLength`)
- `validsql`: Validates SQL queries (uses `SQLMaxLength`)
- `validslug`: Validates slug fields (uses `SlugMinLength`, `SlugMaxLength`)
- `validtoken`: Validates API tokens (uses `TokenPrefix`, `TokenLength`)

## Current Inconsistencies to Fix

Based on the audit of existing code, the following inconsistencies need to be addressed:

1. **URL validation**: `core-api/connectors.go` uses `url` instead of `validurl` ✅ (Fixed)
2. **Missing validation**: `UpdateProfileRequest` had no validation tags ✅ (Fixed)
3. **Hardcoded values**: All struct tags should use values that match the constants above

## Implementation Notes

- Constants are defined in both `constants.go` (for general use) and `validator/constants.go` (for validator-specific use)
- Custom validation functions in `validator/validator.go` use these constants
- Struct tags must use hardcoded values but should match the constants
- All update request types have been verified to not require any fields (only update provided fields)