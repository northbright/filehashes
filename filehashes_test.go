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

func ExampleSum() {
	ctx := context.Background()

	hashAlgs := []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
	}

	files := []string{
		"README.md",
		"filehashes.go",
		"filehashes_test.go",
		"msg.go",
		// Hashing 0-byte files is allowed.
		"go.sum",
		// File does not exist and you'll get error message later.
		"FILE_NOT_EXIST",
	}

	ch := filehashes.Sum(
		ctx,
		filehashes.DefaultConcurrency,
		filehashes.DefaultBufferSize,
		hashAlgs,
		files,
	)

	// Consume messages.
	for m := range ch {
		switch msg := m.(type) {
		case
			filehashes.NoFileError,
			filehashes.SumError,
			filehashes.SumStarted,
			filehashes.SumStopped,
			filehashes.SumProgress:
			// All types of messages have String().
			log.Printf("%v", msg)
		case filehashes.SumDone:
			// Sum single file done.
			log.Printf("sum %s done\n", msg.File)

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
		case filehashes.SumAllDone:
			// All done.
			log.Printf("sum all files done:\n")
			for _, file := range msg.Files {
				log.Printf("%s\n", file)
			}

		default:
			log.Printf("unknown message: %v", msg)
		}
	}

	// Output:
}
