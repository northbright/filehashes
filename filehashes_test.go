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

	ch, err := filehashes.Sum(
		ctx,
		filehashes.DefaultBufferSize,
		"filehashes.go",
		crypto.MD5,
		crypto.SHA1,
	)

	if err != nil {
		log.Printf("Sum error: %v", err)
		return
	}

	for m := range ch {
		switch msg := m.(type) {
		case
			filehashes.TaskError,
			filehashes.TaskStarted,
			filehashes.TaskStopped,
			filehashes.TaskProgress:
			// All types of messages have String(),
			// use default output if you want.
			log.Printf("%v", msg)
		case filehashes.TaskDone:
			// Done.
			// Use customized output instead of default one.
			log.Printf("sum done\n")

			// Get hash algorithms and checksums from the message.
			// msg is value of TaskDone(map[crypto.Hash][]byte).
			for h, checksum := range msg {
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
