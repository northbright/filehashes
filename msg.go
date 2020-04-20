package filehashes

import (
	"crypto"
	"fmt"
)

// Msg represents a message.
type Msg interface {
	String() string
}

// NoFileErrorMsg reprents the error of no file to sum.
type NoFileErrorMsg struct{}

// ErrorMsg represents the sum error message.
type ErrorMsg struct {
	File   string `json:"file"`
	ErrMsg string `json:"err_msg"`
}

// SumDoneMsg represents the sum of single file done message.
type SumDoneMsg struct {
	File      string                 `json:"file"`
	Checksums map[crypto.Hash][]byte `json:"checksums"`
}

// SumStartedMsg represent the sum of single file started message.
type SumStartedMsg struct {
	File string `json:"file"`
}

// SumStoppedMsg represents the sum stopped message.
type SumStoppedMsg struct {
	File   string `json:"file"`
	ErrMsg string `json:"err_msg"`
}

// SumProgressMsg represents the sum of single file progress updated message.
type SumProgressMsg struct {
	File     string `json:"file"`
	Progress int    `json:"progress"`
}

func newErrorMsg(file string, errMsg string) Msg {
	return ErrorMsg{file, errMsg}
}

// String return a formated message string of ErrorMsg.
func (m ErrorMsg) String() string {
	return fmt.Sprintf("sum %s error: %s\n", m.File, m.ErrMsg)
}

func newSumDoneMsg(file string, m map[crypto.Hash][]byte) Msg {
	return SumDoneMsg{file, m}
}

// String return a formated message string of SumDoneMsg.
func (m SumDoneMsg) String() string {
	str := fmt.Sprintf("sum %s done\n", m.File)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf("%v: %X\n", h, checksum)
	}
	return str
}

func newSumStartedMsg(file string) Msg {
	return SumStartedMsg{file}
}

// String return a formated message string of SumStartedMsg.
func (m SumStartedMsg) String() string {
	return fmt.Sprintf("sum %s started\n", m.File)
}

func newSumStoppedMsg(file string, errMsg string) Msg {
	return SumStoppedMsg{file, errMsg}
}

// String return a formated message string of SumStoppedMsg.
func (m SumStoppedMsg) String() string {
	return fmt.Sprintf("sum %s stopped: %s\n", m.File, m.ErrMsg)
}

func newSumProgressMsg(file string, progress int) Msg {
	return SumProgressMsg{file, progress}
}

// String return a formated message string of SumProgressMsg.
func (m SumProgressMsg) String() string {
	return fmt.Sprintf("sum %s progress: %d\n", m.File, m.Progress)
}
