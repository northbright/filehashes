package filehashes

import (
	"crypto"
)

// Request represents the request of sum a single file.
type Request struct {
	File     string        `json:"file"`
	HashAlgs []crypto.Hash `json:"hash_algs"`
}

// NewRequest returns a new request.
// If there's no hash algorithms, use default ones.
func NewRequest(file string, hashAlgs []crypto.Hash) *Request {
	if len(hashAlgs) == 0 {
		hashAlgs = DefaultHashAlgs
	}
	return &Request{file, hashAlgs}
}
