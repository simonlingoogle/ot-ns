// Copyright (c) 2020, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package otcli

import (
	"bufio"
	"context"
	"fmt"
	"github.com/openthread/ot-ns/otoutfilter"
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OtCliRW interface {
	WriteLine(line string)
	ReadLine() (line string, err error)
}

const (
	DefaultCommandTimeout = time.Second * 10
)

var (
	DoneOrErrorRegexp = regexp.MustCompile(`(Done|Error \d+: .*)`)
)

func New(exePath string, id NodeId) (*OtCli, error) {
	flashFile := fmt.Sprintf("tmp/0_%d.flash", id)
	if err := os.RemoveAll(flashFile); err != nil {
		simplelogger.Errorf("Remove flash file %s failed: %+v", flashFile, err)
	}

	exePath, err := filepath.Abs(exePath)
	if err != nil {
		return nil, err
	}
	simplelogger.Debugf("node exe path: %s", exePath)
	cmd := exec.CommandContext(context.Background(), exePath, strconv.Itoa(id))
	pipeIn, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	pipeOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	pipeOutErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	n := &OtCli{
		id:           id,
		cmd:          cmd,
		Input:        pipeIn,
		Output:       bufio.NewReader(pipeOut),
		outputErr:    pipeOutErr,
		pendingLines: make(chan string, 100),
	}

	go n.lineReader()
	n.AssurePrompt()

	return n, nil
}

type OtCli struct {
	id int

	cmd       *exec.Cmd
	Input     io.WriteCloser
	Output    io.Reader
	outputErr io.Reader

	pendingLines chan string
}

func (node *OtCli) String() string {
	return fmt.Sprintf("OtCli<%d>", node.id)
}

func (node *OtCli) Id() int {
	return node.id
}

func (node *OtCli) Start() {
	node.IfconfigUp()
	node.ThreadStart()

	simplelogger.Infof("%v - started, panid=0x%04x, channel=%d, eui64=%#v, extaddr=%#v, state=%s, masterkey=%#v, mode=%v", node,
		node.GetPanid(), node.GetChannel(), node.GetEui64(), node.GetExtAddr(), node.GetState(),
		node.GetMasterKey(), node.GetMode())
}

func (node *OtCli) Stop() {
	node.ThreadStop()
	node.IfconfigDown()
	simplelogger.Debugf("%v - stopped, state = %s", node, node.GetState())
}

func (node *OtCli) Exit() error {
	_, _ = node.Input.Write([]byte("exit\n"))
	node.expectEOF(DefaultCommandTimeout)
	err := node.cmd.Wait()
	if err != nil {
		simplelogger.Warnf("%s exit error: %+v", node, err)
	}
	return err
}

func (node *OtCli) AssurePrompt() {
	_, _ = node.Input.Write([]byte("\n"))
	if found, _ := node.TryExpectLine("", time.Second); found {
		return
	}

	_, _ = node.Input.Write([]byte("\n"))
	if found, _ := node.TryExpectLine("", time.Second); found {
		return
	}

	_, _ = node.Input.Write([]byte("\n"))
	node.expectLine("", DefaultCommandTimeout)
}

func (node *OtCli) CommandNoWait(cmd string, timeout time.Duration) {
	_, _ = node.Input.Write([]byte(cmd + "\n"))
	node.expectLine(cmd, timeout)
}

func (node *OtCli) Command(cmd string, timeout time.Duration) []string {
	_, _ = node.Input.Write([]byte(cmd + "\n"))
	node.expectLine(cmd, timeout)
	output := node.expectLine(DoneOrErrorRegexp, timeout)

	var result string
	output, result = output[:len(output)-1], output[len(output)-1]
	if result != "Done" {
		panic(result)
	}
	return output
}

func (node *OtCli) CommandExpectString(cmd string, timeout time.Duration) string {
	output := node.Command(cmd, timeout)
	if len(output) != 1 {
		simplelogger.Panicf("expected 1 line, but received %d: %#v", len(output), output)
	}

	return output[0]
}

func (node *OtCli) CommandExpectInt(cmd string, timeout time.Duration) int {
	s := node.CommandExpectString(cmd, DefaultCommandTimeout)
	var iv int64
	var err error

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		iv, err = strconv.ParseInt(s[2:], 16, 0)
	} else {
		iv, err = strconv.ParseInt(s, 10, 0)
	}

	if err != nil {
		simplelogger.Panicf("unexpected number: %#v", s)
	}
	return int(iv)
}

func (node *OtCli) CommandExpectHex(cmd string, timeout time.Duration) int {
	s := node.CommandExpectString(cmd, DefaultCommandTimeout)
	var iv int64
	var err error

	iv, err = strconv.ParseInt(s[2:], 16, 0)

	if err != nil {
		simplelogger.Panicf("unexpected number: %#v", s)
	}
	return int(iv)
}

func (node *OtCli) SetChannel(ch int) {
	simplelogger.AssertTrue(11 <= ch && ch <= 26)
	node.Command(fmt.Sprintf("channel %d", ch), DefaultCommandTimeout)
}

func (node *OtCli) GetChannel() int {
	return node.CommandExpectInt("channel", DefaultCommandTimeout)
}

func (node *OtCli) GetChildList() (childlist []int) {
	s := node.CommandExpectString("child list", DefaultCommandTimeout)
	ss := strings.Split(s, " ")

	for _, ids := range ss {
		id, err := strconv.Atoi(ids)
		if err != nil {
			simplelogger.Panicf("unpexpected child list: %#v", s)
		}
		childlist = append(childlist, id)
	}
	return
}

func (node *OtCli) GetChildTable() {
	// todo: not implemented yet
}

func (node *OtCli) GetChildTimeout() int {
	return node.CommandExpectInt("childtimeout", DefaultCommandTimeout)
}

func (node *OtCli) SetChildTimeout(timeout int) {
	node.Command(fmt.Sprintf("childtimeout %d", timeout), DefaultCommandTimeout)
}

func (node *OtCli) GetContextReuseDelay() int {
	return node.CommandExpectInt("contextreusedelay", DefaultCommandTimeout)
}

func (node *OtCli) SetContextReuseDelay(delay int) {
	node.Command(fmt.Sprintf("contextreusedelay %d", delay), DefaultCommandTimeout)
}

func (node *OtCli) GetNetworkName() string {
	return node.CommandExpectString("networkname", DefaultCommandTimeout)
}

func (node *OtCli) SetNetworkName(name string) {
	node.Command(fmt.Sprintf("networkname %s", name), DefaultCommandTimeout)
}

func (node *OtCli) GetEui64() string {
	return node.CommandExpectString("eui64", DefaultCommandTimeout)
}

func (node *OtCli) SetEui64(eui64 string) {
	node.Command(fmt.Sprintf("eui64 %s", eui64), DefaultCommandTimeout)
}

func (node *OtCli) GetExtAddr() uint64 {
	s := node.CommandExpectString("extaddr", DefaultCommandTimeout)
	v, err := strconv.ParseUint(s, 16, 64)
	simplelogger.PanicIfError(err)
	return v
}

func (node *OtCli) SetExtAddr(extaddr uint64) {
	node.Command(fmt.Sprintf("extaddr %016x", extaddr), DefaultCommandTimeout)
}

func (node *OtCli) GetExtPanid() string {
	return node.CommandExpectString("extpanid", DefaultCommandTimeout)
}

func (node *OtCli) SetExtPanid(extpanid string) {
	node.Command(fmt.Sprintf("extpanid %s", extpanid), DefaultCommandTimeout)
}

func (node *OtCli) GetIfconfig() string {
	return node.CommandExpectString("ifconfig", DefaultCommandTimeout)
}

func (node *OtCli) IfconfigUp() {
	node.Command("ifconfig up", DefaultCommandTimeout)
}

func (node *OtCli) IfconfigDown() {
	node.Command("ifconfig down", DefaultCommandTimeout)
}

func (node *OtCli) GetIpAddr() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr", DefaultCommandTimeout)
	return addrs
}

func (node *OtCli) GetIpAddrLinkLocal() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr linklocal", DefaultCommandTimeout)
	return addrs
}

func (node *OtCli) GetIpAddrMleid() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipaddr mleid", DefaultCommandTimeout)
	return addrs
}

func (node *OtCli) GetIpAddrRloc() []string {
	addrs := node.Command("ipaddr rloc", DefaultCommandTimeout)
	return addrs
}

func (node *OtCli) GetIpMaddr() []string {
	// todo: parse IPv6 addresses
	addrs := node.Command("ipmaddr", DefaultCommandTimeout)
	return addrs
}

func (node *OtCli) GetIpMaddrPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("ipmaddr promiscuous", DefaultCommandTimeout)
}

func (node *OtCli) IpMaddrPromiscuousEnable() {
	node.Command("ipmaddr promiscuous enable", DefaultCommandTimeout)
}

func (node *OtCli) IpMaddrPromiscuousDisable() {
	node.Command("ipmaddr promiscuous disable", DefaultCommandTimeout)
}

func (node *OtCli) GetPromiscuous() bool {
	return node.CommandExpectEnabledOrDisabled("promiscuous", DefaultCommandTimeout)
}

func (node *OtCli) PromiscuousEnable() {
	node.Command("promiscuous enable", DefaultCommandTimeout)
}

func (node *OtCli) PromiscuousDisable() {
	node.Command("promiscuous disable", DefaultCommandTimeout)
}

func (node *OtCli) GetRouterEligible() bool {
	return node.CommandExpectEnabledOrDisabled("routereligible", DefaultCommandTimeout)
}

func (node *OtCli) RouterEligibleEnable() {
	node.Command("routereligible enable", DefaultCommandTimeout)
}

func (node *OtCli) RouterEligibleDisable() {
	node.Command("routereligible disable", DefaultCommandTimeout)
}

func (node *OtCli) GetJoinerPort() int {
	return node.CommandExpectInt("joinerport", DefaultCommandTimeout)
}

func (node *OtCli) SetJoinerPort(port int) {
	node.Command(fmt.Sprintf("joinerport %d", port), DefaultCommandTimeout)
}

func (node *OtCli) GetKeySequenceCounter() int {
	return node.CommandExpectInt("keysequence counter", DefaultCommandTimeout)
}

func (node *OtCli) SetKeySequenceCounter(counter int) {
	node.Command(fmt.Sprintf("keysequence counter %d", counter), DefaultCommandTimeout)
}

func (node *OtCli) GetKeySequenceGuardTime() int {
	return node.CommandExpectInt("keysequence guardtime", DefaultCommandTimeout)
}

func (node *OtCli) SetKeySequenceGuardTime(guardtime int) {
	node.Command(fmt.Sprintf("keysequence guardtime %d", guardtime), DefaultCommandTimeout)
}

type LeaderData struct {
	PartitionID       int
	Weighting         int
	DataVersion       int
	StableDataVersion int
	LeaderRouterID    int
}

func (node *OtCli) GetLeaderData() (leaderData LeaderData) {
	var err error
	output := node.Command("leaderdata", DefaultCommandTimeout)
	for _, line := range output {
		if strings.HasPrefix(line, "Partition ID:") {
			leaderData.PartitionID, err = strconv.Atoi(line[14:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Weighting:") {
			leaderData.Weighting, err = strconv.Atoi(line[11:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Data Version:") {
			leaderData.DataVersion, err = strconv.Atoi(line[14:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Stable Data Version:") {
			leaderData.StableDataVersion, err = strconv.Atoi(line[21:])
			simplelogger.PanicIfError(err)
		}

		if strings.HasPrefix(line, "Leader Router ID:") {
			leaderData.LeaderRouterID, err = strconv.Atoi(line[18:])
			simplelogger.PanicIfError(err)
		}
	}
	return
}

func (node *OtCli) GetLeaderPartitionId() int {
	return node.CommandExpectInt("leaderpartitionid", DefaultCommandTimeout)
}

func (node *OtCli) SetLeaderPartitionId(partitionid int) {
	node.Command(fmt.Sprintf("leaderpartitionid 0x%x", partitionid), DefaultCommandTimeout)
}

func (node *OtCli) GetLeaderWeight() int {
	return node.CommandExpectInt("leaderweight", DefaultCommandTimeout)
}

func (node *OtCli) SetLeaderWeight(weight int) {
	node.Command(fmt.Sprintf("leaderweight 0x%x", weight), DefaultCommandTimeout)
}

func (node *OtCli) FactoryReset() {
	simplelogger.Warnf("%v - factoryreset", node)
	_, _ = node.Input.Write([]byte("factoryreset\n"))
	node.AssurePrompt()
	simplelogger.Debugf("%v - ready", node)
}

func (node *OtCli) Reset() {
	simplelogger.Warnf("%v - reset", node)
	_, _ = node.Input.Write([]byte("reset\n"))
	node.AssurePrompt()
	simplelogger.Debugf("%v - ready", node)
}

func (node *OtCli) GetMasterKey() string {
	return node.CommandExpectString("masterkey", DefaultCommandTimeout)
}

func (node *OtCli) SetMasterKey(key string) {
	node.Command(fmt.Sprintf("masterkey %s", key), DefaultCommandTimeout)
}

func (node *OtCli) GetMode() string {
	// todo: return Mode type rather than just string
	return node.CommandExpectString("mode", DefaultCommandTimeout)
}

func (node *OtCli) SetMode(mode string) {
	node.Command(fmt.Sprintf("mode %s", mode), DefaultCommandTimeout)
}

func (node *OtCli) GetPanid() uint16 {
	// todo: return Mode type rather than just string
	return uint16(node.CommandExpectInt("panid", DefaultCommandTimeout))
}

func (node *OtCli) SetPanid(panid uint16) {
	node.Command(fmt.Sprintf("panid 0x%x", panid), DefaultCommandTimeout)
}

func (node *OtCli) GetRloc16() uint16 {
	return uint16(node.CommandExpectHex("rloc16", DefaultCommandTimeout))
}

func (node *OtCli) GetRouterSelectionJitter() int {
	return node.CommandExpectInt("routerselectionjitter", DefaultCommandTimeout)
}

func (node *OtCli) SetRouterSelectionJitter(timeout int) {
	node.Command(fmt.Sprintf("routerselectionjitter %d", timeout), DefaultCommandTimeout)
}

func (node *OtCli) GetRouterUpgradeThreshold() int {
	return node.CommandExpectInt("routerupgradethreshold", DefaultCommandTimeout)
}

func (node *OtCli) SetRouterUpgradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerupgradethreshold %d", timeout), DefaultCommandTimeout)
}

func (node *OtCli) GetRouterDowngradeThreshold() int {
	return node.CommandExpectInt("routerdowngradethreshold", DefaultCommandTimeout)
}

func (node *OtCli) SetRouterDowngradeThreshold(timeout int) {
	node.Command(fmt.Sprintf("routerdowngradethreshold %d", timeout), DefaultCommandTimeout)
}

func (node *OtCli) GetState() string {
	return node.CommandExpectString("state", DefaultCommandTimeout)
}

func (node *OtCli) ThreadStart() {
	node.Command("thread start", DefaultCommandTimeout)
}
func (node *OtCli) ThreadStop() {
	node.Command("thread stop", DefaultCommandTimeout)
}

func (node *OtCli) GetVersion() string {
	return node.CommandExpectString("version", DefaultCommandTimeout)
}

func (node *OtCli) GetSingleton() bool {
	s := node.CommandExpectString("singleton", DefaultCommandTimeout)
	if s == "true" {
		return true
	} else if s == "false" {
		return false
	} else {
		simplelogger.Panicf("expect true/false, but read: %#v", s)
		return false
	}
}

func (node *OtCli) lineReader() {
	// close the line channel after line reader routine exit
	defer close(node.pendingLines)

	scanner := bufio.NewScanner(otoutfilter.NewOTOutFilter(node.Output, node.String()))
	scanner.Split(bufio.ScanLines)

	defer func() {
		if scanner.Err() != nil {
			simplelogger.Errorf("%v read input error: %v", node, scanner.Err())
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()

		select {
		case node.pendingLines <- line:
			break
		default:
			// if we failed to append line, we just read one line to get more space
			select {
			case <-node.pendingLines:
			default:
			}

			node.pendingLines <- line // won't block here
			break
		}
	}
}

func (node *OtCli) TryExpectLine(line interface{}, timeout time.Duration) (bool, []string) {
	var outputLines []string

	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			return false, outputLines
		case readLine, ok := <-node.pendingLines:
			if !ok {
				errmsg, _ := ioutil.ReadAll(node.outputErr)
				simplelogger.Panicf("%s EOF: %s", node, string(errmsg))
				return false, outputLines
			}

			simplelogger.Debugf("%v - %s", node, readLine)

			outputLines = append(outputLines, readLine)
			if node.isLineMatch(readLine, line) {
				// found the exact line
				return true, outputLines
			} else {
				// hack: output scan result here, should have better implementation
				//| J | Network Name     | Extended PAN     | PAN  | MAC Address      | Ch | dBm | LQI |
				if strings.HasPrefix(readLine, "|") || strings.HasPrefix(readLine, "+") {
					fmt.Printf("%s\n", readLine)
				}
			}
		}
	}
}

func (node *OtCli) expectLine(line interface{}, timeout time.Duration) []string {
	found, output := node.TryExpectLine(line, timeout)
	if !found {
		simplelogger.Panicf("expect line timeout: %#v", line)
	}

	return output
}

func (node *OtCli) expectEOF(timeout time.Duration) {
	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			simplelogger.Panicf("expect EOF, but timeout")
		case readLine, ok := <-node.pendingLines:
			if !ok {
				// EOF
				return
			}

			simplelogger.Warnf("%v - %s", node, readLine)
		}
	}
}

func (node *OtCli) CommandExpectEnabledOrDisabled(cmd string, timeout time.Duration) bool {
	output := node.CommandExpectString(cmd, timeout)
	if output == "Enabled" {
		return true
	} else if output == "Disabled" {
		return false
	} else {
		simplelogger.Panicf("expect Enabled/Disabled, but read: %#v", output)
	}
	return false
}

func (node *OtCli) Ping(addr string, payloadSize int, count int, interval int, hopLimit int) {
	cmd := fmt.Sprintf("ping %s %d %d %d %d", addr, payloadSize, count, interval, hopLimit)
	_, _ = node.Input.Write([]byte(cmd + "\n"))
	node.expectLine(cmd, DefaultCommandTimeout)
	node.AssurePrompt()
}

func (node *OtCli) isLineMatch(line string, _expectedLine interface{}) bool {
	switch expectedLine := _expectedLine.(type) {
	case string:
		return line == expectedLine
	case *regexp.Regexp:
		return expectedLine.MatchString(line)
	case []string:
		for _, s := range expectedLine {
			if s == line {
				return true
			}
		}
	default:
		simplelogger.Panic("unknown expected string")
	}
	return false
}

func (node *OtCli) DumpStat() string {
	return fmt.Sprintf("extaddr %016x, addr %04x, state %-6s", node.GetExtAddr(), node.GetRloc16(), node.GetState())
}