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

// skipSpaces skips leading spaces and tabs in the given string and
// returns the number of skipped characters.
func skipSpaces(val string) int {
	if len(val) == 0 {
		return 0
	}

	i := 0
	for i < len(val) && (val[i] == ' ' || val[i] == '\t') {
		i++
	}

	return i
}

// setValue sets the reflect.Value from the given string.
func setValue(rVal reflect.Value, val string) error {
	switch rVal.Kind() {
	case reflect.Slice:
		return setSliceFromLiteral(rVal, val)
	case reflect.Map:
		if rVal.IsNil() {
			rVal.Set(reflect.MakeMap(rVal.Type()))
		}
		return setMapFromLiteral(rVal, val)
	case reflect.Struct:
		return setStructFromLiteral(rVal.Addr(), val)
	case reflect.Interface:
		inferredType, err := inferType(val)
		if err != nil {
			return err
		}

		out := reflect.New(inferredType).Elem()
		if err = setValue(out, val); err != nil {
			return err
		}

		rVal.Set(out)
		return nil
	default:
		return setScalarVal(rVal, val)
	}
}

// setScalarVal sets the scalar value from the given string.
func setScalarVal(rVal reflect.Value, scalarVal string) error {
	switch rVal.Kind() {
	case reflect.String:
		rVal.SetString(scalarVal)
	case reflect.Bool:
		b, err := strconv.ParseBool(scalarVal)
		if err != nil {
			return fmt.Errorf("invalid bool value %q: %w", scalarVal, err)
		}
		rVal.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(scalarVal, 10, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid int value %q: %w", scalarVal, err)
		}
		rVal.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(scalarVal, 10, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid uint value %q: %w", scalarVal, err)
		}
		rVal.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(scalarVal, rVal.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid float value %q: %w", scalarVal, err)
		}
		rVal.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %s", rVal.Type().String())
	}

	return nil
}

// setSliceFromLiteral sets the slice elements from the given literal string.
func setSliceFromLiteral(rSliceVal reflect.Value, raw string) error {
	n := skipSpaces(raw)

	raw = raw[n:]
	if raw == "" {
		return fmt.Errorf("invalid slice value '%s'", raw)
	}

	if raw[0] != '[' {
		return fmt.Errorf("expected [ at the beginning of slice value '%s'", raw)
	}

	raw = raw[1:]
	elemType := rSliceVal.Type().Elem()

	for raw != "" {
		n = skipSpaces(raw)
		raw = raw[n:]

		if len(raw) == 0 {
			return fmt.Errorf("unterminated slice value, missing closing ']'")
		}

		if len(raw) == 1 {
			if raw[0] != ']' {
				return fmt.Errorf("unterminated slice value, missing closing ']'")
			}

			break
		}

		typ := elemType
		if typ.Kind() == reflect.Interface {
			inferTyp, err := inferType(raw)
			if err != nil {
				return err
			}
			typ = inferTyp
		}

		item, l, scanErr := scanValue(typ, raw)
		if scanErr != nil {
			return scanErr
		}

		elem := reflect.New(typ).Elem()
		if err := setValue(elem, item); err != nil {
			return err
		}

		rSliceVal.Set(reflect.Append(rSliceVal, elem))

		raw = raw[l:]
		n = skipSpaces(raw)
		raw = raw[n:]

		if raw == "" {
			return fmt.Errorf("unterminated slice value '%s'", raw)
		}

		if raw[0] == ',' {
			raw = raw[1:]
			continue
		}

	}

	return nil
}

// setMapFromLiteral sets the map entries from the given literal string.
func setMapFromLiteral(rMapVal reflect.Value, raw string) error {
	n := skipSpaces(raw)

	raw = raw[n:]
	if raw == "" {
		return fmt.Errorf("invalid map value '%s'", raw)
	}

	if raw[0] != '{' {
		return fmt.Errorf("expected { at the beginning of map value '%s'", raw)
	}

	raw = raw[1:]

	keyType := rMapVal.Type().Key()
	elemType := rMapVal.Type().Elem()

	for raw != "" {
		n = skipSpaces(raw)
		raw = raw[n:]

		if len(raw) == 0 {
			return fmt.Errorf("unterminated map value, missing closing '}'")
		}

		if len(raw) == 1 {
			if raw[0] != '}' {
				return fmt.Errorf("unterminated map value, missing closing '}'")
			}

			break
		}

		rKeyVal := reflect.New(keyType).Elem()
		keyStr, l, scanErr := scanValue(keyType, raw)
		if scanErr != nil {
			return scanErr
		}

		if err := setValue(rKeyVal, keyStr); err != nil {
			return err
		}

		raw = raw[l:]
		n = skipSpaces(raw)
		raw = raw[n:]
		if raw == "" || raw[0] != ':' {
			return errors.New("invalid map item format, missing ':' separator")
		}

		raw = raw[1:]
		n = skipSpaces(raw)
		raw = raw[n:]

		et := elemType
		if et.Kind() == reflect.Interface {
			inferTyp, err := inferType(raw)
			if err != nil {
				return err
			}
			et = inferTyp
		}

		valStr, vLen, err := scanValue(et, raw)
		if err != nil {
			return err
		}

		rElemVal := reflect.New(et).Elem()
		if err = setValue(rElemVal, valStr); err != nil {
			return err
		}

		rMapVal.SetMapIndex(rKeyVal, rElemVal)

		raw = raw[vLen:]
		n = skipSpaces(raw)
		raw = raw[n:]

		if raw == "" {
			return fmt.Errorf("unterminated map value '%s'", raw)
		}

		if raw[0] == ',' {
			raw = raw[1:]
			continue
		}
	}

	return nil
}

// setStructFromLiteral sets the struct fields from the given literal string.
func setStructFromLiteral(rStructVal reflect.Value, raw string) error {
	n := skipSpaces(raw)

	raw = raw[n:]
	if raw == "" {
		return fmt.Errorf("invalid struct value '%s'", raw)
	}

	if raw[0] != '{' {
		return fmt.Errorf("expected {, but got '%s'", raw)
	}

	raw = raw[1:]

	keyType := reflect.TypeFor[string]()

	for raw != "" {
		n = skipSpaces(raw)
		raw = raw[n:]

		if len(raw) == 0 {
			return fmt.Errorf("unterminated struct value, missing closing '}'")
		}

		if len(raw) == 1 {
			if raw[0] != '}' {
				return fmt.Errorf("unterminated struct value, missing closing '}'")
			}

			break
		}

		prop, l, scanErr := scanValue(keyType, raw)
		if scanErr != nil {
			return scanErr
		}

		raw = raw[l:]
		n = skipSpaces(raw)
		raw = raw[n:]
		if raw == "" || raw[0] != ':' {
			return errors.New("invalid struct item format, missing ':' separator")
		}

		raw = raw[1:]
		n = skipSpaces(raw)
		raw = raw[n:]

		rFieldVal, fieldExists := findStructField(prop, rStructVal)

		var valTyp reflect.Type
		if fieldExists {
			valTyp = rFieldVal.Type()
		} else {
			inferTyp, err := inferType(raw)
			if err != nil {
				return err
			}

			valTyp = inferTyp
		}

		value, vLen, err := scanValue(valTyp, raw)
		if err != nil {
			return err
		}

		if fieldExists {
			if err = setValue(rFieldVal, value); err != nil {
				return err
			}
		}

		raw = raw[vLen:]
		n = skipSpaces(raw)
		raw = raw[n:]

		if raw == "" {
			return fmt.Errorf("unterminated struct value '%s'", raw)
		}

		if raw[0] == ',' {
			raw = raw[1:]
			continue
		}
	}

	return nil
}

// inferType infers the type of the given value string.
func inferType(val string) (reflect.Type, error) {
	i := 0
	for i < len(val) && (val[i] == ' ' || val[i] == '\t') {
		i++
	}

	val = val[i:]
	if val == "" {
		return nil, errors.New("cannot infer type from empty value")
	}

	switch val[0] {
	case '\'':
		return reflect.TypeFor[string](), nil
	case '[':
		return reflect.TypeFor[[]any](), nil
	case '{':
		return reflect.TypeFor[map[string]any](), nil
	}

	end := 0
	for end < len(val) {
		c := val[end]
		if c == ' ' || c == '\t' || c == ';' || c == ']' || c == '}' || c == ':' || c == ',' {
			break
		}
		end++
	}
	token := val[:end]

	if token == "true" || token == "false" {
		return reflect.TypeFor[bool](), nil
	}

	if _, err := strconv.ParseInt(token, 10, 64); err == nil {
		return reflect.TypeFor[int64](), nil
	}

	if _, err := strconv.ParseFloat(token, 64); err == nil {
		return reflect.TypeFor[float64](), nil
	}

	return nil, fmt.Errorf("cannot infer type from value %q", token)
}

// scanValue scans a value of the given type from the string.
func scanValue(typ reflect.Type, val string) (string, int, error) {
	switch typ.Kind() {
	case reflect.String:
		return scanString(val, '\'')
	case reflect.Bool:
		return scanBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return scanInteger(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return scanInteger(val)
	case reflect.Float32, reflect.Float64:
		return scanNumber(val)
	case reflect.Slice:
		return scanSlice(val)
	case reflect.Map, reflect.Struct:
		return scanMap(val)
	case reflect.Interface:
		inferredType, err := inferType(val)
		if err != nil {
			return "", -1, err
		}

		return scanValue(inferredType, val)
	default:
		return "", -1, fmt.Errorf("unsupported type '%s' for scanning value", typ.String())
	}
}

// scanString scans a string literal from the given string.
func scanString(val string, quote byte) (string, int, error) {
	if len(val) == 0 || val[0] != quote {
		return "", -1, fmt.Errorf("string must start with %c", quote)
	}

	i := 1
	for i < len(val) && val[i] != quote {
		if val[i] == '\\' {
			i++
		}
		i++
	}

	if i >= len(val) {
		return "", -1, fmt.Errorf("unterminated string, missing closing %c", quote)
	}

	item := val[1:i]
	return item, i + 1, nil
}

// scanBool scans a boolean literal from the given string.
func scanBool(val string) (string, int, error) {
	i := 0
	for i < len(val) {
		c := val[i]
		if c == ' ' || c == '\t' || c == ';' || c == ']' || c == '}' || c == ':' || c == ',' {
			break
		}
		i++
	}
	if i == 0 {
		return "", -1, errors.New("invalid bool value")
	}
	item := val[:i]
	if item != "true" && item != "false" {
		return "", -1, fmt.Errorf("invalid bool value %q", item)
	}
	return item, i, nil
}

// scanInteger scans an integer literal from the given string.
func scanInteger(val string) (string, int, error) {
	i := 0
	for i < len(val) && ((val[i] >= '0' && val[i] <= '9') || val[i] == '-') {
		i++
	}

	if i == 0 {
		return "", -1, errors.New("invalid integer value")
	}

	item := val[:i]
	return item, i, nil
}

// scanNumber scans a float literal from the given string.
func scanNumber(val string) (string, int, error) {
	i := 0
	for i < len(val) && ((val[i] >= '0' && val[i] <= '9') || val[i] == '-' || val[i] == '.') {
		i++
	}

	if i == 0 {
		return "", -1, errors.New("invalid float value")
	}

	item := val[:i]
	return item, i, nil
}

// scanSlice scans a slice literal from the given string.
func scanSlice(val string) (string, int, error) {
	if len(val) == 0 || val[0] != '[' {
		return "", -1, errors.New("slice must start with '['")
	}

	i := 1
	bracketCount := 1
	for i < len(val) && bracketCount > 0 {
		if val[i] == '[' {
			bracketCount++
		} else if val[i] == ']' {
			bracketCount--
		}
		i++
	}

	if bracketCount != 0 {
		return "", -1, errors.New("unterminated slice, missing closing ']'")
	}

	item := val[:i]
	return item, i, nil

}

// scanMap scans a map literal from the given string.
func scanMap(val string) (string, int, error) {
	if len(val) == 0 || val[0] != '{' {
		return "", -1, errors.New("map must start with '{'")
	}

	i := 1
	bracketCount := 1
	for i < len(val) && bracketCount > 0 {
		if val[i] == '{' {
			bracketCount++
		} else if val[i] == '}' {
			bracketCount--
		}
		i++
	}

	if bracketCount != 0 {
		return "", -1, errors.New("unterminated map, missing closing '}'")
	}

	item := val[:i]
	return item, i, nil
}
