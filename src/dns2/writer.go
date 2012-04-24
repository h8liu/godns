package dns2

import (
	"bytes"
	"encoding/binary"
)

type writer struct {
	buf bytes.Buffer
}

func (w *writer) writeUint8(i uint8) {
	e := binary.Write(&w.buf, binary.BigEndian, byte(i))
	if e != nil {
		panic(e)
	}
}

func (w *writer) writeUint16(i uint16) {
	e := binary.Write(&w.buf, binary.BigEndian, i)
	if e != nil {
		panic(e)
	}
}

func (w *writer) writeUint32(i uint32) {
	e := binary.Write(&w.buf, binary.BigEndian, i)
	if e != nil {
		panic(e)
	}
}

func (w *writer) writeBytes(buf []byte) {
	w.buf.Write(buf)
}

func (w *writer) writeLabel(s string) (n int) {
	var buf = []byte(s)
	m := len(s)
	if m == 0 {
		panic("empty label")
	}
	if m > 63 {
		panic("label too long")
	}
	w.writeUint8(uint8(len(s)))
	w.writeBytes(buf)
	return m + 1
}

func (w *writer) writeName(n *Name) {
	sum := 0
	for _, s := range n.label {
		sum += w.writeLabel(s)
	}
	if sum > 255 {
		panic("name too long")
	}
	w.writeUint8(0)
}

func (w *writer) writeRR(rr *RR) {
	w.writeName(rr.Name)
	w.writeUint16(rr.Type)
	w.writeUint16(rr.Class)
	w.writeUint32(rr.TTL)
	n := len(rr.RData)
	if n > 0xffff {
		panic("Rdata too long")
	}
	w.writeUint16(uint16(n))
	w.writeBytes(rr.RData)
}

func (w *writer) writeQues(q *Ques) {
	w.writeName(q.Name)
	w.writeUint16(q.Type)
	w.writeUint16(q.Class)
}

func (w *writer) writeMsg(m *Msg) {
}
