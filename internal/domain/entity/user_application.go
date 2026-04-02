package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserApplication representa o vínculo de um usuário com uma aplicação
// Permite que o mesmo email seja usado em diferentes aplicações
type UserApplication struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewUserApplication cria uma nova instância de UserApplication
func NewUserApplication(userID, applicationID uuid.UUID) *UserApplication {
	now := time.Now()
	return &UserApplication{
		ID:            uuid.New(),
		UserID:        userID,
		ApplicationID: applicationID,
		Active:        true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

