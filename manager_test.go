package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/northbright/filehashes"
)

func ExampleManager_StartSumFiles() {
	ctx := context.Background()

	// You may set the number of concurrent worker goroutines.
	// It's filehashes.DefaultConcurrency by default.
	concurrency := 1
	bufferSize := filehashes.DefaultBufferSize

	man, ch := filehashes.NewManager(concurrency, bufferSize)

	// Create requests.
	reqs := []*filehashes.Request{
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.MD5}),
		filehashes.NewRequest("filehashes.go", []crypto.Hash{crypto.SHA1}),
		filehashes.NewRequest("go.mod", []crypto.Hash{crypto.MD5, crypto.SHA1}),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}),
	}

	// Start to sum files.
	man.StartSumFiles(ctx, reqs)

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
				filehashes.PROGRESSUPDATED:
				log.Printf("message: %v", m)
			case filehashes.DONE:
				switch checksums := m.Data.(type) {
				case map[crypto.Hash][]byte:
					for h, checksum := range checksums {
						str := ""
						switch h {
						case crypto.MD5:
							str = "MD5: "
						case crypto.SHA1:
							str = "SHA1: "
						default:
							str = fmt.Sprintf("%d: ", h)
						}

						str += fmt.Sprintf("%X\n", checksum)
						log.Printf(str)
					}
				}

			default:
				log.Printf("unknown message: %v", m)
			}
		}
	}

	// Output:
}
