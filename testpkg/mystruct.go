//go:build ignore

package testpkg

// MyStruct example struct to parse.
type MyStruct struct {
	CoarseX    uint8 `bitfield:"bits=5"`
	CoarseY    uint8 `bitfield:"bits=5"`
	NametableX uint8 `bitfield:"bits=1"`
	NametableY uint8 `bitfield:"bits=1"`
	FineY      uint8 `bitfield:"bits=3"`
	_          uint8 `bitfield:"bits=1"`

	Low  uint8 `bitfield:"bits=8,union=hl"`
	High uint8 `bitfield:"bits=7,union=hl"`
}
