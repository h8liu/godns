package dns

type RecordProb struct {
	name  *Name
	types []uint16
    Records []*RR
}

func NewRecordProb(name *Name, types []uint16) *RecordProb {
	ret := &RecordProb{name: name, types: make([]uint16, len(types))}
	copy(ret.types, types)
	return ret
}

func (p *RecordProb) Title() (name string, meta []string) {
	return "record", []string{p.name.String()}
}

func (p *RecordProb) interested(tp uint16) bool {
    for _, t := range p.types {
        if t == tp {
            return true
        }
    }
    return false
}

func (p *RecordProb) collectRecords(recur *RecurProb) {
    for _, r := range recur.History {
        msg := r.Resp.Msg
        records := msg.FilterINRR(func(rr *RR, seg int) bool {
            return rr.Name.Equal(p.name) && p.interested(rr.Type)
        })
        p.Records = append(p.Records, records...)
    }
}

func (p *RecordProb) ExpandVia(a Agent) {
	if len(p.types) == 0 {
		return
	}

	recur := NewRecurProb(p.name, A)
	a.SolveSub(recur)
    p.collectRecords(recur)

	if !(recur.AnsCode == OKAY || recur.AnsCode == NONEXIST) {
		return // error on finding the domain server
	}

	// restart from here
	authZone := recur.AnsZone
	for _, t := range p.types[1:] {
		if t == A {
			continue // already probed
		}

		recur = NewRecurProb(p.name, t)
		recur.StartFrom(authZone)
		a.SolveSub(recur)
        p.collectRecords(recur)
	}
}
