package filehashes

import (
	"crypto"
	"fmt"
)

// Msg represents a message.
type Msg interface {
	String() string
}

// NoFileError reprents the error of no file to sum.
type NoFileError struct{}

// SumError represents the sum error message.
type SumError struct {
	File   string
	ErrMsg string
}

// SumDone represents the sum of single file done message.
type SumDone struct {
	File      string
	Checksums map[crypto.Hash][]byte
}

// SumAllDone represets the sum of all files done messages.
type SumAllDone struct {
	Files []string
}

// SumStarted represent the sum of single file started message.
type SumStarted struct {
	File string
}

// SumStopped represents the sum stopped message.
type SumStopped struct {
	File   string
	ErrMsg string
}

// SumProgress represents the sum of single file progress updated message.
type SumProgress struct {
	File     string
	Progress int
}

func newNoFileError() Msg {
	return NoFileError{}
}

// String return a formated message string of NoFileError.
func (m NoFileError) String() string {
	return "no file to sum"
}

func newSumError(file string, errMsg string) Msg {
	return SumError{file, errMsg}
}

// String return a formated message string of SumError.
func (m SumError) String() string {
	return fmt.Sprintf("sum %s error: %s\n", m.File, m.ErrMsg)
}

func newSumDone(file string, m map[crypto.Hash][]byte) Msg {
	return SumDone{file, m}
}

// String return a formated message string of SumDone.
func (m SumDone) String() string {
	str := fmt.Sprintf("sum %s done\n", m.File)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf("%v: %X\n", h, checksum)
	}
	return str
}

func newSumAllDone(files []string) Msg {
	return SumAllDone{files}
}

// String return a formated message string of SumAllDone.
func (m SumAllDone) String() string {
	return "sum all files done\n"
}

func newSumStarted(file string) Msg {
	return SumStarted{file}
}

// String return a formated message string of SumStarted.
func (m SumStarted) String() string {
	return fmt.Sprintf("sum %s started\n", m.File)
}

func newSumStopped(file string, errMsg string) Msg {
	return SumStopped{file, errMsg}
}

// String return a formated message string of SumStopped.
func (m SumStopped) String() string {
	return fmt.Sprintf("sum %s stopped: %s\n", m.File, m.ErrMsg)
}

func newSumProgress(file string, progress int) Msg {
	return SumProgress{file, progress}
}

// String return a formated message string of SumProgress.
func (m SumProgress) String() string {
	return fmt.Sprintf("sum %s progress: %d\n", m.File, m.Progress)
}
