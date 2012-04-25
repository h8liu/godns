package dns

type Ques struct {
	Name  *Name
	Type  uint16
	Class uint16
}

type RR struct {
	Name  *Name
	Type  uint16
	Class uint16
	TTL   uint32
	RData []byte
	rdata rdata
}

const (
	A     = 1
	NS    = 2
	MD    = 3
	MF    = 4
	CNAME = 5
	SOA   = 6
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
	AAAA  = 28
)

const (
	F_RESPONSE  = 0x1 << 15
	F_OPMASK    = 0x3 << 11
	F_AA        = 0x1 << 10
	F_TC        = 0x1 << 9
	F_RD        = 0x1 << 8
	F_RA        = 0x1 << 7
	F_RCODEMASK = 0xf
)

const (
	OPQUERY  = 0 << 11
	OPIQUERY = 1 << 11
	OPSTATUS = 2 << 11
)

const (
	RCODE_OKAY         = 0
	RCODE_FORMATERROR  = 1
	RCODE_SERVERFAIL   = 2
	RCODE_NAMEERROR    = 3
	RCODE_NOTIMPLEMENT = 4
	RCODE_REFUSED      = 5
)

const (
	IN = 1
	CS = 2
	CH = 3
	HS = 4
)
