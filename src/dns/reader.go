package dns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

// message parser
type reader struct {
	buf    *bytes.Reader
	seeker *bytes.Reader
}

var (
	errEmptyLabel   = errors.New("empty label")
	errDashStart    = errors.New("label starts with dash")
	errDashEnd      = errors.New("label ends with dash")
	errSpecialChars = errors.New("label has special characters")
	errLongLabel    = errors.New("label too long")
	errLongName     = errors.New("name too long")
)

func newReader(wire []byte) *reader {
	return &reader{bytes.NewReader(wire),
		bytes.NewReader(wire)}
}

func (r *reader) readUint8() (ret uint8, err error) {
	e := binary.Read(r.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (r *reader) readUint16() (ret uint16, err error) {
	e := binary.Read(r.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (r *reader) readUint32() (ret uint32, err error) {
	e := binary.Read(r.buf, binary.BigEndian, &ret)
	if e != nil {
		return 0, e
	}
	return
}

func (r *reader) readBytes(buf []byte) (err error) {
	n, e := r.buf.Read(buf)
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
		return "", errEmptyLabel
	}

	if s[0] == '-' {
		return "", errDashStart
	}
	if s[len(s)-1] == '-' {
		return "", errDashEnd
	}

	for _, c := range s {
		switch {
		default:
			return "", errSpecialChars
		case 'a' <= c && c <= 'z':
		case '0' <= c && c <= '9':
		case c == '_':
		case c == '-':
		}
	}

	return s, nil
}

func (r *reader) readName() (n *Name, err error) {
	sum := 0
	labels := make([]string, 0, 5)
	rin := r.buf
	for {
		n, e := rin.ReadByte()
		if e != nil {
			return nil, e
		}
		if n == 0 {
			break
		}
		if n&0xc0 == 0xc0 {
			c2, e := rin.ReadByte()
			if e != nil {
				return nil, e
			}
			off := ((uint16(n) & 0x3f) << 8) + uint16(c2)
			rin = r.seeker
			rin.Seek(int64(off), 0)
			continue
		}
		sum += int(n) + 1
		if n > 63 {
			return nil, errLongLabel
		}
		if sum > 255 {
			return nil, errLongName
		}
		b := make([]byte, n)
		_, e = rin.Read(b)
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

func (r *reader) readRR(ret *RR) (err error) {
	ret.Name, err = r.readName()
	if err != nil {
		return
	}
	ret.Type, err = r.readUint16()
	if err != nil {
		return
	}
	ret.Class, err = r.readUint16()
	if err != nil {
		return
	}
	ret.TTL, err = r.readUint32()
	if err != nil {
		return
	}
	n, err := r.readUint16()
	if err != nil {
		return
	}

	ret.Rdata, err = r.readRdata(ret.Class, ret.Type, n)
	if err != nil {
		return
	}

	return nil
}

func (r *reader) readRdata(c, t, n uint16) (ret Rdata, e error) {
	if c == IN {
		switch t {
		default:
			ret = new(RdBytes)
		case A:
			ret = new(RdIP)
		case CNAME, NS:
			ret = new(RdName)
		case TXT:
			ret = new(RdBytes)
		}
	} else {
		ret = new(RdBytes)
	}

	err := ret.readFrom(r, n)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *reader) readQues(ret *Ques) (err error) {
	ret.Name, err = r.readName()
	if err != nil {
		return
	}
	ret.Type, err = r.readUint16()
	if err != nil {
		return
	}
	ret.Class, err = r.readUint16()
	if err != nil {
		return
	}
	return nil
}

func (r *reader) readMsg(m *Msg) (e error) {
	m.ID, e = r.readUint16()
	if e != nil {
		return
	}
	m.Flags, e = r.readUint16()
	if e != nil {
		return
	}
	qdCount, e := r.readUint16()
	if e != nil {
		return
	}
	anCount, e := r.readUint16()
	if e != nil {
		return
	}
	nsCount, e := r.readUint16()
	if e != nil {
		return
	}
	arCount, e := r.readUint16()
	if e != nil {
		return
	}
	m.Ques = make([]Ques, qdCount)
	m.Answ = make([]RR, anCount)
	m.Auth = make([]RR, nsCount)
	m.Addi = make([]RR, arCount)
	var zero uint16 = 0
	for i := zero; i < qdCount; i++ {
		e = r.readQues(&m.Ques[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < anCount; i++ {
		e = r.readRR(&m.Answ[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < nsCount; i++ {
		e = r.readRR(&m.Auth[i])
		if e != nil {
			return
		}
	}
	for i := zero; i < arCount; i++ {
		e = r.readRR(&m.Addi[i])
		if e != nil {
			return
		}
	}
	return nil
}
