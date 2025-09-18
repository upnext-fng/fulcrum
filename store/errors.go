package store

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ErrorCode represents different types of store errors
type ErrorCode string

const (
	// Entity errors
	ErrEntityNotFound      ErrorCode = "ENTITY_NOT_FOUND"
	ErrEntityAlreadyExists ErrorCode = "ENTITY_ALREADY_EXISTS"
	ErrEntityValidation    ErrorCode = "ENTITY_VALIDATION_FAILED"
	ErrEntityNil           ErrorCode = "ENTITY_NIL"
	ErrEntityIDEmpty       ErrorCode = "ENTITY_ID_EMPTY"

	// Database errors
	ErrDatabaseConnection  ErrorCode = "DATABASE_CONNECTION_FAILED"
	ErrDatabaseTimeout     ErrorCode = "DATABASE_TIMEOUT"
	ErrDatabaseConstraint  ErrorCode = "DATABASE_CONSTRAINT_VIOLATION"
	ErrDuplicateKey        ErrorCode = "DUPLICATE_KEY"
	ErrForeignKeyViolation ErrorCode = "FOREIGN_KEY_VIOLATION"

	// Transaction errors
	ErrTransactionFailed   ErrorCode = "TRANSACTION_FAILED"
	ErrTransactionRollback ErrorCode = "TRANSACTION_ROLLBACK"
	ErrTransactionDeadlock ErrorCode = "TRANSACTION_DEADLOCK"

	// Condition errors
	ErrInvalidCondition ErrorCode = "INVALID_CONDITION"
	ErrFieldNotFound    ErrorCode = "FIELD_NOT_FOUND"
	ErrInvalidOperator  ErrorCode = "INVALID_OPERATOR"

	// Pagination errors
	ErrInvalidPagination ErrorCode = "INVALID_PAGINATION"
	ErrPageOutOfRange    ErrorCode = "PAGE_OUT_OF_RANGE"

	// Permission errors
	ErrAccessDenied      ErrorCode = "ACCESS_DENIED"
	ErrInsufficientPerms ErrorCode = "INSUFFICIENT_PERMISSIONS"

	// Soft delete errors
	ErrAlreadyDeleted       ErrorCode = "ALREADY_DELETED"
	ErrNotDeleted           ErrorCode = "NOT_DELETED"
	ErrSoftDeleteNotEnabled ErrorCode = "SOFT_DELETE_NOT_ENABLED"

	// Generic errors
	ErrInternal        ErrorCode = "INTERNAL_ERROR"
	ErrInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrOperationFailed ErrorCode = "OPERATION_FAILED"
)

// StoreError represents a structured error from the store package
type StoreError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Operation  string                 `json:"operation"`
	EntityID   string                 `json:"entity_id,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	Field      string                 `json:"field,omitempty"`
	Cause      error                  `json:"-"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Error implements the error interface
func (e *StoreError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *StoreError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error
func (e *StoreError) Is(target error) bool {
	if t, ok := target.(*StoreError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Cause, target)
}

// WithMetadata adds metadata to the error
func (e *StoreError) WithMetadata(key string, value interface{}) *StoreError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithField sets the field name for the error
func (e *StoreError) WithField(field string) *StoreError {
	e.Field = field
	return e
}

// WithEntityID sets the entity ID for the error
func (e *StoreError) WithEntityID(id string) *StoreError {
	e.EntityID = id
	return e
}

// WithEntityType sets the entity type for the error
func (e *StoreError) WithEntityType(entityType string) *StoreError {
	e.EntityType = entityType
	return e
}

// NewStoreError creates a new store error
func NewStoreError(code ErrorCode, message, operation string) *StoreError {
	return &StoreError{
		Code:      code,
		Message:   message,
		Operation: operation,
		Timestamp: time.Now(),
	}
}

// NewStoreErrorWithCause creates a new store error with an underlying cause
func NewStoreErrorWithCause(code ErrorCode, message, operation string, cause error) *StoreError {
	return &StoreError{
		Code:      code,
		Message:   message,
		Operation: operation,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// ==========================================
// ERROR CLASSIFICATION FUNCTIONS
// ==========================================

// ClassifyGormError converts GORM errors to store errors
func ClassifyGormError(err error, operation string) *StoreError {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return NewStoreErrorWithCause(ErrEntityNotFound, "Entity not found", operation, err)
	case errors.Is(err, gorm.ErrInvalidTransaction):
		return NewStoreErrorWithCause(ErrTransactionFailed, "Invalid transaction", operation, err)
	case errors.Is(err, gorm.ErrNotImplemented):
		return NewStoreErrorWithCause(ErrOperationFailed, "Operation not implemented", operation, err)
	case errors.Is(err, gorm.ErrMissingWhereClause):
		return NewStoreErrorWithCause(ErrInvalidCondition, "Missing WHERE clause", operation, err)
	case errors.Is(err, gorm.ErrUnsupportedRelation):
		return NewStoreErrorWithCause(ErrOperationFailed, "Unsupported relation", operation, err)
	case errors.Is(err, gorm.ErrPrimaryKeyRequired):
		return NewStoreErrorWithCause(ErrEntityIDEmpty, "Primary key required", operation, err)
	case errors.Is(err, gorm.ErrModelValueRequired):
		return NewStoreErrorWithCause(ErrEntityNil, "Model value required", operation, err)
	case errors.Is(err, gorm.ErrInvalidData):
		return NewStoreErrorWithCause(ErrEntityValidation, "Invalid data", operation, err)
	default:
		// Check for database-specific errors by string matching
		errStr := err.Error()
		switch {
		case containsAny(errStr, []string{"duplicate key", "unique constraint", "UNIQUE constraint"}):
			return NewStoreErrorWithCause(ErrDuplicateKey, "Duplicate key violation", operation, err)
		case containsAny(errStr, []string{"foreign key", "FOREIGN KEY constraint"}):
			return NewStoreErrorWithCause(ErrForeignKeyViolation, "Foreign key constraint violation", operation, err)
		case containsAny(errStr, []string{"connection refused", "connection failed", "no connection"}):
			return NewStoreErrorWithCause(ErrDatabaseConnection, "Database connection failed", operation, err)
		case containsAny(errStr, []string{"timeout", "deadline exceeded"}):
			return NewStoreErrorWithCause(ErrDatabaseTimeout, "Database operation timeout", operation, err)
		case containsAny(errStr, []string{"deadlock", "lock timeout"}):
			return NewStoreErrorWithCause(ErrTransactionDeadlock, "Transaction deadlock detected", operation, err)
		default:
			return NewStoreErrorWithCause(ErrInternal, "Internal database error", operation, err)
		}
	}
}

// ==========================================
// ERROR CHECKING FUNCTIONS
// ==========================================

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code == ErrEntityNotFound
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsDuplicateKeyError checks if the error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code == ErrDuplicateKey
	}
	return false
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code == ErrEntityValidation || storeErr.Code == ErrInvalidCondition
	}
	return false
}

// IsConnectionError checks if the error is a connection error
func IsConnectionError(err error) bool {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code == ErrDatabaseConnection || storeErr.Code == ErrDatabaseTimeout
	}
	return false
}

// IsTransactionError checks if the error is a transaction error
func IsTransactionError(err error) bool {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code == ErrTransactionFailed ||
			storeErr.Code == ErrTransactionRollback ||
			storeErr.Code == ErrTransactionDeadlock
	}
	return false
}

// ==========================================
// HELPER FUNCTIONS
// ==========================================

// containsAny checks if the string contains any of the substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Code
	}
	return ErrInternal
}

// GetErrorMessage extracts a user-friendly message from an error
func GetErrorMessage(err error) string {
	var storeErr *StoreError
	if errors.As(err, &storeErr) {
		return storeErr.Message
	}
	return "An internal error occurred"
}

// ==========================================
// PREDEFINED ERRORS
// ==========================================

// Predefined error instances for common scenarios
var (
	ErrSoftDeleteNotEnabledError = NewStoreError(ErrSoftDeleteNotEnabled, "Soft delete is not enabled for this repository", "soft_delete")
)

// ==========================================
// LEGACY ERROR FUNCTIONS
// ==========================================

// IsRecordNotFound checks if the error is a record not found error
// This function maintains backward compatibility
func IsRecordNotFound(err error) bool {
	return IsNotFoundError(err)
}
