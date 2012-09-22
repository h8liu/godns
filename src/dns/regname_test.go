package dns

import "testing"

func TestRegName(t *testing.T) {
    o := func (n, r1, r2 *Name) {
        reg1, reg2 := RegParts(n)
        if (r1 == nil && reg1 != nil) || !r1.Equal(reg1) ||
            (r2 == nil && reg2 != nil) || !r2.Equal(reg2) {
            t.Error("RegParts(%s) = %s, %s   expecting: %s, %s",
                n, reg1, reg2, r1, r2)
        }
    }

    o(N("."), nil, N("."))
    o(N("com"), nil, N("com"))
    o(N("com.cn"), nil, N("com.cn"))
    o(N("liulonnie.net"), N("liulonnie.net"), N("net"))
    o(N("blog.liulonnie.net"), N("liulonnie.net"), N("net"))
    o(N("wildcard.blog.liulonnie.net"),
        N("liulonnie.net"), N("net"))
    o(N("www.google.com.tw"), N("google.com.tw"), N("com.tw"))
}
