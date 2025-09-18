package store

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/upnext-fng/fulcrum/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ==========================================
// REPOSITORY CONFIGURATION OPTIONS
// ==========================================

// RepositoryOption defines a configuration option for the repository
type RepositoryOption[T Entity] func(*RepositoryImpl[T])

// WithSoftDelete enables soft delete functionality for the repository
func WithSoftDelete[T Entity]() RepositoryOption[T] {
	return func(r *RepositoryImpl[T]) {
		r.softDelete = true
	}
}

// WithCustomTable sets a custom table name for the repository
func WithCustomTable[T Entity](tableName string) RepositoryOption[T] {
	return func(r *RepositoryImpl[T]) {
		r.tableName = tableName
	}
}

// WithHooks sets custom hooks for the repository
func WithHooks[T Entity](hooks *RepositoryHooks[T]) RepositoryOption[T] {
	return func(r *RepositoryImpl[T]) {
		r.hooks = hooks
	}
}

// RepositoryHooks defines custom hooks for repository operations
type RepositoryHooks[T Entity] struct {
	BeforeCreate func(ctx context.Context, entity T) error
	AfterCreate  func(ctx context.Context, entity T) error
	BeforeUpdate func(ctx context.Context, entity T) error
	AfterUpdate  func(ctx context.Context, entity T) error
	BeforeDelete func(ctx context.Context, id string) error
	AfterDelete  func(ctx context.Context, id string) error
}

// ==========================================
// UNIFIED REPOSITORY IMPLEMENTATION
// ==========================================

// RepositoryImpl is a unified repository that handles all operations
// It replaces the complex composition pattern with a single struct
type RepositoryImpl[T Entity] struct {
	db         *gorm.DB
	logger     *logger.Logger
	modelType  reflect.Type
	tableName  string
	softDelete bool
	hooks      *RepositoryHooks[T]
}

// NewRepositoryImpl creates a new repository with the given options
func NewRepositoryImpl[T Entity](db *gorm.DB, logger *logger.Logger, opts ...RepositoryOption[T]) RepositoryImpl[T] {
	var zero T
	modelType := reflect.TypeOf(zero)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	repo := &RepositoryImpl[T]{
		db:         db,
		logger:     logger,
		modelType:  modelType,
		tableName:  modelType.Name(),
		softDelete: false,
		hooks:      nil,
	}

	// Apply options
	for _, opt := range opts {
		opt(repo)
	}

	return *repo
}

// ==========================================
// CORE OPERATIONS IMPLEMENTATION
// ==========================================

// Health checks the database connection
func (r *RepositoryImpl[T]) Health(ctx context.Context) error {
	var result int
	return r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}

// WithContext returns a new repository instance with the given context
func (r *RepositoryImpl[T]) WithContext(ctx context.Context) CoreOperations[T] {
	newRepo := *r
	newRepo.db = r.db.WithContext(ctx)
	return &newRepo
}

// WithTransaction returns a new repository instance with the given transaction
func (r *RepositoryImpl[T]) WithTransaction(tx *gorm.DB) CoreOperations[T] {
	newRepo := *r
	newRepo.db = tx
	return &newRepo
}

// DB returns the underlying database connection
func (r *RepositoryImpl[T]) DB() *gorm.DB {
	return r.db
}

// ==========================================
// READ OPERATIONS IMPLEMENTATION
// ==========================================

// FindByID finds an entity by its ID
func (r *RepositoryImpl[T]) FindByID(ctx context.Context, id string) (T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if IsRecordNotFound(err) {
			r.logger.Debug("Entity not found", zap.String("id", id))
		} else {
			r.logger.WithErr(err).Error("failed to find entity", zap.String("id", id))
		}
		var zero T
		return zero, err
	}
	return entity, nil
}

// FindAll finds all entities with pagination
func (r *RepositoryImpl[T]) FindAll(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []T
	query := r.db.WithContext(ctx).Model(new(T))
	return Paginate(ctx, query, params, &entities)
}

// Count returns the total count of entities
func (r *RepositoryImpl[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Count(&count).Error
	return count, err
}

// Exists checks if an entity exists by its ID
func (r *RepositoryImpl[T]) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Find finds entities matching the given condition
func (r *RepositoryImpl[T]) Find(ctx context.Context, condition Condition[T]) ([]T, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	var entities []T

	err := r.db.WithContext(ctx).Where(sql, args...).Find(&entities).Error
	if err != nil {
		r.logger.WithErr(err).Error("failed to find entities")
		return nil, err
	}

	return entities, nil
}

// FindOne finds a single entity matching the given condition
func (r *RepositoryImpl[T]) FindOne(ctx context.Context, condition Condition[T]) (T, error) {
	if err := condition.Validate(); err != nil {
		var zero T
		return zero, fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	var entity T

	err := r.db.WithContext(ctx).Where(sql, args...).First(&entity).Error
	if err != nil {
		if IsRecordNotFound(err) {
			r.logger.Debug("Entity not found")
		} else {
			r.logger.WithErr(err).Error("failed to find entity")
		}
		var zero T
		return zero, err
	}

	return entity, nil
}

// FindWithPagination finds entities matching the given condition with pagination
func (r *RepositoryImpl[T]) FindWithPagination(ctx context.Context, condition Condition[T], params *PaginationParams) (*PaginationResult, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	if params == nil {
		params = DefaultPaginationParams()
	}

	sql, args := condition.ToSQL()
	var entities []T

	query := r.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}

	return Paginate(ctx, query, params, &entities)
}

// CountWhere returns the count of entities matching the given condition
func (r *RepositoryImpl[T]) CountWhere(ctx context.Context, condition Condition[T]) (int64, error) {
	if err := condition.Validate(); err != nil {
		return 0, fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	var count int64

	query := r.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}

	return count, query.Count(&count).Error
}

// ExistsWhere checks if any entity exists matching the given condition
func (r *RepositoryImpl[T]) ExistsWhere(ctx context.Context, condition Condition[T]) (bool, error) {
	count, err := r.CountWhere(ctx, condition)
	return count > 0, err
}

// ==========================================
// WRITE OPERATIONS IMPLEMENTATION
// ==========================================

// Create creates a new entity
func (r *RepositoryImpl[T]) Create(ctx context.Context, entity T) error {
	if reflect.ValueOf(entity).IsNil() {
		return fmt.Errorf("entity cannot be nil")
	}

	// Execute before create hook if available
	if r.hooks != nil && r.hooks.BeforeCreate != nil {
		if err := r.hooks.BeforeCreate(ctx, entity); err != nil {
			return err
		}
	}

	if (entity).GetID() == "" {
		(entity).SetID(uuid.New().String())
	}

	err := r.db.WithContext(ctx).Create(entity).Error
	if err != nil {
		r.logger.WithErr(err).Error("failed to create entity")
		return err
	}

	// Execute after create hook if available
	if r.hooks != nil && r.hooks.AfterCreate != nil {
		if err := r.hooks.AfterCreate(ctx, entity); err != nil {
			return err
		}
	}

	r.logger.Infof("Created entity with ID %s", (entity).GetID())
	return nil
}

// Update updates an existing entity
func (r *RepositoryImpl[T]) Update(ctx context.Context, entity T) error {
	if reflect.ValueOf(entity).IsNil() {
		return fmt.Errorf("entity cannot be nil")
	}

	// Execute before update hook if available
	if r.hooks != nil && r.hooks.BeforeUpdate != nil {
		if err := r.hooks.BeforeUpdate(ctx, entity); err != nil {
			return err
		}
	}

	if (entity).GetID() == "" {
		return fmt.Errorf("entity ID cannot be empty for update")
	}

	err := r.db.WithContext(ctx).Save(entity).Error
	if err != nil {
		r.logger.WithErr(err).Error("failed to update entity")
		return err
	}

	// Execute after update hook if available
	if r.hooks != nil && r.hooks.AfterUpdate != nil {
		if err := r.hooks.AfterUpdate(ctx, entity); err != nil {
			return err
		}
	}

	r.logger.Infof("Updated entity with ID %s", (entity).GetID())
	return nil
}

// Save creates or updates an entity
func (r *RepositoryImpl[T]) Save(ctx context.Context, entity T) error {
	if reflect.ValueOf(entity).IsNil() {
		return fmt.Errorf("entity cannot be nil")
	}

	if (entity).GetID() == "" {
		return r.Create(ctx, entity)
	}

	return r.Update(ctx, entity)
}

// Delete deletes an entity by its ID
func (r *RepositoryImpl[T]) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	// Execute before delete hook if available
	if r.hooks != nil && r.hooks.BeforeDelete != nil {
		if err := r.hooks.BeforeDelete(ctx, id); err != nil {
			return err
		}
	}

	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to delete entity")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// Execute after delete hook if available
	if r.hooks != nil && r.hooks.AfterDelete != nil {
		if err := r.hooks.AfterDelete(ctx, id); err != nil {
			return err
		}
	}

	r.logger.Infof("Deleted entity with ID %s", id)
	return nil
}

// UpdateWhere updates entities matching the given condition
func (r *RepositoryImpl[T]) UpdateWhere(ctx context.Context, updates map[string]interface{}, condition Condition[T]) error {
	if len(updates) == 0 {
		return fmt.Errorf("updates cannot be empty")
	}

	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	query := r.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}

	result := query.Updates(updates)
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to update entities")
		return result.Error
	}

	r.logger.Infof("Updated %d entities", result.RowsAffected)
	return nil
}

// DeleteWhere deletes entities matching the given condition
func (r *RepositoryImpl[T]) DeleteWhere(ctx context.Context, condition Condition[T]) error {
	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	query := r.db.WithContext(ctx)
	if sql != "" {
		query = query.Where(sql, args...)
	}

	result := query.Delete(new(T))
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to delete entities")
		return result.Error
	}

	r.logger.Infof("Deleted %d entities", result.RowsAffected)
	return nil
}

// CreateBatch creates multiple entities in a batch
func (r *RepositoryImpl[T]) CreateBatch(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return fmt.Errorf("entities slice cannot be empty")
	}

	for _, entity := range entities {
		if (entity).GetID() == "" {
			(entity).SetID(uuid.New().String())
		}
	}

	err := r.db.WithContext(ctx).CreateInBatches(entities, 1000).Error
	if err != nil {
		r.logger.WithErr(err).Error("failed to create batch")
		return err
	}

	r.logger.Infof("Created %d entities", len(entities))
	return nil
}

// UpdateBatch updates multiple entities
func (r *RepositoryImpl[T]) UpdateBatch(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return fmt.Errorf("entities slice cannot be empty")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Save(entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple entities by their IDs
func (r *RepositoryImpl[T]) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("IDs slice cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(new(T), "id IN ?", ids)
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to delete batch")
		return result.Error
	}

	r.logger.Infof("Deleted %d entities", result.RowsAffected)
	return nil
}

// ==========================================
// SOFT DELETE OPERATIONS IMPLEMENTATION
// ==========================================

// SoftDelete soft deletes an entity by its ID
func (r *RepositoryImpl[T]) SoftDelete(ctx context.Context, id string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(new(T)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)

	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to soft delete")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.logger.Info("Entity soft deleted", zap.String("id", id))
	return nil
}

// SoftDeleteBatch soft deletes multiple entities by their IDs
func (r *RepositoryImpl[T]) SoftDeleteBatch(ctx context.Context, ids []string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(new(T)).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Update("deleted_at", now)

	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to soft delete batch")
		return result.Error
	}

	r.logger.Infof("Soft deleted %d entities", result.RowsAffected)
	return nil
}

// SoftDeleteWhere soft deletes entities matching the given condition
func (r *RepositoryImpl[T]) SoftDeleteWhere(ctx context.Context, condition Condition[T]) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	now := time.Now()

	query := r.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}
	query = query.Where("deleted_at IS NULL")

	return query.Update("deleted_at", now).Error
}

// Restore restores a soft deleted entity by its ID
func (r *RepositoryImpl[T]) Restore(ctx context.Context, id string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)

	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to restore")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.logger.Info("Entity restored", zap.String("id", id))
	return nil
}

// RestoreBatch restores multiple soft deleted entities by their IDs
func (r *RepositoryImpl[T]) RestoreBatch(ctx context.Context, ids []string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	result := r.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("id IN ? AND deleted_at IS NOT NULL", ids).
		Update("deleted_at", nil)

	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to restore batch")
		return result.Error
	}

	r.logger.Infof("Restored %d entities", result.RowsAffected)
	return nil
}

// RestoreWhere restores entities matching the given condition
func (r *RepositoryImpl[T]) RestoreWhere(ctx context.Context, condition Condition[T]) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()

	query := r.db.WithContext(ctx).Unscoped().Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}
	query = query.Where("deleted_at IS NOT NULL")

	return query.Update("deleted_at", nil).Error
}

// FindWithDeleted finds all entities including soft deleted ones
func (r *RepositoryImpl[T]) FindWithDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if !r.softDelete {
		return nil, ErrSoftDeleteNotEnabledError
	}

	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []T
	query := r.db.WithContext(ctx).Unscoped().Model(new(T))
	return Paginate(ctx, query, params, &entities)
}

// FindOnlyDeleted finds only soft deleted entities
func (r *RepositoryImpl[T]) FindOnlyDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if !r.softDelete {
		return nil, ErrSoftDeleteNotEnabledError
	}

	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []T
	query := r.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("deleted_at IS NOT NULL")
	return Paginate(ctx, query, params, &entities)
}

// ForceDelete permanently deletes an entity by its ID
func (r *RepositoryImpl[T]) ForceDelete(ctx context.Context, id string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Unscoped().Delete(new(T), "id = ?", id)
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to force delete")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.logger.Info("Entity force deleted", zap.String("id", id))
	return nil
}

// ForceDeleteBatch permanently deletes multiple entities by their IDs
func (r *RepositoryImpl[T]) ForceDeleteBatch(ctx context.Context, ids []string) error {
	if !r.softDelete {
		return ErrSoftDeleteNotEnabledError
	}

	result := r.db.WithContext(ctx).Unscoped().Delete(new(T), "id IN ?", ids)
	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("failed to force delete batch")
		return result.Error
	}

	r.logger.Infof("Force deleted %d entities", result.RowsAffected)
	return nil
}

// CleanupDeleted permanently deletes soft deleted entities older than the specified duration
func (r *RepositoryImpl[T]) CleanupDeleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	if !r.softDelete {
		return 0, ErrSoftDeleteNotEnabledError
	}

	cutoffTime := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoffTime).
		Delete(new(T))

	if result.Error != nil {
		r.logger.WithErr(result.Error).Error("Failed to cleanup deleted entities")
		return 0, result.Error
	}

	r.logger.Infof("Deleted %d entities older than %s", result.RowsAffected, olderThan)
	return result.RowsAffected, nil
}

// ==========================================
// FACTORY FUNCTIONS FOR BACKWARD COMPATIBILITY
// ==========================================

// NewRepository creates a new repository instance for basic operations
func NewRepository[T Entity](db *gorm.DB, logger *logger.Logger) Repository[T] {
	repo := NewRepositoryImpl[T](db, logger)
	return &repo
}

// NewSoftDeleteRepository creates a new repository instance with soft delete enabled
func NewSoftDeleteRepository[T SoftDeletableEntity](db *gorm.DB, logger *logger.Logger) SoftDeleteRepository[T] {
	repo := NewRepositoryImpl[T](db, logger, WithSoftDelete[T]())
	return &repo
}
