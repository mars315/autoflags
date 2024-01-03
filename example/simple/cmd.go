package main

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
// output: main.Config{Name:"default name", Age:133}
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
