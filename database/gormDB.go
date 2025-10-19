package database

import (
	"context"
	"fmt"
	"os"
	"sports-backend-api/util"

	"gorm.io/gorm"
)

// DB is gorm variable
var DB *gorm.DB

// DBConfig represents db configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	DBName   string
	Password string
}

// BuildDBConfig to set value of DBConfig
func BuildConfig() *DBConfig {
	util.LoadEnv()
	dbConfig := DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	return &dbConfig
}

// DbURL to generate connection string
func DBUrl(dbConfig *DBConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)
}

func GetDB(ctx context.Context) *gorm.DB {
	DB = DB.WithContext(ctx)
	return DB
}
