package sql

import (
	"bufio"
	"fmt"
	"github.com/luobote55/java2go/gen"
	"github.com/luobote55/java2go/internal/match"
	"github.com/luobote55/java2go/internal/strs"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	version = "1.0.0"
)

// A Generator represents the state of a single Go source file
// being scanned for generator commands.
type Generator struct {
	r        io.Reader
	path     string // full rooted path name.
	dir      string // full rooted directory of file.
	file     string // base name of file.
	target   string
	goo      string
	pkg      string
	commands map[string][]string
	lineNum  int // current line number.
	env      []string
}

type EntField struct {
	Name        string
	Comment     string
	Typ         string
	DefaultNull bool
	Nillable    bool
}

// run runs the generators in the current file.
func (g *Generator) run() (ok bool) {
	// Processing below here calls g.errorf on failure, which does panic(stop).
	// If we encounter an error, we abort the package.
	defer func() {
		e := recover()
		if e != nil {
			ok = false
			if e != stop {
				panic(e)
			}
		}
	}()
	g.dir, g.file = filepath.Split(g.path)
	g.dir = filepath.Clean(g.dir) // No final separator please.
	g.target = strs.JSONSnakeCase(strings.Replace(strings.Replace(g.file, "DO", "", -1), "java", "go", -1))
	if g.goo == "" {
		g.goo = "./"
	}

	// 检查路径是否存在
	if _, err := os.Stat(g.goo); os.IsNotExist(err) {
		fmt.Println("目标目录并不存在：" + g.goo)
		return false
	}

	// Scan for lines that start "//go:generate".
	// Can't use bufio.Scanner because it can't handle long lines,
	// which are likely to appear when using generate.
	input := bufio.NewReader(g.r)
	var err error
	// One line per loop.

	for {
		g.lineNum++ // 1-indexed.
		var buf []byte
		buf, err = input.ReadSlice('\n')
		if err != nil {
			// Check for marker at EOF without final \n.
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}

		if strings.HasPrefix(string(buf), "CREATE TABLE") {
			g.runTable(buf, input)
		}
	}
	return true
}

// run runs the generators in the current file.
func (g *Generator) runTable(buf []byte, input *bufio.Reader) (ok bool) {
	// Processing below here calls g.errorf on failure, which does panic(stop).
	// If we encounter an error, we abort the package.
	defer func() {
		e := recover()
		if e != nil {
			ok = false
			if e != stop {
				panic(e)
			}
		}
	}()

	tableName := strings.Replace(match.FindBacktick(string(buf)), "`", "", -1)

	// 组合路径和文件名
	filepath := filepath.Join(g.goo, tableName+".go")
	// 检查文件是否存在
	if _, err := os.Stat(filepath); err == nil || os.IsExist(err) {
		fmt.Println("文件已经存在：" + filepath + "， 如要更新先删除")
		return false
	}
	fmt.Println("写入文件：" + filepath)
	// One line per loop.
	file := gen.NewGeneratedFile()
	file.TableName = tableName
	file.StructName = strs.GoCamelCase(tableName)
	var err error
	g.header(file)
	g.structer(file)

	var field *EntField = nil
	for {
		g.lineNum++ // 1-indexed.
		var buf []byte
		buf, err = input.ReadSlice('\n')
		if err != nil {
			// Check for marker at EOF without final \n.
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			break
		}
		if strings.Contains(string(buf), ") ENGINE =") {
			comment := match.FindFix(string(buf), `COMMENT = \'(.*?)\'`)
			file.Replace("commentcommentcommentcommentcomment", comment)
			break
		} else if strings.Contains(string(buf), "AUTO_INCREMENT") {
			file.P("\t\tfield.Int64(\"id\").Comment(\"id\"),")
			field = nil
		} else if strings.Contains(string(buf), "deleted") {
			continue
		} else if strings.Contains(string(buf), "create_time") {
			continue
		} else if strings.Contains(string(buf), "update_time") {
			continue
		} else if strings.Contains(string(buf), "bigint(") {
			g.RunBigInt(file, buf)
		} else if strings.Contains(string(buf), "int(") {
			g.RunInt(file, buf)
		} else if strings.Contains(string(buf), "varchar(") {
			g.RunString(file, buf)
		} else if strings.Contains(string(buf), "char(") {
			g.RunString(file, buf)
		} else if strings.Contains(string(buf), "text") {
			g.RunBytes(file, buf)
		} else if strings.Contains(string(buf), "double") {
			g.RunFloat(file, buf)
		} else if strings.Contains(string(buf), " INDEX ") {
			g.RunIndex(file, buf)
		} else if strings.Contains(string(buf), "datetime(") {
			g.RunTime(file, buf)
		} else {
			if strings.Contains(string(buf), "    `") {
				fmt.Println("暂不支持的类型：" + string(buf))
				field = nil
			}
			continue
		}
	}

	//if err != nil && err != io.EOF {
	//	g.errorf("error reading %s: %s", ShortPath(g.path), err)
	//}

	field = new(EntField)
	field.Comment = "创建时间"
	field.Name = "\"created_at\""
	field.Typ = "Time"
	g.field(file, field)

	field.Comment = "修改时间"
	field.Name = "\"updated_at\""
	field.Typ = "Time"
	g.field(file, field)

	field.Comment = "删除"
	field.Name = "\"deleted_at\""
	field.Typ = "Time"
	g.field(file, field)
	file.P("\t}")
	file.P("}")
	file.P("")

	g.index(file)

	err = file.WriteFile(filepath)
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func (g *Generator) RunBigInt(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Int64"
	defaultString := ".Default(0)"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT`)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			num, err := strconv.Atoi(defaul)
			if err == nil {
				defaultString = fmt.Sprintf(".Default(%d)", num)
			}
			nillableString = nillableString + defaultString
		}
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunInt(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Int32"
	defaultString := ".Default(0)"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT`)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			num, err := strconv.Atoi(defaul)
			if err == nil {
				defaultString = fmt.Sprintf(".Default(%d)", num)
			}
			nillableString = nillableString + defaultString
		}
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunString(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "String"
	defaultString := ".Default(\"\")"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT `)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			defaultString = fmt.Sprintf(".Default(\"%s\")", defaul)
			nillableString = nillableString + defaultString
		}
	}
	stringLen := match.FindFix(string(buf), `char\((.*?)\)`)
	num, err := strconv.Atoi(stringLen)
	if err == nil {
		nillableString = fmt.Sprintf("%s.MaxLen(%d)", nillableString, num)
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunBytes(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Bytes"
	defaultString := ".Default(\"\")"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT `)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			defaultString = fmt.Sprintf(".Default(\"%s\")", defaul)
			nillableString = nillableString + defaultString
		}
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunFloat(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Float"
	defaultString := ".Default(0)"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT`)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			num, err := strconv.Atoi(defaul)
			if err == nil {
				defaultString = fmt.Sprintf(".Default(%d)", num)
			}
			nillableString = nillableString + defaultString
		}
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunFloat32(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Float32"
	defaultString := ".Default(0)"
	nillableString := ""
	nill := match.FindFix(string(buf), `\) (.*?) DEFAULT`)
	if strings.Contains(nill, "NULL") {
		field.Nillable = true
		nillableString = ".Nillable()"
		defaul := match.FindFix(string(buf), `DEFAULT (.*?) `)
		if strings.Contains(defaul, "NULL") {
			field.DefaultNull = true
			defaultString = ""
		} else {
			num, err := strconv.Atoi(defaul)
			if err == nil {
				defaultString = fmt.Sprintf(".Default(%d)", num)
			}
			nillableString = nillableString + defaultString
		}
	}

	file.P("\t\tfield." + field.Typ + "(" + field.Name + ")" + nillableString + ".Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunTime(file *gen.GeneratedFile, buf []byte) (ok bool) {
	var field *EntField = new(EntField)
	strss := match.FindBackticks(string(buf))
	if len(strss) > 0 {
		field.Name = strings.Replace(strss[0], "`", "\"", -1)
	}
	field.Comment = strings.Replace(match.FindQuote(string(buf)), "'", "\"", -1)
	field.Typ = "Time"
	file.P("\t\tfield." + field.Typ + "(" + field.Name + ").")
	file.P("\t\t\tOptional().Nillable().")
	file.P("\t\t\tSchemaType(map[string]string{")
	file.P("\t\t\t\tdialect.MySQL:  \"datetime\",")
	file.P("\t\t\t\tdialect.SQLite: \"datetime\",")
	file.P("\t\t\t}).Comment(" + field.Comment + "),")
	return true
}

func (g *Generator) RunIndex(file *gen.GeneratedFile, buf []byte) (ok bool) {
	gg := gen.GeneratedIndex{
		StorageKey: "",
		Fields:     []string{},
		UNIQUE:     false,
	}
	if strings.Contains(string(buf), "UNIQUE") {
		gg.UNIQUE = true
	}
	indexs := match.FindBackticks(string(buf))
	if len(indexs) < 2 {
		return false
	}
	gg.StorageKey = strings.Replace(indexs[0], "`", "", -1)
	for i := 1; i < len(indexs); i++ {
		gg.Fields = append(gg.Fields, strings.Replace(indexs[i], "`", "", -1))
	}
	file.Index = append(file.Index, gg)
	return true
}

func (g *Generator) header(file *gen.GeneratedFile) {
	file.P("// Code generated by j2g. DO NOT EDIT.")
	file.P("// versions:")
	file.P("// - j2g v", version)
	file.P("package schema                       ")
	file.P("                                     ")
	file.P("import (")
	file.P("\t//\"entgo.io/ent/schema/index\"")
	file.P("\t\"time\"")
	file.P("")
	file.P("\t\"entgo.io/ent\"")
	file.P("\t\"entgo.io/ent/dialect\"")
	file.P("\t\"entgo.io/ent/schema/field\"")
	file.P(")")
	file.P("")
}

func (g *Generator) structer(file *gen.GeneratedFile) {
	if len(file.ApiModel) > 0 {
		comment := file.ApiModel[0]
		for index := 1; index < len(file.ApiModel); index++ {
			comment += ", " + file.ApiModel[index]
		}
		file.P("// " + comment)
	} else {
		file.P("// commentcommentcommentcommentcomment")
	}
	file.P("// " + file.StructName + " holds the schema definition for the " + file.StructName + " entity.")
	file.P("type " + file.StructName + " struct {")
	file.P("\tent.Schema")
	file.P("}")
	file.P("")
	file.P("// Fields of the " + file.StructName + ".")
	file.P("func (" + file.StructName + ") Fields() []ent.Field {")
	file.P("\treturn []ent.Field{")
}

func (g *Generator) index(file *gen.GeneratedFile) {
	if len(file.Index) == 0 {
		return
	}
	file.Replace("\t//\"entgo.io/ent/schema/index\"", "\t\"entgo.io/ent/schema/index\"")

	file.P("// Indexes of the " + file.StructName + " .")
	file.P("func (" + file.StructName + ") Indexes() []ent.Index {")
	file.P("\treturn []ent.Index{")
	for i := 0; i < len(file.Index); i++ {
		buf := "\t\tindex"
		for j := 0; j < len(file.Index[i].Fields); j++ {
			buf += ".Fields(\"" + file.Index[i].Fields[j] + "\")"
		}
		buf += ".StorageKey(\"" + file.Index[i].StorageKey + "\"),"
		file.P(buf)
	}
	file.P("\t}")
	file.P("}")
	file.P("")
}

func (g *Generator) field(file *gen.GeneratedFile, field *EntField) {
	if field.Name == "" {
		return
	}
	if field.Typ == "Int32" {
		file.P("\t\tfield." + field.Typ + "(" + field.Name + ").Default(0).Comment(" + field.Comment + "),")
	} else if field.Typ == "Int64" {
		file.P("\t\tfield." + field.Typ + "(" + field.Name + ").Default(0).Comment(" + field.Comment + "),")
	} else if field.Typ == "Float32" { // 32bit
		file.P("\t\tfield." + field.Typ + "(" + field.Name + ").Default(0).Comment(" + field.Comment + "),")
	} else if field.Typ == "Float" { // 64bit
		file.P("\t\tfield." + field.Typ + "(" + field.Name + ").Default(0).Comment(" + field.Comment + "),")
	} else if field.Typ == "String" {
		file.P("\t\tfield." + field.Typ + "(" + field.Name + ").Default(\"\").Comment(" + field.Comment + "),")
	} else if field.Typ == "Time" {
		if field.Name == "created_at" {
			file.P("\t\tfield." + field.Typ + "(" + field.Name + ").")
			file.P("\t\t\tDefault(time.Now).")
			file.P("\t\t\tSchemaType(map[string]string{")
			file.P("\t\t\t\tdialect.MySQL:  \"datetime\",")
			file.P("\t\t\t\tdialect.SQLite: \"datetime\",")
			file.P("\t\t\t}),")
		} else if field.Name == "updated_at" {
			file.P("\t\tfield." + field.Typ + "(" + field.Name + ").")
			file.P("\t\t\tDefault(time.Now).")
			file.P("\t\t\tSchemaType(map[string]string{")
			file.P("\t\t\t\tdialect.MySQL:  \"datetime\",")
			file.P("\t\t\t\tdialect.SQLite: \"datetime\",")
			file.P("\t\t\t}),")
		} else {
			file.P("\t\tfield." + field.Typ + "(" + field.Name + ").")
			file.P("\t\t\tOptional().Nillable().")
			file.P("\t\t\tDefault(time.Now).")
			file.P("\t\t\tSchemaType(map[string]string{")
			file.P("\t\t\t\tdialect.MySQL:  \"datetime\",")
			file.P("\t\t\t\tdialect.SQLite: \"datetime\",")
			file.P("\t\t\t}),")
		}
	}
}
