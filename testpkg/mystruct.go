package testpkg

// MyStruct example struct to parse.
type MyStruct struct {
	CoarseX    uint8 `bitfield:"5"`
	CoarseY    uint8 `bitfield:"5"`
	NametableX uint8 `bitfield:"1"`
	NametableY uint8 `bitfield:"1"`
	FineY      uint8 `bitfield:"3"`
	_          uint8 `bitfield:"1"`
}
