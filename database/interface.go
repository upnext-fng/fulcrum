package database

import "gorm.io/gorm"

type DatabaseService interface {
	Connection() *gorm.DB
	HealthCheck() error
	Close() error
}
