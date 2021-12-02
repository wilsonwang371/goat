package strategy

import (
	"goalgotrade/nugen/bar"
)

type Analyzer interface {
	BeforeAttach(s Strategy) error
	Attached(s Strategy) error
	BeforeOnBars(s Strategy, bars bar.Bars) error
}
