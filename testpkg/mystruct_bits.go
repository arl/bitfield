package testpkg

type MyStruct uint16

func (s MyStruct) CoarseX() uint8 {
	return uint8((s >> 11) & 0x1f)
}

func (s MyStruct) SetCoarseX(val uint8) MyStruct {
	return s ^ 0x1f<<11 | (MyStruct(val)&0x1f)<<11
}

func (s MyStruct) CoarseY() uint8 {
	return uint8((s >> 6) & 0x1f)
}

func (s MyStruct) SetCoarseY(val uint8) MyStruct {
	return s ^ 0x1f<<6 | (MyStruct(val)&0x1f)<<6
}

func (s MyStruct) NametableX() uint8 {
	return uint8((s >> 5) & 0x1)
}

func (s MyStruct) SetNametableX(val uint8) MyStruct {
	return s ^ 0x1<<5 | (MyStruct(val)&0x1)<<5
}

func (s MyStruct) NametableY() uint8 {
	return uint8((s >> 4) & 0x1)
}

func (s MyStruct) SetNametableY(val uint8) MyStruct {
	return s ^ 0x1<<4 | (MyStruct(val)&0x1)<<4
}

func (s MyStruct) FineY() uint8 {
	return uint8((s >> 1) & 0x7)
}

func (s MyStruct) SetFineY(val uint8) MyStruct {
	return s ^ 0x7<<1 | (MyStruct(val)&0x7)<<1
}
