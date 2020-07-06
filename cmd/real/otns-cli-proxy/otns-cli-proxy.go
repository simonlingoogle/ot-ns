package main

import (
	"context"
	"flag"
	"io"
	"os"
	"strconv"

	"github.com/openthread/ot-ns/cli/runcli"
	"github.com/openthread/ot-ns/progctx"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
	"github.com/tarm/serial"
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
	NodeId     NodeId
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

	options.NodeId, err = strconv.Atoi(args[0])
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
	parseArgs()
	simplelogger.Infof("options %+v", options)

	ctx := progctx.New(context.Background())

	ctx.WaitAdd("readCli", 1)
	go readCli(ctx, options.DevicePath)

	ctx.WaitAdd("RunCli", 1)
	go func() {
		defer ctx.WaitDone("RunCli")
		runcli.RunCli(ctx, &cliHandler{}, runcli.CliOptions{
			EchoInput: true,
		})
	}()

	ctx.Wait()
}

func readCli(ctx *progctx.ProgCtx, devicePath string) {
	var err error

	defer ctx.WaitDone("readCli")

	defer func() {
		if err != nil {
			simplelogger.Errorf("%s", err)
		}
		ctx.Cancel(err)
	}()

	cfg := serial.Config{
		Name: devicePath,
		Baud: 115200,
	}
	_, err = serial.OpenPort(&cfg)
}
