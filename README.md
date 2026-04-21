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

### Padding (reserved) bits

Fields declared with the blank identifier `_` reserve bits in the layout but are
otherwise ignored — no code in `Pack`/`Unpack` references them. Use this to
model "don't care" or hardware-reserved slots without inventing a dummy name:

```go
// Layout (LSB first): ----bbb-ggg-rrr-
type Color struct {
    _ uint8 `bitfield:"1"`
    R uint8 `bitfield:"3"`
    _ uint8 `bitfield:"1"`
    G uint8 `bitfield:"3"`
    _ uint8 `bitfield:"1"`
    B uint8 `bitfield:"3"`
    _ uint8 `bitfield:"4"`
}
```

`Pack` writes zeroes into reserved slots and `UnpackColor` simply does not read
from them. The field's element type only matters in that its native width must
be large enough to hold the declared bit count.

### Unexported types and fields

Both the struct type and its individual fields may be unexported. When the
*type* is unexported, the generator keeps the generated helpers at the same
visibility:

| Source type | Pack method     | Unpack function |
|-------------|-----------------|-----------------|
| `Foo`       | `func (Foo) Pack() …`   | `func UnpackFoo(…) Foo` |
| `foo`       | `func (foo) pack() …`   | `func unpackFoo(…) foo` |

### Output location

The `-output` flag must point to a file inside the source package directory —
the generated file declares `package <sourcePkg>`, so placing it elsewhere
would produce a mismatched file. The tool enforces this.


## License

This project is licensed under the [MIT](LICENSE) - see the [LICENSE](LICENSE) file for details.
