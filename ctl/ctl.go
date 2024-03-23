package ctl

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

// CmdCtl represents the source command.
var CmdCtl = &cobra.Command{
	Use:   "ctl",
	Short: "Generate the protobuf code from xxxController.java",
	Long:  "Generate the protobuf code from xxxController.java. Example: ./j2g.exe ctl ./test/ctl ./test/ctl ",
	Run:   run,
}

var (
	controllerPath string
	voPath         string
	requestPath    string
	protoPath      string
)

func init() {
	CmdCtl.Flags().StringVarP(&controllerPath, "controller_path", "c", "./", "java controller source directory")
	CmdCtl.Flags().StringVarP(&voPath, "vo_path", "v", "./", "java vo source directory")
	CmdCtl.Flags().StringVarP(&requestPath, "request_path", "r", "./", "java request source directory")
	CmdCtl.Flags().StringVarP(&protoPath, "proto_path", "p", "./", "protobuf file directory")
}

func run(_ *cobra.Command, args []string) {
	if len(controllerPath) == 0 {
		fmt.Println("Please enter the controllerPath")
		return
	}
	if len(voPath) == 0 {
		fmt.Println("Please enter the voPath")
		return
	}
	if len(requestPath) == 0 {
		fmt.Println("Please enter the requestPath")
		return
	}
	if len(protoPath) == 0 {
		fmt.Println("Please enter the protoPath")
		return
	}
	var err error
	msgs := make(map[string]*Message, 0)
	ctrlNeedMsgs := make(map[string]*Message, 0)
	err = walkVo(voPath, msgs, ctrlNeedMsgs)
	if err != nil {
		fmt.Println(err)
	}
	err = walkVo(requestPath, msgs, ctrlNeedMsgs)
	if err != nil {
		fmt.Println(err)
	}
	err = walk(controllerPath, protoPath, msgs, ctrlNeedMsgs)
	if err != nil {
		fmt.Println(err)
	}
}

func look(name ...string) error {
	for _, n := range name {
		if _, err := exec.LookPath(n); err != nil {
			return err
		}
	}
	return nil
}

func walkVo(dir string, msgs, ctrlNeedMsgs map[string]*Message) error {
	if dir == "" {
		dir = "."
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext != ".java" {
			return nil
		}
		if info.IsDir() {
			walkVo(path, msgs, ctrlNeedMsgs)
		}
		return generateVo(path, msgs, ctrlNeedMsgs)
	})
}

func generateVo(path string, msgs, ctrlNeedMsgs map[string]*Message) error {
	protoBytes, err := os.ReadFile(path)
	if err != nil && len(protoBytes) < 0 {
	}
	g := &GeneratorMessage{
		r:        bytes.NewReader(protoBytes),
		path:     path,
		dir:      "",
		file:     "",
		target:   "",
		pkg:      "",
		commands: nil,
		lineNum:  0,
		env:      nil,
	}
	g.run(msgs)
	return nil
}

func walk(dir string, goo string, msgs, ctrlNeedMsgs map[string]*Message) error {
	if dir == "" {
		dir = "."
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext != ".java" {
			return nil
		}
		if info.IsDir() {
			walk(path, goo, msgs, ctrlNeedMsgs)
		}
		return generate(path, goo, msgs, ctrlNeedMsgs)
	})
}

// generate is used to execute the generate command for the specified proto file
func generate(java string, goo string, msgs, ctrlNeedMsgs map[string]*Message) error {
	protoBytes, err := os.ReadFile(java)
	if err != nil && len(protoBytes) < 0 {
	}
	g := &Generator{
		r:        bytes.NewReader(protoBytes),
		path:     java,
		dir:      "",
		file:     "",
		target:   "",
		goo:      goo,
		pkg:      "",
		commands: nil,
		lineNum:  0,
		env:      nil,
	}
	g.run(msgs, ctrlNeedMsgs)
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

var stop = fmt.Errorf("error in generation")
