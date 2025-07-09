package irminsdkgo

import "time"

// API client timeout constants
const (
	// DefaultAPITimeout is the default timeout for core API requests
	DefaultAPITimeout = 10 * time.Second
	
	// DefaultConnectorTimeout is the default timeout for connector API requests
	DefaultConnectorTimeout = 120 * time.Second
)