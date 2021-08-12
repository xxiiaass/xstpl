package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

// todo
/**
template 路径, 在应用目录建立一个文件夹
**/

func Write(text, path string) {
	ioutil.WriteFile(path, []byte(text), 0644)
}

// 替换包名后，将模板文件写入目标文件夹中
func CpTemplate(templatePath, templateName, targetPath, packageName string) {
	// str, _ := ioutil.ReadFile(path.Join(CurPath, "/../cli/private_model/template/", templateDir, templateName))
	str, _ := ioutil.ReadFile(templatePath)
	afterStr := strings.Replace(string(str), "package template", "package "+packageName, 1)
	Write(afterStr, path.Join(targetPath, "lib_auto_generate_"+templateName))
}

// 将文件中声明的方法，全部变为私有
func ToPrivate(str string) string {
	matchStr := str

	white := []string{"Get", "First", "GetOne", "Count", "DoneOperate", "Delete"}

	reg := regexp.MustCompile(`\) [A-Z][a-zA-Z]+\(`)
	arr := reg.FindAllStringSubmatch(matchStr, -1)
	for _, s := range arr {
		beforeMethod := s[0]
		ignore := false
		for _, w := range white {
			if strings.Contains(beforeMethod, w) {
				ignore = true
			}
		}
		if ignore {
			continue
		}

		// 第三个字符为方法的首字符
		first := beforeMethod[2:3]
		btStr := []byte(beforeMethod)
		low := strings.ToLower(first)
		btStr[2] = []byte(low)[0]

		matchStr = strings.Replace(matchStr, beforeMethod, string(btStr), 1)
	}
	return matchStr
}

type DefineField struct {
	StructKey string
	Key       string
	Type      string
	Number    bool
}

type Fields struct {
	All      []DefineField
	Number   []DefineField
	Pluck    []DefineField
	PluckUni []DefineField
	Map      []DefineField
}

type Func struct {
	Name    string // 自己的名字
	Proxy   string // 执行的build方法名
	Argus   string
	ToBuild string
}

type Render struct {
	Funcs     []Func
	TypeName  string
	QueryName string
	Driver    string
	Fields    Fields
}

// 渲染
func (t Render) Render(tmp string) string {
	var doc bytes.Buffer
	tm, err := template.New("code").Parse(tmp)
	if err != nil {
		panic(err)
	}
	tm.Execute(&doc, t)
	html := doc.String()
	return html
}

func MethodNameToPrivate(str string) string {
	btStr := []byte(str)
	btStr[0] = []byte(strings.ToLower(str[0:1]))[0]
	return string(btStr)
}

func collectTemplate() string {
	return `
{{ $name := .TypeName }}
type {{$name}}Collect []{{$name}}
{{range .Fields.Pluck}}
func(s  {{$name}}Collect) Pluck{{.StructKey}}() []{{.Type}}{
	list := make([]{{.Type}}, len(s))
	for i, v := range s{
		list[i] = v.{{.StructKey}}
	}
	return list
}
{{end}}

{{range .Fields.PluckUni}}
func(s  {{$name}}Collect) PluckUni{{.StructKey}}() []{{.Type}}{
	uniMap := make(map[{{.Type}}]bool)
	list := make([]{{.Type}} ,0)
	for _, v := range s{
		_, ok := uniMap[v.{{.StructKey}}]
		if !ok {
			uniMap[v.{{.StructKey}}] = true
		    list = append(list, v.{{.StructKey}})
		}
	}
	return list
}
{{end}}

{{range .Fields.Map}}
func(s  {{$name}}Collect) GroupBy{{.StructKey}}() map[{{.Type}}][]{{$name}}{
	m := make(map[{{.Type}}][]{{$name}})
	for _, v := range s{
		if _, ok := m[v.{{.StructKey}}]; !ok{
			m[v.{{.StructKey}}] = make([]{{$name}}, 0)
		}
		m[v.{{.StructKey}}] = append(m[v.{{.StructKey}}], v)
	}
	return m
}
{{end}}

func(s  {{$name}}Collect) Filter( f func(item {{$name}}) bool) {{$name}}Collect{
	m := make({{$name}}Collect, 0)
	for _, v := range s{
		if f(v){
			m = append(m, v)
		}
	}
	return m
}

{{range .Fields.Map}}
func(s  {{$name}}Collect) KeyBy{{.StructKey}}() map[{{.Type}}]{{$name}}{
	m := make(map[{{.Type}}]{{$name}})
	for _, v := range s{
		m[v.{{.StructKey}}] = v
	}
	return m
}
{{end}}
`
}

type Filter func(name string) bool

func defaultFilter(name string) bool {
	return !strings.Contains(name, "_") && name[0] <= 'Z' && name[0] >= 'A'
}

type Task struct {
	FromDirPath     string // 解析的结构体，所在的文件夹
	BuildFileStr    string
	ignoreMethod    []string // 自动解析出来的方法，需要跳过的内容
	PackageName     string   // 包名
	WriteDirPath    string   // 生成的代码，写入的路劲
	DriverName      string   // 驱动名(底层的包名)
	IsPrivate       bool     // 生产的方法，是否是私有
	ModelFilterFunc Filter
}

func (mt Task) topLine() string {
	return "package " + mt.PackageName + "\n"
}

// 代理了build的方法集
func (mt Task) ProxyTemplate() string {
	return `
{{ $name := .TypeName }}
{{ $queryName := .QueryName }}
{{range .Funcs}}
func (m *{{$queryName}}) {{.Name}}({{.Argus}}) *{{$queryName}}{
	m.GetBuild().{{.Proxy}}({{.ToBuild}})
	return m
}
{{end}}

`
}

// 用于构建查询语句的模板, 公有的
func (mt Task) PublicBuildQueryTemplate() string {
	return `


{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}(args ...interface{}) *{{$queryName}}{
	return m.Where("{{.Key}}", args...)
}
{{end}}


{{range .Fields.All}}
func (m *{{$queryName}}) KSet{{.StructKey}}(val interface{}) *{{$queryName}}{
	return m.Set("{{.Key}}", val)
}
{{end}}

{{range .Fields.Number}}
func (m *{{$queryName}}) KInc{{.StructKey}}(num int) *{{$queryName}}{
	return m.Inc("{{.Key}}", num)
}
{{end}}


{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}In(values interface{}) *{{$queryName}}{
	return m.WhereIn("{{.Key}}", values)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) KWhe{{.StructKey}}NotIn(values interface{}) *{{$queryName}}{
	return m.WhereNotIn("{{.Key}}", values)
}
{{end}}
`
}

// 用于构建查询语句的模板, 私有的
func (mt Task) PrivateBuildQueryTemplate() string {
	return `


{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}(args ...interface{}) *{{$queryName}}{
	return m.where("{{.Key}}", args...)
}
{{end}}


{{range .Fields.All}}
func (m *{{$queryName}}) kSet{{.StructKey}}(val interface{}) *{{$queryName}}{
	return m.Set("{{.Key}}", val)
}
{{end}}

{{range .Fields.Number}}
func (m *{{$queryName}}) kInc{{.StructKey}}(num int) *{{$queryName}}{
	return m.Inc("{{.Key}}", num)
}
{{end}}



{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}In(values interface{}) *{{$queryName}}{
	return m.whereIn("{{.Key}}", values)
}
{{end}}

{{range .Fields.All}}
func (m *{{$queryName}}) kWhe{{.StructKey}}NotIn(values interface{}) *{{$queryName}}{
	return m.whereNotIn("{{.Key}}", values)
}
{{end}}
`
}
