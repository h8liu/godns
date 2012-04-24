package dns2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type wireBuf struct {
	buf bytes.Buffer
}

type MsgError struct {
	s string
}

type ParseError struct {
	s string
}

func (w *wireBuf) writeUint8(i uint8) {
	e := binary.Write(&w.buf, binary.BigEndian, byte(i))
	if e != nil {
		panic(e)
	}
}

func (w *wireBuf) writeUint16(i uint16) {
	e := binary.Write(&w.buf, binary.BigEndian, i)
	if e != nil {
		panic(e)
	}
}

func (w *wireBuf) writeUint32(i uint32) {
	e := binary.Write(&w.buf, binary.BigEndian, i)
	if e != nil {
		panic(e)
	}
}

func (w *wireBuf) writeBytes(buf []byte) {
	w.buf.Write(buf)
}

func (w *wireBuf) writeLabel(s string) (n int) {
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

func (w *wireBuf) writeName(n *Name) {
	sum := 0
	for _, s := range n.labels {
		sum += w.writeLabel(s)
	}
	if sum > 255 {
		panic("name too long")
	}
	w.writeUint8(0)
}

func (m *MsgError) Error() string {
	return fmt.Sprintf("dns message: %s", m.s)
}

func (m *ParseError) Error() string {
	return fmt.Sprintf("parse message: %s", m.s)
}

func (w *wireBuf) writeRR(rr *RR) (err error) {
	w.writeName(rr.Name)
	w.writeUint16(rr.Type)
	w.writeUint16(rr.Class)
	w.writeUint32(rr.TTL)
	n := len(rr.RData)
	if n > 0xffff {
		return &MsgError{"Rdata too long"}
	}
	w.writeUint16(uint16(n))
	w.writeBytes(rr.RData)
	return nil
}

func (w *wireBuf) writeQues(q *Ques) {
	w.writeName(q.Name)
	w.writeUint16(q.Type)
	w.writeUint16(q.Class)
}

func (w *wireBuf) writeMsg(m *Msg) (err error) {
	w.writeUint16(m.ID)
	w.writeUint16(m.Flags)

	n := len(m.Ques)
	if n > 0xffff {
		return &MsgError{"too many questions"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Answ)
	if n > 0xffff {
		return &MsgError{"too many answers"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Auth)
	if n > 0xffff {
		return &MsgError{"too many authorities"}
	}
	w.writeUint16(uint16(n))

	n = len(m.Addi)
	if n > 0xffff {
		return &MsgError{"too many additionals"}
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

func (w *wireBuf) wire() []byte {
	return w.buf.Bytes()
}

func (w *wireBuf) fill(wire []byte) {
	w.buf.Reset()
	w.buf.Write(wire)
}

func (w *wireBuf) readUint8() (ret uint8, err error) {
	e := binary.Read(&w.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (w *wireBuf) readUint16() (ret uint16, err error) {
	e := binary.Read(&w.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (w *wireBuf) readUint32() (ret uint32, err error) {
	e := binary.Read(&w.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (w *wireBuf) readBytes(buf []byte) (err error) {
	n, e := w.buf.Read(buf)
	if e != nil {
		return e
	}
	if n != len(buf) {
		return io.EOF
	}
	return nil
}

func fmtLabel(b []byte) (ret string, err error) {
	s := strings.ToLower(string(b))
	n := len(s)
	if n == 0 {
		return "", &ParseError{"empty label"}
	}

	if s[0] == '-' {
		return "", &ParseError{"label starts with dash"}
	}
	if s[len(s)-1] == '-' {
		return "", &ParseError{"label ends with dash"}
	}

	for _, c := range s {
		switch {
		default:
			return "", &ParseError{"label has special characters"}
		case 'a' <= c && c <= 'z':
		case '0' <= c && c <= '9':
		case c == '_':
		case c == '-':
		}
	}

	return s, nil
}

func (w *wireBuf) readName() (n *Name, err error) {
	sum := 0
	labels := make([]string, 0)
	for {
		n, e := w.readUint8()
		if e != nil {
			return nil, e
		}
		if n == 0 {
			break
		}
		sum += int(n) + 1
		if n > 63 {
			return nil, &ParseError{"label too long"}
		}
		if sum > 255 {
			return nil, &ParseError{"name too long"}
		}
		b := make([]byte, n)
		e = w.readBytes(b)
		if e != nil {
			return nil, e
		}
		s, e := fmtLabel(b)
		if e != nil {
			return nil, e
		}
		labels = append(labels, s)
	}

	return &Name{labels}, nil
}

func (w *wireBuf) readRR(ret *RR) (err error) {
	ret.Name, err = w.readName()
	if err != nil {
		return
	}
	ret.Type, err = w.readUint16()
	if err != nil {
		return
	}
	ret.Class, err = w.readUint16()
	if err != nil {
		return
	}
	ret.TTL, err = w.readUint32()
	if err != nil {
		return
	}
	n, err := w.readUint16()
	ret.RData = make([]byte, n)
	err = w.readBytes(ret.RData)
	if err != nil {
		return
	}
	return nil
}

func (w *wireBuf) readQues(ret *Ques) (err error) {
	ret.Name, err = w.readName()
	if err != nil {
		return
	}
	ret.Type, err = w.readUint16()
	if err != nil {
		return
	}
	ret.Class, err = w.readUint16()
	if err != nil {
		return
	}
	return nil
}

func (w *wireBuf) readMsg(m *Msg) (e error) {
	m.ID, e = w.readUint16()
	if e != nil {
		return
	}
	m.Flags, e = w.readUint16()
	if e != nil {
		return
	}
	qdCount, e := w.readUint16()
	if e != nil {
		return
	}
	anCount, e := w.readUint16()
	if e != nil {
		return
	}
	nsCount, e := w.readUint16()
	if e != nil {
		return
	}
	arCount, e := w.readUint16()
	if e != nil {
		return
	}
	m.Ques = make([]Ques, qdCount)
	m.Answ = make([]RR, anCount)
	m.Auth = make([]RR, nsCount)
	m.Addi = make([]RR, arCount)
	var zero uint16 = 0
	for i := zero; i < qdCount; i++ {
		e = w.readQues(&m.Ques[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < anCount; i++ {
		e = w.readRR(&m.Answ[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < nsCount; i++ {
		e = w.readRR(&m.Auth[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < nsCount; i++ {
		e = w.readRR(&m.Addi[i])
		if e != nil {
			return
		}
	}
	return nil
}
