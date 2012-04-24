package dns2

import (
	"math/rand"
	"fmt"
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

func (m *Msg) String() string {
	if len(m.Ques) == 0 {
		return "<no questions>"
	}
	return fmt.Sprintf("msg: %s", m.Ques[0].Name)
}

func (m *Msg) RandID() {
	m.ID = uint16(rand.Uint32())
}

func (m *Msg) ToWire() ([]byte, error) {
	w := new(wireBuf)
	e := w.writeMsg(m)
	if e != nil {
		return nil, e
	}
	return w.wire(), nil
}

func FromWire(buf []byte) (*Msg, error) {
	w := new(wireBuf)
	w.fill(buf)
	ret := new(Msg)
	e := w.readMsg(ret)
	if e != nil {
		return nil, e
	}
	return ret, nil
}
