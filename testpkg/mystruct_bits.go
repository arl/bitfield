package testpkg

// Code generated by github.com/arl/bitfield. DO NOT EDIT.

type MyStruct uint16

func (s MyStruct) CoarseX() uint8 {
	return uint8(s & 0x1f)
}

func (s MyStruct) SetCoarseX(val uint8) MyStruct {
	return s&^0x1f | MyStruct(val&0x1f)
}

func (s MyStruct) CoarseY() uint8 {
	return uint8((s >> 5) & 0x1f)
}

func (s MyStruct) SetCoarseY(val uint8) MyStruct {
	return s&^(0x1f<<5) | (MyStruct(val&0x1f) << 5)
}

func (s MyStruct) NametableX() uint8 {
	return uint8((s >> 10) & 0x1)
}

func (s MyStruct) SetNametableX(val uint8) MyStruct {
	return s&^(0x1<<10) | (MyStruct(val&0x1) << 10)
}

func (s MyStruct) NametableY() uint8 {
	return uint8((s >> 11) & 0x1)
}

func (s MyStruct) SetNametableY(val uint8) MyStruct {
	return s&^(0x1<<11) | (MyStruct(val&0x1) << 11)
}

func (s MyStruct) FineY() uint8 {
	return uint8((s >> 12) & 0x7)
}

func (s MyStruct) SetFineY(val uint8) MyStruct {
	return s&^(0x7<<12) | (MyStruct(val&0x7) << 12)
}

func (s MyStruct) Low() uint8 {
	return uint8(s & 0xff)
}

func (s MyStruct) SetLow(val uint8) MyStruct {
	return s&^0xff | MyStruct(val&0xff)
}

func (s MyStruct) High() uint8 {
	return uint8((s >> 8) & 0x7f)
}

func (s MyStruct) SetHigh(val uint8) MyStruct {
	return s&^(0x7f<<8) | (MyStruct(val&0x7f) << 8)
}

func (s MyStruct) F1() bool {
	return s&0x1 != 0
}

func (s MyStruct) SetF1(val bool) MyStruct {
	var ival MyStruct
	if val {
		ival = 1
	}
	return s&^0x1 | ival<<0
}

func (s MyStruct) F2() bool {
	return s&0x4 != 0
}

func (s MyStruct) SetF2(val bool) MyStruct {
	var ival MyStruct
	if val {
		ival = 1
	}
	return s&^0x4 | ival<<2
}
