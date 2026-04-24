package example

import "github.com/arl/bit"

//go:generate go run .. -type=Flags,Tiny,Wide,Color,padded -output ./generated.go

type Mode uint8

type Flags struct {
	Opcode  uint8 `bitfield:"6"`
	Mode    uint8 `bitfield:"2"`
	Enabled bool  `bitfield:"1"`
	Rsvd    uint8 `bitfield:"7"`
}

type u3 = bit.U3

type Tiny struct {
	A bool  `bitfield:"1"`
	B u3    `bitfield:"3"`
	_ uint8 `bitfield:"3"`
	C bool  `bitfield:"1"`
}

type Wide struct {
	Lo uint32 `bitfield:"32"`
	Hi uint32 `bitfield:"32"`
}

// Color packs three 3-bit channels into a 16-bit word with a single
// padding bit between each channel and four padding bits at the top.
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

// padded is an unexported struct with unexported fields; the generator
// must emit unexported pack/unpack helpers for it.
type padded struct {
	lo  uint8 `bitfield:"4"`
	_   uint8 `bitfield:"2"`
	hi  uint8 `bitfield:"1"`
	set bool  `bitfield:"1"`
}
