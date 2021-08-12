package main

import (
	"fmt"
	"github.com/xxiiaass/iutils"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

// todo
/**
build.go 路径
**/

func (mt *MysqlTask) parseFunction() []Func {
	str := mt.BuildFileStr
	reg := regexp.MustCompile("func \\(build \\*Build\\)\\s+([A-Z]\\S+)\\((.+)?\\) \\*Build \\{")
	x := reg.FindAllStringSubmatch(string(str), -1)
	funcs := make([]Func, 0)
	for _, s := range x {
		if iutils.IsExists(s[1], mt.ignoreMethod) {
			// 手动忽略的方法
			continue
		}
		if s[2] == "" {
			funcs = append(funcs, Func{Name: s[1], Proxy: s[1], Argus: "", ToBuild: ""})
		} else {
			argus := strings.Split(s[2], ",")
			toBuilds := make([]string, len(argus))
			for j, a := range argus {
				arrs := strings.Split(strings.Trim(a, " "), " ")
				if arrs[1][0:3] == "..." {
					toBuilds[j] = arrs[0] + "..."
				} else {
					toBuilds[j] = arrs[0]
				}
			}
			funcs = append(funcs, Func{Name: s[1], Proxy: s[1], Argus: s[2], ToBuild: strings.Join(toBuilds, ",")})
		}
	}

	return funcs
}

func (mt *MysqlTask) parseField(fileTxt string) []DefineField {
	// \s+ 用于去除任何空白字符，其后的 \w+? 用于获取字段名，如果存在使用 // 单行注释字段该字段将会被略过
	reg := regexp.MustCompile("@\\s+(\\w+?)\\s+(\\S+?)\\s+`gorm:\"(\\S+?)\"[^`]+?json:\"(\\S+?)\"`.*?@")
	x := reg.FindAllStringSubmatch(fileTxt, -1)
	names := make([]DefineField, 0)
	for _, field := range x {
		// field[0] Raw Str
		// field[1] Struct Field Name
		// field[2] Struct Field Type
		// field[3] gorm tag
		// field[4] json tag

		// 获取 column name
		if len(field) > 3 {
			// 存在 column
			if pos := strings.Index(field[3], "column:"); pos >= 0 {
				begin := pos + len("column:")
				// 区分 gorm:"column:theater_id;PRIMARY_KEY"
				if sp := strings.Index(field[3], ";"); sp >= 0 && sp > begin {
					field[3] = field[3][begin:sp]
				} else {
					field[3] = field[3][begin:]
				}
			} else {
				continue
			}
		} else {
			continue
		}
		isNum := iutils.IsExists(field[2], []string{"int64", "int", "float64", "float32"})
		names = append(names, DefineField{StructKey: field[1], Key: field[3], Type: field[2], Number: isNum})
	}

	return names
}

func NewMysqlTask(driverName, targetPath, PackageName, buildfilestr string) *MysqlTask {
	return &MysqlTask{
		Task: Task{
			FromDirPath:     targetPath,
			BuildFileStr:    buildfilestr,
			PackageName:     PackageName,
			WriteDirPath:    targetPath,
			DriverName:      driverName,
			ignoreMethod:    []string{"clone", "Clone", "TableName"},
			ModelFilterFunc: defaultFilter,
		},
	}
}

type MysqlTask struct {
	Task
}

func (mt *MysqlTask) Run() {
	funcs := mt.parseFunction()
	debug(fmt.Sprintf("解析驱动文件 parse funcs : %d", len(funcs)))
	if mt.IsPrivate {
		// 如果是渲染私有的方式，替换方法名
		for i, item := range funcs {
			if strings.Contains(strings.ToLower(item.Name), "where") {
				funcs[i].Name = MethodNameToPrivate(item.Name)
				continue
			}
		}
	}
	allTypes, fileds := mt.getAllTypeName()
	debug(fmt.Sprintf("解析模型文件 getAllTypeName， allTypes:%d, fileds:%d", len(allTypes), len(fileds)))

	workTypes := make([]string, 0)
	workFileds := make([][]DefineField, 0)
	fileNums := 0
	for i, _ := range allTypes {
		workTypes = append(workTypes, allTypes[i])
		workFileds = append(workFileds, fileds[i])

		if (i > 0 && i%5 == 0) || i == len(allTypes)-1 {
			fileNums++
			query := mt.topLine() +
				"import \"reflect\"\n" +
				"import \"github.com/xxiiaass/iutils\"\n" +
				"import \"github.com/xxiiaass/" + mt.DriverName + "\"\n"

			for i, tname := range workTypes {
				query += mt.renderQuery(funcs, tname, workFileds[i])
			}

			collect := mt.topLine() +
				`
/*
此文件为自动生成，所有修改都不会生效
*/
`
			for i, tname := range workTypes {
				collect += mt.renderCollect(funcs, tname, workFileds[i])
			}
			Write(query, path.Join(mt.WriteDirPath, fmt.Sprintf("lib_auto_generate_query_%d.go", fileNums)))
			Write(collect, path.Join(mt.WriteDirPath, fmt.Sprintf("lib_auto_generate_collect_%d.go", fileNums)))

			workTypes = make([]string, 0)
			workFileds = make([][]DefineField, 0)
		}
	}

	// 最后写入一个空文件，解决 n个分卷, 切换到 n-1个分卷时，第n分卷的内容重复的问题
	ioutil.WriteFile(path.Join(mt.WriteDirPath, fmt.Sprintf("lib_auto_generate_collect_%d.go", fileNums+1)), []byte(mt.topLine()), 0644)
	ioutil.WriteFile(path.Join(mt.WriteDirPath, fmt.Sprintf("lib_auto_generate_query_%d.go", fileNums+1)), []byte(mt.topLine()), 0644)

}

func (mt *MysqlTask) renderQuery(funcs []Func, typeName string, filed []DefineField) string {
	nums := make([]DefineField, 0)
	for _, i := range filed {
		if i.Number {
			nums = append(nums, i)
		}
	}
	t := Render{funcs, typeName, typeName + "Query", mt.DriverName, Fields{
		All:      filed,
		Pluck:    filed,
		PluckUni: filed,
		Map:      filed,
		Number:   nums,
	}}

	code := t.Render(mt.ExecTemplate())
	if mt.IsPrivate {
		code = ToPrivate(code)
	}
	code += t.Render(mt.ProxyTemplate())
	if mt.IsPrivate {
		code += t.Render(mt.PrivateBuildQueryTemplate())
	} else {
		code += t.Render(mt.PublicBuildQueryTemplate())
	}
	return code
}

func (mt *MysqlTask) renderCollect(funcs []Func, typeName string, filed []DefineField) string {
	t := Render{funcs, typeName, typeName + "Query", mt.DriverName, Fields{
		All:      filed,
		Pluck:    filed,
		PluckUni: filed,
		Map:      filed,
	}}
	return t.Render(collectTemplate())
}

func (mt *MysqlTask) getAllTypeName() ([]string, [][]DefineField) {
	files, _ := ioutil.ReadDir(mt.FromDirPath)
	debug(fmt.Sprintf("从%s读取模型, 共%d个", mt.FromDirPath, len(files)))
	tables := make([]string, 0)
	fileds := make([][]DefineField, 0)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !mt.ModelFilterFunc(f.Name()) {
			debug(fmt.Sprintf("%s因名字没命中筛选规则，跳过", f.Name()))
			continue
		}
		str, err := ioutil.ReadFile(path.Join(mt.FromDirPath, f.Name()))
		if err != nil {
			panic(err)
		}
		name := f.Name()[:len(f.Name())-3]
		debug(mt.DriverName)
		if !regexp.MustCompile("\\n\\s+" + mt.DriverName + ".Query\\s*\\n").MatchString(string(str)) {
			// 校验 xsorm.Query
			debug(fmt.Sprintf("%s因没有驱动申明，跳过", f.Name()))
			continue
		}
		if !regexp.MustCompile(".*type " + name + "Query struct").MatchString(string(str)) {
			// 校验  type AccountQuery struct {
			debug(fmt.Sprintf("%s因驱动申明不合规，跳过", f.Name()))
			continue
		}

		reg := regexp.MustCompile("\\n")
		modelStr := reg.ReplaceAllString(string(str), "@@")
		// } 后增加判断换行以防止误判 interface{} 类型为 struct 结尾
		targetStructStr := regexp.MustCompile("type " + name + " struct {.+?}@@").FindStringSubmatch(modelStr)
		if len(targetStructStr) == 0 {
			debug(fmt.Sprintf("%s因结构定义不合规，跳过", f.Name()))
			continue
		}

		filed := mt.parseField(targetStructStr[0])
		fileds = append(fileds, filed)
		tables = append(tables, name)
	}
	return tables, fileds
}

// 执行后获得结果的方法模板
func (mt *MysqlTask) ExecTemplate() string {
	return `
func New{{.QueryName}}() *{{.QueryName}} {
	s := {{.QueryName}}{}
    s.SetBuild({{.Driver}}.NewBuild(&s))
	i, ok := reflect.ValueOf(&s).Interface().(BeforeHook)
	if ok {
		i.Before()
	}
	return &s
}

{{ $name := .TypeName }}
{{ $queryName := .QueryName }}

type _{{$queryName}}ColsStruct struct{
{{range .Fields.All}}{{.StructKey}} string
{{end}}
}
func Get{{$queryName}}Cols() *_{{$queryName}}ColsStruct {
	return &_{{$queryName}}ColsStruct{
{{range .Fields.All}}{{.StructKey}} : "{{.Key}}",
{{end}}
	}
}

func (m *{{.QueryName}}) First() *{{.TypeName}} {
	s := make([]{{.TypeName}}, 0)
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	m.GetBuild().ModelType(&s).Limit(1).Get()
	if len(s) > 0{
		return &s[0]
	}
	return &{{.TypeName}}{}
}


func (m *{{.QueryName}}) GetOne() *{{.TypeName}} {
	s := make([]{{.TypeName}}, 0)
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	m.GetBuild().ModelType(&s).Limit(1).Get()
	if len(s) > 0{
		return &s[0]
	}
	return nil
}

func (m *{{.QueryName}}) Get() {{.TypeName}}Collect {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := make([]{{.TypeName}}, 0)
	m.GetBuild().ModelType(&s).Get()
	return s
}

func (m *{{.QueryName}}) Clone() *{{.QueryName}} {
	nm := New{{.QueryName}}()
	nm.SetBuild(m.GetBuild().Clone())
	return nm
}

func (m *{{.QueryName}}) Count() int64 {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Count()
}

func (m *{{.QueryName}}) Sum(col string) float64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Sum(col)
}

func (m *{{.QueryName}}) Max(col string) float64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Max(col)
}

func (m *{{.QueryName}}) DoneOperate() int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).DoneOperate()
}

func (m *{{.QueryName}}) Update(h iutils.H) int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Update(h)
}

func (m *{{.QueryName}}) Delete() int64 {
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Delete()
}

func (m *{{.QueryName}}) Save(x *{{.TypeName}}) {
    m.GetBuild().ModelType(x).Save()
}

func (m *{{.QueryName}}) Error() error {
    return m.GetBuild().ModelType(m).Error()
}

//支持分表
func (m *{{.QueryName}}) Insert(argu interface{}) {
	s := {{.TypeName}}{}
	m.GetBuild().ModelType(&s).Insert(argu)
}

func (m *{{.QueryName}}) Increment(column string, amount int) int64 {
	i, ok := reflect.ValueOf(m).Interface().(MustHook)
	if ok {
		i.Must()
	}
	s := {{.TypeName}}{}
	return m.GetBuild().ModelType(&s).Increment(column, amount)
}


`
}
