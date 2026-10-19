package helper

import (
	"encoding/hex"
	"encoding/json"

	"github.com/vault-thirteen/auxie/errors"
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
			return nil, errors.ErrorFromText(Err_InvalidHashSum)
		}

		var sha256Bytes []byte
		sha256Bytes, err = hex.DecodeString(s)
		if err != nil {
			return nil, errors.ErrorFromText(Err_InvalidHashSum)
		}

		if len(sha256Bytes) != 32 {
			return nil, errors.ErrorFromText(Err_InvalidHashSum)
		}

		hashes = append(hashes, HashSum(sha256Bytes))
	}

	return hashes, nil
}
