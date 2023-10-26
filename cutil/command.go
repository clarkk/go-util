package cutil

import (
	"fmt"
	"bytes"
	"strings"
	"regexp"
	"os/exec"
)

type Command struct {
	out 	bytes.Buffer
	err 	bytes.Buffer
}

func (c *Command) Run(command string){
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &c.out
	cmd.Stderr = &c.err
	err := cmd.Run()
	if err != nil {
		//fmt.Println(err.Error()+": "+c.err.String())
		panic("Command "+command+": "+err.Error())
	}
}

func (c *Command) Empty() bool{
	return c.out.Len() == 0
}

func (c *Command) Output_lines() []string{
	out := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimRight(c.out.String(), "\n"), " ")
	lines := strings.Split(out, "\n")
	for i := range lines {
		lines[i] = strings.Trim(lines[i], " ")
	}
	return lines
}

func (c *Command) Output_fields() [][]string{
	fields := [][]string{}
	lines := c.Output_lines()
	for i := range lines {
		fields = append(fields, strings.Split(lines[i], " "))
	}
	return fields
}

func (c *Command) Fscanf(format string, a ...interface{}){
	/*c.Run(fmt.Sprintf("netstat -nlp | grep :%d", *port))
	var (
		pid 	string
	)
	c.Fscanf("%s", &pid)*/
	
	n, err := fmt.Fscanf(&c.out, format, a...)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("n:", n)
}