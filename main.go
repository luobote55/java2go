package main

import (
	"github.com/spf13/cobra"
	"go-tools/java2go/ctl"
	"go-tools/java2go/do"
	"go-tools/java2go/sql"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "java2go",
	Short:   "java2go 2 go.",
	Long:    `java2go 2 go`,
	Version: "1.0.0",
}

func init() {
	rootCmd.AddCommand(do.CmdDo)
	rootCmd.AddCommand(sql.CmdSql)
	rootCmd.AddCommand(ctl.CmdCtl)
}

// help:
// ./j2g.exe do ./test ./test

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
