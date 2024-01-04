// `go run cmd.go --f2 "KK"`
// output:
//	{
//		"F1": "x86xm",
//		"F2": "KK",
//		"F3": 87,
//		"A4": {
//		"F4": 99
//		},
//		"DBUrl": ":27071",
//		"LogFile": "stdout",
//		"Debug": true,
//		"Name": "test",
//		"Short": "s",
//		"Age": 18,
//		"Usage": "usage",
//		"KeepTime": 1000000000,
//		"NoUse": ""
//	}
//
// `go run cmd.go -h`
// output:
//	Flags:
//		--a3.f3 int        f3 (default 87)
//		--age int          age (default 18)
//		--c.f1 string      f1 (default "x86xm")
//		--dburl string     dburl (default ":27071")
//		--debug            enable debug model,false to disable; ,please (default true)
//		--f2 string        f2 (default "ZH")
//		--f4 int           f4 (default 99)
//		-h, --help             help for test
//		--keep duration     (default 1s)
//		--logfile string   udp|udp:UdpAddr|FilePath|redirect:x (default "stdout")
//		--name string      name (default "test")
//		--short string     short (default "s")
//		--usage string     usage (default "usage")
//

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mars315/autoflags"
	"github.com/spf13/cobra"
)

func main() {
	v := new(Flag)
	v.A3 = new(A3)
	rootCmd := &cobra.Command{
		Use: "test auto read flag",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", v)
		},
	}

	if err := autoflags.BindAndExecute(rootCmd, v,
		autoflags.WithTagNameOption("mapstructure"),
		autoflags.WithIgnoreUntaggedFieldsOption(false),
		autoflags.WithSquashOption(false)); err != nil {
		panic(err)
	}
}

type Flag struct {
	A1 `mapstructure:"c"`
	A2 `mapstructure:",squash"`
	*A3
	A4       A4            `mapstructure:",squash"`
	DBUrl    string        `mapstructure:"dburl, desc:dburl, default::27071"`
	LogFile  string        `mapstructure:"logfile, default:stdout, desc:udp|udp:UdpAddr|FilePath|redirect:x"`
	Debug    bool          `mapstructure:"debug, default:true, desc:enable debug model\\,false to disable; \\,please"` // "\\" before ","
	Name     string        `mapstructure:",desc:name, default:test"`
	Short    string        `mapstructure:"short, desc:short, default:s"`
	Age      int           `mapstructure:"age, desc:age, default:18"`
	Usage    string        `mapstructure:"usage, desc:usage, default:usage"`
	KeepTime time.Duration `mapstructure:"keep,omitempty, default:1s""`
	NoUse    string        `mapstructure:"-"`
}

func (f *Flag) String() string {
	str, _ := json.MarshalIndent(f, "", "\t")
	return string(str)
}

type A1 struct {
	F1 string `mapstructure:"f1,desc:f1,default:x86xm"`
}

type A2 struct {
	F2 string `mapstructure:"f2,desc:f2,default:ZH"`
}

type A3 struct {
	F3 int `mapstructure:"f3,desc:f3,default:87"`
}

type A4 struct {
	F4 int `mapstructure:"f4,desc:f4,default:99"`
}
