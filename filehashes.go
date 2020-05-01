package filehashes

import (
	"bufio"
	"context"
	"crypto"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

var (
	DefaultConcurrency = 4
	DefaultBufferSize  = 8 * 1024 * 1024
	DefaultHashAlgs    = []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
	}

	ErrNoFileToHash = errors.New("no file to hash")
	ErrFileIsDir    = errors.New("file is dir")
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

// StartSumFile starts to compute checksums of a file by given hash algorithms.
// You may specify one or more hash algorithm(s) in hashAlgs parameter(s).
// It'll start a new goroutine to compoute checksums.
// It returns a channel to receive the messages,
// the channel will be closed after the goroutine exited.
// You may use for - range to read the messages.
func StartSumFile(ctx context.Context, bufferSize int, file string, hashAlgs []crypto.Hash) <-chan *Message {
	ch := make(chan *Message)
	req := &Request{file, hashAlgs}

	go func() {
		sum(ctx, bufferSize, req, ch)
		close(ch)
	}()

	return ch
}

// StartSumFiles starts to computes checksums of files.
// reqs are the requests which contains files and hash algorithms.
// It'll start a new goroutine to compoute checksums.
// It returns a channel to receive the messages,
// the channel will be closed after the goroutine exited.
// You may use for - range to read the messages.
func StartSumFiles(ctx context.Context, concurrency int, bufferSize int, reqs []*Request) <-chan *Message {
	ch := make(chan *Message)

	go func() {
		sumFiles(ctx, concurrency, bufferSize, reqs, ch)
		close(ch)
	}()

	return ch
}

func sumFiles(ctx context.Context, concurrency int, bufferSize int, reqs []*Request, ch chan *Message) {

	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	count := len(reqs)
	if count <= 0 {
		ch <- newMessage(ERROR, nil, ErrNoFileToHash.Error())
		return
	}

	sem := make(chan struct{}, concurrency)

	for i := 0; i < count; i++ {
		ch <- newMessage(SCHEDULED, reqs[i], nil)

		// After first "concurrency" amount of goroutines started,
		// It'll block starting new goroutines until one running goroutine finishs.
		sem <- struct{}{}

		go func(req *Request) {
			defer func() { <-sem }()
			// Do the work
			sum(ctx, bufferSize, req, ch)
		}(reqs[i])
	}

	// After last goroutine is started,
	// there're still "concurrency" amount of goroutines running.
	// Make sure wait all goroutines to finish.
	for j := 0; j < cap(sem); j++ {
		sem <- struct{}{}
	}

	// All goroutines done.
}

func sum(ctx context.Context, bufferSize int, req *Request, ch chan *Message) {
	// Send sum started message.
	ch <- newMessage(STARTED, req, nil)

	// Open file.
	f, fi, err := openFile(req.File)
	if err != nil {
		ch <- newMessage(ERROR, req, err.Error())
		return
	}
	defer f.Close()

	hashes := map[crypto.Hash]hash.Hash{}
	for _, h := range req.HashAlgs {
		if h.Available() {
			hashes[h] = h.New()
		}
	}

	r := bufio.NewReaderSize(f, bufferSize)
	buf := make([]byte, bufferSize)

	size := fi.Size()
	summedSize := int64(0)
	oldProgress := 0
	progress := 0

LOOP:
	for {
		select {
		case <-ctx.Done():
			// Send stopped message.
			ch <- newMessage(STOPPED, req, ctx.Err().Error())
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				// Send error message.
				ch <- newMessage(ERROR, req, err.Error())
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
				progress = int(summedSize * 100 / size)
			}
			if progress != oldProgress {
				// Send progress updated message.
				ch <- newMessage(PROGRESSUPDATED, req, progress)
				oldProgress = progress
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
		ch <- newMessage(PROGRESSUPDATED, req, 100)
	}

	ch <- newMessage(DONE, req, checksums)
}
