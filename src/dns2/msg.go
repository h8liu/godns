package dns2

import (
	"math/rand"
)

type Msg struct {
	ID    uint16
	Flags uint16
	Ques  []Ques
	Answ  []RR
	Auth  []RR
	Addi  []RR
}

func QuesMsg(n string, t uint16) (ret *Msg, err error) {
	name, e := NewName(n)
	if e != nil {
		return nil, e
	}
	ret = &Msg{0, 0,
		make([]Ques, 0),
		make([]RR, 0),
		make([]RR, 0),
		make([]RR, 0)}
	ret.Ques = append(ret.Ques, Ques{name, t, IN}) // copy in
	ret.RandID()

	return ret, nil
}

func (m *Msg) RandID() {
	m.ID = uint16(rand.Uint32())
}

func (m *Msg) ToWire() (raw []byte, err error) {
	w := new(writer)
	e := w.writeMsg(m)
	if e != nil {
		return nil, e
	}
	return w.toWire(), nil
}
