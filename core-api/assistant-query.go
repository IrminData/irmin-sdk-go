package irmincore

// QueryGenerationRequest represents a request to generate a query from natural language
type QueryGenerationRequest struct {
	// Natural language prompt describing the desired query
	Prompt string `json:"prompt" validate:"required,min=1,max=1000"`

	// Optional repository slug for repository-specific queries
	RepositorySlug *string `json:"repository_slug,omitempty"`

	// Optional repository reference (branch, tag, commit)
	RepositoryRef *string `json:"repository_ref,omitempty"`

	// Optional conversation ID to continue an existing conversation
	ConversationID *uint `json:"conversation_id,omitempty"`

	// Optional metadata for the request
	Metadata map[string]any `json:"metadata,omitempty"`
}
