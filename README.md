# Tag — Declarative Struct Tag Parser for Go

[![Run Tests](https://github.com/codnect/tag/actions/workflows/tag.yaml/badge.svg?branch=main)](https://github.com/codnect/tag/actions/workflows/tag.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=codnect_tag&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=codnect_tag)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=codnect_tag&metric=coverage)](https://sonarcloud.io/summary/new_code?id=codnect_tag)
[![Go Report Card](https://goreportcard.com/badge/codnect.io/tag)](https://goreportcard.com/report/codnect.io/tag)
[![Go Reference](https://pkg.go.dev/badge/codnect.io/tag?status.svg)](https://pkg.go.dev/codnect.io/tag?tab=doc)

**Tag** is a lightweight, zero-dependency library that transforms raw Go struct tag
strings into strongly typed, well-structured Go values.

It eliminates repetitive parsing logic, reduces boilerplate, and provides a clean,
declarative way to define your tag schema — while keeping performance predictable
and implementation overhead minimal.

## 🚀 Installation

To install Tag, run the following command:

```bash
go get github.com/codnect/tag
```

## 🌟 Overview

Go’s `reflect.StructTag` exposes tag values only as raw strings:

```go
prop:"'database.host',default=5432"
```

In most projects, this leads to:

* Manual string splitting
* Repetitive parsing logic
* Custom flag handling
* Default values
* Lists, maps, typed values
* Edge cases everywhere

Tag centralizes this logic into a single, safe, strongly typed mechanism.

You define a struct describing your tag. Tag handles parsing, type inference, nested structures, and value assignment.

## ⚡ Quick Start

Define the structure of your tag and implement the **Tagger** interface:

```go
type PropTag struct {
    Key      string `option:"value"`
    Optional bool   `option:"optional"`
    Default  int    `option:"default"`
}

// Tag identifies the tag name.
func (t PropTag) Tag() string {
    return "prop"
}
```

Parse a raw tag string:

```go
propTag := &PropTag{}
err := tag.Parse(`prop:"'database.host',default=5432"`, propTag)
if err != nil {
    // handle error
}
```

## ✨ Tag Syntax Overview

A tag expression is composed of a primary value followed by zero or more options.
Tag recognizes three option forms:
1.	Primary Value — always the first item
2.	Boolean Flags — option names without a value (e.g., optional)
3.	Key/Value Options — options in the form name=value

This structure applies uniformly to all tags and defines how each parsed item
is mapped onto the fields of your tag struct.

### Primary Value

The first item of a tag expression is always treated as the primary value.

```go
prop:"'database.host'"
```

This value implicitly maps to the field whose option name is `value` — a reserved keyword used
by Tag to represent the primary value.

```go
type PropTag struct {
    Key string `option:"value"` // primary value
}
```

📌 **Important**: `value` is reserved option name and should not be used for other option names.
Only one field in a tag struct can map to `option:"value"`.

### Bool Flag Options

Any option written without an equals sign (=) is parsed as a boolean flag and evaluated to true.

```go
prop:"'database.host',optional"`
```

Flags must always map to boolean fields in your tag struct:

```go
type PropTag struct {
    Key      string `option:"value"`    // primary value
    Optional bool   `option:"optional"` // bool flag
}
```

### Key/Value Options

Options written in the form `key=value` attach typed configuration values.

```go
prop:"'database.host',default=5432"
```

These options map directly to the struct fields that use the same option name:

```go
type PropTag struct {
    Key      string `option:"value"`    // primary value
    Optional bool   `option:"optional"` // bool flag option
    Default  any    `option:"default"`  // key=value option
}
```

Tag automatically infers the type of each option value (string, int, float, bool, slice, map, struct, etc.).

## 📚 Advanced Examples

### Slices

Slices are expressed using square brackets `[...]`.
Strings are written with quotes; numeric values can be written directly:

```go
roles:"['admin','editor','viewer']"
```

This maps to a slice field in your tag struct:

```go
type RolesTag struct {
    Value []string `option:"value"`
}

func (t RolesTag) Tag() string {
    return "roles"
}
```

### Nested Slices

Slices can contain other slices:

```go
matrix:"[[1,2,3],[4,5,6],[7,8,9]]"
```

This maps to a nested slice field in your tag struct:

```go
type MatrixTag struct {
    Value [][]int `option:"value"`
}

func (t MatrixTag) Tag() string {
    return "matrix"
}
```

### Maps
Maps are expressed using curly braces `{...}` with key:value pairs.
Keys must be strings; values can be any supported type:

```go
limits:"{'cpu':'2','memory':'4Gi'}"
```

This maps to a map field in your tag struct:

```go
type LimitsTag struct {
    Value map[string]string `option:"value"`
}

func (t LimitsTag) Tag() string {
    return "limits"
}
```

Maps can mix different value types as well:

```go
config:"{'debug':true,'max_connections':100,'timeout':30.5}"
```

This maps to a map with any values:

```go
type ConfigTag struct {
    Value map[string]any `option:"value"`
}

func (t ConfigTag) Tag() string {
    return "config"
}
```

### Maps with Inner Slices and Nested Maps

Map values can be slices or other maps:
```go
settings:"{'servers':['server1','server2'],'thresholds':{'cpu':80,'memory':70}}"
```

This maps to a complex nested structure:

```go
type SettingsTag struct {
    Value map[string]any `option:"value"`
}

func (t SettingsTag) Tag() string {
    return "settings"
}
```

### Nested Structs (Custom Types)

Custom struct fields use the same syntax as maps — key/value pairs inside {...}.
Tag parses it as a map internally, then binds it into your struct using field option names.

```go
database:"{'host':'localhost','port':5432,'credentials':{'user':'admin','password':'secret'}}"
```

This maps to a nested struct field:

```go
type Credentials struct {
    User     string `option:"user"`
    Password string `option:"password"`
}

type DatabaseTag struct {
    Host        string      `option:"host"`
    Port        int         `option:"port"`
    Credentials Credentials `option:"credentials"`
}

func (t DatabaseTag) Tag() string {
    return "database"
}
```

### Combined Options

You can combine flags, key/value options, slices, maps, and nested structs
in a single tag:

```go
service:"'my-service',optional,replicas=3,labels={'env':'prod','tier':'backend'},ports=[80,443]"
```

This maps to a complex tag struct:

```go
type ServiceTag struct {
    Name     string            `option:"value"`
    Optional bool              `option:"optional"`
    Replicas int               `option:"replicas"`
    Labels   map[string]string `option:"labels"`
    Ports    []int             `option:"ports"`
}

func (t ServiceTag) Tag() string {
    return "service"
}
```

## License

Tag is released under version 2.0 of the Apache License.
