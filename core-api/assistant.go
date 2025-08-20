package irmincore

// CreateAssistantMessageRequest represents the JSON request body for creating assistant messages.
type CreateAssistantMessageRequest struct {
	Message        string `json:"message"                    validate:"required,max=100"              example:"What is the capital of France?"`
	ConversationID string `json:"conversation_id"                    validate:"required,max=100"              example:"conv_1a2b3c"`
}

// CreateAssistantConversationRequest represents the JSON request body for creating assistant conversations.
type CreateAssistantConversationRequest struct {
	Metadata map[string]any `json:"metadata"`
}
