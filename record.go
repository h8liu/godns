package dns

type RecordProb struct {
	name  *Name
	types []uint16
}

func NewRecordProb(name *Name, types []uint16) *RecordProb {
	ret := &RecordProb{name: name, types: make([]uint16, len(types))}
	copy(ret.types, types)
	return ret
}

func (p *RecordProb) Title() (name string, meta []string) {
	return "record", []string{}
}

func (p *RecordProb) ExpandVia(a Agent) {
	recur := NewRecurProb(p.name, A)
	a.SolveSub(recur)

	if recur.Answer == nil {
		return
	}

}
