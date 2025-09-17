package store

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Entity represents a base entity interface
type Entity interface {
	GetID() string
	SetID(string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// SoftDeletableEntity represents an entity that supports soft deletion
type SoftDeletableEntity interface {
	Entity
	GetDeletedAt() *time.Time
	SetDeletedAt(*time.Time)
	IsDeleted() bool
}

//// ReadRepository defines read operations for entities
//type ReadRepository[T Entity] interface {
//	FindByID(ctx context.Context, id string) (*T, error)
//	FindAll(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
//	Count(ctx context.Context) (int64, error)
//	Exists(ctx context.Context, id string) (bool, error)
//
//	// Type-safe condition operations
//	Find(ctx context.Context, condition Condition[T]) ([]*T, error)
//	FindOne(ctx context.Context, condition Condition[T]) (*T, error)
//	FindWithPagination(ctx context.Context, condition Condition[T], params *PaginationParams) (*PaginationResult, error)
//	CountWhere(ctx context.Context, condition Condition[T]) (int64, error)
//	ExistsWhere(ctx context.Context, condition Condition[T]) (bool, error)
//
//	Health(ctx context.Context) error
//}
//
//// WriteRepository defines write operations for entities
//type WriteRepository[T Entity] interface {
//	Create(ctx context.Context, entity *T) error
//	Update(ctx context.Context, entity *T) error
//	Save(ctx context.Context, entity *T) error
//	Delete(ctx context.Context, id string) error
//
//	// Type-safe condition operations
//	UpdateWhere(ctx context.Context, updates map[string]interface{}, condition Condition[T]) error
//	DeleteWhere(ctx context.Context, condition Condition[T]) error
//
//	// Batch operations
//	CreateBatch(ctx context.Context, entities []*T) error
//	UpdateBatch(ctx context.Context, entities []*T) error
//	DeleteBatch(ctx context.Context, ids []string) error
//}
//
//// Repository combines read and write operations
//type Repository[T Entity] interface {
//	ReadRepository[T]
//	WriteRepository[T]
//}
//
//// SoftDeleteRepository extends Repository with soft delete operations
//type SoftDeleteRepository[T SoftDeletableEntity] interface {
//	Repository[T]
//
//	SoftDelete(ctx context.Context, id string) error
//	SoftDeleteBatch(ctx context.Context, ids []string) error
//	SoftDeleteWhere(ctx context.Context, condition Condition[T]) error
//
//	Restore(ctx context.Context, id string) error
//	RestoreBatch(ctx context.Context, ids []string) error
//	RestoreWhere(ctx context.Context, condition Condition[T]) error
//
//	FindWithDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
//	FindOnlyDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
//
//	ForceDelete(ctx context.Context, id string) error
//	ForceDeleteBatch(ctx context.Context, ids []string) error
//
//	CleanupDeleted(ctx context.Context, olderThan time.Duration) (int64, error)
//}

// BaseOperations defines core database operations
type BaseOperations[T Entity] interface {
	Health(ctx context.Context) error
	WithContext(ctx context.Context) BaseOperations[T]
	WithTransaction(tx *gorm.DB) BaseOperations[T]
	DB() *gorm.DB
}

// ReadOperations defines read-only operations
type ReadOperations[T Entity] interface {
	BaseOperations[T]

	// Basic reads
	FindByID(ctx context.Context, id string) (*T, error)
	FindAll(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)

	// Type-safe condition operations
	Find(ctx context.Context, condition Condition[T]) ([]*T, error)
	FindOne(ctx context.Context, condition Condition[T]) (*T, error)
	FindWithPagination(ctx context.Context, condition Condition[T], params *PaginationParams) (*PaginationResult, error)
	CountWhere(ctx context.Context, condition Condition[T]) (int64, error)
	ExistsWhere(ctx context.Context, condition Condition[T]) (bool, error)
}

// WriteOperations defines write operations
type WriteOperations[T Entity] interface {
	BaseOperations[T]

	// Basic writes
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, entity *T) error
	Save(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id string) error

	// Condition-based writes
	UpdateWhere(ctx context.Context, updates map[string]interface{}, condition Condition[T]) error
	DeleteWhere(ctx context.Context, condition Condition[T]) error

	// Batch operations
	CreateBatch(ctx context.Context, entities []*T) error
	UpdateBatch(ctx context.Context, entities []*T) error
	DeleteBatch(ctx context.Context, ids []string) error
}

// SoftDeleteOperations defines soft delete operations
type SoftDeleteOperations[T SoftDeletableEntity] interface {
	BaseOperations[T]

	// Soft delete
	SoftDelete(ctx context.Context, id string) error
	SoftDeleteBatch(ctx context.Context, ids []string) error
	SoftDeleteWhere(ctx context.Context, condition Condition[T]) error

	// Restore
	Restore(ctx context.Context, id string) error
	RestoreBatch(ctx context.Context, ids []string) error
	RestoreWhere(ctx context.Context, condition Condition[T]) error

	// Query with deleted
	FindWithDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
	FindOnlyDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error)

	// Force delete
	ForceDelete(ctx context.Context, id string) error
	ForceDeleteBatch(ctx context.Context, ids []string) error

	// Cleanup
	CleanupDeleted(ctx context.Context, olderThan time.Duration) (int64, error)
}

// ==========================================
// COMPOSED REPOSITORY INTERFACES
// ==========================================

// ReadRepository combines base and read operations
type ReadRepository[T Entity] interface {
	ReadOperations[T]
}

// WriteRepository combines base and write operations
type WriteRepository[T Entity] interface {
	WriteOperations[T]
}

// Repository combines read and write operations
type Repository[T Entity] interface {
	ReadOperations[T]
	WriteOperations[T]
}

// SoftDeleteRepository combines all operations including soft delete
type SoftDeleteRepository[T SoftDeletableEntity] interface {
	ReadOperations[T]
	WriteOperations[T]
	SoftDeleteOperations[T]
}
