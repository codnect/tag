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
	"fmt"
	"reflect"
	"strings"
)

const DefaultOptionName = "value"

// Tagger represents a struct tag.
type Tagger interface {
	// Tag returns the name of the tag
	Tag() string
}

// Parse parses the struct tag string and populates the fields.
func Parse[T Tagger](raw string, target T) error {
	if raw == "" {
		return nil
	}

	rTargetVal := reflect.ValueOf(target)
	if rTargetVal.Kind() != reflect.Ptr || rTargetVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("tag: target must be pointer to struct, got %T", target)
	}

	for raw != "" {
		skip := skipSpaces(raw)
		raw = raw[skip:]

		if raw == "" {
			break
		}

		i := 0
		for i < len(raw) && raw[i] > ' ' && raw[i] != ':' && raw[i] != '"' && raw[i] != 0x7f {
			i++
		}

		if i == 0 {
			return fmt.Errorf("tag: empty tag name in '%s'", raw)
		}

		if i+1 >= len(raw) {
			return fmt.Errorf("tag: missing tag value for tag '%s'", raw[:i])
		}

		if raw[i] != ':' {
			return fmt.Errorf("tag: expected ':' after tag name %q in %q", raw[:i], raw)
		}

		name := raw[:i]
		raw = raw[i+1:]

		value, valueLen, err := scanString(raw, '"')
		if err != nil {
			return fmt.Errorf("tag: invalid tag value: %s", err)
		}

		if target.Tag() == name {
			skip = skipSpaces(value)
			opts := value[skip:]
			if len(opts) == 0 {
				return fmt.Errorf("tag: empty tag value for tag '%s'", name)
			}

			if err = parseOptions(value, rTargetVal); err != nil {
				return fmt.Errorf("tag: failed to parse '%s' options %q: %w", name, value, err)
			}

			return nil
		}

		raw = raw[valueLen:]
	}

	return nil
}

// parseOptions parses the options and sets the struct fields accordingly.
func parseOptions(opts string, structPtr reflect.Value) error {
	val, valLen, err := scanPosOptValue(opts)
	if err != nil {
		return err
	}

	if valLen <= 0 {
		return fmt.Errorf("missing positional option value")
	}

	rFieldVal, exists := findStructField(DefaultOptionName, structPtr)
	if exists {
		if err = setValue(rFieldVal, val); err != nil {
			return err
		}
	}

	opts = opts[valLen:]
	skip := skipSpaces(opts)
	opts = opts[skip:]

	if len(opts) == 0 {
		return nil
	}

	if opts[0] != ',' {
		return fmt.Errorf("missing ',' after positional option value")
	} else {
		opts = opts[1:]
	}

	for len(opts) != 0 {
		val, valLen, err = scanIdentifier(opts)
		if err != nil {
			return err
		}

		opt := val
		opts = opts[valLen:]
		skip = skipSpaces(opts)
		opts = opts[skip:]

		if len(opts) >= 1 && opts[0] == '=' {
			opts = opts[1:]
			skip = skipSpaces(opts)
			opts = opts[skip:]

			if len(opts) == 0 {
				return fmt.Errorf("missing value for option %q", opt)
			}

			val, valLen, err = scanOptValue(opts)
			if err != nil {
				return err
			}

			if valLen <= 0 {
				return fmt.Errorf("missing value for option %q", opt)
			}

			if err = setOption(opt, val, structPtr); err != nil {
				return err
			}

			opts = opts[valLen:]
		} else if len(opts) == 0 {
			if err = setFlagOption(opt, structPtr); err != nil {
				return err
			}

			break
		}

		skip = skipSpaces(opts)
		opts = opts[skip:]

		if len(opts) == 0 {
			break
		}

		if opts[0] != ',' {
			return fmt.Errorf("missing comma after option %q", opt)
		} else {
			opts = opts[1:]
		}
	}

	return nil
}

// findStructField finds the struct field corresponding to the option name.
func findStructField(optName string, structPtr reflect.Value) (reflect.Value, bool) {
	optName = strings.ToLower(optName)
	structType := structPtr.Type().Elem()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		optionTag, ok := field.Tag.Lookup("option")
		if !ok {
			continue
		}

		optionTag = strings.ToLower(optionTag)
		if optionTag != optName {
			continue
		}

		fieldVal := structPtr.Elem().Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		return fieldVal, true
	}

	return reflect.Value{}, false
}

// setFlagOption sets a boolean flag option in the struct.
func setFlagOption(name string, structPtr reflect.Value) error {
	rFieldVal, exists := findStructField(name, structPtr)
	if !exists {
		return nil
	}

	if rFieldVal.Kind() != reflect.Bool {
		return fmt.Errorf("invalid type for flag option %q: want bool", name)
	}

	if err := setValue(rFieldVal, "true"); err != nil {
		return fmt.Errorf("failed to set flag option %q: %w", name, err)
	}

	return nil
}

// setOption sets an option with a value in the struct.
func setOption(name string, value string, structPtr reflect.Value) error {
	rFieldVal, exists := findStructField(name, structPtr)
	if !exists {
		return nil
	}

	if err := setValue(rFieldVal, value); err != nil {
		return fmt.Errorf("failed to set option %q: %w", name, err)
	}

	return nil
}
