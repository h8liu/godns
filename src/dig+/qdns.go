package main

import (
	"dns"
	"fmt"
	"os"
)

func main() {
	client := dns.NewClient()

	fmt.Println("> dig liulonnie.net @74.220.195.131")
	resp, err := client.Query(
		dns.ParseIP("74.220.195.131"), // ns1.hostmonster.com
		dns.Domain("liulonnie.net"),
		dns.A,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.Msg)

	fmt.Println("> dig -recursive liulonnie.net a")
	client.RecurQuery(dns.Domain("liulonnie.net"), dns.A, os.Stdout)

	fmt.Println("(do it again to see if caching works)")
	fmt.Println("> dig -recursive liulonnie.net a")
	client.RecurQuery(dns.Domain("liulonnie.net"), dns.A, os.Stdout)
}
