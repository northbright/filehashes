package filehashes

import (
	"crypto"
	"crypto/md5"
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

// id returns the request's ID.
// It uses md5 hash of file name and hash algorithms as the unique request ID.
func (req *Request) id() string {
	h := md5.New()
	h.Write([]byte(req.File))
	for _, alg := range req.HashAlgs {
		h.Write([]byte(fmt.Sprintf("%d", alg)))
	}
	return fmt.Sprintf("%X", h.Sum(nil))
}
