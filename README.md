# 使用说明

* `xstpl init` 执行上面的命令，会在当前目录生成一个文件夹xsbuild
* 编辑xsbuild/config.ini中的配置
```
# 这个选项是你放置数据库模型文件的地方, 相对目录
DalDir=models

# 这个是定义文件放置的地方，文件中定义了需要再全局使用的一些类型定义, 供生成代码引用
DefineTemplateFile=xsbuild/template.go

# 这个定义了驱动（xsorm）的方法定义的文件，暂时没想到更好的方法，手动获取xsorm包之后，获得目录，填入绝对路径
DriverBuildFile=/Users/kevin/go/pkg/mod/github.com/xxiiaass/xsorm@v0.0.0-20210811114407-2929ca97f8ab/build.go
```

* 执行 `xstpl`