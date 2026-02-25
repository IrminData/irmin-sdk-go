package irminmodels

import "time"

// AsyncJobStatus represents the status of an async job.
type AsyncJobStatus string

const (
	AsyncJobStatusPending   AsyncJobStatus = "pending"
	AsyncJobStatusRunning   AsyncJobStatus = "running"
	AsyncJobStatusCompleted AsyncJobStatus = "completed"
	AsyncJobStatusFailed    AsyncJobStatus = "failed"
)

// AsyncJobType represents the type of async job.
type AsyncJobType string

const (
	AsyncJobTypeZipDownload AsyncJobType = "zip_download"
)

// AsyncJob represents an asynchronous background job response.
type AsyncJob struct {
	ID           string         `json:"id"             validate:"required,validsqid=async-jobs"  example:"aj_2k8n9q1m7p3x4z"`
	CreatedAt    time.Time      `json:"created_at"     validate:"required"                       example:"2025-01-15T10:30:00Z"`
	UpdatedAt    time.Time      `json:"updated_at"     validate:"required"                       example:"2025-01-15T10:35:00Z"`
	Type         AsyncJobType   `json:"type"           validate:"required"                       example:"zip_download"`
	Status       AsyncJobStatus `json:"status"         validate:"required"                       example:"pending"`
	Progress     int            `json:"progress"                                                 example:"50"`
	ErrorMessage string         `json:"error_message,omitempty"                                  example:""`
	ResultExpiry *time.Time     `json:"result_expiry,omitempty"`
}
