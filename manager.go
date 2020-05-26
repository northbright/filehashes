package filehashes

import (
	"bufio"
	"context"
	"crypto"
	"encoding"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

// Manager controls the number of concurrent worker goroutines which compute files hashes.
type Manager struct {
	concurrency int
	bufferSize  int
	sem         chan struct{}
	ch          chan *Message
}

var (
	DefaultConcurrency = 4
	DefaultBufferSize  = 8 * 1024 * 1024

	ErrNoFileToHash         = errors.New("no file to hash")
	ErrNoHashFuncs          = errors.New("no hash functions")
	ErrHashFuncNotAvailable = errors.New("hash function is not available")
	ErrInvalidHashState     = errors.New("invalid hash state")
	ErrFileIsDir            = errors.New("file is dir")
)

func openFile(file string) (*os.File, os.FileInfo, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	if fi.IsDir() {
		return nil, nil, ErrFileIsDir
	}

	return f, fi, nil
}

// NewManager creates a new manager and returns a channel to receive the messages.
// The channel will not be closed.
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

// Start starts to sum file by given request.
// Caller should import hash function packages for the hash functions.
// e.g. import (_ "crypto/md5")
// It'll start a new goroutine for each request.
func (m *Manager) Start(ctx context.Context, req *Request) {
	go func() {
		m.sum(ctx, req)
	}()
}

func (m *Manager) sum(ctx context.Context, req *Request) {
	defer func() {
		<-m.sem
		// Task(goroutine) exited.
		m.ch <- newMessage(EXITED, req, nil)
	}()

	// Task is scheduled.
	m.ch <- newMessage(SCHEDULED, req, nil)

BEFORE_START:
	// This loop make it possible to cancel the task even it's scheduled:
	// blocked at m.mem <- struct{}{}.
	for {
		select {
		case <-ctx.Done():
			// Stopped before start, use previous request.
			resumeReq := req
			// Task is stopped.
			m.ch <- newMessage(STOPPED, req, resumeReq)
			return
		case m.sem <- struct{}{}:
			break BEFORE_START
		}
	}

	// Check hash functions.
	if len(req.HashFuncs) == 0 {
		m.ch <- newMessage(ERROR, req, ErrNoHashFuncs.Error())
		return
	}

	// Create hash map.
	// Key is the crypto.Hash(hash function).
	// Value is the hash.Hash.
	hashes := map[crypto.Hash]hash.Hash{}
	for _, h := range req.HashFuncs {
		if !h.Available() {
			m.ch <- newMessage(ERROR, req, ErrHashFuncNotAvailable.Error())
			return
		}
		hashes[h] = h.New()
	}

	// Open file.
	f, fi, err := openFile(req.File)
	if err != nil {
		m.ch <- newMessage(ERROR, req, err.Error())
		return
	}
	defer f.Close()

	// Create bufio.Reader.
	r := bufio.NewReaderSize(f, m.bufferSize)
	buf := make([]byte, m.bufferSize)

	size := fi.Size()
	summedSize := int64(0)
	oldProgress := 0
	progress := 0

	// Restore previous hash states.
	if req.Stat != nil {
		if !req.StateIsValid() {
			m.ch <- newMessage(ERROR, req, ErrInvalidHashState.Error())
			return
		}

		// Restore saved hash states
		for k, v := range hashes {
			u := v.(encoding.BinaryUnmarshaler)
			data := req.Stat.Datas[k]

			if err := u.UnmarshalBinary(data); err != nil {
				m.ch <- newMessage(ERROR, req, err.Error())
				return
			}
		}

		// Seek file offset and update summed size.
		summedSize = req.Stat.SummedSize
		if _, err := f.Seek(summedSize, os.SEEK_SET); err != nil {
			m.ch <- newMessage(ERROR, req, err.Error())
			return
		}

		// Task is restored.
		m.ch <- newMessage(RESTORED, req, req.Stat)

		// Update progress.
		progress = req.Stat.Progress
		oldProgress = progress
		m.ch <- newMessage(PROGRESS_UPDATED, req, progress)
	} else {
		// Task is started.
		m.ch <- newMessage(STARTED, req, nil)
	}

LOOP:
	for {
		select {
		case <-ctx.Done():
			// Save state and create a request to resume.
			state := newState(summedSize, progress, hashes)
			resumeReq := &Request{req.File, req.HashFuncs, state}
			// Task is stopped.
			m.ch <- newMessage(STOPPED, req, resumeReq)
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				// Send error message.
				m.ch <- newMessage(ERROR, req, err.Error())
				return
			}

			if n == 0 {
				break LOOP
			}

			// Adds more data to the running hash.
			for _, h := range hashes {
				if n, err = h.Write(buf[:n]); err != nil {
					return
				}
			}

			summedSize += int64(n)
			if size != 0 {
				progress = int(float64(summedSize) / float64(size) * 100)
			}
			if progress != oldProgress {
				oldProgress = progress
				// Task progress is updated.
				m.ch <- newMessage(PROGRESS_UPDATED, req, progress)
			}
		}
	}

	// Done. Send the checksums.
	checksums := map[crypto.Hash]string{}
	for k, v := range hashes {
		checksums[k] = fmt.Sprintf("%X", v.Sum(nil))
	}

	// Update progress for 0-size file.
	if size == 0 {
		m.ch <- newMessage(PROGRESS_UPDATED, req, 100)
	}

	// Task is done.
	m.ch <- newMessage(DONE, req, checksums)
}
