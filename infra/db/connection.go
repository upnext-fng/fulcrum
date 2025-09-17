package db

import (
	"errors"
	"strings"

	"github.com/upnext-fng/fulcrum/infra/db/dialects"
	"gorm.io/gorm"
)

type DBType int

type GetDBConn func(uri string) (*gorm.DB, error)

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
	DbTypeMySQL: func(uri string) (*gorm.DB, error) {
		return dialects.MySqlDB(uri)
	},
	DbTypePostgres: func(uri string) (*gorm.DB, error) {
		return dialects.PostgresDB(uri)
	},
	DbTypeNotSupported: func(uri string) (*gorm.DB, error) {
		return nil, nil
	},
}

func NewGormDB(dbType string, uri string) (db *gorm.DB, err error) {
	gormDBType := getDBType(dbType)
	if gormDBType == DbTypeNotSupported {
		return nil, errors.New("database type is not supported")
	}
	return dbTypeMap[gormDBType](uri)
}
