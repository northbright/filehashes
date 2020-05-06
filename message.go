package filehashes

import (
	"crypto"
	"encoding"
	"encoding/json"
	"hash"
)

// MessageType represents the type of messages.
type MessageType uint

const (
	ERROR MessageType = iota
	SCHEDULED
	STARTED
	STOPPED
	RESTORED
	PROGRESSUPDATED
	DONE
	UNKNOWN
	maxTYPE
)

var (
	messageTypeStrs = map[MessageType]string{
		ERROR:           "error",
		SCHEDULED:       "scheduled",
		STARTED:         "started",
		STOPPED:         "stopped",
		RESTORED:        "restored",
		PROGRESSUPDATED: "progress_updated",
		DONE:            "done",
		UNKNOWN:         "unknown",
	}
)

// State represents the hashes states.
type State struct {
	// SummedSize is summed size of file.
	SummedSize int64 `json:"summed_size,string"`
	// Data contains binary datas marshaled from hashes.
	Datas map[crypto.Hash][]byte `json:"datas"`
}

// newState returns the current hash state.
func newState(summedSize int64, hashes map[crypto.Hash]hash.Hash) *State {
	datas := map[crypto.Hash][]byte{}

	for k, v := range hashes {
		m := v.(encoding.BinaryMarshaler)
		data, _ := m.MarshalBinary()
		datas[k] = data
	}

	state := &State{SummedSize: summedSize, Datas: datas}
	return state
}

// Message represents the messages.
type Message struct {
	// Type is the type code(uint) of message.
	Type MessageType `json:"type"`
	// TypeStr is the type in string.
	TypeStr string `json:"type_str"`
	// Req is the request of sum a file.
	Req *Request `json:"request,omitempty"`
	// Data stores the data of message.
	// Each type has its own data type.
	//
	// ERROR: data is a string to store error message.
	// SCHEDULED: data is nil.
	// STARTED: data is nil.
	// STOPPED, RESTORED: data is a State. It can be used to pause / resume hashing.
	// PROGRESSUPDATED: data is a int to store the percent(0 - 100).
	// DONE: data is a map[crypto.Hash]string to store the checksums.
	Data interface{} `json:"data,omitempty"`
}

// newMessage returns a new message.
func newMessage(t MessageType, req *Request, data interface{}) *Message {
	typeStr, ok := messageTypeStrs[t]
	if !ok {
		t = UNKNOWN
		typeStr = messageTypeStrs[t]
	}

	return &Message{t, typeStr, req, data}
}

// JSON serializes a message as JSON.
// All types of messages are JSON-friendly.
// To get the meaning of Message.Data for different types of messages,
// check comments of Message struct.
func (m *Message) JSON() ([]byte, error) {
	return json.Marshal(m)
}
