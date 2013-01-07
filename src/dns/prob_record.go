package dns

type ProbRecord struct {
	name    *Name
	types   []uint16
	Records []*RR
}

func NewProbRecord(name *Name, types []uint16) *ProbRecord {
	ret := &ProbRecord{name: name, types: make([]uint16, len(types))}
	copy(ret.types, types)
	return ret
}

func (p *ProbRecord) Title() (title []string) {
	return []string{"record", p.name.String()}
}

func (p *ProbRecord) interested(tp uint16) bool {
	for _, t := range p.types {
		if t == tp {
			return true
		}
	}
	return false
}

func (p *ProbRecord) collectRecords(recur *ProbRecur) {
	for _, r := range recur.History {
		if r.Resp == nil {
			continue
		}
		msg := r.Resp.Msg
		records := msg.FilterIN(func(rr *RR, seg int) bool {
			return rr.Name.Equal(p.name) && p.interested(rr.Type)
		})
		p.Records = append(p.Records, records...)
	}
}

func (p *ProbRecord) ExpandVia(a Solver) {
	if len(p.types) == 0 {
		return
	}

	recur := NewProbRecur(p.name, A)
	if !a.SolveSub(recur) {
		return
	}
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

		recur = NewProbRecur(p.name, t)
		recur.StartsWith(authZone)
		if !a.SolveSub(recur) {
			return // max depth reached
		}

		p.collectRecords(recur)
	}
}
