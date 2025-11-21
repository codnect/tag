# Tag
Tag is a lightweight, zero-dependency Go library for parsing custom struct tag values into
strongly typed Go structs — safely, consistently, and with minimal reflection overhead.

It eliminates the repetitive boilerplate that appears whenever you need richer struct tag
semantics such as defaults, flags, lists, maps, positional options, or typed values.

## Features
* Reflection-based decoding into your own tag structs
* Support for scalar, slice, map, and nested struct values
* Type inference for any fields
* Boolean flag options (optional)
* Positional (first) option support

## Why This Library Exists

Go’s built-in reflect.StructTag exposes tag values only as raw strings:
```go
`prop:"'db-host',optional,default=5432"`
```

This forces every project to repeat the same work — splitting strings, handling flags, defaults, lists, maps, type conversions.

This leads to duplicated, bug-prone, and hard-to-maintain code across multiple projects.

Tag provides a single, reusable mechanism that:
*	Parses tag values safely
*	Populates fields automatically
*	Supports flags, lists, maps, defaults, and typed values

You define the tag structure once — the library handles the rest.

## How It Works

You define a struct representing the options of a tag:

```go
type PropTag struct {
    Key      string `option:"value"`
    Optional bool   `option:"optional"`
    Default  any    `option:"default"`
}

// Tag returns the name used in struct fields like `prop:"..."`.
func (PropTag) Tag() string { return "prop" }
```

Then you apply it to your own types:
```go
type Config struct {
    Host string `prop:"'db-host',optional,default=5432"`
}
```

## License
Tag is released under version 2.0 of the Apache License.