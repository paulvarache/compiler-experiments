package cmd

import (
	"encoding/json"
	"log"
	"os"

	"compiler/lexer"
	"compiler/parser"

	"github.com/spf13/cobra"
)

var (
	printASTCmd = &cobra.Command{
		Use:   "print-ast",
		Short: "Sample compiler written in go",
		Long:  "This project is not intended for production. This is a learning exercise",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				return
			}
			src := args[0]
			absSrc, err := ResolvePath(src)
			checkErr(err)
			srcFile, err := os.Open(absSrc)
			checkErr(err)
			defer srcFile.Close()
			l := lexer.NewLexer(srcFile)
			p := parser.NewParser(l)
			program, err := p.ParseProgram()
			checkErr(err)
			j, err := json.MarshalIndent(program, "", "  ")
			checkErr(err)
			log.Println(string(j))
		},
	}
)

func init() {
	rootCmd.AddCommand(printASTCmd)
}
