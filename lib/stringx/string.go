// Copyright © 2023 mars315 <254262243@qq.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//

package stringx

import (
	"strconv"
	"strings"

	"github.com/mars315/autoflags/lib/builtin"
)

// SafeTokens 切分字符串，去掉空格
func SafeTokens(cmd string, sep string) []string {
	if cmd = strings.TrimSpace(cmd); len(cmd) == 0 {
		return nil
	}

	tokens := strings.Split(cmd, sep)
	list := tokens[:0]
	for _, token := range tokens {
		str := strings.TrimSpace(token)
		if len(str) > 0 {
			list = append(list, str)
		}
	}
	return list
}

func ToBool(s string) bool {
	if len(s) == 0 {
		return false
	}

	s = strings.ToLower(s)
	return s == "true" || s == "t"
}

// Atof string to float64
func Atof[T builtin.Float](v string) T {
	vF64, _ := strconv.ParseFloat(v, 64)
	return T(vF64)
}

// Atoi string to signed integer
func Atoi[T builtin.SignedInteger](v string) T {
	vInt, _ := strconv.ParseInt(v, 10, 64)
	return T(vInt)
}

// AtoSlice string to signed integer slice
func AtoSlice[T builtin.SignedInteger](s string, sep string) []T {
	ss := SafeTokens(s, sep)
	if len(ss) == 0 {
		return nil
	}

	l := make([]T, 0, len(ss))
	for _, v := range ss {
		l = append(l, Atoi[T](v))
	}
	return l
}

// Split Like strings.Split, but remove the spaces from each string.
func Split(s0, sep string) []string {
	s := strings.TrimSpace(s0)
	l := strings.Split(s, sep)

	r := l[:0]
	for _, str := range l {
		r = append(r, strings.TrimSpace(str))
	}

	return r
}
