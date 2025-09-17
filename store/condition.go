package store

import (
	"fmt"
	"reflect"
	"strings"
)

// Condition represents a type-safe query condition
type Condition[T Entity] interface {
	ToSQL() (string, []any)
	Validate() error
}

// FieldCondition represents a condition on a specific field
type FieldCondition[T Entity, V any] struct {
	field    string
	operator string
	value    V
}

// ComparisonOperator represents SQL comparison operators
type ComparisonOperator string

const (
	OpEqual              ComparisonOperator = "="
	OpNotEqual           ComparisonOperator = "!="
	OpGreaterThan        ComparisonOperator = ">"
	OpGreaterThanOrEqual ComparisonOperator = ">="
	OpLessThan           ComparisonOperator = "<"
	OpLessThanOrEqual    ComparisonOperator = "<="
	OpLike               ComparisonOperator = "LIKE"
	OpILike              ComparisonOperator = "ILIKE"
	OpIn                 ComparisonOperator = "IN"
	OpNotIn              ComparisonOperator = "NOT IN"
	OpIsNull             ComparisonOperator = "IS NULL"
	OpIsNotNull          ComparisonOperator = "IS NOT NULL"
	OpBetween            ComparisonOperator = "BETWEEN"
)

// LogicalOperator represents SQL logical operators
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
	LogicalNot LogicalOperator = "NOT"
)

// ==========================================
// CONDITION BUILDERS
// ==========================================

// Field creates a field selector for building conditions
func Field[T Entity](fieldName string) *FieldSelector[T] {
	return &FieldSelector[T]{
		fieldName: fieldName,
	}
}

// FieldSelector provides fluent API for building field conditions
type FieldSelector[T Entity] struct {
	fieldName string
}

// Equal creates an equality condition
func (f *FieldSelector[T]) Equal(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpEqual),
		value:    value,
	}
}

// NotEqual creates a not-equal condition
func (f *FieldSelector[T]) NotEqual(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpNotEqual),
		value:    value,
	}
}

// GreaterThan creates a greater-than condition
func (f *FieldSelector[T]) GreaterThan(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpGreaterThan),
		value:    value,
	}
}

// GreaterThanOrEqual creates a greater-than-or-equal condition
func (f *FieldSelector[T]) GreaterThanOrEqual(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpGreaterThanOrEqual),
		value:    value,
	}
}

// LessThan creates a less-than condition
func (f *FieldSelector[T]) LessThan(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpLessThan),
		value:    value,
	}
}

// LessThanOrEqual creates a less-than-or-equal condition
func (f *FieldSelector[T]) LessThanOrEqual(value any) *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpLessThanOrEqual),
		value:    value,
	}
}

// Like creates a LIKE condition
func (f *FieldSelector[T]) Like(pattern string) *FieldCondition[T, string] {
	return &FieldCondition[T, string]{
		field:    f.fieldName,
		operator: string(OpLike),
		value:    pattern,
	}
}

// ILike creates an ILIKE condition (case-insensitive)
func (f *FieldSelector[T]) ILike(pattern string) *FieldCondition[T, string] {
	return &FieldCondition[T, string]{
		field:    f.fieldName,
		operator: string(OpILike),
		value:    pattern,
	}
}

// In creates an IN condition
func (f *FieldSelector[T]) In(values ...any) *FieldCondition[T, []any] {
	return &FieldCondition[T, []any]{
		field:    f.fieldName,
		operator: string(OpIn),
		value:    values,
	}
}

// NotIn creates a NOT IN condition
func (f *FieldSelector[T]) NotIn(values ...any) *FieldCondition[T, []any] {
	return &FieldCondition[T, []any]{
		field:    f.fieldName,
		operator: string(OpNotIn),
		value:    values,
	}
}

// IsNull creates an IS NULL condition
func (f *FieldSelector[T]) IsNull() *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpIsNull),
		value:    nil,
	}
}

// IsNotNull creates an IS NOT NULL condition
func (f *FieldSelector[T]) IsNotNull() *FieldCondition[T, any] {
	return &FieldCondition[T, any]{
		field:    f.fieldName,
		operator: string(OpIsNotNull),
		value:    nil,
	}
}

// Between creates a BETWEEN condition
func (f *FieldSelector[T]) Between(start, end any) *BetweenCondition[T] {
	return &BetweenCondition[T]{
		field: f.fieldName,
		start: start,
		end:   end,
	}
}

// ==========================================
// SPECIFIC CONDITION TYPES
// ==========================================

// BetweenCondition represents a BETWEEN condition
type BetweenCondition[T Entity] struct {
	field string
	start any
	end   any
}

func (c *BetweenCondition[T]) ToSQL() (string, []any) {
	return fmt.Sprintf("%s BETWEEN ? AND ?", c.field), []any{c.start, c.end}
}

func (c *BetweenCondition[T]) Validate() error {
	if c.field == "" {
		return fmt.Errorf("field name cannot be empty")
	}
	if c.start == nil || c.end == nil {
		return fmt.Errorf("start and end values cannot be nil")
	}
	return nil
}

// ToSQL converts the condition to SQL
func (c *FieldCondition[T, V]) ToSQL() (string, []any) {
	switch c.operator {
	case string(OpIsNull), string(OpIsNotNull):
		return fmt.Sprintf("%s %s", c.field, c.operator), []any{}
	case string(OpIn), string(OpNotIn):
		if values, ok := any(c.value).([]any); ok {
			placeholders := strings.Repeat("?,", len(values))
			placeholders = strings.TrimSuffix(placeholders, ",")
			return fmt.Sprintf("%s %s (%s)", c.field, c.operator, placeholders), values
		}
		return fmt.Sprintf("%s %s (?)", c.field, c.operator), []any{c.value}
	default:
		return fmt.Sprintf("%s %s ?", c.field, c.operator), []any{c.value}
	}
}

// Validate validates the condition
func (c *FieldCondition[T, V]) Validate() error {
	if c.field == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// Validate field exists in entity type
	var zero T
	entityType := reflect.TypeOf(zero)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	if !hasField(entityType, c.field) {
		return fmt.Errorf("field '%s' does not exist in entity type %s", c.field, entityType.Name())
	}

	switch c.operator {
	case string(OpIn), string(OpNotIn):
		if values, ok := any(c.value).([]any); ok {
			if len(values) == 0 {
				return fmt.Errorf("IN/NOT IN requires at least one value")
			}
		}
	case string(OpIsNull), string(OpIsNotNull):
		// No value validation needed
	default:
		// Check if value is nil using interface conversion
		if c.operator != string(OpIsNull) && c.operator != string(OpIsNotNull) {
			// Convert to interface{} to check for nil
			iface := any(c.value)
			if iface == nil {
				return fmt.Errorf("value cannot be nil for operator %s", c.operator)
			}
			// Check if it's a nil pointer, slice, map, channel, function, or interface
			value := reflect.ValueOf(iface)
			if (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface || value.Kind() == reflect.Map || value.Kind() == reflect.Slice || value.Kind() == reflect.Chan || value.Kind() == reflect.Func) && value.IsNil() {
				return fmt.Errorf("value cannot be nil for operator %s", c.operator)
			}
		}
	}

	return nil
}

// ==========================================
// COMPOSITE CONDITIONS
// ==========================================

// CompositeCondition represents a combination of conditions
type CompositeCondition[T Entity] struct {
	conditions []Condition[T]
	operator   LogicalOperator
}

// And combines conditions with AND operator
func And[T Entity](conditions ...Condition[T]) *CompositeCondition[T] {
	return &CompositeCondition[T]{
		conditions: conditions,
		operator:   LogicalAnd,
	}
}

// Or combines conditions with OR operator
func Or[T Entity](conditions ...Condition[T]) *CompositeCondition[T] {
	return &CompositeCondition[T]{
		conditions: conditions,
		operator:   LogicalOr,
	}
}

// Not negates a condition
func Not[T Entity](condition Condition[T]) *CompositeCondition[T] {
	return &CompositeCondition[T]{
		conditions: []Condition[T]{condition},
		operator:   LogicalNot,
	}
}

// ToSQL converts the composite condition to SQL
func (c *CompositeCondition[T]) ToSQL() (string, []any) {
	if len(c.conditions) == 0 {
		return "", []any{}
	}

	var sqlParts []string
	var args []any

	for _, condition := range c.conditions {
		sql, condArgs := condition.ToSQL()
		if sql != "" {
			if c.operator == LogicalNot {
				sqlParts = append(sqlParts, fmt.Sprintf("NOT (%s)", sql))
			} else {
				sqlParts = append(sqlParts, fmt.Sprintf("(%s)", sql))
			}
			args = append(args, condArgs...)
		}
	}

	if len(sqlParts) == 0 {
		return "", []any{}
	}

	if c.operator == LogicalNot {
		return sqlParts[0], args
	}

	sql := strings.Join(sqlParts, fmt.Sprintf(" %s ", c.operator))
	return sql, args
}

// Validate validates all conditions in the composite
func (c *CompositeCondition[T]) Validate() error {
	if len(c.conditions) == 0 {
		return fmt.Errorf("composite condition must have at least one condition")
	}

	if c.operator == LogicalNot && len(c.conditions) > 1 {
		return fmt.Errorf("NOT operator can only be applied to a single condition")
	}

	for i, condition := range c.conditions {
		if err := condition.Validate(); err != nil {
			return fmt.Errorf("condition %d validation failed: %w", i, err)
		}
	}

	return nil
}
