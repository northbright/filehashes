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

func ExampleManager_StartSumFile() {
	ctx := context.Background()

	concurrency := filehashes.DefaultConcurrency
	bufferSize := filehashes.DefaultBufferSize

	man, ch := filehashes.NewManager(concurrency, bufferSize)

	// Create requests.
	reqs := []*filehashes.Request{
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.MD5}),
		filehashes.NewRequest("filehashes.go", []crypto.Hash{crypto.SHA1}),
		filehashes.NewRequest("go.mod", []crypto.Hash{crypto.MD5, crypto.SHA1}),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}),
	}

	for _, req := range reqs {
		go func(req *filehashes.Request) {
			// Start sum file.
			man.StartSumFile(ctx, req)
		}(req)
	}

	// Create a timeout to exit.
	chTimeout := time.After(5 * time.Second)

	// Consume messages.
	for {
		select {
		case <-chTimeout:
			log.Printf("timeout, exit")
			return
		case m := <-ch:
			switch msg := m.(type) {
			case
				filehashes.SumErrorMsg,
				filehashes.SumScheduledMsg,
				filehashes.SumStartedMsg,
				filehashes.SumStoppedMsg,
				filehashes.SumProgressMsg:
				// All types of messages have String().
				log.Printf("%v", msg)
			case filehashes.SumDoneMsg:
				// Sum single file done.
				log.Printf("sum %v done\n", msg.Request)

				// Get hash algorithms and checksums from the message.
				// msg is value of SumDone(map[crypto.Hash][]byte).
				for h, checksum := range msg.Checksums {
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

			default:
				log.Printf("unknown message: %v", msg)
			}
		}
	}

	// Output:
}
