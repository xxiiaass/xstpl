package template

import (
	"daodao-go/infrastructure/data/customType/dbModels"
	"daodao-go/infrastructure/data/customType/mysql"
	"daodao-go/repository/name"
)

type WhereCb = mysql.WhereCb

type BeforeHook = dbModels.BeforeHook

type MustHook = dbModels.MustHook

type Snuid = name.Snuid
