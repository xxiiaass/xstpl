package models

import (
	"github.com/xxiiaass/xsorm"
)

type Account struct {
	Autokid     int64   `gorm:"column:autokid"    json:"autokid"`
	Uuid        string  `gorm:"column:uuid"    json:"uuid"`
	AccountType int64   `gorm:"column:account_type"    json:"account_type"`
	UserId      int64   `gorm:"column:user_id"    json:"user_id"`
	Content     string  `gorm:"column:content"    json:"content"`
	Name        string  `gorm:"column:name"    json:"name"`
	Money       float64 `gorm:"column:money"    json:"money"`
}

func (t *Account) TableName() string {
	return "account"
}

type AccountQuery struct {
	xsorm.Query
}
