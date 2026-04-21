package example

import (
	"testing"
)

func TestRoundTrip(t *testing.T) {
	// Flags
	f := Flags{Opcode: 42, Mode: 2, Enabled: true, Rsvd: 0x5A}
	p := f.Pack()
	// Opcode=42 (6 bits) | Mode=2 << 6 | Enabled << 8 | Rsvd=0x5A<<9
	want := uint16(42) | uint16(2)<<6 | uint16(1)<<8 | uint16(0x5A)<<9
	if p != want {
		t.Errorf("Flags Pack got=%#x want=%#x", p, want)
	}
	back := UnpackFlags(p)
	if back != f {
		t.Errorf("Flags roundtrip got=%+v want=%+v", back, f)
	}

	// Overflow is masked at pack time.
	f2 := Flags{Opcode: 0xFF, Mode: 0xFF, Enabled: true, Rsvd: 0xFF}
	back2 := UnpackFlags(f2.Pack())
	wantFlags := Flags{Opcode: 0x3F, Mode: 0x3, Enabled: true, Rsvd: 0x7F}
	if back2 != wantFlags {
		t.Errorf("Flags overflow masking")
	}

	// Tiny (uint8 storage)
	ti := Tiny{A: true, B: 5, C: true}
	tb := UnpackTiny(ti.Pack())
	if tb != ti {
		t.Errorf("Tiny roundtrip got=%+v want=%+v", tb, ti)
	}
	// Zero should round-trip.
	var tiny Tiny
	zero := UnpackTiny(tiny.Pack())
	if zero != tiny {
		t.Errorf("Tiny zero roundtrip")
	}

	// Wide (uint64 storage, exact-width fields — no masking should be emitted)
	w := Wide{Lo: 0xDEADBEEF, Hi: 0xCAFEBABE}
	wp := w.Pack()
	if wp != uint64(0xCAFEBABE)<<32|uint64(0xDEADBEEF) {
		t.Errorf("Wide Pack got=%#x", wp)
	}
	wb := UnpackWide(wp)
	if wb != w {
		t.Errorf("Wide roundtrip")
	}
}
