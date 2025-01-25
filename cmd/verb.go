package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	GRAY 		= "1;30"
	GREEN 		= "0;32"
	BLUE 		= "0;36"
	YELLOW		= "1;33"
	RED 		= "0;31"
	PURPLE		= "0;35"
	
	VERB_PLAIN 	= 1
	VERB_COLOR 	= 2
)

var (
	verbose 	= 0
	out 		= log.New(os.Stdout, "", log.LstdFlags | log.Lmicroseconds)
)

func Verbose(v int){
	verbose = v
}

//	Print to stdout without timestamp if verbosity is enabled
func Color(output, color string){
	if verbose == 0 {
		return
	}
	
	output = strings.Replace(output, "\n", "\r\n\t> ", -1)
	if verbose == VERB_COLOR {
		if color != "" {
			output = "\033["+color+"m"+output+"\033[0m"
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