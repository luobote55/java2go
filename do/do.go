package do

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// CmdDo represents the source command.
var CmdDo = &cobra.Command{
	Use:   "do",
	Short: "Generate the ent schema code from xxxDO.java",
	Long:  "Generate the ent schema code from xxxDO.java. Example: ./j2g.exe do ./test/do ./test/do ",
	Run:   run,
}

var javaPath string

func init() {
	CmdDo.Flags().StringVarP(&javaPath, "java_path", "p", "./", "java source directory")
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please enter the java file or directory")
		return
	}
	var (
		err  error
		java = strings.TrimSpace(args[0])
		goo  = strings.TrimSpace(args[1])
	)
	if strings.HasSuffix(java, ".java") {
		err = generate(java, goo, args)
	} else {
		err = walk(java, goo, args)
	}
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

func walk(dir string, goo string, args []string) error {
	if dir == "" {
		dir = "."
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext != ".java" {
			return nil
		}
		return generate(path, goo, args)
	})
}

// generate is used to execute the generate command for the specified proto file
func generate(java string, goo string, args []string) error {
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
	g.run()
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
