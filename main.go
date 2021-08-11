package main

import (
	"daodao-go/modules/command"
	"flag"
	"github.com/xxiiaass/iutils"
	"io/ioutil"
	"path"
	"strings"
)

const dalDir = "/../repository"


var CurPath = ""

func Parse() {
	flag.StringVar(&CurPath, "app-path", "", "app-path")
}

// 获取可执行文件的所在路径, 如果外部指定了路径，则使用指定路径
func GetCurPath() string {
	if CurPath != "" {
		return CurPath
	}
	return iutils.GetCurPath()
}


func main() {
	Parse()
	CurPath := GetCurPath()

	dalFiles, _ := ioutil.ReadDir(path.Join(CurPath, dalDir))
	for _, item := range dalFiles {
		if item.IsDir() {
			if item.Name() == "cache" || item.Name() == "name" {
				continue
			}
			// 二级文件目录，使用私有方式构建
			fullPath := path.Join(CurPath, dalDir, item.Name())

			CpTemplate("mysql", "mysql_define.go", fullPath, item.Name())
			task := NewMysqlTask(fullPath, item.Name())
			task.IsPrivate = true
			task.ModelFilterFunc = func(name string) bool {
				return strings.Contains(name, "Model.go") || defaultFilter(name)
			}
			task.Run()
		}
	}
}
