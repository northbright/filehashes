package filehashes

import (
	"crypto"
	"fmt"
	"os"
)

type MsgType uint8

type Msg interface {
	String() string
}

type SumErrorMsg struct {
	os.FileInfo
	FilePath string
	ErrMsg   string
}

func (m *SumErrorMsg) String() string {
	return fmt.Sprintf("sum %s error: %s\n", m.FilePath, m.ErrMsg)
}

type SumDoneMsg struct {
	os.FileInfo
	FilePath  string
	Checksums map[crypto.Hash][]byte
}

func (m *SumDoneMsg) String() string {
	str := fmt.Sprintf("sum %s done:\n--------------------\n", m.FilePath)

	for h, checksum := range m.Checksums {
		str += fmt.Sprintf("%v: %X\n", h, checksum)
	}
	return str
}

type SumStartedMsg struct {
	os.FileInfo
	FilePath string
}

func (m *SumStartedMsg) String() string {
	return fmt.Sprintf("sum %s started\n", m.FilePath)
}

type SumStoppedMsg struct {
	os.FileInfo
	FilePath string
	ErrMsg   string
}

func (m *SumStoppedMsg) String() string {
	return fmt.Sprintf("sum %s stopped: %s\n", m.FilePath, m.ErrMsg)
}

type SumProgressUpdatedMsg struct {
	os.FileInfo
	FilePath string
	Progress int
}

func (m *SumProgressUpdatedMsg) String() string {
	return fmt.Sprintf("sum %s progress: %d\n", m.FilePath, m.Progress)
}

func newSumErrorMsg(fi os.FileInfo, p string, errMsg string) *SumErrorMsg {
	m := &SumErrorMsg{
		fi,
		p,
		errMsg,
	}
	return m
}

func newSumDoneMsg(fi os.FileInfo, p string, checksums map[crypto.Hash][]byte) *SumDoneMsg {
	m := &SumDoneMsg{
		fi,
		p,
		checksums,
	}
	return m
}

func newSumStartedMsg(fi os.FileInfo, p string) *SumStartedMsg {
	m := &SumStartedMsg{
		fi,
		p,
	}
	return m
}

func newSumStoppedMsg(fi os.FileInfo, p string, errMsg string) *SumStoppedMsg {
	m := &SumStoppedMsg{
		fi,
		p,
		errMsg,
	}
	return m
}

func newSumProgressUpdatedMsg(fi os.FileInfo, p string, progress int) *SumProgressUpdatedMsg {
	m := &SumProgressUpdatedMsg{
		fi,
		p,
		progress,
	}
	return m
}
