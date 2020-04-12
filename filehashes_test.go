package filehashes_test

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
	"fmt"

	"github.com/northbright/filehashes"
)

func ExampleSum() {
	ctx := context.Background()

	ch, err := filehashes.Sum(ctx, "filehashes.go", crypto.MD5, crypto.SHA1)
	if err != nil {
		fmt.Printf("Sum error: %v", err)
		return
	}

	for m := range ch {
		switch msg := m.(type) {
		case *filehashes.SumErrorMsg,
			*filehashes.SumDoneMsg,
			*filehashes.SumStartedMsg,
			*filehashes.SumStoppedMsg,
			*filehashes.SumProgressUpdatedMsg:
			// All types of messages have String()
			fmt.Printf("%v", msg)
		default:
			fmt.Printf("unknown message: %v", msg)
		}
	}

	// Output:
	//sum filehashes.go started
	//sum filehashes.go progress: 100
	//sum filehashes.go done:
	//--------------------
	//2: 7AA3A1024812C042D13BE78E50689265
	//3: 23E1352125DACEA548FB2266E3700F22C7F68C7C
}
