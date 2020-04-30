package filehashes

import (
	"context"
)

// Manager controls the number of concurrent worker goroutines which compute files hashes.
type Manager struct {
	concurrency int
	bufferSize  int
	sem         chan struct{}
	ch          chan *Message
}

// NewManager creates a new manager and returns a channel to receive the messages.
// The channel will not be closed.
// The messages include:
//   SumError* Message: an error occurred.
//   SumScheduled* Message: a file is scheduled to sum.
//   SumStarted* Message: a file is started to sum.
//   SumStopped* Message: a file is stopped to sum.
//   SumProgress* Message: the progress of sum a file is updated.
//   SumDone* Message: it's done to sum a file done. Checksums contains the results.
func NewManager(concurrency int, bufferSize int) (*Manager, <-chan *Message) {

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
		make(chan *Message),
	}
	return m, m.ch
}

// StartSumFile starts to sum a file by given request.
func (m *Manager) StartSumFile(ctx context.Context, req *Request) {
	go func() {
		m.ch <- newMessage(SCHEDULED, req, nil)

		m.sem <- struct{}{}
		sum(ctx, m.bufferSize, req, m.ch)
		<-m.sem
	}()
}

// StartSumFiles starts to sum files by given requests.
func (m *Manager) StartSumFiles(ctx context.Context, reqs []*Request) {
	for _, req := range reqs {
		go func(req *Request) {
			m.ch <- newMessage(SCHEDULED, req, nil)

			m.sem <- struct{}{}
			sum(ctx, m.bufferSize, req, m.ch)
			<-m.sem
		}(req)
	}
}
