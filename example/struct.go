package example

//go:generate go run .. -type=Flags,Tiny,Wide -output ./generated.go

type Mode uint8

type Flags struct {
	Opcode  uint8 `bitfield:"6"`
	Mode    uint8 `bitfield:"2"`
	Enabled bool  `bitfield:"1"`
	Rsvd    uint8 `bitfield:"7"`
}

type Tiny struct {
	A bool  `bitfield:"1"`
	B uint8 `bitfield:"3"`
	C bool  `bitfield:"1"`
}

type Wide struct {
	Lo uint32 `bitfield:"32"`
	Hi uint32 `bitfield:"32"`
}
