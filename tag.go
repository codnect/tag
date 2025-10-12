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
	"reflect"
	"strconv"
)

// Tag represents a struct tag.
type Tag interface {
	// TagName returns the name of the tag
	TagName() string
}

// ParseFor parses the struct tag string and populates the fields.
func ParseFor[T Tag](tags string) (*T, error) {
	rVal := reflect.New(reflect.TypeFor[T]())
	tag := rVal.Interface().(Tag)

	for tags != "" {
		i := 0
		for i < len(tags) && tags[i] == ' ' {
			i++
		}

		tags = tags[i:]
		if tags == "" {
			return nil, errors.New("no tag found")
		}

		i = 0
		for i < len(tags) && tags[i] > ' ' && tags[i] != ':' && tags[i] != '"' && tags[i] != 0x7f {
			i++
		}

		if i == 0 || i+1 >= len(tags) || tags[i] != ':' || tags[i+1] != '"' {
			break
		}

		name := tags[:i]
		tags = tags[i+1:]

		i = 1
		for i < len(tags) && tags[i] != '"' {
			if tags[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(tags) {
			break
		}

		quotedOpts := tags[:i+1]
		tags = tags[i+1:]

		if tag.TagName() == name {
			opts, err := strconv.Unquote(quotedOpts)
			if err != nil {
				return nil, err
			}

			err = parseOpts(opts, rVal)
			if err != nil {
				return nil, err
			}

			return rVal.Interface().(*T), nil
		}
	}

	return nil, errors.New("tag not found")
}

// parseOpts parses the option string and sets the corresponding fields in the struct.
func parseOpts(opts string, rVal reflect.Value) error {
	optIdx := 0

	for opts != "" {
		i := 0
		for i < len(opts) && opts[i] == ' ' {
			i++
		}

		opts = opts[i:]
		if opts == "" {
			break
		}

		i = 0
		for i < len(opts) && opts[i] > ' ' && opts[i] != '=' && opts[i] != ',' && opts[i] != 0x7f {
			i++
		}

		if i == 0 && i+1 < len(opts) {
			if opts[i] == ',' {
				opts = opts[i+1:]
				optIdx++
				continue
			} else if opts[i] != '=' {
				return errors.New("invalid option syntax")
			}
		}

		if i == 0 || i+1 >= len(opts) {
			break
		}

		key := opts[:i]
		nextCh := opts[i]
		opts = opts[i+1:]

		if nextCh == ',' {
			err := setOptWithoutValue(rVal, key, optIdx)
			if err != nil {
				return err
			}
			optIdx++
			continue
		} else if nextCh != '=' {
			return errors.New("invalid option syntax")
		}

		i = 1
		for i < len(opts) && opts[i] != ',' {
			if opts[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(opts) {
			err := setOpt(rVal, key, opts)
			if err != nil {
				return err
			}
			optIdx++
			break
		}

		val := opts[:i]
		opts = opts[i+1:]

		err := setOpt(rVal, key, val)
		if err != nil {
			return err
		}
		optIdx++
	}

	return nil
}

// setOptWithoutValue sets a boolean option.
func setOptWithoutValue(structVal reflect.Value, fieldName string, optIdx int) error {
	if optIdx == 0 {
		fieldVal, found := findField(structVal, "-")

		if found {
			return setValue(fieldVal, fieldName)
		}
	}

	fieldVal, found := findField(structVal, fieldName)
	if !found {
		return nil
	}

	if fieldVal.Kind() != reflect.Bool {
		return errors.New("option without value can only be set to bool fields")
	}

	return setValue(fieldVal, "true")
}

// setOpt sets the option value to the corresponding struct field.
func setOpt(structVal reflect.Value, name, val string) error {
	fieldVal, found := findField(structVal, name)
	if !found {
		return nil
	}

	return setValue(fieldVal, val)
}

// findField finds the struct field with the given option name.
func findField(structVal reflect.Value, name string) (reflect.Value, bool) {
	structTyp := structVal.Type().Elem()

	for i := 0; i < structTyp.NumField(); i++ {
		field := structTyp.Field(i)
		if !field.IsExported() {
			continue
		}

		optTag, ok := field.Tag.Lookup("option")
		if !ok {
			continue
		}

		if optTag != name {
			continue
		}

		fieldVal := structVal.Elem().Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		return fieldVal, true
	}

	return reflect.Value{}, false
}

// setValue sets the value to the reflect.Value based on its kind.
func setValue(rVal reflect.Value, val string) error {
	switch rVal.Kind() {
	case reflect.Slice:
		return setSliceVal(rVal, val)
	case reflect.Map:
		return setMapVal(rVal, val)
	case reflect.Interface:
		l, r := 0, len(val)
		for l < r && val[l] == ' ' {
			l++
		}
		for r > l && val[r-1] == ' ' {
			r--
		}

		val = val[l:r]

		if len(val) >= 2 && val[0] == '{' && val[len(val)-1] == '}' {
			mVal := reflect.MakeMap(reflect.TypeOf(map[string]any{}))
			rVal.Set(mVal)
			return setMapVal(mVal, val)
		} else if len(val) >= 2 && val[0] == '[' && val[len(val)-1] == ']' {
			sVal := reflect.MakeSlice(reflect.TypeOf([]any{}), 0, 0)
			rVal.Set(sVal)
			return setSliceVal(sVal, val)
		}

		rVal.Set(reflect.ValueOf(val))
		return nil
	default:
		return setScalarVal(rVal, val)
	}
}

// setSliceVal parses and sets the slice value from the string.
func setSliceVal(rVal reflect.Value, val string) error {
	lb, rb := 0, len(val)
	for lb < rb && val[lb] == ' ' {
		lb++
	}
	for rb > lb && val[rb-1] == ' ' {
		rb--
	}

	if rb-lb < 2 {
		return fmt.Errorf("missing curly braces")
	} else if val[lb] != '[' {
		return fmt.Errorf("missing opening square brace")
	} else if val[rb-1] != ']' {
		return fmt.Errorf("missing closing square brace")
	}

	val = val[lb+1 : rb-1]

	out := reflect.MakeSlice(rVal.Type(), 0, 0)
	elemType := rVal.Type().Elem()

	i := 0
	for {
		start := i

		for i < len(val) && val[i] != ';' {
			i++
		}

		l, r := start, i
		for l < r && val[l] == ' ' {
			l++
		}

		for r > l && val[r-1] == ' ' {
			r--
		}

		if r > l {
			part := val[l:r]

			ev := reflect.New(elemType).Elem()

			if err := setValue(ev, part); err != nil {
				return err
			}

			out = reflect.Append(out, ev)
		}

		if i >= len(val) {
			break
		}

		i++
	}

	rVal.Set(out)
	return nil
}

// setMapVal parses and sets the map value from the string.
func setMapVal(rVal reflect.Value, val string) error {
	mapType := rVal.Type()
	keyType := mapType.Key()
	elemType := mapType.Elem()

	if rVal.IsNil() {
		rVal.Set(reflect.MakeMap(mapType))
	}

	lb, rb := 0, len(val)
	for lb < rb && val[lb] == ' ' {
		lb++
	}
	for rb > lb && val[rb-1] == ' ' {
		rb--
	}

	if rb-lb < 2 {
		return fmt.Errorf("missing curly braces")
	} else if val[lb] != '{' {
		return fmt.Errorf("missing opening curly brace")
	} else if val[rb-1] != '}' {
		return fmt.Errorf("missing closing curly brace")
	}

	val = val[lb+1 : rb-1]

	var (
		mKey string
		mVal string
	)

	i := 0
	for {
		start := i

		if val[i] == '{' || val[i] == '}' {
			return fmt.Errorf("nested maps are not supported")
		}

		if val[i] == '[' || val[i] == ']' {
			return fmt.Errorf("slices are not supported in map")
		}

		for i < len(val) && val[i] != ';' && val[i] != ':' {
			i++
		}

		l, r := start, i
		for l < r && val[l] == ' ' {
			l++
		}

		for r > l && val[r-1] == ' ' {
			r--
		}

		if r > l {
			part := val[l:r]

			if mKey == "" {
				mKey = part
			} else if mVal == "" {
				mVal = part
			}

			if mKey != "" && mVal != "" {
				kv := reflect.New(keyType).Elem()
				vv := reflect.New(elemType).Elem()

				if err := setValue(kv, mKey); err != nil {
					return fmt.Errorf("%v: %v", mKey, err)
				}

				if err := setValue(vv, mVal); err != nil {
					return fmt.Errorf("%v: %v", mKey, err)
				}

				rVal.SetMapIndex(kv, vv)

				mKey = ""
				mVal = ""

				if i >= len(val) {
					break
				}

				if val[i] == ';' {
					i++
				} else {
					return fmt.Errorf("missing semicolon between key-value pairs")
				}
			}

			if val[i] == ':' || val[i] == ' ' {
				i++
			} else {
				return fmt.Errorf("missing colon between key and value")
			}

		}

	}

	return nil
}

// setScalarVal sets the scalar value to the reflect.Value based on its kind.
func setScalarVal(rVal reflect.Value, val string) error {
	switch rVal.Kind() {
	case reflect.String:
		rVal.SetString(val)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", val, err)
		}
		rVal.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", val, err)
		}
		rVal.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint value %q: %w", val, err)
		}
		rVal.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", val, err)
		}
		rVal.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %s", rVal.Type().String())
	}

	return nil

}
