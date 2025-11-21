// Copyright 2025 Codnect
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tag

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CustomType struct {
	String string            `option:"string"`
	Int    int               `option:"int"`
	Float  float64           `option:"float"`
	Bool   bool              `option:"bool"`
	Slice  []any             `option:"slice"`
	Map    map[string]string `option:"map"`
}

type TestTag[T any] struct {
	Value  T              `option:"value"`
	String string         `option:"string"`
	Int    int            `option:"int"`
	Float  float64        `option:"float"`
	Bool   bool           `option:"bool"`
	Slice  []any          `option:"slice"`
	Map    map[string]any `option:"map"`

	StringSlice []string         `option:"stringSlice"`
	IntSlice    []int            `option:"intSlice"`
	FloatSlice  []float64        `option:"floatSlice"`
	MapSlice    []map[string]any `option:"mapSlice"`

	Struct           CustomType `option:"struct"`
	notExported      string     `option:"notExported"`
	OptionWithoutTag string
	InvalidFlag      string `option:"invalidFlag"`
}

func (t TestTag[T]) Tag() string {
	return "test"
}

func TestParse(t *testing.T) {
	testCases := []struct {
		name          string
		rawTags       string
		tagStruct     Tagger
		wantErr       error
		wantTagStruct Tagger
	}{
		{
			name:      "nil tag struct",
			rawTags:   ``,
			tagStruct: nil,
			wantErr:   nil,
		},
		{
			name:      "empty tags",
			rawTags:   ``,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
		},
		{
			name:      "non pointer tag struct",
			rawTags:   `test:"value"`,
			tagStruct: TestTag[any]{},
			wantErr:   errors.New("tag: target must be pointer to struct, got tag.TestTag[interface {}]"),
		},
		{
			name:      "blank tags",
			rawTags:   `   `,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
		},
		{
			name:      "missing tag name",
			rawTags:   `:"value"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: empty tag name in ':"value"'`),
		},
		{
			name:      "missing tag value",
			rawTags:   `test:`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: missing tag value for tag 'test'`),
		},
		{
			name:      "empty tag value",
			rawTags:   `test:""`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: empty tag value for tag 'test'`),
		},
		{
			name:      "missing positional option value",
			rawTags:   `test:",flag"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options ",flag": missing positional option value`),
		},
		{
			name:      "missing colon after tag name",
			rawTags:   `test"value"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: expected ':' after tag name "test" in "test\"value\""`),
		},
		{
			name:      "missing colon after tag name with space before value",
			rawTags:   `test "value"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: expected ':' after tag name "test" in "test \"value\""`),
		},
		{
			name:      "missing comma after positional option",
			rawTags:   `test:"any int=1141"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any int=1141": missing ',' after positional option value`),
		},
		{
			name:      "tag value with no quotes",
			rawTags:   `test:value`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: invalid tag value: must start with "`),
		},
		{
			name:      "tag value with missing ending quote",
			rawTags:   `test:"value`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: invalid tag value: must end with "`),
		},
		{
			name:      "missing option value after equal sign",
			rawTags:   `test:"any,int="`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any,int=": missing value for option "int"`),
		},
		{
			name:      "missing missing comma after option",
			rawTags:   `test:"any,int=1141 float=11.41"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any,int=1141 float=11.41": missing comma after option "int"`),
		},
		{
			name:      "not exported option",
			rawTags:   `test:"any,notExported='should not be set'"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
			},
		},
		{
			name:      "option without tag",
			rawTags:   `test:"any,optionWithoutTag='should not be set'"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
			},
		},
		{
			name:      "string positional option",
			rawTags:   `test:"any"`,
			tagStruct: &TestTag[string]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[string]{
				Value: "any",
			},
		},
		{
			name:      "string positional option with quotes",
			rawTags:   `test:"'any value'"`,
			tagStruct: &TestTag[string]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[string]{
				Value: "any value",
			},
		},
		{
			name:      "string slice positional option",
			rawTags:   `test:"['any','another']"`,
			tagStruct: &TestTag[[]string]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]string]{
				Value: []string{"any", "another"},
			},
		},
		{
			name:      "int8 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[int8]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int8]{
				Value: 11,
			},
		},
		{
			name:      "int8 slice positional option",
			rawTags:   `test:"[11,-41]"`,
			tagStruct: &TestTag[[]int8]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]int8]{
				Value: []int8{11, -41},
			},
		},
		{
			name:      "int16 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[int16]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int16]{
				Value: 11,
			},
		},
		{
			name:      "int16 slice positional option",
			rawTags:   `test:"[11,-41]"`,
			tagStruct: &TestTag[[]int16]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]int16]{
				Value: []int16{11, -41},
			},
		},
		{
			name:      "int32 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[int32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int32]{
				Value: 11,
			},
		},
		{
			name:      "int32 slice positional option",
			rawTags:   `test:"[11,-41]"`,
			tagStruct: &TestTag[[]int32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]int32]{
				Value: []int32{11, -41},
			},
		},
		{
			name:      "int64 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[int64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int64]{
				Value: 11,
			},
		},
		{
			name:      "int64 slice positional option",
			rawTags:   `test:"[11,-41]"`,
			tagStruct: &TestTag[[]int64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]int64]{
				Value: []int64{11, -41},
			},
		},
		{
			name:      "invalid int positional option",
			rawTags:   `test:"a1"`,
			tagStruct: &TestTag[int8]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "a1": invalid int value "a1"`),
		},
		{
			name:      "int positional option",
			rawTags:   `test:"1141"`,
			tagStruct: &TestTag[int]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int]{
				Value: 1141,
			},
		},
		{
			name:      "int slice positional option",
			rawTags:   `test:"[11,-41]"`,
			tagStruct: &TestTag[[]int]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]int]{
				Value: []int{11, -41},
			},
		},
		{
			name:      "int8 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[int8]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[int8]{
				Value: 11,
			},
		},
		{
			name:      "uint8 slice positional option",
			rawTags:   `test:"[11,41]"`,
			tagStruct: &TestTag[[]uint8]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]uint8]{
				Value: []uint8{11, 41},
			},
		},
		{
			name:      "uint16 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[uint16]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[uint16]{
				Value: 11,
			},
		},
		{
			name:      "uint16 slice positional option",
			rawTags:   `test:"[11,41]"`,
			tagStruct: &TestTag[[]uint16]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]uint16]{
				Value: []uint16{11, 41},
			},
		},
		{
			name:      "uint32 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[uint32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[uint32]{
				Value: 11,
			},
		},
		{
			name:      "uint32 slice positional option",
			rawTags:   `test:"[11,41]"`,
			tagStruct: &TestTag[[]uint32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]uint32]{
				Value: []uint32{11, 41},
			},
		},
		{
			name:      "uint64 positional option",
			rawTags:   `test:"11"`,
			tagStruct: &TestTag[uint64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[uint64]{
				Value: 11,
			},
		},
		{
			name:      "uint64 slice positional option",
			rawTags:   `test:"[11,41]"`,
			tagStruct: &TestTag[[]uint64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]uint64]{
				Value: []uint64{11, 41},
			},
		},
		{
			name:      "invalid uint positional option",
			rawTags:   `test:"a1141"`,
			tagStruct: &TestTag[uint]{},
			wantErr:   fmt.Errorf(`tag: failed to parse 'test' options "a1141": invalid unsigned int value "a1141"`),
		},
		{
			name:      "uint positional option",
			rawTags:   `test:"1141"`,
			tagStruct: &TestTag[uint]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[uint]{
				Value: 1141,
			},
		},
		{
			name:      "uint slice positional option",
			rawTags:   `test:"[11,41]"`,
			tagStruct: &TestTag[[]uint]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]uint]{
				Value: []uint{11, 41},
			},
		},
		{
			name:      "invalid float positional option",
			rawTags:   `test:"11.41a"`,
			tagStruct: &TestTag[float32]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "11.41a": invalid float value "11.41a"`),
		},
		{
			name:      "float32 positional option",
			rawTags:   `test:"11.41"`,
			tagStruct: &TestTag[float32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[float32]{
				Value: 11.41,
			},
		},
		{
			name:      "float32 slice positional option",
			rawTags:   `test:"[11.41,41.11]"`,
			tagStruct: &TestTag[[]float32]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]float32]{
				Value: []float32{11.41, 41.11},
			},
		},
		{
			name:      "float64 positional option",
			rawTags:   `test:"11.41"`,
			tagStruct: &TestTag[float64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[float64]{
				Value: 11.41,
			},
		},
		{
			name:      "float64 slice positional option",
			rawTags:   `test:"[11.41,41.11]"`,
			tagStruct: &TestTag[[]float64]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[]float64]{
				Value: []float64{11.41, 41.11},
			},
		},
		{
			name:      "invalid bool positional option",
			rawTags:   `test:"truea"`,
			tagStruct: &TestTag[bool]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "truea": invalid bool value "truea"`),
		},
		{
			name:      "bool(true) positional option",
			rawTags:   `test:"true"`,
			tagStruct: &TestTag[bool]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[bool]{
				Value: true,
			},
		},
		{
			name:      "bool(false) positional option",
			rawTags:   `test:"false"`,
			tagStruct: &TestTag[bool]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[bool]{
				Value: false,
			},
		},
		{
			name:      "map positional option",
			rawTags:   `test:"{'string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'nested':{'key':'val'}}"`,
			tagStruct: &TestTag[map[string]any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[map[string]any]{
				Value: map[string]any{
					"string": "value",
					"int":    int64(1141),
					"float":  11.41,
					"bool":   true,
					"slice":  []any{"a", "b"},
					"nested": map[string]any{"key": "val"},
				},
			},
		},
		{
			name:      "any positional option for string value",
			rawTags:   `test:"any"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
			},
		},
		{
			name:      "any positional option for string value with quotes",
			rawTags:   `test:"'any value'"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any value",
			},
		},
		{
			name:      "any positional option for int value",
			rawTags:   `test:"141"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: int64(141),
			},
		},
		{
			name:      "any positional option for unsigned int value",
			rawTags:   `test:"-141"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: int64(-141),
			},
		},
		{
			name:      "any positional option for bool(true) value",
			rawTags:   `test:"true"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: true,
			},
		},
		{
			name:      "any positional option for bool(false) value",
			rawTags:   `test:"false"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: false,
			},
		},
		{
			name:      "any positional option for float value",
			rawTags:   `test:"1.41"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: 1.41,
			},
		},
		{
			name:      "any positional option for unsigned float value",
			rawTags:   `test:"-1.41"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: -1.41,
			},
		},
		{
			name:      "any positional option for slice value",
			rawTags:   `test:"['any','another']"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: []any{"any", "another"},
			},
		},
		{
			name:      "any positional option for map value",
			rawTags:   `test:"{'string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'nested':{'key':'val'}}"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: map[string]any{
					"string": "value",
					"int":    int64(1141),
					"float":  11.41,
					"bool":   true,
					"slice":  []any{"a", "b"},
					"nested": map[string]any{"key": "val"},
				},
			},
		},
		{
			name:      "custom type positional option",
			rawTags:   `test:"{'string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'map':{'key':'val'}}"`,
			tagStruct: &TestTag[CustomType]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[CustomType]{
				Value: CustomType{
					String: "value",
					Int:    1141,
					Float:  11.41,
					Bool:   true,
					Slice:  []any{"a", "b"},
					Map:    map[string]string{"key": "val"},
				},
			},
		},
		{
			name:      "flag option",
			rawTags:   `test:"any,bool"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
				Bool:  true,
			},
		},
		{
			name:      "invalid slice positional option missing opening bracket",
			rawTags:   `test:"11,41]"`,
			tagStruct: &TestTag[[]int]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "11,41]": expected [ at the beginning of slice value '11'`),
		},
		{
			name:      "invalid slice positional option missing closing bracket",
			rawTags:   `test:"[11,41"`,
			tagStruct: &TestTag[[]int]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "[11,41": unterminated slice, missing closing ']'`),
		},
		{
			name:      "invalid map positional option missing opening brace",
			rawTags:   `test:"'key':'val'}"`,
			tagStruct: &TestTag[map[string]any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "'key':'val'}": expected { at the beginning of map value 'key'`),
		},
		{
			name:      "invalid map positional option missing closing brace",
			rawTags:   `test:"{'key':'val'"`,
			tagStruct: &TestTag[map[string]any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "{'key':'val'": unterminated map, missing closing '}'`),
		},
		{
			name:      "full options",
			rawTags:   `test:"any,int=1141,float=11.41,bool=true,slice=['a','b'],map={'key':'val'},stringSlice=['str1','str2'],intSlice=[11,22],floatSlice=[1.1,2.2],mapSlice=[{'key1':'val1'},{'key2':'val2'}]"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value:       "any",
				Int:         1141,
				Float:       11.41,
				Bool:        true,
				Slice:       []any{"a", "b"},
				Map:         map[string]any{"key": "val"},
				StringSlice: []string{"str1", "str2"},
				IntSlice:    []int{11, 22},
				FloatSlice:  []float64{1.1, 2.2},
				MapSlice: []map[string]any{
					{"key1": "val1"},
					{"key2": "val2"},
				},
			},
		},
		{
			name:      "slice option with missing opening bracket",
			rawTags:   `test:"any,slice=['a','b'"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any,slice=['a','b'": unterminated slice, missing closing ']'`),
		},
		{
			name:      "nested slice positional option",
			rawTags:   `test:"[['a','b'],['c','d']]"`,
			tagStruct: &TestTag[[][]string]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[[][]string]{
				Value: [][]string{
					{"a", "b"},
					{"c", "d"},
				},
			},
		},
		{
			name:      "unknown option flag",
			rawTags:   `test:"any,unknownFlag"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
			},
		},
		{
			name:      "invalid option flag",
			rawTags:   `test:"any,invalidFlag"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any,invalidFlag": invalid type for flag option "invalidFlag": want bool`),
		},
		{
			name:      "struct option",
			rawTags:   `test:"any,struct={'string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'map':{'key':'val'}}"`,
			tagStruct: &TestTag[any]{},
			wantErr:   nil,
			wantTagStruct: &TestTag[any]{
				Value: "any",
				Struct: CustomType{
					String: "value",
					Int:    1141,
					Float:  11.41,
					Bool:   true,
					Slice:  []any{"a", "b"},
					Map:    map[string]string{"key": "val"},
				},
			},
		},
		{
			name:      "invalid struct option",
			rawTags:   `test:"any,struct='string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'map':{'key':'val'}}"`,
			tagStruct: &TestTag[any]{},
			wantErr:   errors.New(`tag: failed to parse 'test' options "any,struct='string':'value','int':1141,'float':11.41,'bool':true,'slice':['a','b'],'map':{'key':'val'}}": failed to set option "struct": expected {, but got 'string'`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Parse(tc.rawTags, tc.tagStruct)
			if tc.wantErr != nil {
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				return
			}

			assert.NoError(t, err)
			if tc.wantTagStruct != nil {
				assert.Equal(t, tc.wantTagStruct, tc.tagStruct)
			}
		})
	}

}
