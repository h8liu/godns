package dns2

type Msg struct {
	ID    uint16
	Flags uint16
	Ques  *Ques
	Answ  []RR
	Auth  []RR
	Addi  []RR
}

func QuesMsg(n string, t uint16) (ret *Msg, err error) {
    name, e := NewName(n)
    if e != nil { return nil, e }
    q := &Ques{name, t, IN}
    ret = &Msg{32, 0, q,
        make([]RR, 0),
        make([]RR, 0),
        make([]RR, 0)}

    return ret, nil
}

func (m *Msg) ToRaw() (raw []byte, err error) {
    w = new(writer)
    w.writeMsg(m)
    // TODO:
}
