package dns

type Rdata interface {
	pson() ([]string, func(p *Pson))
	writeTo(w *writer) error
	readFrom(r *reader, n uint16) error
}

// for rdata of a string of a byte array, like txt records
type RdBytes struct {
	Data []byte
}

func (rd *RdBytes) pson() ([]string, func(p *Pson)) {
	return []string{}, nil
}

func (rd *RdBytes) writeTo(w *writer) error {
	w.writeUint16(0)
	return nil
}

func (rd *RdBytes) readFrom(r *reader, n uint16) error {
	rd.Data = make([]byte, n)
	return r.readBytes(rd.Data)
}

// for rdatas of a single ip address, like a records
type RdIP struct {
	Ip *IPv4
}

func (rd *RdIP) pson() ([]string, func(p *Pson)) {
	return []string{rd.Ip.String()}, nil
}

func (rd *RdIP) writeTo(w *writer) error {
	w.writeBytes(rd.Ip.Bytes())
	return nil
}

func (rd *RdIP) readFrom(r *reader, n uint16) (err error) {
	if n != 4 {
		return &ParseError{"A rdata: wrong size"}
	}
	buf := make([]byte, 4)
	if err = r.readBytes(buf); err != nil {
		return err
	}
	if rd.Ip = IPFromBytes(buf); rd.Ip == nil {
		return &ParseError{"make ip from bytes"}
	}
	return nil
}

// for rdatas of a single name, like ns records
type RdName struct {
	Name *Name
}

func (r *RdName) pson() ([]string, func(p *Pson)) {
	return []string{r.Name.String()}, nil
}

func (rd *RdName) writeTo(w *writer) error {
	panic("not implemented")
	w.writeName(rd.Name)
	return nil
}

func (rd *RdName) readFrom(r *reader, n uint16) (err error) {
	rd.Name, err = r.readName()
	if err != nil {
		return err
	}
	return nil
}
