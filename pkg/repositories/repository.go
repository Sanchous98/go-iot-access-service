package repositories

import (
	"github.com/google/uuid"
)

type WithResource interface {
	GetResource() string
}

type Repository[T WithResource] interface {
	Find(id uuid.UUID) T
	FindAll() []T
}
