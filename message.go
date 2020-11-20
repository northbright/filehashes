package filehashes

import (
	"encoding/json"
)

// MessageType represents the type of messages,
// which are generated when computing file hashes.
type MessageType uint

const (
	// An error occured.
	ERROR MessageType = iota
	// Task is scheduled.
	SCHEDULED
	// Task is started.
	STARTED
	// Task is stopped.
	STOPPED
	// Task is restored.
	RESTORED
	// Progress of task is updated.
	PROGRESS_UPDATED
	// Task is done.
	DONE
	// Goroutine of the task exited.
	// It's send after one of ERROR / STOPPED / DONE message.
	EXITED
	// Unknown message type.
	UNKNOWN
	maxTYPE
)

var (
	messageTypeStrs = map[MessageType]string{
		ERROR:            "error",
		SCHEDULED:        "scheduled",
		STARTED:          "started",
		STOPPED:          "stopped",
		RESTORED:         "restored",
		PROGRESS_UPDATED: "progress_updated",
		EXITED:           "exited",
		DONE:             "done",
		UNKNOWN:          "unknown",
	}
)

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
	// STOPPED, RESTORED: data is a updated *Request with *State.
	// It can be used to pause / resume hashing.
	// PROGRESS_UPDATED: data is a int to store the percent(0 - 100).
	// DONE: data is a map[crypto.Hash]string to store the checksums.
	// EXITED: data is nil.
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
