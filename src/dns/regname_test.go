package dns

import "testing"

func TestRegName(t *testing.T) {
	o := func(n, r1, r2 *Name) {
		reg1, reg2 := RegParts(n)
		if (r1 == nil && reg1 != nil) || !r1.Equal(reg1) ||
			(r2 == nil && reg2 != nil) || !r2.Equal(reg2) {
			t.Errorf("RegParts(%s) = %s, %s   expecting: %s, %s",
				n, reg1, reg2, r1, r2)
		}
	}

	p := func(n *Name, b bool) {
		r := IsRegistrar(n)
		if r != b {
			t.Errorf("IsRegistrar(%s) = %t,   expecting: %t", n, r, b)
		}
	}

	o(Domain("."), nil, Domain("."))
	o(Domain("com"), nil, Domain("com"))
	o(Domain("com.cn"), nil, Domain("com.cn"))
	o(Domain("liulonnie.net"), Domain("liulonnie.net"), Domain("net"))
	o(Domain("blog.liulonnie.net"), Domain("liulonnie.net"), Domain("net"))
	o(Domain("wildcard.blog.liulonnie.net"),
		Domain("liulonnie.net"), Domain("net"))
	o(Domain("www.google.com.tw"), Domain("google.com.tw"), Domain("com.tw"))

	p(Domain("."), true)
	p(Domain("net"), true)
	p(Domain("com.tw"), true)
	p(Domain("google.com.tw"), false)
	p(Domain("www.google.com.tw"), false)
	p(Domain("google.com"), false)
	p(Domain("google"), false)
}
