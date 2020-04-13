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
	DefaultBufferSize = 8 * 1024 * 1024
	ErrFileIsDir      = fmt.Errorf("file is dir")
	ErrFileSizeIsZero = fmt.Errorf("file size is 0")
)

// Sum computes the file checksums by given hash algorithms.
// You may specify one or more hash algorithm(s) in hashAlgs parameter(s).
// Sum will start a new goroutine to compoute checksums.
// It returns a channel to receive the messages,
// you may use for - range to read the messages.
func Sum(ctx context.Context, bufferSize int, filePath string, hashAlgs ...crypto.Hash) (<-chan Msg, error) {
	ch := make(chan Msg)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, ErrFileIsDir
	}

	if fi.Size() == 0 {
		return nil, ErrFileSizeIsZero
	}

	hashes := map[crypto.Hash]hash.Hash{}
	for _, h := range hashAlgs {
		hashes[h] = h.New()
	}

	go func() {
		sum(ctx, bufferSize, filePath, fi, f, hashes, ch)
	}()

	return ch, nil
}

func sum(ctx context.Context, bufferSize int, filePath string, fi os.FileInfo, f *os.File, hashes map[crypto.Hash]hash.Hash, ch chan Msg) {
	defer func() {
		f.Close()
		close(ch)
	}()

	r := bufio.NewReaderSize(f, bufferSize)
	buf := make([]byte, bufferSize)

	size := fi.Size()
	summedSize := int64(0)
	oldProgress := 0
	progress := 0

	// Sum started.
	ch <- newTaskStarted()
LOOP:
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			// Send stopped message.
			ch <- newTaskStopped(err.Error())
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				// Send error message.
				ch <- newTaskError(err.Error())
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
				ch <- newTaskProgress(progress)
				oldProgress = progress
			}
		}

	}

	// Done. Send the checksums.
	checksums := map[crypto.Hash][]byte{}
	for k, v := range hashes {
		checksums[k] = v.Sum(nil)
	}

	ch <- newTaskDone(checksums)
}
