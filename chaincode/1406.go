package main

import (
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"go.uber.org/zap"

	"github.com/hyperledger-labs/cckit/serialize"
	"github.com/hyperledger-labs/cckit/state"
)

// DefaultParam default parameter name
const DefaultParam = `default`

type (
	// Context of chaincode invoke
	Context interface {
		Clone() Context

		Stub() shim.ChaincodeStubInterface

		Serializer() serialize.Serializer

		// Client returns invoker ClientIdentity
		Client() (cid.ClientIdentity, error)

		// Response returns response builder
		Response() Response
		Logger() *zap.Logger
		Path() string
		Handler() *HandlerMeta
		SetHandler(*HandlerMeta)
		State() state.State
		UseState(state.State) Context

		// Time returns txTimestamp
		Time() (time.Time, error)

		ReplaceArgs(args [][]byte) Context // replace args, for usage in preMiddleware
		GetArgs() [][]byte

		// Params returns parameter values.
		Params() InterfaceMap

		// Param returns parameter value.
		Param(name ...string) interface{}

		// ParamString returns parameter value as string.
		ParamString(name string) string

		// ParamBytes returns parameter value as bytes.
		ParamBytes(name string) []byte

		// ParamInt returns parameter value as int.
		ParamInt(name string) int

		// ParamInt32 returns parameter value as int32.
		ParamInt32(name string) int32

		// SetParam sets parameter value.
		SetParam(name string, value interface{})

		Event() state.Event
		UseEvent(state.Event) Context
	}

	context struct {
		stub       shim.ChaincodeStubInterface
		handler    *HandlerMeta
		logger     *zap.Logger
		state      state.State
		event      state.Event
		args       [][]byte
		params     InterfaceMap
		serializer serialize.Serializer
	}
)

// NewContext creates new instance of router.Context
func NewContext(stub shim.ChaincodeStubInterface, serializer serialize.Serializer, logger *zap.Logger) *context {
	return &context{
		stub:       stub,
		logger:     logger,
		serializer: serializer,
	}
}

func (c *context) Clone() Context {
	ctx := NewContext(c.stub, c.serializer, c.logger)
	if c.state != nil {
		ctx.state = c.state.Clone()
	}

	return ctx
}

func (c *context) Stub() shim.ChaincodeStubInterface {
	return c.stub
}

func (c *context) Client() (cid.ClientIdentity, error) {
	return cid.New(c.Stub())
}

func (c *context) Serializer() serialize.Serializer {
	return c.serializer
}

func (c *context) Response() Response {
	return &ContextResponse{c}
}

func (c *context) Logger() *zap.Logger {
	return c.logger
}

func (c *context) Path() string {
	if len(c.GetArgs()) == 0 {
		return ``
	}
	return string(c.GetArgs()[0])
}

func (c *context) Handler() *HandlerMeta {
	return c.handler
}

func (c *context) SetHandler(h *HandlerMeta) {
	c.handler = h
}

func (c *context) State() state.State {
	if c.state == nil {
		c.state = state.NewState(c.stub, c.logger)
	}
	return c.state
}

func (c *context) UseState(s state.State) Context {
	c.state = s
	return c
}

func (c *context) Event() state.Event {
	if c.event == nil {
		c.event = state.NewEvent(c.stub)
	}
	return c.event
}

func (c *context) UseEvent(e state.Event) Context {
	c.event = e
	return c
}

func (c *context) Time() (time.Time, error) {
	txTimestamp, err := c.stub.GetTxTimestamp()
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(txTimestamp.GetSeconds(), int64(txTimestamp.GetNanos())), nil
}

// ReplaceArgs replace args, for usage in preMiddleware
func (c *context) ReplaceArgs(args [][]byte) Context {
	c.args = args
	return c
}

func (c *context) GetArgs() [][]byte {
	if c.args != nil {
		return c.args
	}
	return c.stub.GetArgs()
}

func (c *context) Params() InterfaceMap {
	return c.params
}

func (c *context) SetParam(name string, value interface{}) {
	if c.params == nil {
		c.params = make(InterfaceMap)
	}
	c.params[name] = value
}

// Deprecated: Use Param instead.
func (c *context) Arg(name string) interface{} {
	return c.Param(name)
}

func (c *context) Param(name ...string) interface{} {
	var pName = DefaultParam
	if len(name) > 0 {
		pName = name[0]
	}
	return c.params[pName]
}

// Deprecated: Use ParamString instead.
func (c *context) ArgString(name string) string {
	return c.ParamString(name)
}

func (c *context) ParamString(name string) string {
	out, _ := c.Param(name).(string)
	return out
}

// Deprecated: Use ParamBytes instead.
func (c *context) ArgBytes(name string) []byte {
	return c.ParamBytes(name)
}

func (c *context) ParamBytes(name string) []byte {
	out, _ := c.Param(name).([]byte)
	return out
}

// Deprecated: Use ParamInt instead.
func (c *context) ArgInt(name string) int {
	return c.ParamInt(name)
}

func (c *context) ParamInt(name string) int {
	out, _ := c.Param(name).(int)
	return out
}

// ParamInt32 returns parameter value as int32.
func (c *context) ParamInt32(name string) int32 {
	out, _ := c.Param(name).(int32)
	return out
}

//func (c *context) Set(key string, val interface{}) {
//	if c.store == nil {
//		c.store = make(InterfaceMap)
//	}
//	c.store[key] = val
//}

//func (c *context) Get(key string) interface{} {
//	return c.store[key]
//}

func (c *context) SetEvent(name string, payload interface{}) error {
	return c.Event().Set(name, payload)
}

func ContextWithStateCache(ctx Context) Context {
	clone := ctx.Clone()
	return clone.UseState(state.WithCache(clone.State()))
}
