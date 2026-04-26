// bitfield is a tool to generate Pack/Unpack code for struct types whose
// fields are tagged with bit widths.
//
// Given a type whose fields carry `bitfield:"<width>"` struct tags:
//
//	type Flags struct {
//	    Opcode  uint8 `bitfield:"6"`
//	    Mode    uint8 `bitfield:"2"`
//	    Enabled bool  `bitfield:"1"`
//	    Rsvd    uint8 `bitfield:"7"`
//	}
//
// running this command in the same directory
//
//	bitfield -type=Flags
//
// creates the file flags_fields.go containing:
//
//	func (v Flags) Pack() uint16
//	func UnpackFlags(raw uint16) Flags
//
// The bit layout places the first field at the LSB and subsequent fields at
// increasing offsets. The storage type is the smallest of uint8, uint16,
// uint32, uint64 that holds the total width.
//
// Supported field types: bool (always exactly 1 bit) and any type whose
// underlying kind is uint8, uint16, uint32, or uint64. Named types are
// preserved in the emitted code, so `type Mode uint8` round-trips as Mode.
//
// Fields (exported or not) may be declared with the blank identifier `_`
// to reserve bits without contributing a name to Pack/Unpack:
//
//	type Color struct {
//	    _ uint8 `bitfield:"1"` // padding
//	    R uint8 `bitfield:"3"`
//	    _ uint8 `bitfield:"1"`
//	    G uint8 `bitfield:"3"`
//	}
//
// When the target type itself is unexported, the generated methods follow
// suit: `pack` and `unpack<Type>` instead of `Pack`/`Unpack<Type>`.
//
// # Typical go:generate wiring
//
// Add a go:generate directive in your package:
//
//	//go:generate go run github.com/arl/bitfield/v2 -type=Flags
//
// Then `go generate ./...` (re)produces flags_fields.go with Pack and
// Unpack<Type> for each listed type.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// fieldSpec captures a single field's generator-relevant facts.
type fieldSpec struct {
	Name      string       // as declared; "_" for blank (padding) fields
	TypeName  string       // e.g. "uint8" or "Mode"; used in emitted code
	Kind      reflect.Kind // Bool, Uint8, Uint16, Uint32, Uint64
	Width     uint         // declared bits
	Offset    uint         // LSB offset within storage
	KindWidth uint         // native width of Kind: 1 for bool, 8/16/32/64 otherwise
	Blank     bool         // declared as `_`; reserves bits but emits no pack/unpack code
}

// typeSpec captures a struct's generator-relevant facts.
type typeSpec struct {
	Name     string
	Fields   []fieldSpec
	Total    uint   // sum of widths
	Storage  string // "uint8", "uint16", "uint32", "uint64"
	Exported bool   // whether the type is exported; controls visibility of generated funcs
}

// packName returns the name of the generated Pack method for t.
// Exported types use "Pack"; unexported types use "pack" to keep visibility
// aligned with the type itself.
func (t typeSpec) packName() string {
	if t.Exported {
		return "Pack"
	}
	return "pack"
}

// unpackName returns the name of the generated Unpack function for t.
// For an exported "Foo" the name is "UnpackFoo"; for an unexported "foo"
// the name is "unpackFoo" (the type-name part is always capitalized so the
// function reads naturally, while the leading verb follows the type's
// visibility).
func (t typeSpec) unpackName() string {
	// if t.Name == "" {
	// 	return "Unpack"
	// }
	// head := strings.ToUpper(t.Name[:1]) + t.Name[1:]
	if t.Exported {
		return "Unpack" //+ head
	}
	return "unpack" //+ head
}

func parseWidth(tag string) (uint, error) {
	// Extend here later if multi-value tags are needed.
	raw := strings.TrimSpace(tag)
	n, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid bitfield tag %q: %w", tag, err)
	}
	if n == 0 {
		return 0, fmt.Errorf("bitfield tag %q: width must be at least 1", tag)
	}
	return uint(n), nil
}

func storageFor(totalBits uint) (string, error) {
	switch {
	case totalBits == 0:
		return "", errors.New("struct has no bitfield-tagged fields")
	case totalBits <= 8:
		return "uint8", nil
	case totalBits <= 16:
		return "uint16", nil
	case totalBits <= 32:
		return "uint32", nil
	case totalBits <= 64:
		return "uint64", nil
	default:
		return "", fmt.Errorf("total width %d exceeds 64 bits", totalBits)
	}
}

func writeType(buf *bytes.Buffer, t typeSpec) {
	writePack(buf, t)
	writeUnpack(buf, t)
}

func writePack(buf *bytes.Buffer, t typeSpec) {
	fmt.Fprintf(buf, "// %s returns the bit-packed %s representation of v.\n", t.packName(), t.Storage)
	fmt.Fprintf(buf, "func (v %s) %s() %s {\n", t.Name, t.packName(), t.Storage)
	fmt.Fprintf(buf, "\tvar out %s\n", t.Storage)
	for _, f := range t.Fields {
		if f.Blank {
			continue
		}
		writePackField(buf, t.Storage, f)
	}
	fmt.Fprintln(buf, "\treturn out")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

func writePackField(buf *bytes.Buffer, storage string, f fieldSpec) {
	if f.Kind == reflect.Bool {
		if f.Offset == 0 {
			fmt.Fprintf(buf, "\tif v.%s {\n\t\tout |= 1\n\t}\n", f.Name)
		} else {
			fmt.Fprintf(buf, "\tif v.%s {\n\t\tout |= 1 << %d\n\t}\n", f.Name, f.Offset)
		}
		return
	}

	// Integer field.
	//
	// Three questions drive the emitted form:
	//   1. Does the field's declared type equal the storage type? If yes, no
	//      outer conversion is needed.
	//   2. Does the declared width equal the field's native type width? If
	//      yes, the per-field mask is redundant.
	//   3. Is the offset zero? If yes, drop the shift.
	sameType := f.TypeName == storage
	needMask := f.Width != f.KindWidth
	mask := (uint64(1) << f.Width) - 1

	// Build the field's value expression in the storage type. needsParens
	// tracks whether the expression must be parenthesized when shifted,
	// because `&` binds more loosely than `<<` in Go.
	var expr string
	var needsParens bool
	switch {
	case sameType && !needMask:
		expr = "v." + f.Name
	case sameType && needMask:
		expr = fmt.Sprintf("v.%s & %s", f.Name, hexLit(mask))
		needsParens = true
	case !sameType && !needMask:
		expr = fmt.Sprintf("%s(v.%s)", storage, f.Name)
	default:
		// Mask in the field's declared type, then widen. The mask constant
		// takes the type of v.Name for the inner expression, and the outer
		// conversion produces the storage type.
		expr = fmt.Sprintf("%s(v.%s & %s)", storage, f.Name, hexLit(mask))
	}

	switch {
	case f.Offset == 0:
		fmt.Fprintf(buf, "\tout |= %s\n", expr)
	case needsParens:
		fmt.Fprintf(buf, "\tout |= (%s) << %d\n", expr, f.Offset)
	default:
		fmt.Fprintf(buf, "\tout |= %s << %d\n", expr, f.Offset)
	}
}

func writeUnpack(buf *bytes.Buffer, t typeSpec) {
	fmt.Fprintf(buf, "// %s decodes a packed %s into a %s.\n", t.unpackName(), t.Storage, t.Name)
	fmt.Fprintf(buf, "func (v *%s) %s(raw %s) {\n", t.Name, t.unpackName(), t.Storage)
	fmt.Fprintf(buf, "\t*v = %s{\n", t.Name)
	for _, f := range t.Fields {
		if f.Blank {
			continue
		}
		writeUnpackField(buf, t.Storage, f)
	}
	fmt.Fprintln(buf, "\t}")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf)
}

func writeUnpackField(buf *bytes.Buffer, storage string, f fieldSpec) {
	if f.Kind == reflect.Bool {
		if f.Offset == 0 {
			fmt.Fprintf(buf, "\t\t%s: raw&1 != 0,\n", f.Name)
		} else {
			fmt.Fprintf(buf, "\t\t%s: raw>>%d&1 != 0,\n", f.Name, f.Offset)
		}
		return
	}

	// Integer field. Build a shifted+masked expression in the storage type,
	// then convert to the declared field type.
	needMask := f.Width != f.KindWidth
	sameType := f.TypeName == storage
	mask := (uint64(1) << f.Width) - 1

	// The raw expression in storage type.
	var raw string
	switch {
	case f.Offset == 0 && !needMask:
		raw = "raw"
	case f.Offset == 0 && needMask:
		raw = fmt.Sprintf("raw & %s", hexLit(mask))
	case f.Offset != 0 && !needMask:
		raw = fmt.Sprintf("raw >> %d", f.Offset)
	default:
		raw = fmt.Sprintf("raw >> %d & %s", f.Offset, hexLit(mask))
	}

	if sameType {
		fmt.Fprintf(buf, "\t\t%s: %s,\n", f.Name, raw)
	} else {
		// Wrap the raw expression in a conversion to the declared type.
		// Parenthesize the operand only if it contains whitespace (i.e. an
		// operator). This keeps emission of trivial cases readable.
		if strings.ContainsAny(raw, " ") {
			fmt.Fprintf(buf, "\t\t%s: %s(%s),\n", f.Name, f.TypeName, raw)
		} else {
			fmt.Fprintf(buf, "\t\t%s: %s(%s),\n", f.Name, f.TypeName, raw)
		}
	}
}

// hexLit formats a mask constant: small values in binary-friendly hex without
// an awkward leading zero run.
func hexLit(v uint64) string {
	return fmt.Sprintf("0x%x", v)
}
