# Overview

Combine [viper](https://github.com/spf13/viper) and [cobra](https://github.com/spf13/cobra) to build a command-line tool, binding command-line arguments to a struct and supporting value retrieval from `viper`.

The first tag of the field represents the name of the flag. If no tag is specified, the field name is used as the flag name.

>> FIELD  FIELD_TYPE `FLAG:FLAG_NAME,LABEL,OTHER_LABEL:OTHER_VALUE"`
* Label:
    - short: short flag name
    - desc: description
    - default: default value
    - squash: squash all anonymous structs
    - `-`: skip this field

* Supports anonymous structs and pointers to anonymous structs.
* Supports long and short flags; supports default values and descriptions for flags.
* Supports automatic retrieval of values from `viper` into the struct.
* Supports specifying the name of the tag.
* Supports types (string, bool, int, int32, int64, float32, float64, []string, []int, struct, struct pointer).

# WHY
Currently, using `github.com/spf13` to build apps and bind command-line arguments is already convenient. `viper` allows specifying parameters from multiple sources, and values can be retrieved using `viper.GetXXXX`. However, it is not convenient enough for me, and I would like to:
1. Specify command-line parameters through a struct using tags and automatically retrieve values from `viper`.
   This way, I can define all the parameters in one struct and directly use them without worrying about where the parameters come from.
2. Avoid hardcoding the names of command function parameters throughout the code, such as `viper.GetString("xxx")`. If the parameter name changes, modifying the code would be necessary.

# HOW

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

anonymous struct
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

// If you want to set the flag for `Name` in `Child` as `--child.name`, you can do it like this:

...
    if err := autoflags.BindAndExecute(rootCmd, &cfg, autoflags.WithSquashOption(false)) {
...

// `go run main.go -h` output:
//Flags:
//    --addr string          (default "localhost")
//    --base.age int         (default 18)
//    --child.name string    (default "default name")
//    -h, --help                help for this command

// If you want to set the flag for `Name` in `Child` as `--child.name`, and the flag for `Age` in `Base` is not `--base.age`, you can do it like this:
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

// If you want to exclude certain fields from being specified as flags, you can do it like this:
type Config struct {
	...
    Ignore string `flag:"-"`
	...
}

// If you want to use a different tag name like `mapstructure`, you can do it like this:
type Config struct {
    ...
	Addr string `mapstructure:"addr,default:localhost"`
    ...
}

...
if err := autoflags.BindAndExecute(rootCmd, &cfg, autoflags.WithTagNameOption("mapstructure")) {
...

```


# Installing
First, use `go get` to install the latest version  of the library.

```
go get -u github.com/mars315/autoflags@latest
```

Next, include autoflags in your application:

```go
import "github.com/mars315/autoflags"
```
# License

autoflags is released under MIT License. See [LICENSE](LICENSE)
