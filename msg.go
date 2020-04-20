package filehashes

import (
	"crypto"
	"fmt"
)

// Msg represents a message.
type Msg interface {
	String() string
}

// ErrorMsg represents the sum error message.
type ErrorMsg struct {
	*Request
	ErrMsg string `json:"err_msg"`
}

// SumStartedMsg represent the sum of single file started message.
type SumStartedMsg struct {
	*Request
}

// SumStoppedMsg represents the sum stopped message.
type SumStoppedMsg struct {
	*Request
	ErrMsg string `json:"err_msg"`
}

// SumProgressMsg represents the sum of single file progress updated message.
type SumProgressMsg struct {
	*Request
	Progress int `json:"progress"`
}

// SumDoneMsg represents the sum of single file done message.
type SumDoneMsg struct {
	*Request
	Checksums map[crypto.Hash][]byte `json:"checksums"`
}

func newErrorMsg(r *Request, errMsg string) Msg {
	return ErrorMsg{r, errMsg}
}

// String return a formated message string of ErrorMsg.
func (m ErrorMsg) String() string {
	return fmt.Sprintf("sum %v error: %s", m.Request, m.ErrMsg)
}

func newSumStartedMsg(r *Request) Msg {
	return SumStartedMsg{r}
}

// String return a formated message string of SumStartedMsg.
func (m SumStartedMsg) String() string {
	return fmt.Sprintf("sum %v started", m.Request)
}

func newSumStoppedMsg(r *Request, errMsg string) Msg {
	return SumStoppedMsg{r, errMsg}
}

// String return a formated message string of SumStoppedMsg.
func (m SumStoppedMsg) String() string {
	return fmt.Sprintf("sum %v stopped", m.Request)
}

func newSumProgressMsg(r *Request, progress int) Msg {
	return SumProgressMsg{r, progress}
}

// String return a formated message string of SumProgressMsg.
func (m SumProgressMsg) String() string {
	return fmt.Sprintf("sum %v progress: %d", m.Request, m.Progress)
}

func newSumDoneMsg(r *Request, m map[crypto.Hash][]byte) Msg {
	return SumDoneMsg{r, m}
}

// String return a formated message string of SumDoneMsg.
func (m SumDoneMsg) String() string {
	str := fmt.Sprintf("sum %v done, hashes:", m.Request)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf(" %v: %X", h, checksum)
	}
	return str
}
