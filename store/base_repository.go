package store

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/upnext-fng/fulcrum/logger"
	"gorm.io/gorm"
)

// ==========================================
// REPOSITORY IMPLEMENTATIONS
// ==========================================

// BaseRepository provides core database functionality
type BaseRepository[T Entity] struct {
	db        *gorm.DB
	logger    logger.Logger
	modelType reflect.Type
	tableName string
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T Entity](db *gorm.DB, logger logger.Logger) *BaseRepository[T] {
	var zero T
	modelType := reflect.TypeOf(zero)

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	//stmt := &gorm.Statement{DB: db}
	//var tableName string
	//if err := stmt.Parse(zero); err != nil {
	//	logger.Error().Err(err).Str("model_type", modelType.Name()).Msg("Failed to parse model, using default table name")
	//	// Use a fallback table name based on the model type
	//	tableName = modelType.Name()
	//} else {
	//	tableName = stmt.Schema.Table
	//}

	return &BaseRepository[T]{
		db:        db,
		logger:    logger.WithComponent("repository").WithModule(modelType.Name()),
		modelType: modelType,
		tableName: modelType.Name(),
	}
}

// Base operations implementation
func (r *BaseRepository[T]) WithContext(ctx context.Context) BaseOperations[T] {
	return &BaseRepository[T]{
		db:        r.db.WithContext(ctx),
		logger:    r.logger,
		modelType: r.modelType,
		tableName: r.tableName,
	}
}

func (r *BaseRepository[T]) WithTransaction(tx *gorm.DB) BaseOperations[T] {
	return &BaseRepository[T]{
		db:        tx,
		logger:    r.logger,
		modelType: r.modelType,
		tableName: r.tableName,
	}
}

func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepository[T]) Health(ctx context.Context) error {
	var result int
	return r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}

// ==========================================
// READ OPERATIONS IMPLEMENTATION
// ==========================================

// ReadOnlyRepository implements only read operations
type ReadOnlyRepository[T Entity] struct {
	*BaseRepository[T]
}

// NewReadOnlyRepository creates a read-only repository
func NewReadOnlyRepository[T Entity](db *gorm.DB, logger logger.Logger) ReadRepository[T] {
	return &ReadOnlyRepository[T]{
		BaseRepository: NewBaseRepository[T](db, logger),
	}
}

// Implement ReadOperations interface
func (r *ReadOnlyRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if IsRecordNotFound(err) {
			r.logger.Debug().Str("id", id).Msg("Entity not found")
		} else {
			r.logger.Error().Err(err).Str("id", id).Msg("Failed to find entity")
		}
		return nil, err
	}
	return &entity, nil
}

func (r *ReadOnlyRepository[T]) FindAll(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []*T
	query := r.db.WithContext(ctx).Model(new(T))
	return Paginate(ctx, query, params, &entities)
}

func (r *ReadOnlyRepository[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Count(&count).Error
	return count, err
}

func (r *ReadOnlyRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *ReadOnlyRepository[T]) Find(ctx context.Context, condition Condition[T]) ([]*T, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	var entities []*T

	err := r.db.WithContext(ctx).Where(sql, args...).Find(&entities).Error
	if err != nil {
		r.logger.Error().Err(err).Str("sql", sql).Msg("Failed to find entities")
		return nil, err
	}

	return entities, nil
}

func (r *ReadOnlyRepository[T]) FindOne(ctx context.Context, condition Condition[T]) (*T, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	var entity T

	err := r.db.WithContext(ctx).Where(sql, args...).First(&entity).Error
	if err != nil {
		if IsRecordNotFound(err) {
			r.logger.Debug().Str("sql", sql).Msg("Entity not found")
		} else {
			r.logger.Error().Err(err).Str("sql", sql).Msg("Failed to find entity")
		}
		return nil, err
	}

	return &entity, nil
}

func (r *ReadOnlyRepository[T]) FindWithPagination(ctx context.Context, condition Condition[T], params *PaginationParams) (*PaginationResult, error) {
	if err := condition.Validate(); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	if params == nil {
		params = DefaultPaginationParams()
	}

	sql, args := condition.ToSQL()
	var entities []*T

	query := r.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}

	return Paginate(ctx, query, params, &entities)
}

func (r *ReadOnlyRepository[T]) CountWhere(ctx context.Context, condition Condition[T]) (int64, error) {
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

func (r *ReadOnlyRepository[T]) ExistsWhere(ctx context.Context, condition Condition[T]) (bool, error) {
	count, err := r.CountWhere(ctx, condition)
	return count > 0, err
}

// Override context methods to return correct type
func (r *ReadOnlyRepository[T]) WithContext(ctx context.Context) BaseOperations[T] {
	return &ReadOnlyRepository[T]{
		BaseRepository: &BaseRepository[T]{
			db:        r.db.WithContext(ctx),
			logger:    r.logger,
			modelType: r.modelType,
			tableName: r.tableName,
		},
	}
}

func (r *ReadOnlyRepository[T]) WithTransaction(tx *gorm.DB) BaseOperations[T] {
	return &ReadOnlyRepository[T]{
		BaseRepository: &BaseRepository[T]{
			db:        tx,
			logger:    r.logger,
			modelType: r.modelType,
			tableName: r.tableName,
		},
	}
}

// ==========================================
// WRITE OPERATIONS IMPLEMENTATION
// ==========================================

// WriteOnlyRepository implements only write operations
type WriteOnlyRepository[T Entity] struct {
	*BaseRepository[T]
}

// NewWriteOnlyRepository creates a write-only repository
func NewWriteOnlyRepository[T Entity](db *gorm.DB, logger logger.Logger) WriteRepository[T] {
	return &WriteOnlyRepository[T]{
		BaseRepository: NewBaseRepository[T](db, logger),
	}
}

func (r *WriteOnlyRepository[T]) Create(ctx context.Context, entity *T) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}

	if (*entity).GetID() == "" {
		(*entity).SetID(uuid.New().String())
	}

	err := r.db.WithContext(ctx).Create(entity).Error
	if err != nil {
		r.logger.Error().Err(err).Str("id", (*entity).GetID()).Msg("Failed to create entity")
		return err
	}

	r.logger.Info().Str("id", (*entity).GetID()).Msg("Entity created successfully")
	return nil
}

func (r *WriteOnlyRepository[T]) Update(ctx context.Context, entity *T) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}

	if (*entity).GetID() == "" {
		return fmt.Errorf("entity ID cannot be empty for update")
	}

	err := r.db.WithContext(ctx).Save(entity).Error
	if err != nil {
		r.logger.Error().Err(err).Str("id", (*entity).GetID()).Msg("Failed to update entity")
		return err
	}

	r.logger.Info().Str("id", (*entity).GetID()).Msg("Entity updated successfully")
	return nil
}

func (r *WriteOnlyRepository[T]) Save(ctx context.Context, entity *T) error {
	if entity == nil {
		return fmt.Errorf("entity cannot be nil")
	}

	if (*entity).GetID() == "" {
		return r.Create(ctx, entity)
	}

	return r.Update(ctx, entity)
}

func (r *WriteOnlyRepository[T]) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(new(T), "id = ?", id)
	if result.Error != nil {
		r.logger.Error().Err(result.Error).Str("id", id).Msg("Failed to delete entity")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.logger.Info().Str("id", id).Msg("Entity deleted successfully")
	return nil
}

func (r *WriteOnlyRepository[T]) UpdateWhere(ctx context.Context, updates map[string]interface{}, condition Condition[T]) error {
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
		r.logger.Error().Err(result.Error).Msg("Failed to update entities")
		return result.Error
	}

	r.logger.Info().Int64("rows_affected", result.RowsAffected).Msg("Entities updated")
	return nil
}

func (r *WriteOnlyRepository[T]) DeleteWhere(ctx context.Context, condition Condition[T]) error {
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
		r.logger.Error().Err(result.Error).Msg("Failed to delete entities")
		return result.Error
	}

	r.logger.Info().Int64("rows_affected", result.RowsAffected).Msg("Entities deleted")
	return nil
}

func (r *WriteOnlyRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return fmt.Errorf("entities slice cannot be empty")
	}

	for _, entity := range entities {
		if (*entity).GetID() == "" {
			(*entity).SetID(uuid.New().String())
		}
	}

	err := r.db.WithContext(ctx).CreateInBatches(entities, 1000).Error
	if err != nil {
		r.logger.Error().Err(err).Int("count", len(entities)).Msg("Failed to create batch")
		return err
	}

	r.logger.Info().Int("count", len(entities)).Msg("Batch created successfully")
	return nil
}

func (r *WriteOnlyRepository[T]) UpdateBatch(ctx context.Context, entities []*T) error {
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

func (r *WriteOnlyRepository[T]) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("IDs slice cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(new(T), "id IN ?", ids)
	if result.Error != nil {
		r.logger.Error().Err(result.Error).Strs("ids", ids).Msg("Failed to delete batch")
		return result.Error
	}

	r.logger.Info().Strs("ids", ids).Int64("rows_affected", result.RowsAffected).Msg("Batch deleted")
	return nil
}

// Override context methods to return correct type
func (r *WriteOnlyRepository[T]) WithContext(ctx context.Context) BaseOperations[T] {
	return &WriteOnlyRepository[T]{
		BaseRepository: &BaseRepository[T]{
			db:        r.db.WithContext(ctx),
			logger:    r.logger,
			modelType: r.modelType,
			tableName: r.tableName,
		},
	}
}

func (r *WriteOnlyRepository[T]) WithTransaction(tx *gorm.DB) BaseOperations[T] {
	return &WriteOnlyRepository[T]{
		BaseRepository: &BaseRepository[T]{
			db:        tx,
			logger:    r.logger,
			modelType: r.modelType,
			tableName: r.tableName,
		},
	}
}

// ==========================================
// FULL REPOSITORY IMPLEMENTATION
// ==========================================

// FullRepository implements both read and write operations
type FullRepository[T Entity] struct {
	*ReadOnlyRepository[T]
	*WriteOnlyRepository[T]
}

// DB returns the underlying database connection
func (r *FullRepository[T]) DB() *gorm.DB {
	return r.ReadOnlyRepository.DB()
}

// Health checks the database connection
func (r *FullRepository[T]) Health(ctx context.Context) error {
	return r.ReadOnlyRepository.Health(ctx)
}

// NewRepository creates a full repository with read and write operations
func NewRepository[T Entity](db *gorm.DB, logger logger.Logger) Repository[T] {
	baseRepo := NewBaseRepository[T](db, logger)

	return &FullRepository[T]{
		ReadOnlyRepository: &ReadOnlyRepository[T]{
			BaseRepository: baseRepo,
		},
		WriteOnlyRepository: &WriteOnlyRepository[T]{
			BaseRepository: baseRepo,
		},
	}
}

// Override context methods to maintain composition
func (r *FullRepository[T]) WithContext(ctx context.Context) BaseOperations[T] {
	baseRepo := &BaseRepository[T]{
		db:        r.ReadOnlyRepository.db.WithContext(ctx),
		logger:    r.ReadOnlyRepository.logger,
		modelType: r.ReadOnlyRepository.modelType,
		tableName: r.ReadOnlyRepository.tableName,
	}

	return &FullRepository[T]{
		ReadOnlyRepository:  &ReadOnlyRepository[T]{BaseRepository: baseRepo},
		WriteOnlyRepository: &WriteOnlyRepository[T]{BaseRepository: baseRepo},
	}
}

func (r *FullRepository[T]) WithTransaction(tx *gorm.DB) BaseOperations[T] {
	baseRepo := &BaseRepository[T]{
		db:        tx,
		logger:    r.ReadOnlyRepository.logger,
		modelType: r.ReadOnlyRepository.modelType,
		tableName: r.ReadOnlyRepository.tableName,
	}

	return &FullRepository[T]{
		ReadOnlyRepository:  &ReadOnlyRepository[T]{BaseRepository: baseRepo},
		WriteOnlyRepository: &WriteOnlyRepository[T]{BaseRepository: baseRepo},
	}
}

// ==========================================
// SOFT DELETE REPOSITORY IMPLEMENTATION
// ==========================================

// SoftDeleteFullRepository implements all operations including soft delete
type SoftDeleteFullRepository[T SoftDeletableEntity] struct {
	*FullRepository[T]
	softDeleteOps *SoftDeleteOps[T]
}

// DB returns the underlying database connection
func (r *SoftDeleteFullRepository[T]) DB() *gorm.DB {
	return r.FullRepository.DB()
}

// Health checks the database connection
func (r *SoftDeleteFullRepository[T]) Health(ctx context.Context) error {
	return r.FullRepository.Health(ctx)
}

// SoftDeleteOps contains soft delete specific operations
type SoftDeleteOps[T SoftDeletableEntity] struct {
	*BaseRepository[T]
}

// NewSoftDeleteRepository creates a repository with soft delete capabilities
func NewSoftDeleteRepository[T SoftDeletableEntity](db *gorm.DB, logger logger.Logger) SoftDeleteRepository[T] {
	baseRepo := NewBaseRepository[T](db, logger)
	fullRepo := &FullRepository[T]{
		ReadOnlyRepository:  &ReadOnlyRepository[T]{BaseRepository: baseRepo},
		WriteOnlyRepository: &WriteOnlyRepository[T]{BaseRepository: baseRepo},
	}

	return &SoftDeleteFullRepository[T]{
		FullRepository: fullRepo,
		softDeleteOps:  &SoftDeleteOps[T]{BaseRepository: baseRepo},
	}
}

// Implement soft delete operations
func (r *SoftDeleteFullRepository[T]) SoftDelete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	now := time.Now()
	result := r.softDeleteOps.db.WithContext(ctx).Model(new(T)).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)

	if result.Error != nil {
		r.softDeleteOps.logger.Error().Err(result.Error).Str("id", id).Msg("Failed to soft delete")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.softDeleteOps.logger.Info().Str("id", id).Msg("Entity soft deleted")
	return nil
}

func (r *SoftDeleteFullRepository[T]) Restore(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	result := r.softDeleteOps.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)

	if result.Error != nil {
		r.softDeleteOps.logger.Error().Err(result.Error).Str("id", id).Msg("Failed to restore")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.softDeleteOps.logger.Info().Str("id", id).Msg("Entity restored")
	return nil
}

func (r *SoftDeleteFullRepository[T]) ForceDelete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	result := r.softDeleteOps.db.WithContext(ctx).Unscoped().Delete(new(T), "id = ?", id)
	if result.Error != nil {
		r.softDeleteOps.logger.Error().Err(result.Error).Str("id", id).Msg("Failed to force delete")
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	r.softDeleteOps.logger.Info().Str("id", id).Msg("Entity force deleted")
	return nil
}

func (r *SoftDeleteFullRepository[T]) FindWithDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []*T
	query := r.softDeleteOps.db.WithContext(ctx).Unscoped().Model(new(T))
	return Paginate(ctx, query, params, &entities)
}

func (r *SoftDeleteFullRepository[T]) FindOnlyDeleted(ctx context.Context, params *PaginationParams) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	var entities []*T
	query := r.softDeleteOps.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("deleted_at IS NOT NULL")
	return Paginate(ctx, query, params, &entities)
}

func (r *SoftDeleteFullRepository[T]) CleanupDeleted(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result := r.softDeleteOps.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoffTime).
		Delete(new(T))

	if result.Error != nil {
		r.softDeleteOps.logger.Error().Err(result.Error).Msg("Failed to cleanup deleted entities")
		return 0, result.Error
	}

	r.softDeleteOps.logger.Info().
		Int64("deleted_count", result.RowsAffected).
		Dur("older_than", olderThan).
		Msg("Cleanup completed")

	return result.RowsAffected, nil
}

// Implement other soft delete methods...
func (r *SoftDeleteFullRepository[T]) SoftDeleteBatch(ctx context.Context, ids []string) error {
	// Implementation similar to SoftDelete but for multiple IDs
	now := time.Now()
	result := r.softDeleteOps.db.WithContext(ctx).Model(new(T)).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Update("deleted_at", now)
	return result.Error
}

func (r *SoftDeleteFullRepository[T]) SoftDeleteWhere(ctx context.Context, condition Condition[T]) error {
	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()
	now := time.Now()

	query := r.softDeleteOps.db.WithContext(ctx).Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}
	query = query.Where("deleted_at IS NULL")

	return query.Update("deleted_at", now).Error
}

func (r *SoftDeleteFullRepository[T]) RestoreBatch(ctx context.Context, ids []string) error {
	result := r.softDeleteOps.db.WithContext(ctx).Unscoped().Model(new(T)).
		Where("id IN ? AND deleted_at IS NOT NULL", ids).
		Update("deleted_at", nil)
	return result.Error
}

func (r *SoftDeleteFullRepository[T]) RestoreWhere(ctx context.Context, condition Condition[T]) error {
	if err := condition.Validate(); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	sql, args := condition.ToSQL()

	query := r.softDeleteOps.db.WithContext(ctx).Unscoped().Model(new(T))
	if sql != "" {
		query = query.Where(sql, args...)
	}
	query = query.Where("deleted_at IS NOT NULL")

	return query.Update("deleted_at", nil).Error
}

func (r *SoftDeleteFullRepository[T]) ForceDeleteBatch(ctx context.Context, ids []string) error {
	result := r.softDeleteOps.db.WithContext(ctx).Unscoped().Delete(new(T), "id IN ?", ids)
	return result.Error
}
