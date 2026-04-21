package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeFlag   = flag.String("type", "", "comma-separated list of type names; must be set")
	outputFlag = flag.String("output", "", "output file name; default srcdir/<type>_fields.go")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of bitfield:\n")
	fmt.Fprintf(os.Stderr, "\tbitfield [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tbitfield [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://pkg.go.dev/github.com/arl/bitfield/v2\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("bitfield: ")
	flag.Usage = Usage
	flag.Parse()

	if len(*typeFlag) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	names := strings.Split(*typeFlag, ",")

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var dir string
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		dir = filepath.Dir(args[0])
	}

	pkg := loadPackage(args)

	specs, err := analyzePackage(pkg, names)
	if err != nil {
		log.Fatal(err)
	}

	outputName := *outputFlag
	if outputName == "" {
		baseName := strings.ToLower(names[0]) + "_fields.go"
		outputName = filepath.Join(dir, baseName)
	}

	// The generated file declares `package <pkg.Name>`, so it must live in
	// the same directory as the source package; otherwise the Go toolchain
	// will reject the result.
	outAbs, err := filepath.Abs(outputName)
	if err != nil {
		log.Fatalf("resolving output path: %s", err)
	}
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolving source directory: %s", err)
	}
	if filepath.Dir(outAbs) != dirAbs {
		log.Fatalf("-output must live in the source package directory %s, got %s", dirAbs, outAbs)
	}

	f, err := os.Create(outputName)
	if err != nil {
		log.Fatalf("creating output: %s", err)
	}
	defer f.Close()

	if err := cliGenerate(f, pkg.Name, specs); err != nil {
		log.Fatalf("generating code: %s", err)
	}
}

// isDirectory reports whether the named file is a directory.
func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

// loadPackage loads and returns the single package described by patterns.
// It exits on error.
func loadPackage(patterns []string) *packages.Package {
	// go/packages requires relative paths to start with "./" or "../".
	// Normalize plain directory names (e.g. "foo") to "./foo".
	normalized := make([]string, len(patterns))
	for i, p := range patterns {
		if !filepath.IsAbs(p) && !strings.HasPrefix(p, ".") {
			normalized[i] = "./" + p
		} else {
			normalized[i] = p
		}
	}

	cfg := &packages.Config{
		// NeedSyntax/NeedTypesInfo force the loader to type-check from source
		// rather than pulling export data, so unexported types and fields are
		// visible in the package scope.
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, normalized...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) == 0 {
		log.Fatalf("no packages found matching %v", strings.Join(patterns, " "))
	}

	var hasErrors bool
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			fmt.Fprintln(os.Stderr, e)
			hasErrors = true
		}
	}
	if hasErrors {
		os.Exit(1)
	}

	// Return the first non-test package.
	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.Name, "_test") {
			return pkg
		}
	}
	return pkgs[0]
}

// analyzePackage builds a typeSpec for each named type found in pkg.
func analyzePackage(pkg *packages.Package, typeNames []string) ([]typeSpec, error) {
	specs := make([]typeSpec, 0, len(typeNames))
	for _, name := range typeNames {
		s, err := analyzeType(pkg, name)
		if err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, nil
}

// analyzeType returns a typeSpec for the named struct type in pkg.
func analyzeType(pkg *packages.Package, typeName string) (typeSpec, error) {
	if pkg.Types == nil {
		return typeSpec{}, fmt.Errorf("type information unavailable for package %s", pkg.Name)
	}
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return typeSpec{}, fmt.Errorf("type %s not found in package %s", typeName, pkg.Name)
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return typeSpec{}, fmt.Errorf("%s is not a type name", typeName)
	}
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return typeSpec{}, fmt.Errorf("%s is not a named type", typeName)
	}
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return typeSpec{}, fmt.Errorf("%s is not a struct", typeName)
	}

	spec := typeSpec{Name: typeName, Exported: tn.Exported()}
	var offset uint

	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		tag := st.Tag(i)

		// Blank-identifier fields (`_`) are allowed as explicit padding: they
		// reserve bits in the layout but emit no code referencing the field
		// (which wouldn't compile anyway since `_` isn't addressable).
		fieldLabel := field.Name()
		if fieldLabel == "_" {
			fieldLabel = fmt.Sprintf("_ (field %d)", i)
		}

		bitfieldTag, ok := reflect.StructTag(tag).Lookup("bitfield")
		if !ok {
			return typeSpec{}, fmt.Errorf("bitfield: %s.%s is missing a bitfield tag", typeName, fieldLabel)
		}
		width, err := parseWidth(bitfieldTag)
		if err != nil {
			return typeSpec{}, fmt.Errorf("bitfield: %s.%s: %w", typeName, fieldLabel, err)
		}
		fs, err := inspectFieldFromTypes(field, width)
		if err != nil {
			return typeSpec{}, fmt.Errorf("bitfield: %s.%s: %w", typeName, fieldLabel, err)
		}
		fs.Offset = offset
		spec.Fields = append(spec.Fields, fs)
		offset += width
	}

	spec.Total = offset
	storage, err := storageFor(offset)
	if err != nil {
		return typeSpec{}, fmt.Errorf("bitfield: %s: %w", typeName, err)
	}
	spec.Storage = storage
	return spec, nil
}

// inspectFieldFromTypes builds a fieldSpec from a types.Var and its declared width.
func inspectFieldFromTypes(field *types.Var, width uint) (fieldSpec, error) {
	fs := fieldSpec{Name: field.Name(), Blank: field.Name() == "_"}
	ft := field.Type()

	// For named types (e.g. type Mode uint8), preserve the declared name.
	// For unnamed built-ins (e.g. uint8 directly), use the basic type name.
	if named, ok := ft.(*types.Named); ok {
		fs.TypeName = named.Obj().Name()
	} else if basic, ok := ft.(*types.Basic); ok {
		fs.TypeName = basic.Name()
	} else {
		return fs, fmt.Errorf("unsupported unnamed type")
	}

	basic, ok := ft.Underlying().(*types.Basic)
	if !ok {
		return fs, fmt.Errorf("unsupported type %s; must have underlying bool, uint8, uint16, uint32, or uint64", ft)
	}

	switch basic.Kind() {
	case types.Bool:
		if width != 1 {
			return fs, fmt.Errorf("bool fields must have width 1, got %d", width)
		}
		fs.Kind = reflect.Bool
		fs.KindWidth = 1
	case types.Uint8:
		fs.Kind = reflect.Uint8
		fs.KindWidth = 8
	case types.Uint16:
		fs.Kind = reflect.Uint16
		fs.KindWidth = 16
	case types.Uint32:
		fs.Kind = reflect.Uint32
		fs.KindWidth = 32
	case types.Uint64:
		fs.Kind = reflect.Uint64
		fs.KindWidth = 64
	default:
		return fs, fmt.Errorf("unsupported type %s; supported: bool, uint8, uint16, uint32, uint64 and named types over them", ft)
	}

	if width > fs.KindWidth {
		return fs, fmt.Errorf("width %d exceeds field type %s width of %d bits", width, ft, fs.KindWidth)
	}
	fs.Width = width
	return fs, nil
}

// cliGenerate writes generated Go source for the given type specs to w.
func cliGenerate(w io.Writer, pkgName string, specs []typeSpec) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)
	fmt.Fprintf(&buf, "// Code generated by \"bitfield %s\"; DO NOT EDIT.\n\n", strings.Join(os.Args[1:], " "))
	for _, s := range specs {
		writeType(&buf, s)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		// Return the unformatted source so the user can inspect the problem.
		_, _ = w.Write(buf.Bytes())
		return fmt.Errorf("generated code failed to format: %w", err)
	}
	_, err = w.Write(src)
	return err
}
