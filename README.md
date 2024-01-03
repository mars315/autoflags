# 总览

结合 [viper](https://github.com/spf13/viper) 和 [cobra](https://github.com/spf13/cobra) 来构建命令行工具，通过一个struct绑定命令行参数，同时支持从`viper`中获取值
tag的第一个标签为 flag的名字，如果没有指定tag，则使用字段名字作为flag的名字

>> FIELD  FIELD_TYPE `FLAG:FLAG_NAME,LABEL,OTHER_LABEL:OTHER_VALUE"`
* Label:
 - short: short flag name
 - desc: description
 - default: default value
 - squash: squash all anonymous structs
 - `-`: skip this field

* 支持匿名struct以及匿名struct的指针
* 支持long flag和short flag；支持flag的默认值、描述
* 支持自动从`viper`中获取值到struct中
* 支持指定tag的名字
* 支持类型(string, bool, int, int32, int64, me.Duration oat32, float64, []string, []int, struct, struct pointer)


# 为什么
当前使用 `github.com/spf13` 来构建app以及绑定命令行参数已经很方便，`viper`可以指定参数来自于多个源，在使用时可以用 `viper:GetXXXX`来获取值。
但是对于我来说还不足够方便,我希望
1. 通过一个结构体，同过指定标签的方式来指定命令行参数，同时支持自动从`viper`中获取值；
这样就可以在一个结构体中定义所有的参数并且直接使用，而且不用关心参数来自哪里。
2. 可以避免将命令函的参数名字硬编码到代码的各个地方,比如 `viper.GetString("xxx")`，这样如果参数名字变了，就需要修改代码

# 怎么做

```go
import (
    "fmt"
    "github.com/mars315/autoflags"
    "github.com/spf13/cobra"
)

type Config struct {
    Name string `flag:"name,short:N,default:default name,desc:your name"`
    Age  int    `flag:"age,short:A,default:18,desc:your age"`
}

// main.go
// `go run main.go -h`
// output:
// Flags:
//	-A, --age int       your age (default 18)
//	-h, --help          help for this command
//	-N, --name string   your name (default "default name")
//
// `go run cmd.go --age 133`
// output: main.Config{Name:"default name", Age:13}
//

func main() {
    var cfg Config
    rootCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Printf("%#v\n", cfg)
        },
    }
    if err := autoflags.BindAndExecute(rootCmd, &cfg); err != nil {
		panic(err)
    }
}

```

使用匿名结构体
```go
type Child struct {
    Name string `flag:"name,default:default name"`
}

type Base struct {
    Age int `flag:"age,default:18"`
}

type Config struct {
    Child
	Base
    Addr string `flag:"addr,default:localhost"`
}

// main.go
// `go run main.go -h`
// output:
// Flags:
//    --addr string    (default "localhost")
//    --age int        (default 18)
//    -h, --help          help for this command
//    --name string    (default "default name")

func main() {
    var cfg Config
    rootCmd := &cobra.Command{
        Run: func(cmd *cobra.Command, args []string) {
         fmt.Printf("%#v\n", cfg)
        },
    }
    if err := autoflags.BindAndExecute(rootCmd, &cfg); err != nil {
        panic(err)
    }
}

// 如果想让`Child`中的`Name`的flag为`--child.name`，可以这样:

...
    if err := autoflags.BindAndExecute(rootCmd, &cfg, autoflags.WithSquashOption(false)) {
...

// `go run main.go -h` output:
//Flags:
//    --addr string          (default "localhost")
//    --base.age int         (default 18)
//    --child.name string    (default "default name")
//    -h, --help                help for this command

// 如果想让`Child`中的`Name`的flag为`--child.name`，而`Base`的`Age`的flag不是`--base.age`，可以这样:
type Config struct {
    Child
    Base   `flag:",squash"` 
    Addr string `flag:"addr,default:localhost"`
}

...
if err := autoflags.BindAndExecute(rootCmd, &cfg, autoflags.WithSquashOption(false)) {
...

// `go run main.go -h` output:
//Flags:
//    --addr string          (default "localhost")
//    --age int         (default 18)
//    --child.name string    (default "default name")
//    -h, --help                help for this command

// 如果想让某些字段不指定为flag，可以这样:
type Config struct {
	...
    Ignore string `flag:"-"`
	...
}

// 如果想用其他tag名字比如`mapstructure`，可以这样:
type Config struct {
    ...
	Addr string `mapstructure:"addr,default:localhost"`
    ...
}

...
if err := autoflags.BindAndExecute(rootCmd, &cfg, autoflags.WithTagNameOption("mapstructure")) {
...

```


# 安装
使用 `go get` 安装.

```
go get -u github.com/mars315/autoflags@latest
```

然后在代码中引入:

```go
import "github.com/mars315/autoflags"
```
# 许可

autoflags is released under MIT License. See [LICENSE](LICENSE)
