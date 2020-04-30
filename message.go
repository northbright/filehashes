package filehashes

// MessageType represents the type of messages.
type MessageType uint

const (
	ERROR MessageType = iota
	SCHEDULED
	STARTED
	STOPPED
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
		PROGRESSUPDATED: "progress_updated",
		DONE:            "done",
		UNKNOWN:         "unknown",
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
	// STOPPED: data is a string to store the reason of stopped.
	// PROGRESSUPDATED: data is a int to store the percent(0 - 100).
	// DONE: data is a map[crypto.Hash][]byte to store the checksums.
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
