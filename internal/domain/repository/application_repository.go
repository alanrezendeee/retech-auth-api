package repository

import (
	"context"

	"github.com/theretech/retechauth-api/internal/domain/entity"
	"github.com/google/uuid"
)

// ApplicationRepository define a interface para operações com aplicações
type ApplicationRepository interface {
	Create(ctx context.Context, app *entity.Application) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Application, error)
	FindByCode(ctx context.Context, code string) (*entity.Application, error)
	Update(ctx context.Context, app *entity.Application) error
	UpsertByCode(ctx context.Context, app *entity.Application) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entity.Application, error)
}

