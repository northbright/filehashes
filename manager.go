package filehashes

import (
	"context"
)

// Manager controls the number of concurrent worker goroutines which compute files hashes.
type Manager struct {
	concurrency int
	bufferSize  int
	sem         chan struct{}
	ch          chan Msg
}

// NewManager creates a new manager and returns a channel to receive the messages.
// The channel will not be closed.
// The messages include:
//   SumErrorMsg: an error occurred.
//   SumScheduledMsg: a file is scheduled to sum.
//   SumStartedMsg: a file is started to sum.
//   SumStoppedMsg: a file is stopped to sum.
//   SumProgressMsg: the progress of sum a file is updated.
//   SumDoneMsg: it's done to sum a file done. Checksums contains the results.
func NewManager(concurrency int, bufferSize int) (*Manager, <-chan Msg) {

	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	m := &Manager{
		concurrency,
		bufferSize,
		make(chan struct{}, concurrency),
		make(chan Msg),
	}
	return m, m.ch
}

// StartSumFile starts to sum a file by given request.
func (m *Manager) StartSumFile(ctx context.Context, req *Request) {
	go func() {
		m.ch <- newSumScheduledMsg(req)

		m.sem <- struct{}{}
		sum(ctx, m.bufferSize, req, m.ch)
		<-m.sem
	}()
}
