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

// ==========================================
// SIMPLIFIED CORE INTERFACES
// ==========================================

// CoreOperations defines essential database operations
// This is the minimal interface that all repositories must implement
type CoreOperations[T Entity] interface {
	Health(ctx context.Context) error
	WithContext(ctx context.Context) CoreOperations[T]
	WithTransaction(tx *gorm.DB) CoreOperations[T]
	DB() *gorm.DB
}

// ReadOperations extends CoreOperations with read capabilities
type ReadOperations[T Entity] interface {
	CoreOperations[T]

	// Basic reads
	FindByID(ctx context.Context, id string) (T, error)
	FindAll(ctx context.Context, params *PaginationParams) (*PaginationResult, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)

	// Type-safe condition operations
	Find(ctx context.Context, condition Condition[T]) ([]T, error)
	FindOne(ctx context.Context, condition Condition[T]) (T, error)
	FindWithPagination(ctx context.Context, condition Condition[T], params *PaginationParams) (*PaginationResult, error)
	CountWhere(ctx context.Context, condition Condition[T]) (int64, error)
	ExistsWhere(ctx context.Context, condition Condition[T]) (bool, error)
}

// WriteOperations extends CoreOperations with write capabilities
type WriteOperations[T Entity] interface {
	CoreOperations[T]

	// Basic writes
	Create(ctx context.Context, entity T) error
	Update(ctx context.Context, entity T) error
	Save(ctx context.Context, entity T) error
	Delete(ctx context.Context, id string) error

	// Condition-based writes
	UpdateWhere(ctx context.Context, updates map[string]interface{}, condition Condition[T]) error
	DeleteWhere(ctx context.Context, condition Condition[T]) error

	// Batch operations
	CreateBatch(ctx context.Context, entities []T) error
	UpdateBatch(ctx context.Context, entities []T) error
	DeleteBatch(ctx context.Context, ids []string) error
}

// SoftDeleteOperations extends CoreOperations with soft delete capabilities
type SoftDeleteOperations[T SoftDeletableEntity] interface {
	CoreOperations[T]

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

// ==========================================
// BACKWARD COMPATIBILITY INTERFACES
// ==========================================

// These interfaces maintain backward compatibility with existing code

// ReadRepository is an alias for ReadOperations
type ReadRepository[T Entity] = ReadOperations[T]

// WriteRepository is an alias for WriteOperations
type WriteRepository[T Entity] = WriteOperations[T]

// BaseOperations is an alias for CoreOperations (backward compatibility)
type BaseOperations[T Entity] = CoreOperations[T]
