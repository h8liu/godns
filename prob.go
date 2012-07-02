package dns

// the functions that a problem must implement
type Prob interface {
	IndentSub() bool
	ExpandVia(a Agent)
	Title() (name string, meta []string)
}

type ProbCase interface {
	Prob() Prob
}
