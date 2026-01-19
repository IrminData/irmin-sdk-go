package schema

import (
	"fmt"
	"slices"
	"strings"

	irminmodels "github.com/IrminData/irmin-sdk-go/models"
)

// CompareSchemas compares two ObjectSchemas and returns a detailed diff.
// The source is typically the current/existing schema, target is the new/expected schema.
func CompareSchemas(source, target *irminmodels.ObjectSchema) *irminmodels.SchemaDiff {
	diff := &irminmodels.SchemaDiff{
		Compatible:         true,
		BreakingChanges:    []irminmodels.SchemaFieldDiff{},
		NonBreakingChanges: []irminmodels.SchemaFieldDiff{},
	}

	if source == nil && target == nil {
		diff.Summary = "Both schemas are nil"
		return diff
	}

	if source == nil {
		diff.Summary = "Source schema is nil, target has schema"
		return diff
	}

	if target == nil {
		diff.Summary = "Target schema is nil, source has schema"
		return diff
	}

	// Set schema paths
	sourcePath := source.Path
	targetPath := target.Path
	diff.SourceSchemaPath = &sourcePath
	diff.TargetSchemaPath = &targetPath

	// Compare based on schema type
	switch {
	case source.Type == irminmodels.ObjectTypeGroup && target.Type == irminmodels.ObjectTypeGroup:
		compareGroupSchemas(source, target, "", diff)
	case source.Type == irminmodels.ObjectTypeStructured && target.Type == irminmodels.ObjectTypeStructured:
		if source.Schema != nil && target.Schema != nil {
			compareJSONSchemas(source.Schema, target.Schema, source.Name, diff)
		}
	default:
		if source.Type != target.Type {
			typeChange := irminmodels.SchemaFieldDiff{
				FieldPath:  source.Name,
				ChangeType: irminmodels.SchemaChangeTypeChanged,
				SourceType: ptr(string(source.Type)),
				TargetType: ptr(string(target.Type)),
			}
			desc := fmt.Sprintf("Schema type changed from %s to %s", source.Type, target.Type)
			typeChange.Description = &desc
			diff.BreakingChanges = append(diff.BreakingChanges, typeChange)
		}
	}

	// Update compatibility flag
	diff.Compatible = len(diff.BreakingChanges) == 0

	// Generate summary
	diff.Summary = generateSummary(diff)

	return diff
}

// compareGroupSchemas compares two group schemas by comparing their children.
func compareGroupSchemas(source, target *irminmodels.ObjectSchema, basePath string, diff *irminmodels.SchemaDiff) {
	// Build maps of children by name for efficient lookup
	sourceChildren := make(map[string]*irminmodels.ObjectSchema)
	for i := range source.Children {
		child := &source.Children[i]
		sourceChildren[child.Name] = child
	}

	targetChildren := make(map[string]*irminmodels.ObjectSchema)
	for i := range target.Children {
		child := &target.Children[i]
		targetChildren[child.Name] = child
	}

	// Find removed children (in source but not in target)
	for name, sourceChild := range sourceChildren {
		if _, exists := targetChildren[name]; !exists {
			fieldPath := joinPath(basePath, name)
			removed := irminmodels.SchemaFieldDiff{
				FieldPath:  fieldPath,
				ChangeType: irminmodels.SchemaChangeRemoved,
				SourceType: ptr(string(sourceChild.Type)),
			}
			desc := fmt.Sprintf("Schema '%s' was removed", name)
			removed.Description = &desc
			// Removing a schema is breaking if it contained required data
			diff.BreakingChanges = append(diff.BreakingChanges, removed)
		}
	}

	// Find added children (in target but not in source)
	for name, targetChild := range targetChildren {
		if _, exists := sourceChildren[name]; !exists {
			fieldPath := joinPath(basePath, name)
			added := irminmodels.SchemaFieldDiff{
				FieldPath:  fieldPath,
				ChangeType: irminmodels.SchemaChangeAdded,
				TargetType: ptr(string(targetChild.Type)),
			}
			desc := fmt.Sprintf("Schema '%s' was added", name)
			added.Description = &desc
			diff.NonBreakingChanges = append(diff.NonBreakingChanges, added)
		}
	}

	// Compare children that exist in both
	for name, sourceChild := range sourceChildren {
		if targetChild, exists := targetChildren[name]; exists {
			childPath := joinPath(basePath, name)
			if sourceChild.Type == irminmodels.ObjectTypeStructured &&
				targetChild.Type == irminmodels.ObjectTypeStructured &&
				sourceChild.Schema != nil && targetChild.Schema != nil {
				compareJSONSchemas(sourceChild.Schema, targetChild.Schema, childPath, diff)
			}
		}
	}
}

// compareJSONSchemas compares two JSONSchemas and populates the diff.
func compareJSONSchemas(source, target *irminmodels.JSONSchema, basePath string, diff *irminmodels.SchemaDiff) {
	// Compare types
	if source.Type != target.Type {
		typeChange := irminmodels.SchemaFieldDiff{
			FieldPath:  basePath,
			ChangeType: irminmodels.SchemaChangeTypeChanged,
			SourceType: &source.Type,
			TargetType: &target.Type,
		}
		desc := fmt.Sprintf("Type changed from %s to %s", source.Type, target.Type)
		typeChange.Description = &desc

		if isBreakingTypeChange(source.Type, target.Type) {
			diff.BreakingChanges = append(diff.BreakingChanges, typeChange)
		} else {
			diff.NonBreakingChanges = append(diff.NonBreakingChanges, typeChange)
		}
	}

	// Compare required fields
	compareRequiredFields(source, target, basePath, diff)

	// Compare properties for object types
	if source.Type == "object" && target.Type == "object" {
		compareProperties(source.Properties, target.Properties, source.Required, target.Required, basePath, diff)
	}

	// Compare items for array types
	if source.Type == "array" && target.Type == "array" {
		if source.Items != nil && target.Items != nil {
			itemPath := basePath + "[]"
			compareJSONSchemas(source.Items, target.Items, itemPath, diff)
		}
	}
}

// compareRequiredFields compares the required field lists between schemas.
func compareRequiredFields(source, target *irminmodels.JSONSchema, basePath string, diff *irminmodels.SchemaDiff) {
	// Fields that became required (breaking)
	for _, field := range target.Required {
		if !slices.Contains(source.Required, field) {
			fieldPath := joinPath(basePath, field)
			change := irminmodels.SchemaFieldDiff{
				FieldPath:   fieldPath,
				ChangeType:  irminmodels.SchemaChangeRequiredChanged,
				WasRequired: ptr(false),
				IsRequired:  ptr(true),
			}
			desc := fmt.Sprintf("Field '%s' is now required", field)
			change.Description = &desc
			diff.BreakingChanges = append(diff.BreakingChanges, change)
		}
	}

	// Fields that became optional (non-breaking)
	for _, field := range source.Required {
		if !slices.Contains(target.Required, field) {
			fieldPath := joinPath(basePath, field)
			change := irminmodels.SchemaFieldDiff{
				FieldPath:   fieldPath,
				ChangeType:  irminmodels.SchemaChangeRequiredChanged,
				WasRequired: ptr(true),
				IsRequired:  ptr(false),
			}
			desc := fmt.Sprintf("Field '%s' is no longer required", field)
			change.Description = &desc
			diff.NonBreakingChanges = append(diff.NonBreakingChanges, change)
		}
	}
}

// compareProperties compares property definitions between two object schemas.
func compareProperties(
	sourceProps, targetProps map[string]irminmodels.JSONSchema,
	_, targetRequired []string,
	basePath string,
	diff *irminmodels.SchemaDiff,
) {
	// Find removed properties
	for name, sourceProp := range sourceProps {
		if _, exists := targetProps[name]; !exists {
			fieldPath := joinPath(basePath, name)
			removed := irminmodels.SchemaFieldDiff{
				FieldPath:  fieldPath,
				ChangeType: irminmodels.SchemaChangeRemoved,
				SourceType: &sourceProp.Type,
			}
			desc := fmt.Sprintf("Field '%s' was removed", name)
			removed.Description = &desc
			// Removing a field is breaking
			diff.BreakingChanges = append(diff.BreakingChanges, removed)
		}
	}

	// Find added properties
	for name, targetProp := range targetProps {
		if _, exists := sourceProps[name]; !exists {
			fieldPath := joinPath(basePath, name)
			added := irminmodels.SchemaFieldDiff{
				FieldPath:  fieldPath,
				ChangeType: irminmodels.SchemaChangeAdded,
				TargetType: &targetProp.Type,
			}
			desc := fmt.Sprintf("Field '%s' was added", name)
			added.Description = &desc

			// Adding a required field is breaking
			if slices.Contains(targetRequired, name) {
				added.IsRequired = ptr(true)
				diff.BreakingChanges = append(diff.BreakingChanges, added)
			} else {
				diff.NonBreakingChanges = append(diff.NonBreakingChanges, added)
			}
		}
	}

	// Compare properties that exist in both
	for name, sourceProp := range sourceProps {
		if targetProp, exists := targetProps[name]; exists {
			fieldPath := joinPath(basePath, name)
			compareJSONSchemas(&sourceProp, &targetProp, fieldPath, diff)
		}
	}
}

// isBreakingTypeChange determines if a type change is breaking.
func isBreakingTypeChange(sourceType, targetType string) bool {
	// Compatible type changes (widening)
	compatibleChanges := map[string][]string{
		"integer": {"number"},        // integer -> number is ok
		"null":    {"string", "any"}, // null -> string/any is ok
	}

	if compatible, ok := compatibleChanges[sourceType]; ok {
		return !slices.Contains(compatible, targetType)
	}

	// All other type changes are breaking
	return true
}

// IsBreakingChange determines if a single field diff represents a breaking change.
func IsBreakingChange(fieldDiff irminmodels.SchemaFieldDiff) bool {
	switch fieldDiff.ChangeType {
	case irminmodels.SchemaChangeRemoved:
		return true
	case irminmodels.SchemaChangeTypeChanged:
		if fieldDiff.SourceType != nil && fieldDiff.TargetType != nil {
			return isBreakingTypeChange(*fieldDiff.SourceType, *fieldDiff.TargetType)
		}
		return true
	case irminmodels.SchemaChangeRequiredChanged:
		// Becoming required is breaking
		if fieldDiff.IsRequired != nil && *fieldDiff.IsRequired {
			return true
		}
		return false
	case irminmodels.SchemaChangeNullabilityChanged:
		// Becoming non-nullable is breaking
		if fieldDiff.IsNullable != nil && !*fieldDiff.IsNullable {
			return true
		}
		return false
	case irminmodels.SchemaChangeAdded:
		// Adding a required field is breaking
		if fieldDiff.IsRequired != nil && *fieldDiff.IsRequired {
			return true
		}
		return false
	case irminmodels.SchemaChangeModified:
		// Generic modification is not necessarily breaking
		return false
	}
	return false
}

// generateSummary creates a human-readable summary of the diff.
func generateSummary(diff *irminmodels.SchemaDiff) string {
	breaking := len(diff.BreakingChanges)
	nonBreaking := len(diff.NonBreakingChanges)

	if breaking == 0 && nonBreaking == 0 {
		return "No schema changes detected"
	}

	var parts []string
	if breaking > 0 {
		parts = append(parts, fmt.Sprintf("%d breaking change(s)", breaking))
	}
	if nonBreaking > 0 {
		parts = append(parts, fmt.Sprintf("%d non-breaking change(s)", nonBreaking))
	}

	return strings.Join(parts, ", ")
}

// joinPath joins path segments with a dot separator.
func joinPath(base, field string) string {
	if base == "" {
		return field
	}
	return base + "." + field
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}
