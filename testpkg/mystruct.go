//go:build ignore

package testpkg

// MyStruct example struct to parse.
type MyStruct struct {
	CoarseX    uint8 `bitfield:"5"`
	CoarseY    uint8 `bitfield:"5"`
	NametableX uint8 `bitfield:"1"`
	NametableY uint8 `bitfield:"1"`
	FineY      uint8 `bitfield:"3"`
	_          uint8 `bitfield:"1"`

	Low  uint8 `bitfield:"8,union=hl"`
	High uint8 `bitfield:"7,union=hl"`

	F1 bool `bitfield:"1,union=flags"`
	_  bool `bitfield:"1,union=flags"`
	F2 bool `bitfield:"1,union=flags"`
}
