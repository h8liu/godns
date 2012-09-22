package main

import (
    "os"
    "bufio"
    "strings"
    "io"
    "dns"
    "fmt"
)

func main() {
    fin, err := os.Open("moz-reglist")
    if err != nil { panic(err) }
    reader := bufio.NewReader(fin)

    regs := make([]*dns.Name, 0, 5000)
    supers := make([]*dns.Name, 0, 5000)
    negs := make([]*dns.Name, 0, 5000)

    for {
        line, err := reader.ReadString('\n')
        if err != nil && err != io.EOF {
            panic(err)
        }
        
        line = strings.TrimSpace(line)
        if len(line) == 0 {
            if err == io.EOF {
                break
            }
            continue
        }
        if strings.HasPrefix(line, "//") {
            continue
        }

        if strings.HasPrefix(line, "!") {
            d := line[1:]
            name, e := dns.NewName(d)
            if e != nil {
                // fmt.Printf("%s : %s\n", line, e)
            } else {
                negs = append(negs, name)
            }
        } else if strings.HasPrefix(line, "*.") {
            d := line[2:]
            name, e := dns.NewName(d)
            if e != nil {
                // fmt.Printf("%s : %s\n", line, e)
            } else {
                supers = append(supers, name)
            }
        } else {
            name, e := dns.NewName(line)
            if e != nil {
                // fmt.Printf("%s : %s\n", line, e)
            } else {
                regs = append(regs, name)
            }
        }
        
        if err == io.EOF {
            break
        }
    }

    o := func (varName string, reglist []*dns.Name) {
        fmt.Printf("var %s = []*Name {\n", varName);
        for _, name := range regs {
            fmt.Printf("    _n(\"%s\"),\n", name)
        }

        fmt.Printf("}\n")
        fmt.Printf("\n")
    };

    fmt.Println("package dns")
    fmt.Println()

    o("superNames", supers);
    o("regNames", regs);
    o("notRegNames", negs);
}
