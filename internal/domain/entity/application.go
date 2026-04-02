package entity

import (
	"time"

	"github.com/google/uuid"
)

// Application representa uma aplicação que usa o sistema de autenticação
type Application struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"` // Ex: "retech-fin-admin"
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewApplication cria uma nova instância de Application
func NewApplication(name, code, description string) *Application {
	now := time.Now()
	return &Application{
		ID:          uuid.New(),
		Name:        name,
		Code:        code,
		Description: description,
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

