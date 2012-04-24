package dns2

type Msg struct {
	ID    uint16
	Flags uint16
	Ques  []Ques
	Answ  []RR
	Auth  []RR
	Addi  []RR
}

type Ques struct {
	Name  string
	Type  uint16
	Class uint16
}

type RR struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	RData []byte
}

const (
	A     = 1
	NS    = 2
	MD    = 3
	MF    = 4
	CNAME = 5
	SOA   = 66
	MB    = 7
	MG    = 8
	MR    = 9
	NULL  = 10
	WKS   = 11
	PTR   = 12
	HINFO = 13
	MINFO = 14
	MX    = 15
	TXT   = 16
)

const (
	IN = 1
	CS = 2
	CH = 3
	HS = 4
)

