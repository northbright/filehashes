package filehashes

import (
	"crypto"
	"encoding"
	"hash"
)

// State represents the hashes states.
type State struct {
	// SummedSize is summed size of file.
	SummedSize int64 `json:"summed_size,string"`
	// Progress stores the last progress.
	Progress int `json:"progress"`
	// Data contains binary datas marshaled from hashes.
	Datas map[crypto.Hash][]byte `json:"datas"`
}

// newState returns the current hash state.
func newState(summedSize int64, progress int, hashes map[crypto.Hash]hash.Hash) *State {
	datas := map[crypto.Hash][]byte{}

	for k, v := range hashes {
		m := v.(encoding.BinaryMarshaler)
		data, _ := m.MarshalBinary()
		datas[k] = data
	}

	state := &State{SummedSize: summedSize, Progress: progress, Datas: datas}
	return state
}
