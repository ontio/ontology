/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ontio/ontology/vm/wasmvm/exec"
	"github.com/ontio/ontology/vm/wasmvm/validate"
	"github.com/ontio/ontology/vm/wasmvm/wasm"
)

const (
	nonSpecTestsDir = "./test_data"
	specTestsDir    = "./test_data/spec"
)

type testCase struct {
	Function string   `json:"function"`
	Args     []string `json:"args"`
	Return   string   `json:"return"`
	Trap     string   `json:"trap"`
}

type file struct {
	FileName string     `json:"file"`
	Tests    []testCase `json:"tests"`
}

var reValue = regexp.MustCompile(`(.+)\:(.+)`)

func parseFloat(str string, bitSize int) float64 {
	hexFloat, err := regexp.MatchString("0x", str)
	if err != nil {
		panic(err)
	}
	isNanOrInf, err := regexp.MatchString("(nan|inf)", str)
	if err != nil {
		panic(err)
	}

	if isNanOrInf {
		if strings.HasPrefix(str, "-") {
			str = strings.TrimPrefix(str, "-")
			if str == "inf" {
				return math.Inf(-1)
			} else if str == "nan" {
				return math.NaN()
			}
		}
		if str == "inf" {
			return math.Inf(+1)
		} else if str == "nan" {
			return math.NaN()
		}
	}

	if hexFloat {
		if strings.HasPrefix(str, "-0x") {
			str = strings.TrimPrefix(str, "-0x")
			if str == "inf" {
				return math.Inf(-1)
			}
			str = "-" + str
		} else {
			if str == "inf" {
				return math.Inf(+1)
			} else if str == "nan" {
				return math.NaN()
			}

			str = strings.TrimPrefix(str, "0x")
		}

		f, _, err := big.ParseFloat(str, 16, big.MaxPrec, big.ToNearestEven)
		if err != nil {
			panic(err)
		}

		n, _ := f.Float64()
		return n
	}

	n, err := strconv.ParseFloat(str, bitSize)
	if err != nil {
		panic(err)
	}
	return n
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		str  string
		want float64
	}{
		{"0xf32", 3890.0},
		{"3.4", 3.4},
		{"0.0", 0.0},
		{"-123.0", -123.0},
		{"0x1.fffffep+127", 340282346638528859811704183484516925440.0},
		{"nan", math.NaN()},
		{"inf", math.Inf(+1)},
		{"-inf", math.Inf(-1)},
	}
	for _, test := range tests {
		f := parseFloat(test.str, 64)
		if f != test.want && !(math.IsNaN(test.want) && math.IsNaN(f)) {
			t.Errorf("unexpected value returned by parseFloat: got=%f want=%f", f, test.want)
		}
	}
}

func parseInt(str string, bitSize int) uint64 {
	isHex, _ := regexp.MatchString("0x", str)
	base := 10

	if isHex {
		if strings.HasPrefix(str, "-") {
			str = strings.TrimPrefix(str, "-0x")
			str = "-" + str
		} else {
			str = strings.TrimPrefix(str, "0x")
		}
		base = 16
	}

	n, err := strconv.ParseUint(str, base, bitSize)
	if err != nil {
		// try parsing as an int
		n2, err := strconv.ParseInt(str, base, bitSize)
		if err != nil {
			panic(err)
		}
		n = uint64(n2)
	}
	return n
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		str  string
		want uint64
	}{
		{"45", 45},
		{"0", 0},
		{"0xABADCAFEDEAD1DEA", 0xABADCAFEDEAD1DEA},
		{"-1", 18446744073709551615},
	}
	for _, test := range tests {
		i := parseInt(test.str, 64)
		if i != test.want {
			t.Errorf("unexpected value returned by parseInt: got=%d want=%d", i, test.want)
		}
	}
}

func parseValue(str string) interface{} {
	if str == "" {
		return nil
	}
	matches := reValue.FindStringSubmatch(str)
	if matches == nil {
		panic("Invalid value expression: " + str)
	}

	switch matches[1] {
	case "i32":
		n := parseInt(matches[2], 32)
		return uint32(n)
	case "i64":
		n := parseInt(matches[2], 64)
		return n
	case "f32":
		return float32(parseFloat(matches[2], 32))
	case "f64":
		return parseFloat(matches[2], 64)
	default:
		panic("invalid value_type prefix " + matches[1])
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		str  string
		want interface{}
	}{
		{"i64:45", uint64(45)},
		{"f32:0.0", float32(0)},
		{"f64:0x1.fffffep+127", float64(340282346638528859811704183484516925440.0)},
		{"f64:inf", float64(math.Inf(+1))},
		{"f32:inf", float32(math.Inf(+1))},
	}
	for _, test := range tests {
		v := parseValue(test.str)
		if !reflect.DeepEqual(test.want, v) {
			t.Errorf("unexpected value returned by parseInt: got=%v want=%v", v, test.want)
		}
	}
}

func parseArgs(args []string) (arr []uint64) {
	for _, str := range args {
		v := parseValue(str)
		var n uint64
		switch v.(type) {
		case uint64:
			n = v.(uint64)
		case uint32:
			n = uint64(v.(uint32))
		case float32:
			n = uint64(math.Float32bits(v.(float32)))
		case float64:
			n = math.Float64bits(v.(float64))
		default:
			panic(fmt.Sprintf("invalid value type: %v(%v)", reflect.TypeOf(v), v))
		}

		arr = append(arr, n)
	}

	return arr
}

func fnString(fn string, args []string) string {
	if len(args) == 0 {
		return fn
	}
	return fmt.Sprintf("%s(%v)", fn, args)
}

func panics(fn func()) (panicked bool, msg string) {
	defer func() {
		r := recover()
		panicked = r != nil
		msg = fmt.Sprint(r)
	}()

	fn()
	return
}

func runTest(fileName string, testCases []testCase, t testing.TB) {
	file, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	module, err := wasm.ReadModule(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = validate.VerifyModule(module); err != nil {
		t.Fatalf("%s: %v", fileName, err)
	}

	vm, err := exec.NewVM(module)
	if err != nil {
		t.Fatalf("%s: %v", fileName, err)
	}

	b, ok := t.(*testing.B)
	for _, testCase := range testCases {
		var expected interface{}

		index := module.Export.Entries[testCase.Function].Index
		args := parseArgs(testCase.Args)

		if testCase.Return != "" {
			expected = parseValue(testCase.Return)
		}

		if testCase.Trap != "" {
			// don't benchmark tests that involve trapping the VM
			fn := func() {
				_, err := vm.ExecCode(false, int64(index), args...)
				if err != nil {
					t.Fatalf("%s, %s: %v", fileName, testCase.Function, err)
				}
			}
			if p, msg := panics(fn); p && msg != testCase.Trap {
				t.Errorf("%s, %s: unexpected trap message: got=%s, want=%s", fileName, fnString(testCase.Function, testCase.Args), msg, testCase.Trap)
			}
			continue
		}

		times := 1

		if ok {
			times = b.N
			b.ResetTimer()
		}

		var res interface{}
		var err error

		for i := 0; i < times; i++ {
			res, err = vm.ExecCode(false, int64(index), args...)
		}
		if ok {
			b.StopTimer()
		}

		if err != nil {
			t.Fatalf("%s, %s: %v", fileName, testCase.Function, err)
		}

		nanEq := false
		if reflect.TypeOf(res) == reflect.TypeOf(expected) {
			switch v := expected.(type) {
			case float32:
				nanEq = math.IsNaN(float64(v)) && math.IsNaN(float64(res.(float32)))
			case float64:
				nanEq = math.IsNaN(v) && math.IsNaN(res.(float64))
			}
		}
		if nanEq {
			continue
		}

		if !reflect.DeepEqual(res, expected) {
			t.Fatalf("%s, %s (%d): unexpected return value: got=%v(%v), want=%v(%v) (%s)", fileName, fnString(testCase.Function, testCase.Args), index, reflect.TypeOf(res), res, reflect.TypeOf(expected), expected, testCase.Return)
		}
	}
}

func testModules(t *testing.T, dir string) {
	files := []file{}
	file, err := os.Open(filepath.Join(dir, "modules.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&files)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		fileName := filepath.Join(dir, file.FileName)
		testCases := file.Tests
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(fileName)
			if err != nil {
				t.Fatal(err)
			}
			runTest(path, testCases, t)
		})
	}
}

func BenchmarkModules(b *testing.B) {
	files := []file{}
	file, err := os.Open(filepath.Join("test_data/spec", "modules.json"))
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&files)
	if err != nil {
		b.Fatal(err)
	}

	for _, file := range files {
		fileName := filepath.Join("test_data/spec", file.FileName)
		testCases := file.Tests
		b.Run(fileName, func(b *testing.B) {
			path, err := filepath.Abs(fileName)
			if err != nil {
				b.Fatal(err)
			}
			runTest(path, testCases, b)
		})
	}
}

func TestNonSpec(t *testing.T) {
	testModules(t, nonSpecTestsDir)
}

func TestSpec(t *testing.T) {
	testModules(t, specTestsDir)
}
