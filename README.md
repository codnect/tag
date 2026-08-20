<img width="150" height="150" alt="tag-logo" src="https://go.codnect.io/tag-logo.png" />

# Tag — Declarative Struct Tag Parser for Go

[![Run Tests](https://github.com/codnect/tag/actions/workflows/tag.yaml/badge.svg?branch=main)](https://github.com/codnect/tag/actions/workflows/tag.yaml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=codnect_tag&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=codnect_tag)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=codnect_tag&metric=coverage)](https://sonarcloud.io/summary/new_code?id=codnect_tag)
[![Go Reference](https://pkg.go.dev/badge/codnect.io/tag?status.svg)](https://pkg.go.dev/codnect.io/tag?tab=doc)

**Tag** is a lightweight, zero-dependency library that transforms raw Go struct tag
strings into strongly typed, well-structured Go values.

It eliminates repetitive parsing logic, reduces boilerplate, and provides a clean,
declarative way to define your tag schema — while keeping performance predictable
and implementation overhead minimal.

## Installation

To install Tag, run the following command:

```bash
go get go.codnect.io/tag
```

## Overview

Go’s `reflect.StructTag` exposes tag values only as raw strings:

```go
prop:"'database.host',default=5432"
```

In most projects, this leads to manual string splitting, repetitive parsing logic,
custom flag handling, default value handling, and edge cases around typed values.
Tag centralizes this logic into a declarative schema. You define a struct that
describes your tag, and Tag handles parsing, type inference, nested structures,
and value assignment.

## Quick Start

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

## What it supports
- Primary values with option:"value"
- Boolean flags such as optional
- Key/value options such as default=5432
- Primitive values, slices, maps, and nested structs
- Typed binding into your own schema structs

## Documentation
Read the full documentation at [go.codnect.io/tag](https://go.codnect.io/tag).

## License

Tag is released under version 2.0 of the Apache License.
