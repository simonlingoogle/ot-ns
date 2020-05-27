package otns_main

import (
	"github.com/simonlingoogle/go-simplelogger"
	"os"
	"testing"
	"time"
)

func TestOTNSMain(t *testing.T) {
	os.Args = append(os.Args, "-log", "debug", "-autogo=false")

	stdin, inputWriter, err := os.Pipe()
	simplelogger.FatalIfError(err)

	go func() {
		for true {
			time.Sleep(time.Second)
			inputWriter.WriteString("go 1\n")
		}
	}()

	Main(nil, true, stdin)
}
