package helper

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

const (
	Err_InvalidHashSum = "invalid hash sum"
)

type HashSum []byte

func ParseFileHashes(text string) (hashes []HashSum, err error) {
	var strings []string
	err = json.Unmarshal([]byte(text), &strings)
	if err != nil {
		return nil, err
	}

	hashes = make([]HashSum, 0, len(strings))

	for _, s := range strings {
		if len(s) != 64 {
			return nil, errors.New(Err_InvalidHashSum)
		}

		var sha256Bytes []byte
		sha256Bytes, err = hex.DecodeString(s)
		if err != nil {
			return nil, errors.New(Err_InvalidHashSum)
		}

		if len(sha256Bytes) != 32 {
			return nil, errors.New(Err_InvalidHashSum)
		}

		hashes = append(hashes, HashSum(sha256Bytes))
	}

	return hashes, nil
}
