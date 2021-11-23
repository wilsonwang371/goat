package strategy

import "goalgotrade/common"

type Analyzer interface {
	BeforeAttach(s Strategy) error
	Attached(s Strategy) error
	BeforeOnBars(s Strategy, bars common.Bars) error
}
