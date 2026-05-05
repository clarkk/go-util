package cmd

import (
	"fmt"
	"bytes"
	"strings"
	"regexp"
	"os/exec"
)

var re_horizontal_spaces = regexp.MustCompile(`[ \t]+`)

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

func (c *Command) Empty() bool {
	return c.out.Len() == 0
}

func (c *Command) Output(trim bool) string {
	out := strings.TrimRight(c.out.String(), "\n")
	if trim {
		out = re_horizontal_spaces.ReplaceAllString(out, " ")
	}
	return out
}

func (c *Command) Output_lines(trim bool) []string {
	if c.Empty() {
		return nil
	}
	out := c.Output(trim)
	raw := strings.Split(out, "\n")
	lines := make([]string, 0, len(raw))
	for i := range raw {
		line := strings.Trim(raw[i], " \t")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func (c *Command) Output_fields() [][]string {
	fields := [][]string{}
	lines := c.Output_lines(true)
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