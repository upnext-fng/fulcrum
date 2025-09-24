package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type manager struct {
	db     *gorm.DB
	config Config
}

func NewManager(config Config) DatabaseService {
	return &manager{
		config: config,
	}
}

func (m *manager) Connection() *gorm.DB {
	if m.db == nil {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			m.config.Host,
			m.config.Username,
			m.config.Password,
			m.config.Database,
			m.config.Port,
			m.config.SSLMode,
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(fmt.Sprintf("failed to connect database: %v", err))
		}
		m.db = db
	}
	return m.db
}

func (m *manager) HealthCheck() error {
	sqlDB, err := m.Connection().DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (m *manager) Close() error {
	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
