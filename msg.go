package filehashes

import (
	"crypto"
	"fmt"
)

// Msg represents a message.
type Msg interface {
	String() string
}

// TaskError represents the task error message.
type TaskError string

// TaskDone represents the task done message.
type TaskDone map[crypto.Hash][]byte

// TaskStarted represent the task started message.
type TaskStarted struct{}

// TaskStopped represents the task stopped message.
type TaskStopped string

// TaskProgress represents the task progress updated message.
type TaskProgress int

func newTaskError(errMsg string) Msg {
	return TaskError(errMsg)
}

func (err TaskError) String() string {
	return fmt.Sprintf("error: %s\n", string(err))
}

func newTaskDone(m map[crypto.Hash][]byte) Msg {
	return TaskDone(m)
}

func (done TaskDone) String() string {
	str := "done\n"

	for h, checksum := range done {
		str += fmt.Sprintf("%v: %X\n", h, checksum)
	}
	return str
}

func newTaskStarted() Msg {
	return TaskStarted{}
}

func (started TaskStarted) String() string {
	return "started\n"
}

func newTaskStopped(errMsg string) Msg {
	return TaskStopped(errMsg)
}

func (stopped TaskStopped) String() string {
	return fmt.Sprintf("stopped: %s\n", string(stopped))
}

func newTaskProgress(progress int) Msg {
	return TaskProgress(progress)
}

func (progress TaskProgress) String() string {
	return fmt.Sprintf("progress: %d\n", int(progress))
}
