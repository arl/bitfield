package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const (
	methodTemplate = `type {{.StructName}} uint16

{{range .Fields}}
func (m {{$.StructName}}) {{.Name}}() uint8 {
	return uint8((m >> {{.Offset}}) & {{.Mask}})
}

func (m *{{$.StructName}}) Set{{.Name}}(value uint8) {
	*m = ({{$.StructName}})((*m & ^({{.Mask}} << {{.Offset}})) | (uint16(value) << {{.Offset}}))
}
{{end}}
`
)

var tmpl = template.Must(template.New("methods").Parse(methodTemplate))

type Field struct {
	Name    string
	Offset  int
	BitSize int
	Mask    uint16
}

type StructInfo struct {
	StructName string
	Fields     []Field
}

func main() {
	var (
		in, out string
		tname   string
	)

	flag.StringVar(&in, "in", "", "input file name")
	flag.StringVar(&out, "out", "", "output file name (defaults to [in]_bits.go)")
	flag.StringVar(&tname, "type", "all", "struct name to read tags from (or all)")
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
					val := strings.Trim(b[9:], `"`)
					bitSize, _ := strconv.Atoi(val)
					mask := (1 << bitSize) - 1
					if field.Names[0].Name != "_" {
						structInfo.Fields = append(structInfo.Fields, Field{
							Name:    field.Names[0].Name,
							Offset:  offset,
							BitSize: bitSize,
							Mask:    uint16(mask),
						})
					}
					offset += bitSize
				}
			}
		}
		structs = append(structs, structInfo)
		return false
	})

	for _, structInfo := range structs {
		file, err := os.Create(out)
		checkf(err, "failed to create %s", out)
		defer file.Close()

		if err := tmpl.Execute(file, structInfo); err != nil {
			panic(err)
		}
	}
}

func checkf(err error, format string, args ...any) {
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "bitfield. fatal error:")
	fmt.Fprintf(os.Stderr, "\n\t%s: %s\n", fmt.Sprintf(format, args...), err)
	os.Exit(1)
}
