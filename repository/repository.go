package repository

import "context"

// Repository interface to be used as a port
type Repository interface {
	Create(ctx context.Context, entity interface{}) (string, error)
	Get(ctx context.Context, filter map[string]interface{}, skip, take *int) ([]interface{}, error)
	GetByID(ctx context.Context, ID string) (interface{}, error)
	Update(ctx context.Context, ID string, entity interface{}) error
	Delete(ctx context.Context, ID string) error
}
