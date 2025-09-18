package dialects

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySqlDB Ex: user:password@/db_name?charset=utf8&parseTime=True&loc=Local
func MySqlDB(dns string, gormConfig *gorm.Config) (db *gorm.DB, err error) {
	return gorm.Open(mysql.Open(dns), gormConfig)
}
