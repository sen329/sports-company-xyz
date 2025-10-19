package database

import (
	"gorm.io/driver/mysql"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func Database() error {
	var err error
	DB, err = gorm.Open(mysql.Open(DBUrl(BuildConfig())), &gorm.Config{
		SkipDefaultTransaction: true, // Improves performance by avoiding auto-transactions.
		PrepareStmt:            true, // Caches compiled statements for performance and helps prevent SQL injection.
	})
	if err != nil {
		return err
	}

	return nil
}
