package filehashes

import (
	"crypto"
	"encoding/json"
)

// Request represents the request of sum a single file.
type Request struct {
	File     string        `json:"file"`
	HashAlgs []crypto.Hash `json:"hash_algs"`
	Stat     *State        `json:"stat"`
}

// NewRequest returns a new request.
// file is the path of file to be compute hash checksum.
// hashAlgs contains the hash functions for hashing the file.
// stat contains the previous hash states.
// If stat is not nil, it'll restore saved hash states and resume hashing.
// If stat is nil, it'll initialize the hashes and start / restart hashing the file.
func NewRequest(file string, hashAlgs []crypto.Hash, stat *State) *Request {
	return &Request{file, hashAlgs, stat}
}

// JSON marshals a request as a JSON.
func (req *Request) JSON() ([]byte, error) {
	return json.Marshal(req)
}
