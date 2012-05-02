package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// message packer
type writer struct {
	buf bytes.Buffer
}

// message packing error
type PackError struct {
	s string
}

func (m *PackError) Error() string {
	return fmt.Sprintf("dns message: %s", m.s)
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
	for _, s := range n.labels {
		sum += w.writeLabel(s)
	}
	if sum > 255 {
		panic("name too long")
	}
	w.writeUint8(0)
}

func (w *writer) writeRR(rr *RR) (err error) {
	w.writeName(rr.Name)
	w.writeUint16(rr.Type)
	w.writeUint16(rr.Class)
	w.writeUint32(rr.TTL)
	err = w.writeRdata(rr.Rdata)
	if err != nil {
		return err
	}
	return nil
}

func (w *writer) writeRdata(rd rdata) (err error) {
	panic("not implemented")
}

func (w *writer) writeQues(q *Ques) {
	w.writeName(q.Name)
	w.writeUint16(q.Type)
	w.writeUint16(q.Class)
}

func (w *writer) writeMsg(m *Msg) (err error) {
	w.writeUint16(m.ID)
	w.writeUint16(m.Flags)

	n := len(m.Ques)
	if n > 0xffff {
		return &PackError{"too many questions"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Answ)
	if n > 0xffff {
		return &PackError{"too many answers"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Auth)
	if n > 0xffff {
		return &PackError{"too many authorities"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Addi)
	if n > 0xffff {
		return &PackError{"too many additionals"}
	}
	w.writeUint16(uint16(n))

	for _, q := range m.Ques {
		w.writeQues(&q)
	}
	for _, rr := range m.Answ {
		e := w.writeRR(&rr)
		if e != nil {
			return e
		}
	}
	for _, rr := range m.Auth {
		e := w.writeRR(&rr)
		if e != nil {
			return e
		}
	}
	for _, rr := range m.Addi {
		e := w.writeRR(&rr)
		if e != nil {
			return e
		}
	}

	return nil
}

func (w *writer) wire() []byte {
	return w.buf.Bytes()
}