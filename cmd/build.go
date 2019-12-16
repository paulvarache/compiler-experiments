package cmd

import (
	"log"
	"os"
	"path/filepath"

	"compiler/generator"
	"compiler/lexer"
	"compiler/parser"

	"github.com/spf13/cobra"
)

var (
	output   string
	buildCmd = &cobra.Command{
		Use:   "build",
		Short: "Sample compiler written in go",
		Long:  "This project is not intended for production. This is a learning exercise",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				return
			}
			src := args[0]
			if output == "" {
				output = filepath.Base(src)
			}
			absSrc, err := ResolvePath(src)
			checkErr(err)
			absOutput, err := ResolvePath(output)
			checkErr(err)
			srcFile, err := os.Open(absSrc)
			checkErr(err)
			defer srcFile.Close()
			l := lexer.NewLexer(srcFile)
			p := parser.NewParser(l)
			program, err := p.ParseProgram()
			checkErr(err)
			gen := generator.NewAssemblyGenerator()
			s, err := gen.FromProgram(program)
			checkErr(err)
			log.Println(s)
			outFile, err := os.Create(absOutput)
			checkErr(err)
			defer outFile.Close()
			outFile.WriteString(s)
		},
	}
)

func init() {
	buildCmd.Flags().StringVarP(&output, "output", "o", "", "Output")
	rootCmd.AddCommand(buildCmd)
}
