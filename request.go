package filehashes

import (
	"crypto"
	"encoding/json"
)

// Request represents the request of sum a single file.
type Request struct {
	File     string        `json:"file"`
	HashAlgs []crypto.Hash `json:"hash_algs"`
}

// NewRequest returns a new request.
// If there's no hash algorithms, use default ones.
func NewRequest(file string, hashAlgs []crypto.Hash) *Request {
	return &Request{file, hashAlgs}
}

// JSON marshals a request as a JSON.
func (req *Request) JSON() ([]byte, error) {
	return json.Marshal(req)
}
