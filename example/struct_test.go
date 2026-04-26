package example

import "testing"

func TestRoundTrip(t *testing.T) {
	// Flags
	f := Flags{Opcode: 42, Mode: 2, Enabled: true, Rsvd: 0x5A}
	p := f.Pack()
	// Opcode=42 (6 bits) | Mode=2 << 6 | Enabled << 8 | Rsvd=0x5A<<9
	want := uint16(42) | uint16(2)<<6 | uint16(1)<<8 | uint16(0x5A)<<9
	if p != want {
		t.Errorf("Flags Pack got=%#x want=%#x", p, want)
	}
	var back Flags
	back.Unpack(p)
	if back != f {
		t.Errorf("Flags roundtrip got=%+v want=%+v", back, f)
	}

	// Overflow is masked at pack time.
	f2 := Flags{Opcode: 0xFF, Mode: 0xFF, Enabled: true, Rsvd: 0xFF}
	var back2 Flags
	back2.Unpack(f2.Pack())
	wantFlags := Flags{Opcode: 0x3F, Mode: 0x3, Enabled: true, Rsvd: 0x7F}
	if back2 != wantFlags {
		t.Errorf("Flags overflow masking")
	}

	// Tiny (uint8 storage)
	ti := Tiny{A: true, B: 5, C: true}
	var tb Tiny
	tb.Unpack(ti.Pack())
	if tb != ti {
		t.Errorf("Tiny roundtrip got=%+v want=%+v", tb, ti)
	}
	// Zero should round-trip.
	var tiny, zero Tiny
	zero.Unpack(tiny.Pack())
	if zero != tiny {
		t.Errorf("Tiny zero roundtrip")
	}

	// Wide (uint64 storage, exact-width fields — no masking should be emitted)
	w := Wide{Lo: 0xDEADBEEF, Hi: 0xCAFEBABE}
	wp := w.Pack()
	if wp != uint64(0xCAFEBABE)<<32|uint64(0xDEADBEEF) {
		t.Errorf("Wide Pack got=%#x", wp)
	}
	var wb Wide
	wb.Unpack(wp)
	if wb != w {
		t.Errorf("Wide roundtrip")
	}
}

func TestBitAlias(t *testing.T) {
	tiny := Tiny{
		A: true,
		B: u3(15), // out of range
		C: true,
	}

	want := uint8(0b10001111)
	tp := tiny.Pack()
	if tp != want {
		t.Errorf("packed tiny = 0b%0b, want 0b%0b", tp, want)
	}

	var t2 Tiny
	t2.Unpack(tp)
	if t2.B = t2.B.Add(1); t2.B != 0 {
		t.Errorf("b = 0b%0b, want 0b%0b", t2.B, 0)
	}

	want = uint8(0b10000001)
	tp = t2.Pack()
	if tp != want {
		t.Errorf("packed tiny = 0b%0b, want 0b%0b", tp, want)
	}
}

func TestColorPadding(t *testing.T) {
	// Layout (LSB first): [pad:1][R:3][pad:1][G:3][pad:1][B:3][pad:4]
	c := Color{R: 5, G: 2, B: 7}
	p := c.Pack()
	want := uint16(5)<<1 | uint16(2)<<5 | uint16(7)<<9
	if p != want {
		t.Fatalf("Color.Pack() = %#x, want %#x", p, want)
	}
	// Padding bits must be preserved as reserved on the wire and ignored on
	// unpack: flipping them in the raw word must not influence channel values.
	noisy := p | 1 | (1 << 4) | (1 << 8) | (0xF << 12)
	var back Color
	back.Unpack(noisy)
	if back != c {
		t.Fatalf("Color.Unpack ignored padding incorrectly: got %+v want %+v", back, c)
	}
	// Overflow is masked per-channel.
	over := Color{R: 0xFF, G: 0xFF, B: 0xFF}
	over.Unpack(over.Pack())
	if over != (Color{R: 7, G: 7, B: 7}) {
		t.Fatalf("Color overflow masking")
	}
}

func TestUnexportedStruct(t *testing.T) {
	// Round-trip through the unexported pack/unpack helpers. This also
	// asserts they exist with the expected (unexported) names.
	p := padded{lo: 0xA, hi: 1, set: true}
	var back padded
	back.unpack(p.pack())
	if back != p {
		t.Fatalf("padded roundtrip: got %+v want %+v", back, p)
	}
	// The 2-bit gap between lo and hi must stay at zero after pack and must
	// not leak into any field on unpack when set in the raw value.
	raw := p.pack()
	if raw&0x30 != 0 {
		t.Fatalf("padded.pack leaked bits into the reserved slot: %#x", raw)
	}
	var back2 padded
	back2.unpack(raw | 0x30)
	if back2 != p {
		t.Fatalf("unpackPadded read from reserved slot")
	}
}
