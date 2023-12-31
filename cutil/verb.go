package cutil

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
)

var (
	verbose 	= 0
	out 		= log.New(os.Stdout, "", log.LstdFlags | log.Lmicroseconds)
)

func Verbose(v int){
	verbose = v
}

//	Print do display with color and no timestamp
func Color(output string, color string){
	if verbose == 0 {
		return
	}
	
	output = strings.Replace(output, "\n", "\r\n\t> ", -1)
	if verbose == 2 {
		if color != "" {
			output = "\033["+color+"m"+output+"\033[0m"
		}
	}
	fmt.Println(output)
}

func Out(output string){
	out.Println(output)
}