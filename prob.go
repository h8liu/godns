package dns

type Prob interface {
	IndentSub() bool
	ExpandVia(a *Agent)
	Title() (name string, meta []string)
}
