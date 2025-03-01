package cmd

import (
	"fmt"
	"bytes"
	"strings"
	"regexp"
	"os/exec"
)

var (
	re_whitespaces = regexp.MustCompile(`\s+`)
)

type Command struct {
	out 	bytes.Buffer
	err 	bytes.Buffer
}

func (c *Command) Run(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &c.out
	cmd.Stderr = &c.err
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, c.err.String())
	}
	return nil
}

func (c *Command) Empty() bool{
	return c.out.Len() == 0
}

func (c *Command) Output_lines() []string{
	out := re_whitespaces.ReplaceAllString(strings.TrimRight(c.out.String(), "\n"), " ")
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

func (c *Command) Fscanf(format string, a... any){
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