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

		if i == 0 || i+1 >= len(raw) || raw[i] != ':' {
			return fmt.Errorf("tag: invalid tag format, expect ':' but got %c", raw[i])
		}

		tagName := raw[:i]
		raw = raw[i+1:]

		tagValue, consumed, err := scanString(raw, '"')
		if err != nil {
			return fmt.Errorf("tag: invalid tag value: %s", err)
		}

		if target.Tag() == tagName {
			v := reflect.ValueOf(target)
			if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
				return fmt.Errorf("tag: target must be pointer to struct, got %T", target)
			}

			if err = parseOptions(tagValue, v); err != nil {
				return fmt.Errorf("tag: failed to parse to options: %s", err)
			}

			return nil
		}

		raw = raw[consumed:]
	}

	return nil
}

// parseOptions parses the options and sets the struct fields accordingly.
func parseOptions(options string, structPtr reflect.Value) error {
	optionIndex := 0

	for options != "" {
		skip := skipSpaces(options)
		options = options[skip:]
		if len(options) == 0 {
			break
		}

		var (
			name string
			cons int
		)

		if optionIndex == 0 && options[0] == '\'' {
			item, length, err := scanString(options, '\'')
			if err != nil {
				return err
			}
			if length == 0 {
				return errors.New("tag: empty quoted positional option")
			}

			name = item
			cons = length

			rest := options[cons:]
			restSkip := skipSpaces(rest)
			options = rest[restSkip:]
		} else {
			i := 0
			for i < len(options) && options[i] > ' ' && options[i] != '=' && options[i] != ',' && options[i] != 0x7f {
				i++
			}

			if i == 0 {
				ch := byte('?')
				if len(options) > 0 {
					ch = options[0]
				}
				return fmt.Errorf("tag: invalid option format, expected flag or key, got %q", ch)
			}

			name = options[:i]
			rest := options[i:]
			restSkip := skipSpaces(rest)
			options = rest[restSkip:]
		}

		if len(options) == 0 {
			if err := setFlagOption(optionIndex, name, structPtr); err != nil {
				return err
			}

			break
		}

		switch options[0] {
		case ',':
			if err := setFlagOption(optionIndex, name, structPtr); err != nil {
				return err
			}
			options = options[1:]
			optionIndex++
			continue

		case '=':
			options = options[1:]
			skip = skipSpaces(options)
			options = options[skip:]

			fieldVal, fieldExists := findStructField(name, structPtr)

			var valueType reflect.Type
			if fieldExists {
				valueType = fieldVal.Type()
			} else {
				inferred, err := inferType(options)
				if err != nil {
					return err
				}
				valueType = inferred
			}

			rawValue, length, err := scanValue(valueType, options)
			if err != nil {
				return err
			}

			if fieldExists {
				if err = setValue(fieldVal, rawValue); err != nil {
					return err
				}
			}

			options = options[length:]
			skip = skipSpaces(options)
			options = options[skip:]

			if len(options) == 0 {
				break
			}

			if options[0] != ',' {
				return fmt.Errorf("tag: expected ',' after option %q", name)
			}

			options = options[1:]
			optionIndex++
			continue

		default:
			return fmt.Errorf("tag: expected '=' or ',' after option %q", name)
		}
	}

	return nil
}

// setFlagOption sets a boolean flag option or positional value in the struct.
func setFlagOption(index int, name string, structPtr reflect.Value) error {
	fieldName := name
	if index == 0 {
		fieldName = DefaultOptionName
	}

	fieldVal, exists := findStructField(fieldName, structPtr)
	if !exists {
		return nil
	}

	if fieldName == DefaultOptionName {
		return setValue(fieldVal, name)
	}

	if fieldVal.Kind() != reflect.Bool {
		return fmt.Errorf("tag: invalid flag option %q: target field is %s, want bool", fieldName, fieldVal.Kind())
	}

	return setValue(fieldVal, "true")
}

// findStructField finds the struct field corresponding to the option name.
func findStructField(optionName string, structPtr reflect.Value) (reflect.Value, bool) {
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

		if optionTag != optionName {
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
