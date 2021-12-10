package authenticator

import (
	crand "crypto/rand"
	"encoding/base32"
	"math/big"
)

func generateSecret() (string, error) {
	randStr, err := getRandomString(8, "alphanum")
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString([]byte(randStr)), nil
}

func generateScratchCodes(num int, alphabet string) (scratchCodes []string, err error) {
	scratchCodes = make([]string, num)
	for i := 0; i < num; i++ {
		scratchCodes[i], err = getRandomString(8, alphabet)
		if err != nil {
			return nil, err
		}
	}
	return scratchCodes, err
}

func getRandomString(length int, alphabet string) (string, error) {
	switch alphabet {
	case "alpha":
		alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	case "num":
		alphabet = "0123456789"
	case "alphanum":
		alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	}

	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := crand.Int(crand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		ret[i] = alphabet[num.Int64()]
	}

	return string(ret), nil
}
