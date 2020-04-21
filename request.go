package filehashes

import (
	"crypto"
	"fmt"
)

// Request represents the request of sum a single file.
type Request struct {
	File     string        `json:"file"`
	HashAlgs []crypto.Hash `json:"hash_algs"`
}

// String returns a formated string for request.
func (req *Request) String() string {
	str := fmt.Sprintf("file: %s(hash algs:", req.File)
	for _, h := range req.HashAlgs {
		str += fmt.Sprintf(" %v", h)
	}
	str += ")"
	return str
}
