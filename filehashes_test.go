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

	ch := filehashes.SumFile(
		ctx,
		filehashes.DefaultBufferSize,
		"README.md",
		hashAlgs,
	)

	// Consume messages.
	for m := range ch {
		switch msg := m.(type) {
		case
			filehashes.ErrorMsg,
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

	// Output:
}
