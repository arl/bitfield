[![Tests](https://github.com/arl/bitfield/actions/workflows/test.yml/badge.svg)](https://github.com/arl/bitfield/actions/workflows/test.yml)

Bitfield
========

Package bitfield generates Pack/Unpack code for struct types whose fields
are tagged with bit widths.

Given a type whose fields carry `bitfield:"<width>"` struct tags:

```go
//go:generate go run github.com/arl/bitfield/v2 -type Flags

type Flags struct {
    Opcode  uint8 `bitfield:"6"`
    Mode    uint8 `bitfield:"2"`
    Enabled bool  `bitfield:"1"`
    Rsvd    uint8 `bitfield:"7"`
}
```

Calling `go generate` will produce a new Go file containing:

```go
func (v Flags) Pack() uint16
func UnpackFlags(raw uint16) Flags
```

The bit layout places the first field at the LSB and subsequent fields at
increasing offsets. The storage type is the smallest of `uint8`, `uint16`, `uint32`,
`uint64` that holds the total width.

Supported field types: `bool` (always exactly 1 bit) and any type whose underlying
kind is `uint8`, `uint16`, `uint32`, or `uint64`. Named types are preserved in the
emitted code, so `type Mode uint8` round-trips as `Mode`.


## License

This project is licensed under the [MIT](LICENSE) - see the [LICENSE](LICENSE) file for details.
