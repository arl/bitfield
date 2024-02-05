package testpkg

import (
	"testing"
)

func TestCoarseX(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b00101, want: 0b101},
		{s: 0b11111, want: 0b11111},
		{s: 0b101111, want: 0b01111},
	}

	for _, tt := range tests {
		if got := tt.s.CoarseX(); got != tt.want {
			t.Errorf("MyStruct(%b).CoarseX() = %b, want %b", tt.s, got, tt.want)
		}
	}
}

func TestSetCoarseX(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  uint8
		want MyStruct
	}{
		{s: 0, want: 0},
		{s: 0, val: 0b11111, want: 0b11111},
		{s: 0b11_1111, val: 0b1_0000, want: 0b11_0000},
	}

	for _, tt := range tests {
		got := tt.s
		if got.SetCoarseX(tt.val); got != tt.want {
			t.Errorf("MyStruct(%b).SetCoarseX(%b) = %b, want %b", tt.s, tt.val, got, tt.want)
		}
	}
}

func TestCoarseY(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b11_1111_1111, want: 0b11111},
		{s: 0b111_1111_1001, want: 0b11111},
		{s: 0b011_1110_1111, want: 0b11111},
	}

	for _, tt := range tests {
		if got := tt.s.CoarseY(); got != tt.want {
			t.Errorf("MyStruct(%b).CoarseY() = %b, want %b", tt.s, got, tt.want)
		}
	}
}

func TestSetCoarseY(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  uint8
		want MyStruct
	}{
		{s: 0, want: 0},
		{s: 0, val: 0b11111, want: 0b11111 << 5},
	}

	for _, tt := range tests {
		got := tt.s
		if got.SetCoarseY(tt.val); got != tt.want {
			t.Errorf("MyStruct(%b).SetCoarseY(%b) = %b, want %b", tt.s, tt.val, got, tt.want)
		}
	}
}

func TestLow(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b1111_1111, want: 0b1111_1111},
		{s: 0b1_1111_1111, want: 0b1111_1111},
		{s: 0b00_1010_1010, want: 0b1010_1010},
		{s: 0b10_0010_1010, want: 0b10_1010},
	}

	for _, tt := range tests {
		if got := tt.s.Low(); got != tt.want {
			t.Errorf("MyStruct(%b).Low() = %b, want %b", tt.s, got, tt.want)
		}
	}
}

func TestSetLow(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  uint8
		want MyStruct
	}{
		{s: 0, val: 0, want: 0},
		{s: 0, val: 0b1000_0001, want: 0b1000_0001},
		{s: 0b1111_1111, val: 0b11_1100, want: 0b0011_1100},
	}

	for _, tt := range tests {
		got := tt.s
		if got.SetLow(tt.val); got != tt.want {
			t.Errorf("MyStruct(%b).SetLow(%b) = %b, want %b", tt.s, tt.val, got, tt.want)
		}
	}
}

func TestHigh(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b1111_1111, want: 0b0},
		{s: 0b1_1111_1111, want: 0b1},
		{s: 0b1010_1100_1010_1010, want: 0b10_1100},
		{s: 0b10_1101_1010_1010, want: 0b10_1101},
	}

	for _, tt := range tests {
		if got := tt.s.High(); got != tt.want {
			t.Errorf("MyStruct(%b).High() = %b, want %b", tt.s, got, tt.want)
		}
	}
}

func TestF1(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want bool
	}{
		{s: 0, want: false},
		{s: 0b1111_1111, want: true},
		{s: 0b1_1111_1110, want: false},
		{s: 0b1000, want: false},
		{s: 0b1101, want: true},
		{s: 0b1, want: true},
		{s: 0b10, want: false},
		{s: 0b110, want: false},
	}

	for _, tt := range tests {
		if got := tt.s.F1(); got != tt.want {
			t.Errorf("MyStruct(%b).F1() = %t, want %t", tt.s, got, tt.want)
		}
	}
}

func TestSetF1(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  bool
		want MyStruct
	}{
		{s: 0, val: false, want: 0},
		{s: 0, val: true, want: 1},
		{s: 0b10, val: false, want: 0b10},
		{s: 0b1000_0001, val: false, want: 0b1000_0000},
		{s: 0b1000_0001, val: true, want: 0b1000_0001},
		{s: 0b1111_1111, val: true, want: 0b1111_1111},
		{s: 0b1111_1111, val: false, want: 0b1111_1110},
		{s: 0b1111_1011, val: true, want: 0b1111_1011},
		{s: 0b1111_1011, val: false, want: 0b1111_1010},
	}

	for _, tt := range tests {
		got := tt.s
		if got.SetF1(tt.val); got != tt.want {
			t.Errorf("MyStruct(%b).SetF1(%t) = %b, want %b", tt.s, tt.val, got, tt.want)
		}
	}
}

func TestF2(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want bool
	}{
		{s: 0, want: false},
		{s: 0b1111_1111, want: true},
		{s: 0b1_1111_1011, want: false},
		{s: 0b1000, want: false},
		{s: 0b1100, want: true},
		{s: 0b1, want: false},
		{s: 0b10, want: false},
		{s: 0b101, want: true},
	}

	for _, tt := range tests {
		if got := tt.s.F2(); got != tt.want {
			t.Errorf("MyStruct(%b).F2() = %t, want %t", tt.s, got, tt.want)
		}
	}
}

func TestSetF2(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  bool
		want MyStruct
	}{
		{s: 0, val: false, want: 0},
		{s: 0, val: true, want: 0b100},
		{s: 0, val: false, want: 0},
		{s: 0b1000_0001, val: false, want: 0b1000_0001},
		{s: 0b1000_0001, val: true, want: 0b1000_0101},
		{s: 0b1111_1111, val: true, want: 0b1111_1111},
		{s: 0b1111_1111, val: false, want: 0b1111_1011},
		{s: 0b1111_1011, val: true, want: 0b1111_1111},
		{s: 0b1111_1011, val: false, want: 0b1111_1011},
	}

	for _, tt := range tests {
		got := tt.s
		if got.SetF2(tt.val); got != tt.want {
			t.Errorf("MyStruct(%b).SetF2(%t) = %b, want %b", tt.s, tt.val, got, tt.want)
		}
	}
}
