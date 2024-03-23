package ctl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/luobote55/java2go/internal/match"
	"github.com/luobote55/java2go/internal/strs"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
	"strings"
)

var sortNum int

type Message struct {
	sortNum    int
	num        int
	ApiModel   []string
	StructName string
	buf        bytes.Buffer
	WithPage   bool
	Child      []string
}

func NewMessage() *Message {
	sortNum++
	return &Message{
		sortNum:    sortNum,
		ApiModel:   nil,
		StructName: "",
		buf:        bytes.Buffer{},
		WithPage:   false,
		Child:      []string{},
	}
}

func GenMessage(msg *Message, reply string) *Message {
	sortNum++
	return &Message{
		sortNum:    sortNum,
		ApiModel:   nil,
		StructName: reply,
		buf:        bytes.Buffer{},
		WithPage:   false,
		Child:      []string{},
	}
}

func (i *Message) GenSort() *Message {
	i.sortNum = sortNum
	sortNum++
	return i
}

// P prints a line to the generated output. It converts each parameter to a
// string following the same rules as fmt.Print. It never inserts spaces
// between parameters.
func (g *Message) P(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		default:
			fmt.Fprint(&g.buf, x)
		}
	}
	fmt.Fprintln(&g.buf)
}

func (i *Message) SetChild(msg string) *Message {
	i.Child = append(i.Child, msg)
	return i
}

type MessageField struct {
	Name     string
	Comment  string
	Typ      string
	Repeated string
}

// A GeneratorMessage represents the state of a single Go source file
// being scanned for generator commands.
type GeneratorMessage struct {
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

// run runs the generators in the current file.
func (g *GeneratorMessage) run(msgs map[string]*Message) (ok bool) {
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

	// Scan for lines that start "//go:generate".
	// Can't use bufio.Scanner because it can't handle long lines,
	// which are likely to appear when using generate.
	input := bufio.NewReader(g.r)
	var err error
	// One line per loop.
	msg := NewMessage()

	var field *MessageField = nil
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

		if strings.HasPrefix(string(buf), "@ApiModel") {
			msg.ApiModel = match.FindDoubleQuotes(string(buf))
		} else if strings.HasPrefix(string(buf), "public class ") {
			msg.StructName = match.FindFix(string(buf), `public class (.*?) `)
			rpcStructer(msg)
			//msg.StructName = "asd"
		} else if strings.Contains(string(buf), "@ApiModelProperty") {
			field = new(MessageField)
			field.Comment = strings.Replace(match.FindDoubleQuote(string(buf)), "\"", "", -1)
		} else if strings.Contains(string(buf), "private ") {
			if field == nil {
				continue
			}
			index := strings.LastIndex(string(buf), " ")
			field.Name = match.FindFix(string(buf[index:]), ` (.*?);`)
			if strings.Contains(string(buf), "List<") {
				field.Repeated = "repeated "
				value := match.FindFix(string(buf), `List<(.*?)>`)
				field.Typ, err = JaveType(value)
				if err != nil {
					field.Typ = value
					msg.SetChild(value)
				}
			} else {
				field.Typ, err = JaveType(string(buf))
				if err != nil {
					field.Typ = strings.Split(match.FindFix(string(buf), `private (.*?);`), " ")[0]
					msg.SetChild(field.Typ)
				}
			}
			rpcField(msg, field)
			field = nil
		}
	}
	//if err != nil && err != io.EOF {
	//	g.errorf("error reading %s: %s", ShortPath(g.path), err)
	//}
	msg.P("}")
	msgs[msg.StructName] = msg

	return true
}

func JaveType(value string) (string, error) {
	if strings.Contains(value, "int") {
		return "int32", nil
	} else if strings.Contains(value, "Integer") {
		return "int32", nil
	} else if strings.Contains(value, "String") {
		return "string", nil
	} else if strings.Contains(value, "string") {
		return "string", nil
	} else if strings.Contains(value, "JSONObject") {
		return "string", nil
	} else if strings.Contains(value, "JSONArray") {
		return "string", nil
	} else if strings.Contains(value, "MultipartFile") {
		return "string", nil
	} else if strings.Contains(value, "HttpServletRequest") {
		return "string", nil
	} else if strings.Contains(value, "?") {
		return "string", nil
	} else if strings.Contains(value, "Long") {
		return "int64", nil
	} else if strings.Contains(value, "int64") {
		return "int64", nil
	} else if strings.Contains(value, "Float") {
		return "float", nil
	} else if strings.Contains(value, "Double") {
		return "double", nil
	} else if strings.Contains(value, "Boolean") {
		return "bool", nil
	} else if strings.Contains(value, "boolean") {
		return "bool", nil
	} else if strings.Contains(value, "Date") {
		return "google.protobuf.Timestamp", nil
	}
	return "", errors.New("没有这个类型：" + value)
}

func rpcStructer(msg *Message) {
	if len(msg.ApiModel) > 0 {
		comment := msg.ApiModel[0]
		for index := 1; index < len(msg.ApiModel); index++ {
			comment += ", " + msg.ApiModel[index]
		}
		msg.P("// " + comment)
	}
	msg.P("message " + match.Remove(msg.StructName, "VO") + " {")
}

func rpcField(msg *Message, field *MessageField) {
	if field.Name == "" {
		return
	}
	commontIndex := 50
	msg.num++
	buf := fmt.Sprintf("  %s%s %s = %d;",
		field.Repeated,
		match.Remove(field.Typ, "VO"),
		strs.LetterCamelCase(match.Remove(field.Name, "VO")),
		msg.num)
	printBufLen := len(buf)
	if printBufLen < commontIndex {
		printBufLen = commontIndex
	}
	// 定义初始值
	initialValue := byte(' ')
	printBuf := bytes.Repeat([]byte{initialValue}, printBufLen)
	copy(printBuf, buf)
	msg.P(string(printBuf) + "// " + match.Remove(field.Comment, "VO"))
}
