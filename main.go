package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

type Field struct {
	Name    string
	Offset  int
	BitSize int
	Mask    string
}

type StructInfo struct {
	StructName string
	Fields     []Field
	Bits       int
}

func main() {
	var (
		in, out string
		tname   string
		pkgname string
	)

	flag.StringVar(&in, "in", "", "input file name")
	flag.StringVar(&out, "out", "", "output file name (defaults to [in]_bits.go)")
	flag.StringVar(&tname, "type", "all", "struct name to read tags from (or all)")
	flag.StringVar(&pkgname, "pkg", "", "package name (defaults to input file package)")
	flag.Parse()

	if in == "" {
		fmt.Fprintf(os.Stderr, "input file must be provided\n")
		os.Exit(1)
	}
	if out == "" {
		s, _ := strings.CutSuffix(in, ".go")
		out = s + "_bits.go"
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, in, nil, parser.ParseComments)
	checkf(err, "failed to parse %s", in)

	if pkgname == "" {
		pkgname = node.Name.Name
	}

	// TODO this should be the number of bits in the type (next power of 2)
	const intsize = 16

	// Process the AST
	var structs []StructInfo
	ast.Inspect(node, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structInfo := StructInfo{StructName: t.Name.Name}
		offset := 0
		for _, field := range s.Fields.List {
			blocks := strings.Fields(strings.Trim(field.Tag.Value, "`"))
			for _, b := range blocks {
				if strings.HasPrefix(b, "bitfield:") {
					fieldName := field.Names[0].Name
					val := strings.Trim(b[9:], `"`)
					bitSize, err := strconv.Atoi(val)
					checkf(err, "invalid bit count for field %s: %s", fieldName, val)
					if fieldName != "_" {
						structInfo.Fields = append(structInfo.Fields, Field{
							Name:    fieldName,
							Offset:  offset,
							BitSize: bitSize,
							Mask:    fmt.Sprintf("0x%x", 1<<uint64(bitSize)-1),
						})
					}
					offset += bitSize
				}
			}
		}
		structInfo.Bits = offset
		structs = append(structs, structInfo)
		return false
	})

	if len(structs) == 0 {
		fmt.Fprintf(os.Stderr, "nothing to generate")
		return
	}

	fmt.Fprintf(bb, "package %s\n\n", pkgname)
	for _, si := range structs {
		gprintf(`type %s uint16`, si.StructName)
		for _, fi := range si.Fields {
			// Getter
			gprintf(`func (s %s) %s() uint8 {`, si.StructName, fi.Name)
			gprintf(`return uint8((s >> %d) & %s)`, intsize-(fi.Offset+fi.BitSize), fi.Mask)
			gprintf(`}`)
			gprintf(``)

			// Setter
			shift := intsize - (fi.Offset + fi.BitSize)
			gprintf(`func (s %s) Set%s(val uint8) %s {`, si.StructName, fi.Name, si.StructName)
			mask := fmt.Sprintf("%s << %d", fi.Mask, shift)
			gprintf(`return s ^ %s | (%s(val)&%s)<< %d`, mask, si.StructName, fi.Mask, shift)
			gprintf(`}`)
			gprintf(``)
		}
	}

	buf, err := format.Source(bb.Bytes())
	checkf(err, "go format failed:\n\n%s", bb.String())

	checkf(os.WriteFile(out, buf, 0666), "write failed")
}

var bb = &bytes.Buffer{}

func gprintf(format string, args ...any) {
	fmt.Fprintf(bb, "%s\n", fmt.Sprintf(format, args...))
}

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "bitfield, fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}
