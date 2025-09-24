package lang

import (
	"golang.org/x/text/message"
)

type Printer struct {
	*message.Printer
}

func (p *Printer) Number_int64(i, factor int64) string {
	return p.Sprintf("%.2f", float64(i / factor))
}