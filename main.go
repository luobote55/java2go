package main

import (
	"github.com/luobote55/java2go/ctl"
	"github.com/luobote55/java2go/do"
	"github.com/luobote55/java2go/sql"
	"github.com/spf13/cobra"
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
