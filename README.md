# 使用说明

* `go get github.com/xxiiaass/xsorm` 获取依赖的驱动项目
* `xstpl init` 执行上面的命令，会在当前目录生成一个文件夹xsbuild, 在.gitignore中添加忽略, 不许随代码提交
* 编辑xsbuild/config.ini中的配置
```
# 这个选项是你放置数据库模型文件的地方, 相对目录
DalDir=models

# 这个是定义文件放置的地方，文件中定义了需要再全局使用的一些类型定义, 供生成代码引用
DefineTemplateFile=xsbuild/template.go

```

* 执行 `xstpl`