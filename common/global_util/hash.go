package global_util

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func Sha1File(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	return hex.EncodeToString(hashInBytes), nil
}
