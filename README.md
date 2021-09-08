# 使用说明

* `go get github.com/xxiiaass/xsorm` 获取依赖的驱动项目
* `xstpl init` 命令，会在当前目录生成一个文件夹xsbuild, 里面是配置文件
* 编辑xsbuild/config.ini中的配置
```
# 这个选项是你放置数据库模型文件的地方, 相对目录
DalDir=models

# 这个是定义文件放置的地方，文件中定义了需要再全局使用的一些类型定义, 供生成代码引用
DefineTemplateFile=xsbuild/template.go

```

* 执行 `xstpl`
* 在.gitignore中，忽略`*lib_auto_generate*`文件