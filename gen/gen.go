package gen

import (
	"bytes"
	"fmt"
	"github.com/luobote55/java2go/internal/strs"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

// A GoIdent is a Go identifier, consisting of a name and import path.
// The name is a single identifier and may not be a dot-qualified selector.
type GoIdent struct {
	GoName       string
	GoImportPath GoImportPath
}

func (id GoIdent) String() string { return fmt.Sprintf("%q.%v", id.GoImportPath, id.GoName) }

// newGoIdent returns the Go identifier for a descriptor.
func newGoIdent(f *File, d protoreflect.Descriptor) GoIdent {
	name := strings.TrimPrefix(string(d.FullName()), string(f.Desc.Package())+".")
	return GoIdent{
		GoName:       strs.GoCamelCase(name),
		GoImportPath: f.GoImportPath,
	}
}

// A GoPackageName is the name of a Go package. e.g., "protobuf".
type GoPackageName string

// cleanPackageName converts a string to a valid Go package name.
func cleanPackageName(name string) GoPackageName {
	return GoPackageName(strs.GoSanitized(name))
}

// A GoImportPath is the import path of a Go package.
// For example: "google.golang.org/protobuf/compiler/protogen"
type GoImportPath string

func (p GoImportPath) String() string { return strconv.Quote(string(p)) }

// Ident returns a GoIdent with s as the GoName and p as the GoImportPath.
func (p GoImportPath) Ident(s string) GoIdent {
	return GoIdent{GoName: s, GoImportPath: p}
}

type pathType int

const (
	pathTypeImport pathType = iota
	pathTypeSourceRelative
)

// A File describes a .proto source file.
type File struct {
	Desc  protoreflect.FileDescriptor
	Proto *descriptorpb.FileDescriptorProto

	GoDescriptorIdent GoIdent       // name of Go variable for the file descriptor
	GoPackageName     GoPackageName // name of this file's Go package
	GoImportPath      GoImportPath  // import path of this file's Go package
}

// An Annotation provides semantic detail for a generated proto element.
//
// See the google.protobuf.GeneratedCodeInfo.Annotation documentation in
// descriptor.proto for details.
type Annotation struct {
	// Location is the source .proto file for the element.
	Location Location

	// Semantic is the symbol's effect on the element in the original .proto file.
	Semantic *descriptorpb.GeneratedCodeInfo_Annotation_Semantic
}

// A Location is a location in a .proto source file.
//
// See the google.protobuf.SourceCodeInfo documentation in descriptor.proto
// for details.
type Location struct {
	SourceFile string
	Path       protoreflect.SourcePath
}

// appendPath add elements to a Location's path, returning a new Location.
func (loc Location) appendPath(num protoreflect.FieldNumber, idx int) Location {
	loc.Path = append(protoreflect.SourcePath(nil), loc.Path...) // make copy
	loc.Path = append(loc.Path, int32(num), int32(idx))
	return loc
}

type GeneratedIndex struct {
	StorageKey string
	Fields     []string
	UNIQUE     bool
}

type GeneratedFile struct {
	ApiModel         []string
	ServiceName      string
	TableName        string
	StructName       string
	Url              string
	Urls             []string
	Index            []GeneratedIndex
	goImportPath     GoImportPath
	buf              bytes.Buffer
	packageNames     map[GoImportPath]GoPackageName
	usedPackageNames map[GoPackageName]bool
	manualImports    map[GoImportPath]bool
	annotations      map[string][]Annotation
}

func NewGeneratedFile() *GeneratedFile {
	return &GeneratedFile{
		ApiModel:         nil,
		TableName:        "",
		StructName:       "",
		Url:              "",
		Urls:             nil,
		Index:            []GeneratedIndex{},
		goImportPath:     "",
		buf:              bytes.Buffer{},
		packageNames:     nil,
		usedPackageNames: nil,
		manualImports:    nil,
		annotations:      nil,
	}
}

// P prints a line to the generated output. It converts each parameter to a
// string following the same rules as fmt.Print. It never inserts spaces
// between parameters.
func (g *GeneratedFile) P(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case GoIdent:
			fmt.Fprint(&g.buf, g.QualifiedGoIdent(x))
		default:
			fmt.Fprint(&g.buf, x)
		}
	}
	fmt.Fprintln(&g.buf)
}

// QualifiedGoIdent returns the string to use for a Go identifier.
//
// If the identifier is from a different Go package than the generated file,
// the returned name will be qualified (package.name) and an import statement
// for the identifier's package will be included in the file.
func (g *GeneratedFile) QualifiedGoIdent(ident GoIdent) string {
	if ident.GoImportPath == g.goImportPath {
		return ident.GoName
	}
	if packageName, ok := g.packageNames[ident.GoImportPath]; ok {
		return string(packageName) + "." + ident.GoName
	}
	packageName := cleanPackageName(path.Base(string(ident.GoImportPath)))
	for i, orig := 1, packageName; g.usedPackageNames[packageName]; i++ {
		packageName = orig + GoPackageName(strconv.Itoa(i))
	}
	g.packageNames[ident.GoImportPath] = packageName
	g.usedPackageNames[packageName] = true
	return string(packageName) + "." + ident.GoName
}

func (g *GeneratedFile) WriteFile(filepath string) error {
	return ioutil.WriteFile(filepath, g.buf.Bytes(), 0644)
}

func (g *GeneratedFile) Replace(src, des string) error {
	g.buf = *bytes.NewBufferString(strings.Replace(g.buf.String(), src, des, -1))
	return nil
}

func (g *GeneratedFile) SetUrl(url string) error {
	g.Url = strings.Replace(url, "\"", "", -1)
	urls := strings.Split(g.Url, "/")
	if urls[0] == "" {
		g.Urls = urls[1:]
	} else {
		g.Urls = urls
	}
	if len(g.Urls) == 0 {
		fmt.Println("url为空:" + g.ApiModel[0])
		return errors.New("url为空:" + g.ApiModel[0])
	}
	g.StructName = strs.GoCamelCase(strings.Replace(strings.Replace(g.Urls[len(g.Urls)-1], "-", "", -1), "\"", "", -1))
	return nil
}

func (g *GeneratedFile) Timestamp() {
	if !strings.Contains(string(g.buf.String()), "google.protobuf.Timestamp") {
		return
	}
	g.Replace("//import \"google/protobuf/timestamp.proto\";", "import \"google/protobuf/timestamp.proto\";")
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
