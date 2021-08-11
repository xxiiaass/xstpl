package main

import (
	"github.com/xxiiaass/iutils"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

var XstplBuildDir = "./xsbuild" // 构建文件夹
var cfgPath = "./xsbuild/config.ini"

var defaultCfg = `DalDir=Models
DefineTemplateFile=xsbuild/template.go
`

// 获取可执行文件的所在路径, 如果外部指定了路径，则使用指定路径
func GetCurPath() string {
	return iutils.GetCurPath()
}

func GetCmd() string {
	if len(os.Args) == 1 {
		return "build"
	}
	return os.Args[1]
}

func main() {

	if GetCmd() == "init" {
		// 初始化目录
		if !exists(XstplBuildDir) {
			mkdir(XstplBuildDir)
		}
		if !exists(cfgPath) {
			gopath := os.Getenv("GOPATH")
			// todo
			defaultCfg += "DriverBuildFile=" + path.Join(gopath)
			Write(defaultCfg, cfgPath)
		}

		if !exists(DefineTemplateFile) {
			defaultTmp := `package template

import (
	"github.com/xxiiaass/xsorm"
)

type WhereCb = xsorm.WhereCb

type BeforeHook interface {
	Before()
}

type MustHook interface {
	Must()
}

`
			Write(defaultTmp, DefineTemplateFile)
		}

		return
	}

	CurPath := "." // GetCurPath()
	initCfg(cfgPath)

	CpTemplate(DefineTemplateFile, "define.go", DalDir, DalDir)
	task := NewMysqlTask("xsorm", path.Join(CurPath, DalDir), DalDir)
	task.Run()

	dalFiles, err := ioutil.ReadDir(path.Join(CurPath, DalDir))
	if err != nil {
		panic(err)
	}
	for _, item := range dalFiles {
		if item.IsDir() {
			if item.Name() == "cache" || item.Name() == "name" {
				continue
			}
			// 二级文件目录，使用私有方式构建
			fullPath := path.Join(CurPath, DalDir, item.Name())

			CpTemplate(DefineTemplateFile, "define.go", fullPath, item.Name())
			task := NewMysqlTask("xsorm", fullPath, item.Name())
			task.IsPrivate = true
			task.ModelFilterFunc = func(name string) bool {
				return strings.Contains(name, "Model.go") || defaultFilter(name)
			}
			task.Run()
		} else {

		}
	}
}

func mkdir(path string) {
	_ = os.Mkdir(path, os.ModeDir)
	err := os.Chmod(path, 0755)
	if err != nil {
		panic(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func debug(str string){
	// fmt.Println(str)
}