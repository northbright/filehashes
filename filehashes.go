package filehashes

import (
	"bufio"
	"context"
	"crypto"
	"fmt"
	"hash"
	"io"
	"os"
)

var (
	DefaultConcurrency = 4
	DefaultBufferSize  = 8 * 1024 * 1024
	ErrFileIsDir       = fmt.Errorf("file is dir")
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

// Sum computes the file checksums by given hash algorithms.
// You may specify one or more hash algorithm(s) in hashAlgs parameter(s).
// Sum will start a new goroutine to compoute checksums.
// It returns a channel to receive the messages,
// you may use for - range to read the messages.
func Sum(ctx context.Context, concurrency int, bufferSize int, hashAlgs []crypto.Hash, files []string) <-chan Msg {
	ch := make(chan Msg)

	go func() {
		sumAll(ctx, concurrency, bufferSize, hashAlgs, files, ch)
	}()

	return ch
}

func sumAll(ctx context.Context, concurrency int, bufferSize int, hashAlgs []crypto.Hash, files []string, ch chan Msg) {
	defer func() {
		close(ch)
	}()

	if concurrency <= 0 {
		concurrency = DefaultConcurrency
	}

	count := len(files)
	if count <= 0 {
		ch <- newNoFileError()
		return
	}

	sem := make(chan struct{}, concurrency)

	for i := 0; i < count; i++ {
		// After first "concurrency" amount of goroutines started,
		// It'll block starting new goroutines until one running goroutine finishs.
		sem <- struct{}{}

		go func(file string) {
			defer func() { <-sem }()
			// Do the work
			sum(ctx, bufferSize, hashAlgs, file, ch)
		}(files[i])
	}

	// After last goroutine is started,
	// there're still "concurrency" amount of goroutines running.
	// Make sure wait all goroutines to finish.
	for j := 0; j < cap(sem); j++ {
		sem <- struct{}{}
	}

	// All goroutines done.
	ch <- newSumAllDone(files)
}

func sum(ctx context.Context, bufferSize int, hashAlgs []crypto.Hash, file string, ch chan Msg) {
	f, fi, err := openFile(file)
	if err != nil {
		ch <- newSumError(file, err.Error())
		return
	}
	defer f.Close()

	hashes := map[crypto.Hash]hash.Hash{}
	for _, h := range hashAlgs {
		hashes[h] = h.New()
	}

	r := bufio.NewReaderSize(f, bufferSize)
	buf := make([]byte, bufferSize)

	size := fi.Size()
	summedSize := int64(0)
	oldProgress := 0
	progress := 0

	// Sum started.
	ch <- newSumStarted(file)
LOOP:
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			// Send stopped message.
			ch <- newSumStopped(file, err.Error())
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				// Send error message.
				ch <- newSumError(file, err.Error())
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
			progress = int(summedSize * 100 / size)
			if progress != oldProgress {
				// Send progress updated message.
				ch <- newSumProgress(file, progress)
				oldProgress = progress
			}
		}

	}

	// Done. Send the checksums.
	checksums := map[crypto.Hash][]byte{}
	for k, v := range hashes {
		checksums[k] = v.Sum(nil)
	}

	ch <- newSumDone(file, checksums)
}
