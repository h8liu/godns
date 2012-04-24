package main

import (
	"fmt"
	"net"
)

func server(conn net.PacketConn) (err error) {
	b := make([]byte, 512)

	// read message
	n, addr, err := conn.ReadFrom(b)
	if err != nil {
		return err
	}
	fmt.Printf("server recv from: %d %s\n", n, addr)

	// send back
	n, err = conn.WriteTo(b[:n], addr)
	if err != nil {
		return err
	}
	fmt.Printf("sent back: %d\n", n)

	conn.Close()
	return nil
}

func test() (err error) {
	// server connection
	conn, err := net.ListenPacket("udp4", "localhost:5344")
	if err != nil {
		fmt.Printf("listen: %s\n", err)
		return err
	}
	go server(conn) // send to background to handle connections

	// client connection
	conn, err = net.ListenPacket("udp4", ":0")
	if err != nil {
		return err
	}

	raddr := &net.UDPAddr{net.ParseIP("127.0.0.1"), 5344}
	if err != nil {
		return err
	}

	buf := []byte("hello, world")

	// send out message
	n, err := conn.WriteTo(buf[:], raddr)
	if err != nil {
		return err
	}
	fmt.Printf("sent to server: %d\n", n)

	buf = make([]byte, 512)
	// read back
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		return err
	}
	fmt.Printf("back from: %d %s\n", n, addr)
	fmt.Printf("msg: %s\n", string(buf[:n]))

	conn.Close()

	fmt.Printf("succ: %d\n", n)

	return nil
}

func main() {
	err := test()
	if err != nil {
		fmt.Printf("%s", err)
	}
}
