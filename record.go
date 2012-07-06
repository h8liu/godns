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

func (p *RecordProb) collectRecords(a Agent) {

}

func (p *RecordProb) ExpandVia(a Agent) {
    defer p.collectRecords(a)

    // reorder types
    if len(p.types) == 0 {
        return
    }

	recur := NewRecurProb(p.name, p.types[0])
	a.SolveSub(recur)

	if !(recur.AnsCode == OKAY || recur.AnsCode == NONEXIST) {
		return // error on finding the domain server
	}

    if len(p.types) == 1 {
        return
    }

    // restart from here
    authZone := recur.AnsZone
    for _, t := range p.types[1:] {
        recur = NewRecurProb(p.name, t)
        recur.StartFrom(authZone.Zone, authZone.Servers) 
        // TODO: continue here
            
    }

}
