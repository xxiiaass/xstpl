package main

import (
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/xxiiaass/xsorm"
)

type TableName struct {
	Column1 int `gorm:"column:column_1" json:"column_1"`
}

func (TableName) TableName() string {
	return "table_name"
}

func main() {
	xsorm.AddConnect(xsorm.XConfig{
		Config: mysql.Config{
			User:                 "root",
			Passwd:               "1234567",
			Addr:                 "localhost:3306",
			DBName:               "cdn",
			AllowNativePasswords: true,
			Net:                  "tcp",
		},
		Debug:   true,
	})
	xsorm.Init()

	tb := new(TableName)
	xsorm.NewBuild(tb).Where("column_1", 2).First()

	fmt.Println(tb.Column1)
}
