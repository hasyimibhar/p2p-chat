package message

import (
	"reflect"
	"sync"
)

type Opcode byte

const (
	OpcodeNull = iota
	OpcodeHandshake
	OpcodeChat
	OpcodeNotify
	OpcodeStabilizeRequest
	OpcodeStabilizeResponse
)

var opcodes map[Opcode]Message
var messages map[reflect.Type]Opcode
var mtx sync.Mutex

func init() {
	opcodes = map[Opcode]Message{}
	messages = map[reflect.Type]Opcode{}

	registerMessage(OpcodeHandshake, (*Handshake)(nil))
	registerMessage(OpcodeChat, (*Chat)(nil))
	registerMessage(OpcodeNotify, (*Notify)(nil))
	registerMessage(OpcodeStabilizeRequest, (*StabilizeRequest)(nil))
	registerMessage(OpcodeStabilizeResponse, (*StabilizeResponse)(nil))
}

func registerMessage(o Opcode, m interface{}) Opcode {
	typ := reflect.TypeOf(m).Elem()

	mtx.Lock()
	defer mtx.Unlock()

	if opcode, registered := messages[typ]; registered {
		return opcode
	}

	opcodes[o] = reflect.New(typ).Elem().Interface().(Message)
	messages[typ] = o

	return o
}
