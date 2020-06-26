package otreal

import (
	. "github.com/openthread/ot-ns/types"
	"time"
)

type OtRealNode struct {
	id int
}

func NewNode(nodeid NodeId) (*OtRealNode, error) {
	return &OtRealNode{id: nodeid}, nil
}

func (node *OtRealNode) Id() int {
	return node.id
}

func (node *OtRealNode) Ping(addr string, payloadSize int, count int, interval int, hopLimit int) {
	panic("not implemented")
}
func (node *OtRealNode) AssurePrompt()                                      { panic("not implemented") }
func (node *OtRealNode) Command(cmd string, timeout time.Duration) []string { panic("not implemented") }
func (node *OtRealNode) CommandNoWait(cmd string, timeout time.Duration)    { panic("not implemented") }
func (node *OtRealNode) Exit() error                                        { panic("not implemented") }
func (node *OtRealNode) GetIpAddrLinkLocal() []string                       { panic("not implemented") }
func (node *OtRealNode) GetIpAddrMleid() []string                           { panic("not implemented") }
func (node *OtRealNode) GetIpAddrRloc() []string                            { panic("not implemented") }
func (node *OtRealNode) RouterEligibleDisable()                             { panic("not implemented") }
func (node *OtRealNode) SetChannel(ch int)                                  { panic("not implemented") }
func (node *OtRealNode) SetMasterKey(key string)                            { panic("not implemented") }
func (node *OtRealNode) SetMode(mode string)                                { panic("not implemented") }
func (node *OtRealNode) SetPanid(panid uint16)                              { panic("not implemented") }
func (node *OtRealNode) SetRouterSelectionJitter(timeout int)               { panic("not implemented") }
func (node *OtRealNode) Start()                                             { panic("not implemented") }
