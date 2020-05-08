package filehashes

import (
	"crypto"
	"encoding/json"
)

// Request represents the request of sum a single file.
type Request struct {
	File      string        `json:"file"`
	HashFuncs []crypto.Hash `json:"hash_funcs"`
	Stat      *State        `json:"stat"`
}

// NewRequest returns a new request.
// file is the path of file to be compute hash checksum.
// hashFuncs contains the hash functions.
// stat contains the previous hash states.
//
// If stat is nil, it starts to read the beginning of the file and compute checksums.
// otherwise, it'll restore saved hash states and resume reading file from the given offset.
func NewRequest(file string, hashFuncs []crypto.Hash, stat *State) *Request {
	return &Request{file, hashFuncs, stat}
}

// JSON marshals a request as a JSON.
func (req *Request) JSON() ([]byte, error) {
	return json.Marshal(req)
}
