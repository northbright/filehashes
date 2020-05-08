package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	_ "crypto/sha256"
	"log"

	"github.com/northbright/filehashes"
)

func ExampleStart() {
	ctx := context.Background()
	// Set sum concurrency, defalut value's filehashes.DefaultConcurrency.
	concurrency := 1
	// Set buffer size.
	bufferSize := filehashes.DefaultBufferSize

	// reqs contains filehashes.Request,
	// which specify the file to hash and the hash function(s).
	reqs := []*filehashes.Request{
		filehashes.NewRequest("filehashes.go", []crypto.Hash{crypto.MD5}, nil),
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.SHA1}, nil),
		filehashes.NewRequest("go.mod", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
	}

	// Start returns a channel to receive messages.
	ch := filehashes.Start(
		ctx,
		concurrency,
		bufferSize,
		reqs,
	)

	// Consume messages.
	for m := range ch {
		switch m.Type {
		case filehashes.ERROR,
			filehashes.SCHEDULED,
			filehashes.STARTED,
			filehashes.STOPPED,
			filehashes.RESTORED,
			filehashes.PROGRESSUPDATED,
			filehashes.DONE:
			buf, _ := m.JSON()
			log.Printf("message: %v", string(buf))
		default:
			log.Printf("unknown message: %v", m)
		}
	}

	// sum goroutine exited.
	log.Printf("sum goroutine exited")

	// Output:
}
