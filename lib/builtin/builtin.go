// Copyright © 2023 mars315 <254262243@qq.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//

package builtin

type (
	// UnsignedInteger .
	UnsignedInteger interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}

	// SignedInteger .
	SignedInteger interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}

	// Integer .
	Integer interface {
		UnsignedInteger | SignedInteger
	}

	// Float .
	Float interface {
		~float32 | ~float64
	}

	Any = any
)
