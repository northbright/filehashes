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
	BufferSize        = 8 * 1024 * 1024
	ErrFileIsDir      = fmt.Errorf("file is dir")
	ErrFileSizeIsZero = fmt.Errorf("file size is 0")
)

func Sum(ctx context.Context, filePath string, hashTypes ...crypto.Hash) (<-chan Msg, error) {
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
	for _, h := range hashTypes {
		hashes[h] = h.New()
	}

	go func() {
		sum(ctx, filePath, fi, f, hashes, ch)
	}()

	return ch, nil
}

func sum(ctx context.Context, filePath string, fi os.FileInfo, f *os.File, hashes map[crypto.Hash]hash.Hash, ch chan Msg) {
	defer func() {
		f.Close()
		close(ch)
	}()

	r := bufio.NewReaderSize(f, BufferSize)
	buf := make([]byte, BufferSize)

	size := fi.Size()
	summedSize := int64(0)
	oldProgress := 0
	progress := 0
	errMsg := ""

	ch <- newSumStartedMsg(fi, filePath)
LOOP:
	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			errMsg = fmt.Sprintf("%v", err)
			ch <- newSumStoppedMsg(fi, filePath, errMsg)
			return
		default:
			n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				ch <- newSumErrorMsg(fi, filePath, errMsg)
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
				ch <- newSumProgressUpdatedMsg(
					fi,
					filePath,
					progress,
				)
				oldProgress = progress
			}
		}

	}

	checksums := map[crypto.Hash][]byte{}
	for k, v := range hashes {
		checksums[k] = v.Sum(nil)
	}

	ch <- newSumDoneMsg(fi, filePath, checksums)
}
