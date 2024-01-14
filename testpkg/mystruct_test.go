package testpkg

import "testing"

func TestCoarseX(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b11111 << 11, want: 0b11111},
		{s: 0b11111 << 11, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, want: 0b11111},
	}

	for _, tt := range tests {
		if got := tt.s.CoarseX(); got != tt.want {
			t.Errorf("MyStruct.CoarseX() = %v, want %v", got, tt.want)
		}
	}
}

func TestSetCoarseX(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  uint8
		want MyStruct
	}{
		{s: 0, val: 0b11111, want: 0b11111 << 11},
		{s: 0b11111 << 11, val: 0b11111, want: 0b11111 << 11},
		{s: 0b11111<<11 | 0b11111<<6, val: 0b11111, want: 0b11111<<11 | 0b11111<<6},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5, val: 0b11111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, val: 0b11111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, val: 0b1111111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0},
	}

	for _, tt := range tests {
		if got := tt.s.SetCoarseX(tt.val); got != tt.want {
			t.Errorf("MyStruct.SetCoarseX(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

func TestCoarseY(t *testing.T) {
	tests := []struct {
		s    MyStruct
		want uint8
	}{
		{s: 0, want: 0},
		{s: 0b11111 << 6, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5, want: 0b11111},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, want: 0b11111},
	}

	for _, tt := range tests {
		if got := tt.s.CoarseY(); got != tt.want {
			t.Errorf("MyStruct.CoarseY() = %v, want %v", got, tt.want)
		}
	}
}

func TestSetCoarseY(t *testing.T) {
	tests := []struct {
		s    MyStruct
		val  uint8
		want MyStruct
	}{
		{s: 0, val: 0b11111, want: 0b11111 << 6},
		{s: 0b11111 << 6, val: 0b11111, want: 0b11111 << 6},
		{s: 0b11111<<11 | 0b11111<<6, val: 0b11111, want: 0b11111<<11 | 0b11111<<6},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5, val: 0b11111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, val: 0b11111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0},
		{s: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0, val: 0b1111111, want: 0b11111<<11 | 0b11111<<6 | 0b1<<5 | 0b1<<0},
	}

	for _, tt := range tests {
		if got := tt.s.SetCoarseY(tt.val); got != tt.want {
			t.Errorf("MyStruct.SetCoarseY(%v) = %v, want %v", tt.val, got, tt.want)
		}
	}
}
