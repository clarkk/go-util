package lang

import (
	"fmt"
	"math"
	"golang.org/x/text/message"
)

type Printer struct {
	*message.Printer
}

func (p *Printer) Number_int64(i, factor int64) string {
	decimals	:= int(math.Log10(float64(factor)))
	format		:= fmt.Sprintf("%%.%df", decimals)
	return p.Sprintf(format, float64(i) / float64(factor))
}