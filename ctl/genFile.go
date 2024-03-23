package ctl

import (
	"bufio"
	"fmt"
	"github.com/luobote55/java2go/gen"
	"github.com/luobote55/java2go/internal/match"
	"github.com/luobote55/java2go/internal/strs"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"sort"
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

type Rpc struct {
	Name       string
	Service    string
	Comment    string
	Http       string
	HttpUrl    string
	Rpc        string
	RequestTyp string
	ReplyTyp   string
}

// run runs the generators in the current file.
func (g *Generator) run(msgs, ctrlNeedMsgs map[string]*Message) (ok bool) {
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
	g.target = strs.JSONSnakeCase(strings.Replace(strings.Replace(g.file, "DO", "", -1), "java", "proto", -1))
	if g.goo == "" {
		g.goo = "./"
	}

	// 检查路径是否存在
	if _, err := os.Stat(g.goo); os.IsNotExist(err) {
		fmt.Println("目标目录并不存在：" + g.goo)
		return false
	}

	replyMsgs := make(map[string]*Message)   // 组合路径和文件名
	requestMsgs := make(map[string]*Message) // 组合路径和文件名
	filepath := filepath.Join(g.goo, g.target)
	// 检查文件是否存在
	if _, err := os.Stat(filepath); err == nil || os.IsExist(err) {
		fmt.Println("文件已经存在：" + filepath + "， 如要更新先删除")
		return false
	}
	// Scan for lines that start "//go:generate".
	// Can't use bufio.Scanner because it can't handle long lines,
	// which are likely to appear when using generate.
	input := bufio.NewReader(g.r)
	var err error
	// One line per loop.
	file := gen.NewGeneratedFile()

	var rpc *Rpc = nil
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
		if strings.HasPrefix(string(buf), "@Api(tags = ") {
			file.ApiModel = match.FindDoubleQuotes(string(buf))
		} else if strings.HasPrefix(string(buf), "@RequestMapping(") {
			file.SetUrl(match.FindDoubleQuote(string(buf)))
		} else if strings.HasPrefix(string(buf), "public class ") {
			file.ServiceName = strings.Replace(match.FindFix(string(buf), `public class (.*?) {`), "Controller", "", 1)
			g.header(file)
		} else if strings.HasPrefix(string(buf), "    @GetMapping(") {
			rpc = new(Rpc)
			rpc.Http = "get"
			rpc.HttpUrl = strings.Replace(match.FindDoubleQuote(string(buf)), "\"", "", -1)
		} else if strings.HasPrefix(string(buf), "    @PostMapping(") {
			rpc = new(Rpc)
			rpc.Http = "post"
			rpc.HttpUrl = strings.Replace(match.FindDoubleQuote(string(buf)), "\"", "", -1)
		} else if strings.HasPrefix(string(buf), "    @ApiOperation(") {
			if rpc == nil {
				continue
			}
			rpc.Comment = match.FindDoubleQuote(string(buf))
		} else if strings.Contains(string(buf), "@RequestBody") {
			if rpc == nil {
				continue
			}
			if rpc.Rpc == "" {
				g.runRpc(rpc, buf)
			}
			if rpc.ReplyTyp == "" {
				g.runReply(rpc, msgs, ctrlNeedMsgs, replyMsgs, buf)
			}
			g.runRequestBody(rpc, msgs, ctrlNeedMsgs, requestMsgs, buf)
			g.rpc(file, rpc)
			rpc = nil
		} else if strings.Contains(string(buf), "@RequestParam") {
			if rpc == nil {
				continue
			}
			if rpc.Rpc == "" {
				g.runRpc(rpc, buf)
			}
			if rpc.ReplyTyp == "" {
				g.runReply(rpc, msgs, ctrlNeedMsgs, replyMsgs, buf)
			}
			g.runRequestParam(rpc, msgs, ctrlNeedMsgs, requestMsgs, buf)
			g.rpc(file, rpc)
			rpc = nil
		} else if strings.Contains(string(buf), "@PathVariable") {
			if rpc == nil {
				continue
			}
			if rpc.Rpc == "" {
				g.runRpc(rpc, buf)
			}
			if rpc.ReplyTyp == "" {
				g.runReply(rpc, msgs, ctrlNeedMsgs, replyMsgs, buf)
			}
			g.runPathVariable(rpc, msgs, ctrlNeedMsgs, requestMsgs, buf)
			g.rpc(file, rpc)
			rpc = nil
		} else if strings.Contains(string(buf), "()") {
			if rpc == nil {
				continue
			}
			if rpc.Rpc == "" {
				g.runRpc(rpc, buf)
			}
			if rpc.ReplyTyp == "" {
				g.runReply(rpc, msgs, ctrlNeedMsgs, replyMsgs, buf)
			}
			g.runRequestVoid(rpc, msgs, ctrlNeedMsgs, requestMsgs, buf)
			g.rpc(file, rpc)
			rpc = nil
		} else {
			if rpc == nil {
				continue
			}
			if rpc.Comment == "" {
				continue
			}
			if strings.Contains(string(buf), "@EventLog") {
				continue
			}
			if strings.Contains(string(buf), "@RequireRole") {
				continue
			}
			g.runRpc(rpc, buf)
			g.runReply(rpc, msgs, ctrlNeedMsgs, replyMsgs, buf)
			requests := match.FindFix(string(buf), `\((.*?)\) {`)
			if requests == "" {
				continue
			}
			request := strings.Split(requests, " ")[0]
			typ, err := JaveType(request)
			if err != nil {
				typ = request
				rpc.RequestTyp = typ
				g.needMsg(msgs, ctrlNeedMsgs, requestMsgs, request)
			} else {
				rpc.RequestTyp = typ
				g.needRequest(requestMsgs, rpc, request)
			}
			g.rpc(file, rpc)
			rpc = nil
		}
	}
	file.P("}")
	file.P("")
	g.message(file, requestMsgs, replyMsgs)

	err = file.WriteFile(filepath)
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func (g *Generator) runRpc(rpc *Rpc, buf []byte) {
	rpcBufs := match.FindFix(string(buf), ` (.*?)\(@RequestBody `)
	if rpcBufs == "" {
		rpcBufs = match.FindFix(string(buf), ` (.*?)\(@RequestParam `)
	}
	if rpcBufs == "" {
		rpcBufs = match.FindFix(string(buf), ` (.*?)\(@PathVariable `)
	}
	if rpcBufs == "" {
		rpcBufs = match.FindFix(string(buf), ` (.*?)\(\) `)
	}
	if rpcBufs == "" {
		rpcBufs = match.FindFix(string(buf), ` (.*?)\(`)
	}
	rpcBufss := strings.Split(rpcBufs, " ")
	rpc.Rpc = strs.GoCamelCase(rpcBufss[len(rpcBufss)-1])
}

func (g *Generator) runReply(rpc *Rpc, msgs, ctrlNeedMsgs map[string]*Message, replyMsgs map[string]*Message, buf []byte) {
	replyBufs := match.FindFix(string(buf), `    public (.*?) `)
	replyBufss := strings.Split(replyBufs, " ")
	reply := replyBufss[0]
	if strings.Contains(reply, "DataGrid<") {
		pageStr := match.FindFix(reply, `(.*?)<`)
		if pageStr == "DataGrid" {
		} else {
			fmt.Println("未识别的变量：" + pageStr)
		}
		value := match.FindFix(reply, `<(.*?)>`)
		typ, err := JaveType(value)
		if err != nil {
			typ = value
		}
		pageMsg, err := g.needPageMsg(msgs, ctrlNeedMsgs, replyMsgs, typ)
		if err != nil {
			return
		}
		rpc.ReplyTyp = rpc.Rpc + "Reply"
		replyMsg := GenMessage(pageMsg, rpc.ReplyTyp).SetChild(pageMsg.StructName)
		replyMsg.ApiModel = []string{rpc.ReplyTyp}
		replyMsg.StructName = rpc.ReplyTyp
		rpcStructer(replyMsg)
		field := &MessageField{
			Name:     pageMsg.StructName,
			Comment:  pageMsg.StructName,
			Typ:      pageMsg.StructName,
			Repeated: "",
		}
		rpcField(replyMsg, field)
		replyMsg.P("}")
		replyMsgs[rpc.ReplyTyp] = replyMsg
	} else if strings.Contains(reply, "List<") {
		pageStr := match.FindFix(reply, `(.*?)<`)
		value := match.FindFix(reply, `<(.*?)>`)
		if pageStr == "List" {
		} else {
			fmt.Println("未识别的变量：" + pageStr)
		}
		typ, err := JaveType(value)
		if err != nil {
			typ = value
		}
		pageMsg, err := g.needListMsg(msgs, ctrlNeedMsgs, replyMsgs, typ)
		if err != nil {
			return
		}
		rpc.ReplyTyp = rpc.Rpc + "Reply"
		replyMsg := GenMessage(pageMsg, rpc.ReplyTyp).SetChild(typ)
		replyMsg.ApiModel = []string{rpc.ReplyTyp}
		replyMsg.StructName = rpc.ReplyTyp
		rpcStructer(replyMsg)
		field := &MessageField{
			Name:     pageMsg.StructName,
			Comment:  pageMsg.StructName,
			Typ:      pageMsg.StructName,
			Repeated: "",
		}
		rpcField(replyMsg, field)
		replyMsg.P("}")
		replyMsgs[rpc.ReplyTyp] = replyMsg
	} else if reply == "int" {
		g.needReply(replyMsgs, rpc, "int32")
	} else if reply == "Integer" {
		g.needReply(replyMsgs, rpc, "int32")
	} else if reply == "String" {
		g.needReply(replyMsgs, rpc, "string")
	} else if reply == "Long" {
		g.needReply(replyMsgs, rpc, "int64")
	} else if reply == "Float" {
		g.needReply(replyMsgs, rpc, "float")
	} else if reply == "Double" {
		g.needReply(replyMsgs, rpc, "double")
	} else if reply == "Boolean" {
		g.needReply(replyMsgs, rpc, "bool")
	} else if reply == "HttpWrapper<?>" {
		g.needReply(replyMsgs, rpc, "string")
	} else if reply == "void" {
		g.needReply(replyMsgs, rpc, "string")
	} else {
		var msg *Message
		replyTyp, err := JaveType(reply)
		if err != nil {
			replyTyp = reply
			msg, err = g.needMsg(msgs, ctrlNeedMsgs, replyMsgs, replyTyp)
			if err != nil {
				return
			}
		}
		rpc.ReplyTyp = rpc.Rpc + "Reply"
		replyMsg := GenMessage(msg, rpc.ReplyTyp).SetChild(replyTyp)
		replyMsg.ApiModel = []string{rpc.ReplyTyp}
		replyMsg.StructName = rpc.ReplyTyp
		rpcStructer(replyMsg)
		if msg == nil {
			field := &MessageField{
				Name:     "data",
				Comment:  "",
				Typ:      replyTyp,
				Repeated: "",
			}
			rpcField(replyMsg, field)
		} else {
			field := &MessageField{
				Name:     msg.StructName,
				Comment:  msg.StructName,
				Typ:      msg.StructName,
				Repeated: "",
			}
			rpcField(replyMsg, field)
		}
		replyMsg.P("}")
		replyMsgs[rpc.ReplyTyp] = replyMsg
	}
	return
}

func (g *Generator) runRequestBody(rpc *Rpc, msgs, ctrlNeedMsgs map[string]*Message, requestMsgs map[string]*Message, buf []byte) {
	request := match.FindFix(string(buf), `@RequestBody (.*?) `)
	if request == "" {
		request = findParam("@RequestBody", string(buf))
	}
	typ, err := JaveType(request)
	if err != nil {
		typ = request
		rpc.RequestTyp = typ
		g.needMsg(msgs, ctrlNeedMsgs, requestMsgs, request)
	} else {
		rpc.RequestTyp = typ
		g.needRequest(requestMsgs, rpc, request)
	}
}

func (g *Generator) runRequestParam(rpc *Rpc, msgs, ctrlNeedMsgs map[string]*Message, requestMsgs map[string]*Message, buf []byte) {
	request := match.FindFix(string(buf), `@RequestParam (.*?) `)
	if request == "" {
		request = findParam("@RequestParam", string(buf))
	}
	typ, err := JaveType(request)
	if err != nil {
		typ = request
		rpc.RequestTyp = typ
		g.needMsg(msgs, ctrlNeedMsgs, requestMsgs, request)
	} else {
		rpc.RequestTyp = typ
		g.needRequest(requestMsgs, rpc, request)
	}
}

func (g *Generator) runPathVariable(rpc *Rpc, msgs, ctrlNeedMsgs map[string]*Message, requestMsgs map[string]*Message, buf []byte) {
	request := match.FindFix(string(buf), `@PathVariable (.*?) `)
	if request == "" {
		request = findParam("@PathVariable", string(buf))
	}
	typ, err := JaveType(request)
	if err != nil {
		typ = request
		rpc.RequestTyp = typ
		g.needMsg(msgs, ctrlNeedMsgs, requestMsgs, request)
	} else {
		rpc.RequestTyp = typ
		g.needRequest(requestMsgs, rpc, request)
	}
}

func findParam(fix string, s string) string {
	ii := strings.Index(s, fix)
	ss := strings.Split(s[ii+len(fix):], " ")
	return ss[1]
}

func (g *Generator) runRequestVoid(rpc *Rpc, msgs, ctrlNeedMsgs map[string]*Message, requestMsgs map[string]*Message, buf []byte) {
	g.needRequest(requestMsgs, rpc, rpc.Name)
}

func (g *Generator) needRequest(requestMsgs map[string]*Message, rpc *Rpc, s string) {
	rpc.RequestTyp = rpc.Rpc + "Request"
	reqMsg := NewMessage()
	reqMsg.ApiModel = []string{rpc.RequestTyp}
	reqMsg.StructName = rpc.RequestTyp
	rpcStructer(reqMsg)
	reqMsg.P("}")
	requestMsgs[rpc.RequestTyp] = reqMsg
}

func (g *Generator) needReply(replyMsgs map[string]*Message, rpc *Rpc, s string) {
	rpc.ReplyTyp = rpc.Rpc + "Reply"
	replyMsg := NewMessage()
	replyMsg.ApiModel = []string{rpc.ReplyTyp}
	replyMsg.StructName = rpc.ReplyTyp
	rpcStructer(replyMsg)
	field := &MessageField{
		Name:     "data",
		Comment:  "",
		Typ:      s,
		Repeated: "",
	}
	rpcField(replyMsg, field)
	replyMsg.P("}")
	replyMsgs[rpc.ReplyTyp] = replyMsg
}

func (g *Generator) needPageMsg(msgs, ctrlNeedMsgs map[string]*Message, needMsgs map[string]*Message, reply string) (*Message, error) {
	msg, err := g.needMsg(msgs, ctrlNeedMsgs, needMsgs, reply)
	if err != nil {
		return nil, err
	}
	pageReply := "Page" + reply
	pageMsg, ok := needMsgs[pageReply]
	if ok {
		return pageMsg, nil
	}
	pageMsg, ok = ctrlNeedMsgs[pageReply]
	if ok {
		return pageMsg, nil
	}
	pageMsg = GenMessage(msg, pageReply).SetChild(reply)
	pageMsg.ApiModel = []string{pageReply}
	pageMsg.StructName = pageReply
	rpcStructer(pageMsg)
	field := &MessageField{
		Name:     msg.StructName,
		Comment:  msg.StructName,
		Typ:      msg.StructName,
		Repeated: "repeated ",
	}
	rpcField(pageMsg, field)
	rpcPage(pageMsg, field)
	pageMsg.P("}")
	needMsgs[pageReply] = pageMsg
	ctrlNeedMsgs[pageReply] = pageMsg
	return pageMsg, nil
}

func (g *Generator) needListMsg(msgs, ctrlNeedMsgs map[string]*Message, needMsgs map[string]*Message, reply string) (*Message, error) {
	var msg *Message
	replyTyp, err := JaveType(reply)
	if err != nil {
		replyTyp = reply
		msg, err = g.needMsg(msgs, ctrlNeedMsgs, needMsgs, replyTyp)
		if err != nil {
			return nil, err
		}
	}
	pageReply := "List" + replyTyp
	pageMsg, ok := needMsgs[pageReply]
	if ok {
		return pageMsg, nil
	}
	pageMsg, ok = ctrlNeedMsgs[pageReply]
	if ok {
		for index := 1; index < 100; index++ {
			pageReplyIndex := fmt.Sprintf("%s%d", pageReply, index)
			pageMsg, ok = ctrlNeedMsgs[pageReplyIndex]
			if !ok {
				pageReply = pageReplyIndex
				break
			}
		}
	}
	pageMsg = GenMessage(msg, pageReply).SetChild(replyTyp)
	pageMsg.ApiModel = []string{pageReply}
	pageMsg.StructName = pageReply
	rpcStructer(pageMsg)
	if msg == nil {
		field := &MessageField{
			Name:     "data",
			Comment:  "",
			Typ:      replyTyp,
			Repeated: "repeated ",
		}
		rpcField(pageMsg, field)
	} else {
		field := &MessageField{
			Name:     msg.StructName,
			Comment:  msg.StructName,
			Typ:      msg.StructName,
			Repeated: "repeated ",
		}
		rpcField(pageMsg, field)
	}
	pageMsg.P("}")
	needMsgs[pageReply] = pageMsg
	ctrlNeedMsgs[pageReply] = pageMsg
	return pageMsg, nil
}

func (g *Generator) needMsg(msgs, ctrlNeedMsgs map[string]*Message, needMsgs map[string]*Message, reply string) (*Message, error) {
	msg, ok := needMsgs[reply]
	if ok {
		return msg, nil
	}
	msg, ok = ctrlNeedMsgs[reply]
	if ok {
		return msg, nil
	}
	_, err := JaveType(reply)
	if err == nil {
		return nil, nil
	}

	msg, ok = msgs[reply]
	if !ok {
		fmt.Println("没有找到这个message：" + reply)
		return nil, errors.New("没有找到这个message：" + reply)
	}
	needMsgs[reply] = msg.GenSort()
	ctrlNeedMsgs[reply] = msg
	for _, s := range msg.Child {
		g.needMsg(msgs, ctrlNeedMsgs, needMsgs, s)
	}
	return msg, nil
}

func rpcPage(pageMsg *Message, field *MessageField) {
	field = &MessageField{
		Name:     "pages",
		Comment:  "pages",
		Typ:      "int32",
		Repeated: "",
	}
	rpcField(pageMsg, field)
	field = &MessageField{
		Name:     "offset",
		Comment:  "offset",
		Typ:      "int32",
		Repeated: "",
	}
	rpcField(pageMsg, field)
	field = &MessageField{
		Name:     "total",
		Comment:  "total",
		Typ:      "int32",
		Repeated: "",
	}
	rpcField(pageMsg, field)
	field = &MessageField{
		Name:     "prePage",
		Comment:  "prePage",
		Typ:      "int32",
		Repeated: "",
	}
	rpcField(pageMsg, field)
	field = &MessageField{
		Name:     "nextPage",
		Comment:  "nextPage",
		Typ:      "int32",
		Repeated: "",
	}
	rpcField(pageMsg, field)
}

func (g *Generator) header(file *gen.GeneratedFile) {
	file.P("// Code generated by j2g. DO NOT EDIT.")
	file.P("// versions:")
	file.P("// - j2g v", version)
	file.P("syntax = \"proto3\";")
	file.P("")
	file.P("package api." + file.Urls[0] + ".v1;")
	file.P("")
	file.P("import \"google/api/annotations.proto\";")
	file.P("//import \"google/protobuf/timestamp.proto\";")
	file.P("")
	file.P("option go_package = \"api/" + file.Urls[0] + "/v1;v1\";")
	file.P("option java_multiple_files = true;")
	file.P("option java_package = \"api." + file.Urls[0] + "\";")
	file.P("")
	file.P("service " + file.ServiceName + " {")
}

func (g *Generator) rpc(file *gen.GeneratedFile, rpc *Rpc) {
	file.P("  // " + rpc.Comment)
	if rpc.HttpUrl == "" {
		file.P("  rpc " + rpc.Rpc + "(" + rpc.RequestTyp + ") returns (" + rpc.ReplyTyp + ");")
		return
	}
	file.P("  rpc " + rpc.Rpc + "(" + rpc.RequestTyp + ") returns (" + rpc.ReplyTyp + "){")
	file.P("    option (google.api.http) = {")
	file.P("      " + rpc.Http + ": \"" + file.Url + rpc.HttpUrl + "\"")
	if rpc.Http == "post" {
		file.P("      body: \"*\"")
	}
	file.P("    };")
	file.P("  };")
}

func (g *Generator) message(file *gen.GeneratedFile, requestMsgs map[string]*Message, replyMsgs map[string]*Message) {
	msgs := make([]*Message, 0)
	for _, msg := range requestMsgs {
		msgs = append(msgs, msg)
	}
	for _, msg := range replyMsgs {
		msgs = append(msgs, msg)
	}
	sort.Slice(msgs, func(ii, jj int) bool {
		return msgs[ii].sortNum < msgs[jj].sortNum
	})
	for _, msg := range msgs {
		file.P(msg.buf.String())
	}
	file.Timestamp()
}

func (g *Generator) fill(file *gen.GeneratedFile, msg *Message) {
}
