//go:build ignore

package mypkg

//go:generate bitfield -out scroll_bitfield.go

type Scroll struct {
	X uint8 `bitfield:"5"`
	Y uint8 `bitfield:"5"`
}
