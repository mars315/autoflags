// Copyright © 2023 mars315 <254262243@qq.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// Auto-bind application flags
// supported type:
//
//	string, bool,
//	int, int32, int64,
//	time.Duration
//	float32, float64,
//	[]string, []int
//	struct, struct pointer
//
// first label is the flag name
//
// FIELD  FIELD_TYPE `FLAG:FLAG_NAME,LABEL,OTHER_LABEL:OTHER_VALUE"`
// Label:
// - short: short name
// - desc: description
// - default: default value
// - squash: squash all anonymous structs
// - `-` skip this field
//
// e.g.
// LongName string flag:"name"` -> --name
// LongName string flag:",short:N"` -> --longname, -N
// *****************************
//
// Define a struct:
//
//	type GFlag struct {
//			Port int32 `flag:port,short:P,desc:port,default:20001"`
//	}
//
// And then use `BindAndExecute` or `BindFlags` to bind the flags like this:
//
// cmd := &cobra.Command{}
// BindAndExecute(cmd, &GFlag{})
// BindFlags(cmd, &GFlag{})
//
// Use a different tag name like this:
//
//	type GFlag struct {
//			Port int32 `mapstructure:"port,short:P,desc:port,default:20001"`
//	}
//
// BindFlags(cmd, &GFlag{}, WithTagNameOption("mapstructure"))
//
// Now you can use:
// `go run main.go --port=20002` to change the port
// `go run main.go -P=20002` to change the port
//
// If some values of the flags come from sources supported by Viper, enable WithAutoUnMarshalOption().
//
// ReadFlags(&v)
// ReadFlags(&v, WithTagNameOption("mapstructure"))
// UnmarshalFlags(&v)
// UnmarshalFlags(&v, WithTagNameOption("mapstructure"))
// BindAndExecute(cmd,&GFlag{}, WithAutoUnMarshalOption())
// BindAndExecute(cmd,&GFlag{}, WithAutoUnMarshalOption(), WithTagNameOption("mapstructure"))
//

package autoflags

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mars315/autoflags/lib/builtin"
	"github.com/mars315/autoflags/lib/stringx"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	TagName         = "flag"
	TagLabelShort   = "short"
	TagLabelDesc    = "desc"
	TagLabelDefault = "default"
	TagLabelSquash  = "squash"
	TagLabelSkip    = "-"
	TagLabelSep     = ","
)

type (
	FlagOption func(*FlagConfig)
	FlagConfig struct {
		// ignoreUntaggedFields ignores all struct fields without explicit, default is "false"
		ignoreUntaggedFields bool
		// Squash all anonymous structs. default is "true"
		// A squash tag may also be added to an individual struct field using a tag. For example:
		//
		//  type Parent struct {
		//		Stand `flag:"stand"`
		//      Child `flag:",squash"`
		//  }
		//
		squash bool
		// parent (squash == false)
		parent []string
		// read the flag value from viper
		autoUnMarshalFlag bool
		// run pre auto marshal flags
		preAutoUnMarshal func(cmd *cobra.Command, args []string)
		// run pre auto marshal flags with error
		preAutoUnMarshalE func(cmd *cobra.Command, args []string) error
		// The tag name that flag reads for field names, default is "flag"
		tagName string
		// The tag label separator, default is  ","
		tagLabelSep string
	}
)

// BindAndExecute automatically bind flag and execute
func BindAndExecute(cmd *cobra.Command, v0 builtin.Any, opts ...FlagOption) error {
	if err := BindFlags(cmd, v0, opts...); err != nil {
		return err
	}

	return cmd.Execute()
}

// BindFlags v0 must be a pointer and the structure where the variable is located
// supported type: string, bool, int, int32, int64, float32, float64, []string, []int time.Duration
//
//	struct and struct pointer
func BindFlags(cmd *cobra.Command, v0 builtin.Any, opts ...FlagOption) error {
	autoMarshalOption(cmd, v0, opts...)
	if err := bindFlags(cmd, v0, defaultFlagConfig(opts...)); err != nil {
		return err
	}

	return viper.BindPFlags(cmd.Flags())
}

// ReadFlags read flag value from viper
// supported type: string, bool, int, int32, int64, float32, float64, []string, []int time.Duration
//
//	struct and struct pointer
func ReadFlags(v0 builtin.Any, opts ...FlagOption) error {
	cfg := defaultFlagConfig(opts...)
	return readFlags(v0, cfg)
}

// UnmarshalFlags unmarshal flag value from viper
// use `mapstructure` to unmarshal
func UnmarshalFlags(v0 builtin.Any, opts ...FlagOption) error {
	defaultOpts := castConfigOptions(defaultFlagConfig(opts...))
	return viper.Unmarshal(v0, defaultOpts...)
}

/////////////////////////////////////////////////////// option ///////////////////////////////////////////////////////

// WithTagNameOption custom tag name
func WithTagNameOption(tag string) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.tagName = tag
	}
}

// WithTagLabelSepOption  tag label separator
func WithTagLabelSepOption(sep string) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.tagLabelSep = sep
	}
}

// WithAutoUnMarshalOption auto unmarshal flag value from viper
// In particular, the flag value comes from different sources (e.g. viper)
func WithAutoUnMarshalOption() FlagOption {
	return func(cfg *FlagConfig) {
		cfg.autoUnMarshalFlag = true
	}
}

// WithIgnoreUntaggedFieldsOption .
func WithIgnoreUntaggedFieldsOption(ignore bool) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.ignoreUntaggedFields = ignore
	}
}

// WithSquashOption if true all embedded structs will be flattened
func WithSquashOption(squash bool) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.squash = squash
	}
}

// WithPreAutoUnMarshalOption executed before `UnmarshalFlags`, can be used to add the data source of `viper`
func WithPreAutoUnMarshalOption(pre func(cmd *cobra.Command, args []string)) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.preAutoUnMarshal = pre
	}
}

// WithPreAutoUnMarshalEOption executed before `UnmarshalFlags`, can be used to add the data source of `viper`
func WithPreAutoUnMarshalEOption(preE func(cmd *cobra.Command, args []string) error) FlagOption {
	return func(cfg *FlagConfig) {
		cfg.preAutoUnMarshalE = preE
	}
}

/////////////////////////////////////////////////////// implement ///////////////////////////////////////////////////////

func bindFlags(cmd *cobra.Command, v0 builtin.Any, cfg *FlagConfig) error {
	if reflect.TypeOf(v0).Kind() != reflect.Pointer {
		return fmt.Errorf("v0 must be pointer")
	}

	v := reflect.ValueOf(v0).Elem()
	t := v.Type()
	flagSet := cmd.Flags()
	for i := 0; i < v.NumField(); i++ {
		fValue := v.Field(i)
		field := t.Field(i)
		tag := parseTag(field, cfg)
		if tag == nil {
			continue
		}
		switch fValue.Kind() {
		case reflect.String:
			flagSet.StringVarP(fValue.Addr().Interface().(*string), tag.Name, tag.Short, tag.Default, tag.Desc)
		case reflect.Bool:
			flagSet.BoolVarP(fValue.Addr().Interface().(*bool), tag.Name, tag.Short, stringx.ToBool(tag.Default), tag.Desc)
		case reflect.Float32:
			flagSet.Float32VarP(fValue.Addr().Interface().(*float32), tag.Name, tag.Short, stringx.Atof[float32](tag.Default), tag.Desc)
		case reflect.Float64:
			flagSet.Float64VarP(fValue.Addr().Interface().(*float64), tag.Name, tag.Short, stringx.Atof[float64](tag.Default), tag.Desc)
		case reflect.Int:
			flagSet.IntVarP(fValue.Addr().Interface().(*int), tag.Name, tag.Short, stringx.Atoi[int](tag.Default), tag.Desc)
		case reflect.Int32:
			flagSet.Int32VarP(fValue.Addr().Interface().(*int32), tag.Name, tag.Short, stringx.Atoi[int32](tag.Default), tag.Desc)
		case reflect.Int64:
			bindInt64(flagSet, fValue, tag)
		case reflect.Slice:
			if err := bindSlice(flagSet, fValue, field, tag); err != nil {
				return err
			}
		case reflect.Struct:
			if err := bindStruct(cmd, fValue, field, cfg); err != nil {
				return err
			}
		case reflect.Pointer:
			if err := bindPointer(cmd, fValue, field, cfg); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type: %s|%s", field.Name, fValue.Kind())
		}
	}
	return nil
}

func readFlags(v0 builtin.Any, cfg *FlagConfig) error {
	v := reflect.ValueOf(v0).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fValue := v.Field(i)
		field := t.Field(i)
		tag := getTag(field, cfg)
		if tag == nil {
			continue
		}
		switch fValue.Kind() {
		case reflect.String:
			fValue.Set(reflect.ValueOf(viper.GetString(tag.Name)))
		case reflect.Bool:
			fValue.Set(reflect.ValueOf(viper.GetBool(tag.Name)))
		case reflect.Float32:
			fValue.Set(reflect.ValueOf(float32(viper.GetFloat64(tag.Name))))
		case reflect.Float64:
			fValue.Set(reflect.ValueOf(viper.GetFloat64(tag.Name)))
		case reflect.Int:
			fValue.Set(reflect.ValueOf(viper.GetInt(tag.Name)))
		case reflect.Int32:
			fValue.Set(reflect.ValueOf(viper.GetInt32(tag.Name)))
		case reflect.Int64:
			readInt64(fValue, tag)
		case reflect.Slice:
			if err := readSlice(fValue, tag); err != nil {
				return err
			}
		case reflect.Struct:
			if err := readStruct(fValue, field, cfg); err != nil {
				return err
			}
		case reflect.Pointer:
			if err := readPointer(fValue, field, cfg); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported type: %s|%s", field.Name, fValue.Kind())
		}
	}
	return nil
}

/////////////////////////////////////////////////////// cast ///////////////////////////////////////////////////////

// alias
type decoderConfigOption = viper.DecoderConfigOption

func castConfigOptions(cfg *FlagConfig) []decoderConfigOption {
	return []decoderConfigOption{
		withSquashOption(true),
		withTagNameOption(cfg.tagName),
		withIgnoreUntaggedFieldsOption(cfg.ignoreUntaggedFields),
	}
}

func withSquashOption(squash bool) decoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.Squash = squash
	}
}

// 自定义tag
func withTagNameOption(tag string) decoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = tag
	}
}

// ignore undefined tag fields
func withIgnoreUntaggedFieldsOption(ignore bool) decoderConfigOption {
	return func(config *mapstructure.DecoderConfig) {
		config.IgnoreUntaggedFields = ignore
	}
}

/////////////////////////////////////////////////////// helper ///////////////////////////////////////////////////////

// set  auto marshal function
func autoMarshalOption(cmd *cobra.Command, v0 builtin.Any, opts ...FlagOption) {
	cfg := defaultFlagConfig(opts...)
	if !cfg.autoUnMarshalFlag {
		return
	}

	if cmd.PreRun != nil {
		handler := cmd.PreRun
		cmd.PreRun = func(cmd *cobra.Command, args []string) {
			if cfg.preAutoUnMarshal != nil {
				cfg.preAutoUnMarshal(cmd, args)
			}
			_ = UnmarshalFlags(v0, opts...)

			handler(cmd, args)
		}
	} else if cmd.PreRunE != nil {
		handler := cmd.PreRunE
		cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
			if cfg.preAutoUnMarshalE != nil {
				if err := cfg.preAutoUnMarshalE(cmd, args); err != nil {
					return err
				}
			}
			if err := UnmarshalFlags(v0, opts...); err != nil {
				return err
			}

			return handler(cmd, args)
		}
	}
}

func defaultFlagConfig(opts ...FlagOption) *FlagConfig {
	cfg := &FlagConfig{
		tagName:     TagName,
		tagLabelSep: TagLabelSep,
		squash:      true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func isStepInto(field reflect.StructField) bool {
	return field.Type.Kind() == reflect.Struct ||
		(field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct)
}

func tryStepOut(field reflect.StructField, cfg *FlagConfig) {
	if len(cfg.parent) == 0 {
		return
	}

	tag := getTag(field, cfg)
	squash := cfg.squash || tag != nil && tag.squash
	if squash {
		return
	}

	cfg.parent = cfg.parent[:len(cfg.parent)-1]
}

// ///////////////////////////////////////////////////// tag ///////////////////////////////////////////////////////

// private
type tagData struct {
	origin  string
	Name    string
	Short   string
	Desc    string
	Default string
	squash  bool
}

func parseTag(field reflect.StructField, cfg *FlagConfig) *tagData {
	if !field.IsExported() {
		return nil
	}
	data := getTag(field, cfg)
	if data == nil {
		return nil
	}

	// add prefix
	// type Base struct {Name string}
	// type Top struct {Base; Level int}
	// skip `Base` field // ignoreUntaggedFields == true
	// --name // ignoreUntaggedFields == false && (cfg.Squash == true || ".squash" in tag)
	// --base.name // ignoreUntaggedFields == false && squash == false
	if !cfg.squash && !data.squash && isStepInto(field) {
		cfg.parent = append(cfg.parent, data.origin)
	}

	return data
}

// getTag .
func getTag(field reflect.StructField, cfg *FlagConfig) *tagData {
	fulls, ok := field.Tag.Lookup(cfg.tagName)

	// ignore untagged field
	if cfg.ignoreUntaggedFields && !ok {
		return nil
	}

	names := strings.Split(strings.TrimSpace(fulls), cfg.tagLabelSep)
	settings := make(map[string]string)
	for i := 0; i < len(names); i++ {
		j := i
		if j == 0 {
			settings[cfg.tagName] = strings.TrimSpace(names[j])
			continue
		}

		for i < len(names) {
			if names[j][len(names[j])-1] != '\\' {
				break
			}
			i++
			names[j] = names[j][0:len(names[j])-1] + cfg.tagLabelSep + names[i]
			names[i] = ""
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(values[0])
		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	// skip `-`
	if settings[cfg.tagName] == TagLabelSkip {
		return nil
	}
	data := &tagData{
		Name:    settings[cfg.tagName],
		Short:   settings[TagLabelShort],
		Desc:    settings[TagLabelDesc],
		Default: settings[TagLabelDefault],
	}

	// untagged field use field name as the flag name
	if len(data.Name) == 0 {
		data.Name = strings.ToLower(field.Name)
	}

	_, squashLabel := settings[TagLabelSquash]
	data.squash = squashLabel && isStepInto(field)

	data.origin = data.Name

	// add prefix
	// type Base struct {Name string}
	// type Top struct {Base; Level int}
	// skip `Base` field // ignoreUntaggedFields == true
	// --name // ignoreUntaggedFields == false && (cfg.Squash == true || ".squash" in tag)
	// --base.name // ignoreUntaggedFields == false && squash == false
	if !cfg.squash {
		if len(cfg.parent) > 0 {
			data.Name = strings.Join(cfg.parent, ".") + "." + data.Name
		}
	}

	return data
}

/////////////////////////////////////////////////////// struct ///////////////////////////////////////////////////////

func bindStruct(cmd *cobra.Command, fValue reflect.Value, field reflect.StructField, cfg *FlagConfig) error {
	defer tryStepOut(field, cfg)
	return bindFlags(cmd, fValue.Addr().Interface(), cfg)
}

func readStruct(fValue reflect.Value, field reflect.StructField, cfg *FlagConfig) error {
	defer tryStepOut(field, cfg)
	return readFlags(fValue.Addr().Interface(), cfg)
}

/////////////////////////////////////////////////////// pointer ///////////////////////////////////////////////////////

func bindPointer(cmd *cobra.Command, fValue reflect.Value, field reflect.StructField, cfg *FlagConfig) error {
	if fValue.IsNil() {
		return fmt.Errorf("nil value of *%s", field.Name)
	}

	if fValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type: %s|%s(%s)", field.Name, fValue.Kind(), fValue.Elem().Kind())
	}
	defer tryStepOut(field, cfg)
	return bindFlags(cmd, fValue.Interface(), cfg)
}

func readPointer(fValue reflect.Value, field reflect.StructField, cfg *FlagConfig) error {
	if fValue.IsNil() {
		return fmt.Errorf("nil value of *%s", field.Name)
	}

	if fValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("unsupported type: %s|%s(%s)", field.Name, fValue.Kind(), fValue.Elem().Kind())
	}

	defer tryStepOut(field, cfg)
	return readFlags(fValue.Interface(), cfg)
}

/////////////////////////////////////////////////////// int64 ///////////////////////////////////////////////////////

func bindInt64(flagSet *flag.FlagSet, fValue reflect.Value, tag *tagData) {
	switch fValue.Addr().Interface().(type) {
	case *time.Duration:
		duration, _ := time.ParseDuration(tag.Default)
		flagSet.DurationVarP(fValue.Addr().Interface().(*time.Duration), tag.Name, tag.Short, duration, tag.Desc)
	default:
		flagSet.Int64VarP(fValue.Addr().Interface().(*int64), tag.Name, tag.Short, stringx.Atoi[int64](tag.Default), tag.Desc)
	}
}

func readInt64(fValue reflect.Value, tag *tagData) {
	i := fValue.Addr().Interface()
	switch i.(type) {
	case *time.Duration:
		fValue.Set(reflect.ValueOf(viper.GetDuration(tag.Name)))
	default:
		fValue.Set(reflect.ValueOf(viper.GetInt64(tag.Name)))
	}
}

/////////////////////////////////////////////////////// slice ///////////////////////////////////////////////////////

func bindSlice(flagSet *flag.FlagSet, fValue reflect.Value, field reflect.StructField, tag *tagData) error {
	switch fValue.Type().Elem().Kind() {
	case reflect.String:
		bindStringSlice(flagSet, fValue, tag)
	case reflect.Int:
		bindIntSlice(flagSet, fValue, tag)
	default:
		return fmt.Errorf("field `%s` unsupported slice type %s", field.Name, fValue.Type().Elem().Kind())
	}
	return nil
}

func readSlice(fValue reflect.Value, tag *tagData) error {
	switch fValue.Type().Elem().Kind() {
	case reflect.String:
		readStringSlice(fValue, tag)
	case reflect.Int:
		readIntSlice(fValue, tag)
	default:
		return fmt.Errorf("unsupported slice type: %s|%s", fValue.Type().Elem().Name(), fValue.Type().Elem().Kind())
	}
	return nil
}

func bindIntSlice(flagSet *flag.FlagSet, fValue reflect.Value, tag *tagData) {
	flagSet.IntSliceVarP(fValue.Addr().Interface().(*[]int), tag.Name, tag.Short, stringx.AtoSlice[int](tag.Default, ","), tag.Desc)
}

func readIntSlice(fValue reflect.Value, tag *tagData) {
	fValue.Set(reflect.ValueOf(viper.GetIntSlice(tag.Name)))
}

func bindStringSlice(flagSet *flag.FlagSet, fValue reflect.Value, tag *tagData) {
	flagSet.StringSliceVarP(fValue.Addr().Interface().(*[]string), tag.Name, tag.Short, stringx.Split(tag.Default, ","), tag.Desc)
}

func readStringSlice(fValue reflect.Value, tag *tagData) {
	fValue.Set(reflect.ValueOf(viper.GetStringSlice(tag.Name)))
}
