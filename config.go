package main

import (
	"github.com/go-ini/ini"
	"log"
	"path"
)

var Cfg *ini.File

func getIni(ConfigFile string) *ini.File {
	if Cfg == nil {
		var err error
		if ConfigFile == "" {
			ConfigFile = "app.ini"
		}
		Cfg, err = ini.Load(ConfigFile)
		if err != nil {
			log.Fatalf("Fail to parse 'app.ini': %v", err)
		}
	}
	return Cfg
}

var DalDir = ""                                // 模型定义文件夹
var DefineTemplateFile = "xsbuild/template.go" // 类型定义文件

func initCfg(ConfigFile string) {
	Cfg := getIni(ConfigFile)
	DalDir = Cfg.Section("").Key("DalDir").MustString("./models")
	DefineTemplateFile = Cfg.Section("").Key("DefineTemplateFile").MustString(path.Join(XstplBuildDir, "template.go"))
}
