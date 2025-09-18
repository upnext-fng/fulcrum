package db

import (
	"errors"
	"strings"
	"time"

	"github.com/upnext-fng/fulcrum/infra/db/dialects"
	"github.com/upnext-fng/fulcrum/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type DBType int

type GetDBConn func(uri string, gormConfig *gorm.Config) (*gorm.DB, error)

const (
	DbTypeMySQL DBType = iota + 1
	DbTypePostgres
	DbTypeNotSupported
)

func getDBType(dbType string) DBType {
	switch strings.ToLower(dbType) {
	case "mysql":
		return DbTypeMySQL
	case "postgres":
		return DbTypePostgres
	}

	return DbTypeNotSupported
}

var dbTypeMap = map[DBType]GetDBConn{
	DbTypeMySQL: func(dns string, gormConfig *gorm.Config) (*gorm.DB, error) {
		return dialects.MySqlDB(dns, gormConfig)
	},
	DbTypePostgres: func(dns string, gormConfig *gorm.Config) (*gorm.DB, error) {
		return dialects.PostgresDB(dns, gormConfig)
	},
	DbTypeNotSupported: func(uri string, gormConfig *gorm.Config) (*gorm.DB, error) {
		return nil, nil
	},
}

func configureGormLogger(isDevelopment bool) gormLogger.Interface {
	if isDevelopment {
		dbLogger := logger.NewLog("database", logger.WithDevelopment(isDevelopment))
		return NewGormLogger(dbLogger, gormLogger.Config{
			SlowThreshold:             time.Second,     // Slow SQL threshold
			LogLevel:                  gormLogger.Info, // Log level
			IgnoreRecordNotFoundError: false,           // Not ignore not found error
			ParameterizedQueries:      false,           // Include params in the SQL log
			Colorful:                  true,            // Disable color
		})
	}

	// Production configuration - minimal logging
	dbLogger := logger.NewLog("database")
	return NewGormLogger(dbLogger, gormLogger.Config{
		SlowThreshold:             time.Duration(200) * time.Millisecond, // 200ms threshold
		LogLevel:                  gormLogger.Warn,                       // Only warnings and errors
		IgnoreRecordNotFoundError: true,                                  // Ignore record not found errors
		ParameterizedQueries:      true,                                  // Show parameterized queries
		Colorful:                  false,                                 // No color in production
	})
}

func NewGormDB(dbType string, dsn string, config *Config) (db *gorm.DB, err error) {
	gormDBType := getDBType(dbType)
	if gormDBType == DbTypeNotSupported {
		return nil, errors.New("database type is not supported")
	}
	return dbTypeMap[gormDBType](dsn, newGormConfig())
}
