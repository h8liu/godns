package dns

// the functions that a problem must implement
type Prob interface {
	Title() (title []string)
	ExpandVia(a ProbAgent)
}
