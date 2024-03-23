package sql

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// CmdSql represents the source command.
var CmdSql = &cobra.Command{
	Use:   "sql",
	Short: "Generate the ent schema code from init-schema.sql",
	Long:  "Generate the ent schema code from init-schema.sql. Example: ./j2g.exe sql ./test/sql ./test/sql ",
	Run:   run,
}

var sqlPath string

func init() {
	CmdSql.Flags().StringVarP(&sqlPath, "sql_path", "p", "./", "sql source directory")
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please enter the sql file or directory")
		return
	}
	var (
		err error
		sql = strings.TrimSpace(args[0])
		goo = strings.TrimSpace(args[1])
	)
	if strings.HasSuffix(sql, ".sql") {
		err = generate(sql, goo, args)
	} else {
		err = walk(sql, goo, args)
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
		if ext := filepath.Ext(path); ext != ".sql" {
			return nil
		}
		return generate(path, goo, args)
	})
}

// generate is used to execute the generate command for the specified proto file
func generate(sql string, goo string, args []string) error {
	protoBytes, err := os.ReadFile(sql)
	if err != nil && len(protoBytes) < 0 {
	}
	g := &Generator{
		r:        bytes.NewReader(protoBytes),
		path:     sql,
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
