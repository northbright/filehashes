package filehashes

import (
	"crypto"
	"fmt"
)

// Msg represents a message.
type Msg interface {
	String() string
}

// SumErrorMsg represents the error message.
type SumErrorMsg struct {
	*Request
	Msg string `json:"message"`
}

// SumScheduledMsg represents the message that the request is scheduled.
type SumScheduledMsg struct {
	*Request
}

// SumStartedMsg represent the message that a file is started to sum.
type SumStartedMsg struct {
	*Request
}

// SumStoppedMsg represents the message that a file is stopped to sum.
type SumStoppedMsg struct {
	*Request
	ErrMsg string `json:"err_msg"`
}

// SumProgressMsg represents the message that progress of sum a file is updated.
type SumProgressMsg struct {
	*Request
	Progress int `json:"progress"`
}

// SumDoneMsg represents the message that sum a file is done.
type SumDoneMsg struct {
	*Request
	Checksums map[crypto.Hash][]byte `json:"checksums"`
}

func newSumErrorMsg(r *Request, e error) Msg {
	return &SumErrorMsg{r, e.Error()}
}

// String returns a formated message string of SumErrorMsg.
func (m SumErrorMsg) String() string {
	return fmt.Sprintf("sum %v error: %s", m.Request, m.Msg)
}

func newSumScheduledMsg(r *Request) Msg {
	return SumScheduledMsg{r}
}

// String returns a formated message string of SumScheduledMsg.
func (m SumScheduledMsg) String() string {
	return fmt.Sprintf("sum %v scheduled", m.Request)
}

func newSumStartedMsg(r *Request) Msg {
	return SumStartedMsg{r}
}

// String returns a formated message string of SumStartedMsg.
func (m SumStartedMsg) String() string {
	return fmt.Sprintf("sum %v started", m.Request)
}

func newSumStoppedMsg(r *Request, errMsg string) Msg {
	return SumStoppedMsg{r, errMsg}
}

// String returns a formated message string of SumStoppedMsg.
func (m SumStoppedMsg) String() string {
	return fmt.Sprintf("sum %v stopped", m.Request)
}

func newSumProgressMsg(r *Request, progress int) Msg {
	return SumProgressMsg{r, progress}
}

// String returns a formated message string of SumProgressMsg.
func (m SumProgressMsg) String() string {
	return fmt.Sprintf("sum %v progress: %d", m.Request, m.Progress)
}

func newSumDoneMsg(r *Request, m map[crypto.Hash][]byte) Msg {
	return SumDoneMsg{r, m}
}

// String returns a formated message string of SumDoneMsg.
func (m SumDoneMsg) String() string {
	str := fmt.Sprintf("sum %v done, hashes:", m.Request)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf(" %v: %X", h, checksum)
	}
	return str
}
