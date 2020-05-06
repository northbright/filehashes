package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	"log"
	"time"

	"github.com/northbright/filehashes"
)

func ExampleManager_Start() {
	ctx := context.Background()

	// You may set the number of concurrent worker goroutines.
	// It's filehashes.DefaultConcurrency by default.
	concurrency := 1
	bufferSize := filehashes.DefaultBufferSize

	man, ch := filehashes.NewManager(concurrency, bufferSize)

	// Create requests.
	reqs := []*filehashes.Request{
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.MD5}, nil),
		filehashes.NewRequest("filehashes.go", []crypto.Hash{crypto.SHA1}, nil),
		filehashes.NewRequest("go.mod", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
	}

	// Start to sum files.
	man.Start(ctx, reqs)

	// Create a timeout to exit.
	chTimeout := time.After(5 * time.Second)

	// Consume messages.
	for {
		select {
		case <-chTimeout:
			log.Printf("timeout")
			return
		case m := <-ch:
			switch m.Type {
			case filehashes.ERROR,
				filehashes.SCHEDULED,
				filehashes.STARTED,
				filehashes.STOPPED,
				filehashes.RESTORED,
				filehashes.PROGRESSUPDATED,
				filehashes.DONE:
				// All messages can be marshaled to JSON.
				buf, _ := m.JSON()
				log.Printf("message: %v", string(buf))

			default:
				log.Printf("unknown message: %v", m)
			}
		}
	}

	// Output:
}
