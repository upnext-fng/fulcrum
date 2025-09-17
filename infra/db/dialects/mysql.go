package dialects

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySqlDB Ex: user:password@/db_name?charset=utf8&parseTime=True&loc=Local
func MySqlDB(uri string) (db *gorm.DB, err error) {
	return gorm.Open(mysql.Open(uri), &gorm.Config{})
}
