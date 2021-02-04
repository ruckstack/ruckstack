package global_util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func HashFile(filePath string) (string, error) {
	chartFileObj, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot open %s for hashing: %s", filePath, err)
	}

	hash := sha1.New()
	_, err = io.Copy(hash, chartFileObj)
	if err != nil {
		return "", fmt.Errorf("cannot compute hash for %s: %s", filePath, err)
	}

	return hex.EncodeToString(hash.Sum(nil)[:20]), nil
}
