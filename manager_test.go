package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	"log"
	"math/rand"
	"time"

	"github.com/northbright/filehashes"
)

func ExampleManager_Start() {
	// You may set the number of concurrent worker goroutines.
	// It's filehashes.DefaultConcurrency by default.
	concurrency := 2
	bufferSize := filehashes.DefaultBufferSize

	man, ch := filehashes.NewManager(concurrency, bufferSize)

	// Create requests.
	reqs := []*filehashes.Request{
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.MD5}, nil),
		filehashes.NewRequest("manager.go", []crypto.Hash{crypto.SHA1}, nil),
		filehashes.NewRequest("manager_test.go", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}, nil),
	}

	// Use a timeout to exit the program.
	tm := time.After(5 * time.Second)

	// Record the cancel func for each request's file.
	cancelMap := map[string]context.CancelFunc{}

	for _, req := range reqs {
		ctx, cancel := context.WithCancel(context.Background())
		cancelMap[req.File] = cancel
		// Start to sum files.
		man.Start(ctx, req)
	}

	// Stop 1 task randomly.
	// The stopped task will be resume with previous progress in message loop.
	go func() {
		time.Sleep(1 * time.Millisecond)
		i := rand.Intn(len(reqs))
		cancel := cancelMap[reqs[i].File]
		cancel()
	}()

	// Consume messages.
	for {
		select {
		case <-tm:
			log.Printf("timeout")
			return
		case m := <-ch:
			switch m.Type {
			case filehashes.ERROR,
				filehashes.SCHEDULED,
				filehashes.STARTED,
				filehashes.RESTORED,
				filehashes.PROGRESS_UPDATED,
				filehashes.DONE,
				filehashes.EXITED:
				// All messages can be marshaled to JSON.
				buf, _ := m.JSON()
				log.Printf("message: %v", string(buf))

			case filehashes.STOPPED:
				buf, _ := m.JSON()
				log.Printf("message: %v", string(buf))

				// Resume the task if it's stopped.
				resumeReq := m.Data.(*filehashes.Request)
				ctx, cancel := context.WithCancel(context.Background())
				cancelMap[resumeReq.File] = cancel

				man.Start(ctx, resumeReq)

			default:
				log.Printf("unknown message: %v", m)
			}
		}
	}

	// Output:
}
