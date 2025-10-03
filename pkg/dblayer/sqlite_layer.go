package dblayer

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sqliteLayer struct {
	*gorm.DB
}

func init() {
	regist("sqlite", createSqliteClient)
}

func createSqliteClient(dsn string) (dblayer, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}
	return &sqliteLayer{DB: db}, nil
}
