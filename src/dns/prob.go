package dns

// the functions that a problem must implement
type Prob interface {
	Title() (name string, meta []string)
	ExpandVia(a Agent)
}
