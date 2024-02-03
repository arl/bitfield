[![Tests](https://github.com/arl/bitfield/actions/workflows/test.yml/badge.svg)](https://github.com/arl/bitfield/actions/workflows/test.yml)

Bitfield
=======

This is a Go code generation tool to emulate C bitfields and unions.

## Quickstart

Install with:

```sh
go install github.com/arl/bitfield@latest
```

Say you want to generate, in Go, the following C struct:

```c
struct scroll
{
    unsigned X : 5;
    unsigned Y : 5;
};
```

You'd declare the following Go struct in a file called `scroll_gen.go`, only used to define the bit field, that's why it's guarded with an `go:build ignore` tag.

```go
//go:build ignore

package mypkg

//go:generate bitfield -out scroll.go

type Scroll struct {
	X uint8 `bitfield:"5"`
	Y uint8 `bitfield:"5"`
}
```

And then run:

```sh
go generate scroll_gen.go
```

This creates the following `scroll.go`:

```go
type Scroll uint16

func (s Scroll) X() uint8 {
	return uint8(s & 0x1f)
}

func (s Scroll) SetX(val uint8) Scroll {
	return s&^0x1f | Scroll(val&0x1f)
}

func (s Scroll) Y() uint8 {
	return uint8((s >> 5) & 0x1f)
}

func (s Scroll) SetY(val uint8) Scroll {
	return s&^(0x1f<<5) | (Scroll(val&0x1f) << 5)
}
```

Note that both getter and setter methods are defined on by-value receiver. That's why setters return the modified value.


### Supported field types

These are the allowed types for the fields of the input struct:
 - bool
 - uint8
 - uint16
 - uint32
 - uint64

`bool` can only be defined on a 1-bit wide field.

For the generated type, `bitfield` automatically uses the smallest unsigned integer to accomodate for all the bits of the field.


### Bit fields with anonymous unions

The following C union:

```c
union Addr
{
    struct
    {
        unsigned cX : 5;  // Coarse X.
        unsigned cY : 5;  // Coarse Y.
        unsigned nt : 2;  // Nametable.
        unsigned fY : 3;  // Fine Y.
    };
    struct
    {
        unsigned l : 8;
        unsigned h : 7;
    };

    unsigned val : 14;
};
```

can be defined with:

```go
type Addr struct {
	cX uint8 `bitfield:"5,union=scroll"`
	cY uint8 `bitfield:"5,union=scroll"`
	nt uint8 `bitfield:"2,union=scroll"`
	fY uint8 `bitfield:"3,union=scroll"`

	l uint8 `bitfield:"8,union=lohi"`
	h uint8 `bitfield:"7,union=lohi"`

	// When not specified, implicitely uses the 'default' union.
	val uint8 `bitfield:"14"`
}
```

## Usage

```
bitfield -h
Usage of bitfield:
  -in string
        INPUT file name (necessary unless within a go:generate comment)
  -out string
        output file name (defaults to standard output)
  -pkg string
        package name (defaults to INPUT file package)
  -type string
        name of the type to convert (defaults to all structs) (default "all")
```

## License

This project is licensed under the [MIT](LICENSE) - see the [LICENSE](LICENSE) file for details.
