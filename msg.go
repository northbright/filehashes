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
	Request
	ErrMsg string `json:"err_msg"`
}

// SumStartedMsg represent the sum of single file started message.
type SumStartedMsg struct {
	Request
}

// SumStoppedMsg represents the sum stopped message.
type SumStoppedMsg struct {
	Request
	ErrMsg string `json:"err_msg"`
}

// SumProgressMsg represents the sum of single file progress updated message.
type SumProgressMsg struct {
	Request
	Progress int `json:"progress"`
}

// SumDoneMsg represents the sum of single file done message.
type SumDoneMsg struct {
	Request
	Checksums map[crypto.Hash][]byte `json:"checksums"`
}

func newErrorMsg(r Request, errMsg string) Msg {
	return ErrorMsg{r, errMsg}
}

// String return a formated message string of ErrorMsg.
func (m ErrorMsg) String() string {
	return fmt.Sprintf("sum %s error: %s\n", m.File, m.ErrMsg)
}

func newSumStartedMsg(r Request) Msg {
	return SumStartedMsg{r}
}

// String return a formated message string of SumStartedMsg.
func (m SumStartedMsg) String() string {
	return fmt.Sprintf("sum %s started\n", m.File)
}

func newSumStoppedMsg(r Request, errMsg string) Msg {
	return SumStoppedMsg{r, errMsg}
}

// String return a formated message string of SumStoppedMsg.
func (m SumStoppedMsg) String() string {
	return fmt.Sprintf("sum %s stopped: %s\n", m.File, m.ErrMsg)
}

func newSumProgressMsg(r Request, progress int) Msg {
	return SumProgressMsg{r, progress}
}

// String return a formated message string of SumProgressMsg.
func (m SumProgressMsg) String() string {
	return fmt.Sprintf("sum %s progress: %d\n", m.File, m.Progress)
}

func newSumDoneMsg(r Request, m map[crypto.Hash][]byte) Msg {
	return SumDoneMsg{r, m}
}

// String return a formated message string of SumDoneMsg.
func (m SumDoneMsg) String() string {
	str := fmt.Sprintf("sum %s done\n", m.File)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf("%v: %X\n", h, checksum)
	}
	return str
}
