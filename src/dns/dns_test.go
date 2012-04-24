package dns

import (
	"fmt"
	"testing"
)

func TestTest(t *testing.T) {
	const i = 0
	if i != 0 {
		t.Errorf("just a test")
	}
}

func ExampleTest() {
	fmt.Println("hello")
	// Output:
	//     hello
}
