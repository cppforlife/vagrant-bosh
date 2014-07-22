package util

import (
	"encoding/base64"

	bosherr "bosh/errors"
)

func DecodePossibleBase64Str(str string) (string, error) {
	// Cheating until yaml library provides proper support for !!binary
	if str[len(str)-1:] == "=" {
		bytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return "", bosherr.WrapError(err, "Decoding base64 encoded str '%s'", str)
		}

		return string(bytes), nil
	}

	return str, nil
}
