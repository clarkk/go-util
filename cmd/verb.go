package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	GRAY Color 		= "1;30"
	GREEN Color 	= "0;32"
	BLUE Color 		= "0;36"
	YELLOW Color	= "1;33"
	RED Color 		= "0;31"
	PURPLE Color	= "0;35"
	
	VERB_PLAIN 	= 1
	VERB_COLOR 	= 2
)

type Color string

var (
	verbose 	= 0
	out 		= log.New(os.Stdout, "", log.LstdFlags | log.Lmicroseconds)
)

func Verbose(v int){
	verbose = v
}

//	Print to stdout without timestamp if verbosity is enabled
func Out_color(output string, color Color){
	if verbose == 0 {
		return
	}
	
	output = strings.Replace(output, "\n", "\r\n\t> ", -1)
	if verbose == VERB_COLOR {
		if color != "" {
			output = "\033["+string(color)+"m"+output+"\033[0m"
		}
	}
	fmt.Println(output)
}

//	Print to stdout with timestamp
func Out(s string){
	out.Println(s)
}

func Outf(s string, args... any){
	out.Printf(s, args...)
}