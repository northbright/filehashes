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

func ExampleStartSumFile() {
	ctx := context.Background()

	// Set buffer size.
	bufferSize := filehashes.DefaultBufferSize

	// Specify the file to hash.
	file := "README.md"

	// Specify the hash algorithm(s).
	hashAlgs := []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
		crypto.SHA256,
	}

	// StartSumFile returns a channel to receive message.
	ch := filehashes.StartSumFile(
		ctx,
		bufferSize,
		file,
		hashAlgs,
	)

	// Consume messages.
	for m := range ch {
		switch m.Type {
		case filehashes.ERROR,
			filehashes.SCHEDULED,
			filehashes.STARTED,
			filehashes.STOPPED,
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

func ExampleStartSumFiles() {
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

	// StartSumFiles returns a channel to receive messages.
	ch := filehashes.StartSumFiles(
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
			filehashes.PROGRESSUPDATED,
			filehashes.DONE:
			buf, _ := m.JSON()
			log.Printf("message: %v", string(buf))
		/*
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
		*/
		default:
			log.Printf("unknown message: %v", m)
		}
	}

	// sum goroutine exited.
	log.Printf("sum goroutine exited")

	// Output:
}
