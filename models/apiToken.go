package irminmodels

import "time"

type APIToken struct {
	ID        string    `json:"id" validate:"required"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
	Name      string    `json:"name" validate:"required,min=1,max=100"`
	Token     string    `json:"token,omitempty" validate:"min=64,validtoken"`
	ExpiresAt time.Time `json:"expiry" validate:"required,gtfield=CreatedAt"`
}
