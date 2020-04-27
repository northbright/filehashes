package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	"fmt"
	"log"

	"github.com/northbright/filehashes"
)

func ExampleSumFile() {
	ctx := context.Background()

	// Set buffer size.
	bufferSize := filehashes.DefaultBufferSize

	// Specify the file to hash.
	file := "README.md"

	// Specify the hash algorithm(s).
	hashAlgs := []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
	}

	// SumFile returns a channel to receive message.
	ch := filehashes.SumFile(
		ctx,
		bufferSize,
		file,
		hashAlgs,
	)

	// Consume messages.
	for m := range ch {
		switch msg := m.(type) {
		case
			filehashes.SumErrorMsg,
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
	// sum goroutine exited.
	log.Printf("sum goroutine exited")

	// Output:
}

func ExampleSumFiles() {
	ctx := context.Background()
	// Set sum concurrency, defalut value's filehashes.DefaultConcurrency.
	concurrency := 2
	// Set buffer size.
	bufferSize := filehashes.DefaultBufferSize

	// reqs contains filehashes.Request,
	// which specify the file to hash and the hash algorithm(s).
	reqs := []*filehashes.Request{
		filehashes.NewRequest("filehashes.go", []crypto.Hash{crypto.MD5}),
		filehashes.NewRequest("README.md", []crypto.Hash{crypto.SHA1}),
		filehashes.NewRequest("go.mod", []crypto.Hash{crypto.MD5, crypto.SHA1}),
		filehashes.NewRequest("go.sum", []crypto.Hash{crypto.MD5, crypto.SHA1}),
	}

	// SumFiles returns a channel to receive messages.
	ch := filehashes.SumFiles(
		ctx,
		concurrency,
		bufferSize,
		reqs,
	)

	// Consume messages.
	for m := range ch {
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
	// sum goroutine exited.
	log.Printf("sum goroutine exited")

	// Output:
}
