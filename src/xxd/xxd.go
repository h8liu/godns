package xxd

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func bufPrint(buf []byte) *bytes.Buffer {
	ret := new(bytes.Buffer)

	for i, b := range buf {
		if i == 0 {
			// no trialing
		} else if i%16 == 0 {
			fmt.Fprintln(ret)
		} else if i%8 == 0 {
			fmt.Fprint(ret, "  ")
		} else if i%4 == 0 {
			fmt.Fprint(ret, " ")
		}

		fmt.Fprintf(ret, "%02x", b)
	}

	fmt.Fprintln(ret)

	return ret
}

func Fprint(w io.Writer, buf []byte) (int, error) {
	return w.Write(bufPrint(buf).Bytes())
}

func Print(buf []byte) (int, error) {
	return Fprint(os.Stdout, buf)
}

func String(buf []byte) string {
	return bufPrint(buf).String()
}
