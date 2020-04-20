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
	DefaultHashAlgs    = []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
	}
	ErrFileIsDir = fmt.Errorf("file is dir")
)

// Request represents the request of sum a single file.
type Request struct {
	File     string        `json:"file"`
	HashAlgs []crypto.Hash `json:"hash_algs"`
}

func (req *Request) String() string {
	str := fmt.Sprintf("file: %s(hash algs:", req.File)
	for _, h := range req.HashAlgs {
		str += fmt.Sprintf(" %v", h)
	}
	str += ")"
	return str
}

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
func Sum(ctx context.Context, bufferSize int, hashAlgs []crypto.Hash, file string) <-chan Msg {
	ch := make(chan Msg)

	go func() {
		sum(ctx, bufferSize, hashAlgs, file, ch)
	}()

	return ch
}

func sum(ctx context.Context, bufferSize int, hashAlgs []crypto.Hash, file string, ch chan Msg) {
	defer func() {
		close(ch)
	}()

	// Use default hash algorithms if it's not set.
	if len(hashAlgs) == 0 {
		hashAlgs = DefaultHashAlgs
	}

	// Send sum started message.
	req := &Request{file, hashAlgs}
	ch <- newSumStartedMsg(req)

	// Open file.
	f, fi, err := openFile(file)
	if err != nil {
		ch <- newErrorMsg(req, err.Error())
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

LOOP:
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			// Send stopped message.
			ch <- newSumStoppedMsg(req, err.Error())
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				// Send error message.
				ch <- newErrorMsg(req, err.Error())
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
				ch <- newSumProgressMsg(req, progress)
				oldProgress = progress
			}
		}
	}

	// Done. Send the checksums.
	checksums := map[crypto.Hash][]byte{}
	for k, v := range hashes {
		checksums[k] = v.Sum(nil)
	}

	ch <- newSumDoneMsg(req, checksums)
}
