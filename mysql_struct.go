package main

import (
	"bytes"
	"fmt"
	"github.com/xxiiaass/iutils"
	"github.com/xxiiaass/xsorm"
	"html/template"
	"regexp"
	"strconv"
)

// 用于生成结构体定义代码


type Field struct {
	Field   string `gorm:"column:Field"`
	Type    string `gorm:"column:Type"`
	Null    string `gorm:"column:Null"`
	Key     string `gorm:"column:Key"`
	Default string `gorm:"column:Default"`
	Extra   string `gorm:"column:Extra"`
	Name    string `gorm:"-"`
	Tagstr  string `gorm:"-"`
}

func toStruct(name string) {
	fields := make([]Field, 0)
	xsorm.NewBuild(&fields).TableName("disable").Raw("DESC " + name)
	maxKeyLength := 0
	for i, field := range fields {
		tp := xsorm.FieldTypeToGolangType(field.Type)
		if tp == "" {
			panic("无法识别类型, key:" + field.Field)
		}
		fields[i].Type = tp
		fields[i].Name = iutils.UnderLineToCamel(field.Field)
		if len(field.Field) > maxKeyLength {
			maxKeyLength = len(field.Field)
		}
	}
	for i, field := range fields {
		fName := field.Field
		if field.Key == "PRI" {
			fName = fName + ";PRIMARY_KEY"
		}
		fields[i].Tagstr = "`gorm:\"column:" + fmt.Sprintf("%s\"%"+strconv.Itoa(maxKeyLength+1-len(fName))+"s", fName, "") +
			"json:\"" + field.Field + "\"`"
	}
	type T struct {
		IsSplit   bool
		TableName string
		Name      string
		Fields    []Field
	}
	t := new(T)
	t.Fields = fields
	regs := regexp.MustCompile("^(.+)_[0-9]+$").FindStringSubmatch(name)
	if len(regs) == 0 {
		t.TableName = iutils.UnderLineToCamel(name)
		t.Name = name
		t.IsSplit = false
	} else {
		t.TableName = iutils.UnderLineToCamel(regs[1])
		t.Name = regs[1]
		t.IsSplit = true
	}

	temp := `

import "github.com/xxiiaass/xsorm"

type {{.TableName}} struct {
{{range .Fields}}
	{{.Name}}  {{.Type}} {{.Tagstr}}{{end}}
}
{{if .IsSplit}}
func (t * {{.TableName}}) BaseName() {
	t.BaseTableName = "{{.Name}}"
}
{{else}}
func ({{.TableName}}) TableName() string {
	return "{{.Name}}"
}
{{end}}

type {{.TableName}}Query struct {
	xsorm.Query
}

`
	var doc bytes.Buffer
	tm, err := template.New("create_model").Parse(temp)
	if err != nil {
		panic(err)
	}
	tm.Execute(&doc, t)
	html := doc.String()
	fmt.Println(regexp.MustCompile("&#34;").ReplaceAllString(html, "\""))
}
