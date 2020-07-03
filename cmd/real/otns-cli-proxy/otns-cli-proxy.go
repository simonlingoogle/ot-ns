package main

import (
	"flag"
	"github.com/openthread/ot-ns/cli/runcli"
	"github.com/simonlingoogle/go-simplelogger"
	"io"
	"os"
	"strconv"
)

type cliHandler struct{}

func (c cliHandler) GetPrompt() string {
	return "> "
}

func (c cliHandler) HandleCommand(cmd string, output io.Writer) error {
	if _, err := output.Write([]byte("Done\n")); err != nil {
		return err
	}

	if cmd == "exit" {
		os.Exit(0)
	}

	return nil
}

var options struct {
	DevicePath string
}

func parseArgs() {
	var err error
	var args []string

	flag.StringVar(&options.DevicePath, "dev", "", "set device path")

	flag.Parse()

	if options.DevicePath == "" {
		simplelogger.Errorf("device path is not specified")
		goto failed
	}

	args = flag.Args()
	if len(args) != 1 {
		simplelogger.Errorf("must specify one node ID")
		goto failed
	}

	_, err = strconv.Atoi(args[0])
	if err != nil {
		simplelogger.Errorf("must specify one node ID")
		goto failed
	}

	return

failed:
	flag.Usage()
	os.Exit(1)
}

func main() {
	args := parseArgs()
	simplelogger.Infof("options %v, args %v", options, args)
	go readCli()

	err := runcli.RunCli(&cliHandler{}, runcli.CliOptions{
		EchoInput: true,
	})

	if err != nil {
		simplelogger.Error(err)
	}
}

func readCli() {

}
