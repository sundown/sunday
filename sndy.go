package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"sundown/sunday/ir"
	"sundown/sunday/parser"
	"sundown/sunday/util"
)

func main() {
	if len(os.Args) < 2 {
		panic("Need args")
	}

	filecontents, err := ioutil.ReadFile(os.Args[2])

	if err != nil {
		panic(err)
	}

	prog := &parser.Program{}

	err = parser.Parser.ParseString(os.Args[2], string(filecontents), prog)

	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "analyse":
		s := &ir.State{}

		s.Analyse(prog)
		fmt.Println(s.String())
	default:
		util.Error("invalid subcommand" + os.Args[1])
		os.Exit(1)
	}
}
