package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type config struct {
	in, out string
	tname   string
	pkgname string
}

func parseFlags(args []string) (*config, string, error) {
	var cfg config

	flags := flag.NewFlagSet("bitfield", flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)
	flags.StringVar(&cfg.in, "in", "", "INPUT file name (necessary unless within a go:generate comment)")
	flags.StringVar(&cfg.out, "out", "", "output file name (defaults to standard output)")
	flags.StringVar(&cfg.tname, "type", "all", "name of the type to convert (defaults to all structs)")
	flags.StringVar(&cfg.pkgname, "pkg", "", "package name (defaults to INPUT file package)")
	if err := flags.Parse(args); err != nil {
		return nil, buf.String(), err
	}

	return &cfg, buf.String(), nil
}

func main() {
	cfg, output, err := parseFlags(os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if goFile := os.Getenv("GOFILE"); cfg.in == "" && goFile != "" {
		cfg.in = goFile
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "bitfield, fatal error:\n")
		fmt.Fprintf(os.Stderr, "\t%s\n", err)
		os.Exit(1)
	}
}

type structInfo struct {
	name       string
	width      uint8 // type width in bits
	unions     map[string]*union
	unionOrder []string // unions in file order
}

func newStructInfo(name string) *structInfo {
	return &structInfo{
		name:   name,
		unions: make(map[string]*union),
	}
}

func (si *structInfo) union(name string) *union {
	if u, ok := si.unions[name]; ok {
		return u
	}
	si.unionOrder = append(si.unionOrder, name)
	var u union
	si.unions[name] = &u
	return &u
}

type union struct {
	fields []fieldInfo
	bits   int // bits actually used
}

type fieldInfo struct {
	Name   string
	Mask   uint64
	Offset int
	Type   string // org field type
}

// returns the type bit-width and a boolean indicating if we support it.
func typeWidth(tname string) (int, bool) {
	switch tname {
	case "bool":
		return 1, true
	case "uint8":
		return 8, true
	case "uint16":
		return 16, true
	case "uint32":
		return 32, true
	case "uint64":
		return 64, true
	}
	return 0, false
}

func nextpow2(n uint8) uint8 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n++
	return n
}

func run(cfg *config) error {
	if cfg.in == "" {
		return fmt.Errorf("input file must be provided")
	}

	var out io.Writer = os.Stdout
	if cfg.out != "" {
		f, err := os.Create(cfg.out)
		if err != nil {
			return fmt.Errorf("output file: %s", err)
		}
		defer f.Close()
		out = f
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, cfg.in, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse input file: %s", err)
	}

	var (
		structs []*structInfo
		tErr    error
	)
	ast.Inspect(node, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		if t.Name.Name != cfg.tname && cfg.tname != "all" {
			return true
		}

		offsets := make(map[string]int)
		structInfo := newStructInfo(t.Name.Name)
		for _, field := range s.Fields.List {
			if field.Tag == nil {
				continue
			}
			tags := strings.Fields(strings.Trim(field.Tag.Value, "`"))
			union := "default"
			for _, tag := range tags {
				if !strings.HasPrefix(tag, "bitfield:") {
					continue
				}

				fieldName := field.Names[0].Name
				kvs := strings.Split(strings.Trim(tag[9:], `"`), ",")
				bits := 0
				for _, tag := range kvs {
					k, v, ok := strings.Cut(tag, "=")
					if !ok {
						if bits != 0 {
							tErr = fmt.Errorf("field '%s' has a malformed struct tag", fieldName)
							return false
						}
						k = "bits"
						v = tag
					}
					switch k {
					case "bits":
						ibits, err := strconv.Atoi(v)
						if err != nil {
							tErr = fmt.Errorf("failed to parse bit count for field '%s'", fieldName)
							return false
						}

						if ibits <= 0 || ibits > 64 {
							tErr = fmt.Errorf("field '%s' has an invalid bit count (%d), must be (0, 64]", fieldName, ibits)
							return false
						}
						bits = ibits
					case "union":
						union = v
					}
				}

				if bits == 0 {
					tErr = fmt.Errorf("missing bit count for field '%s': %s", fieldName, kvs)
					return false
				}

				tname := field.Type.(*ast.Ident).Name
				twidth, ok := typeWidth(tname)
				if !ok {
					tErr = fmt.Errorf("field '%s' has an unsupported type %s", fieldName, tname)
					return false
				}
				if twidth < bits {
					tErr = fmt.Errorf("field '%s' can't represent %d bits with type %s", fieldName, bits, tname)
					return false
				}
				if fieldName != "_" {
					u := structInfo.union(union)
					off := offsets[union]
					mask := uint64(1<<uint64(bits) - 1)
					if tname == "bool" {
						mask = 1 << uint64(off)
					}
					u.fields = append(u.fields, fieldInfo{
						Name:   fieldName,
						Offset: off,
						Mask:   mask,
						Type:   tname,
					})
				}
				offsets[union] += bits
			}
		}

		for n, u := range structInfo.unions {
			u.bits = offsets[n]
			structInfo.width = max(structInfo.width, uint8(offsets[n]))
		}
		structInfo.width = nextpow2(structInfo.width)
		structs = append(structs, structInfo)
		return true
	})
	if tErr != nil {
		// Return the error set during the AST traversal.
		return tErr
	}

	if !slices.ContainsFunc(structs, func(si *structInfo) bool {
		return len(si.unions) > 0
	}) {
		return fmt.Errorf("nothing to generate")
	}

	// Generate the file.
	if cfg.pkgname == "" {
		cfg.pkgname = node.Name.Name
	}

	var g generator
	g.printf("package %s\n\n", cfg.pkgname)
	g.printf("// Code generated by github.com/arl/bitfield. DO NOT EDIT.\n")

	for _, si := range structs {
		if len(si.unions) == 0 {
			// skip structs without any fields
			continue
		}
		g.printf(`type %s uint%d`, si.name, si.width)
		for _, uname := range si.unionOrder {
			union := si.unions[uname]
			// Define the final type
			if union.bits > 64 {
				if uname == "default" {
					return fmt.Errorf("struct '%s' has too many bits (%d)", si.name, union.bits)
				}
				return fmt.Errorf("struct '%s' has too many bits in union '%s' (%d)", si.name, uname, union.bits)
			}

			for _, fi := range union.fields {
				// Getter
				g.printf(`func (s %s) %s() %s {`, si.name, fi.Name, fi.Type)
				switch {
				case fi.Type == "bool":
					g.printf(`	return s&0x%x != 0`, fi.Mask)
				case fi.Offset > 0:
					g.printf(`	return %s((s >> %d) & 0x%x)`, fi.Type, fi.Offset, fi.Mask)
				default:
					g.printf(`	return %s(s & 0x%x)`, fi.Type, fi.Mask)
				}
				g.printf(`}`)
				g.printf(``)

				// Setter
				g.printf(`func (s %s) Set%s(val %s) %s {`, si.name, fi.Name, fi.Type, si.name)
				switch {
				case fi.Type == "bool":
					// The generated assembly doesn't branch.
					g.printf(`	var ival %s`, si.name)
					g.printf(`	if val {`)
					g.printf(`		ival = 1`)
					g.printf(`	}`)
					g.printf(`return s&^0x%x | ival<<%d`, fi.Mask, fi.Offset)
				case fi.Offset > 0:
					g.printf(`	return s &^ (0x%x<<%d) | (%s(val&0x%x)<<%d)`, fi.Mask, fi.Offset, si.name, fi.Mask, fi.Offset)
				default:
					g.printf(`	return s &^ 0x%x | %s(val&0x%x)`, fi.Mask, si.name, fi.Mask)
				}
				g.printf(`}`)
				g.printf(``)
			}
		}
	}

	return g.format(out)
}

type generator struct {
	buf bytes.Buffer
}

func (g *generator) printf(format string, args ...any) {
	fmt.Fprintf(&g.buf, format+"\n", args...)
}

func (g *generator) format(w io.Writer) error {
	buf, err := format.Source(g.buf.Bytes())
	if err != nil {
		return fmt.Errorf("go format failed: %s", err)
	}
	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("write failed: %s", err)
	}
	return nil
}
