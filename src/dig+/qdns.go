package main

import (
    "fmt"
    "dns"
    "dns/dnsprob"
    "os"
)

func main() {
    client := dns.NewClient()

    fmt.Println("> dig liulonnie.net @74.220.195.131")
    resp, err := client.Query(
        dns.ParseIP("74.220.195.131"), // ns1.hostmonster.com
        dns.MakeName("liulonnie.net"),
        dns.A,
    )
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(resp.Msg)
    
    fmt.Println("> dig -recursive liulonnie.net a")
    rp := dnsprob.NewRecursive(
        dns.MakeName("liulonnie.net"), 
        dns.A,
    )
    client.Solve(rp, os.Stdout);

    fmt.Println("(do it again to see if caching works)")
    fmt.Println("> dig -recursive liulonnie.net a")
    rp2 := dnsprob.NewRecursive(
        dns.MakeName("liulonnie.net"), 
        dns.A,
    )
    client.Solve(rp2, os.Stdout);
}
