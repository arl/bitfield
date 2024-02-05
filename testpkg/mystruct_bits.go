package testpkg

// Code generated by github.com/arl/bitfield. DO NOT EDIT.

type MyStruct uint16

func (m MyStruct) CoarseX() uint8 {
	return uint8(m & 0x1f)
}

func (m *MyStruct) SetCoarseX(val uint8) {
	*m &^= 0x1f
	*m |= MyStruct(val & 0x1f)
}

func (m MyStruct) CoarseY() uint8 {
	return uint8((m >> 5) & 0x1f)
}

func (m *MyStruct) SetCoarseY(val uint8) {
	*m &^= 0x1f << 5
	*m |= MyStruct(val&0x1f) << 5
}

func (m MyStruct) NametableX() uint8 {
	return uint8((m >> 10) & 0x1)
}

func (m *MyStruct) SetNametableX(val uint8) {
	*m &^= 0x1 << 10
	*m |= MyStruct(val&0x1) << 10
}

func (m MyStruct) NametableY() uint8 {
	return uint8((m >> 11) & 0x1)
}

func (m *MyStruct) SetNametableY(val uint8) {
	*m &^= 0x1 << 11
	*m |= MyStruct(val&0x1) << 11
}

func (m MyStruct) FineY() uint8 {
	return uint8((m >> 12) & 0x7)
}

func (m *MyStruct) SetFineY(val uint8) {
	*m &^= 0x7 << 12
	*m |= MyStruct(val&0x7) << 12
}

func (m MyStruct) Low() uint8 {
	return uint8(m & 0xff)
}

func (m *MyStruct) SetLow(val uint8) {
	*m &^= 0xff
	*m |= MyStruct(val & 0xff)
}

func (m MyStruct) High() uint8 {
	return uint8((m >> 8) & 0x7f)
}

func (m *MyStruct) SetHigh(val uint8) {
	*m &^= 0x7f << 8
	*m |= MyStruct(val&0x7f) << 8
}

func (m MyStruct) F1() bool {
	return m&0x1 != 0
}

func (m *MyStruct) SetF1(val bool) {
	var ival MyStruct
	if val {
		ival = 1
	}
	*m &^= 0x1
	*m |= ival
}

func (m MyStruct) F2() bool {
	return m&0x4 != 0
}

func (m *MyStruct) SetF2(val bool) {
	var ival MyStruct
	if val {
		ival = 1
	}
	*m &^= 0x4
	*m |= ival << 2
}
