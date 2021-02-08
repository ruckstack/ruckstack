package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type License struct {
	Name    string
	Email   string
	Expires int64
}

var ActiveLicense *License

var validateKeyText = "LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCk1JSUNJakFOQmdrcWhraUc5dzBCQVFFRkFBT0NBZzhBTUlJQ0NnS0NBZ0VBemhrbG01OFZEdUJlV1hlZEYraHcKQnRFVkV2UzFMSzVzWDIwa0p1MkhlWVFweVpoaEZiVDlxVWVTTDl3MEFyWkNFSHEyQm4yQ01CUkhnNFREVFFKYgpxODRmWXg3ck14Qi9TVWJZWmF6SE1wbDRyVGJLWk1OUkhaK0NPQWM2a2grT25CWW9sQlByNUNxclIvbUtsUzNSCjZXeXYrcFpYVGZ5U0tRZ0RZWnpkMGwzVTRxc2dpaGhIbWVveUFuWXBFVHU1T1Z5RjFaR0cwNXpqczBjNno4YXIKdS83eURDbUVRSS85REtTeHRTZzdOQzE0WkM2YUU2MWtta29jL2M1b0ZuZGxkUEo5SzlHa2lZUVFsQzVjRTdVbAo5bGdYQnpGRDdLeithYUxWQ0xsdVNsWElOQ01vdDlmUE1WME9LRjd2SHFpd2lNc2wzc2swRDFEQlpjTitzTWdoCjJQWU1iVmVjWjhjT0NZdWFaU00yZExUeGVySnFlbzZ4elU2Tmo3WHlTR2VyTHhsZEkrOVAzZCs0TkJINWhzOEMKQUZvQ1ZuMVVxRlIxcWs4QlJKdDZqUXplVURpRGtFQng1bCtHWlVPV1dmR21kTXFLVWJTQVQ5KzhIUDg5VWhKdwpvSEIwTzN2cmdtOGZNd2IwcTdoME1Ca0s0WHFobnl2b3RIQzBnejN5NkEzTS91T1BGOWhzelJCTUliZHNJVWM5Ck9ZS0VmOEFJcE0vcTU5d1RnaFR4NjJmSHhhNndhcGpCWTdlV1pLSFFaT0hla1hWcXhOOUY4c0VpSzVkZHNPL2EKT29DdUVydmZBaFB4bzZkV0c3U0srLzkzK3Bsb1NvUENNM09qcDJQblVqU1ZtQ1R4c1BrYXg5MUxWUFJwVyt4SApmYTBySzhGSVRONy92bzdYSnNRR1dnTUNBd0VBQVE9PQotLS0tLUVORCBSU0EgUFVCTElDIEtFWS0tLS0t"
var validateKey *rsa.PublicKey
var licenseFilePath = filepath.Join(environment.RuckstackHome, "license.txt")

func init() {
	decodedText, err := base64.StdEncoding.DecodeString(validateKeyText)
	if err != nil {
		ui.Fatalf("Cannot decode license validating key: %s", err)
	}

	pemBlock, _ := pem.Decode(decodedText)
	if pemBlock == nil {
		ui.Fatalf("failed to parse license validating key")
		return
	}

	validateKeyInterface, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		ui.Fatalf("failed to parse license validating key: %s", err)
		return
	}

	validateKey = validateKeyInterface.(*rsa.PublicKey)

	_, err = os.Stat(licenseFilePath)
	if err == nil {
		licenseText, err := ioutil.ReadFile(licenseFilePath)
		if err != nil {
			ui.VPrintf("Error reading license.txt: %s", err)
		}

		ActiveLicense, err = parse(string(licenseText))

		if err != nil {
			ui.VPrintf("Error parsing license: %s", err)
		}
	} else {
		if os.IsExist(err) {
			ui.VPrintf("No license.txt found")
		} else {
			ui.VPrintf("Error checking license.txt: %s", err)
		}
	}
}

func SetLicense(licenseText string) error {
	var err error

	ActiveLicense, err = parse(licenseText)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(licenseFilePath, []byte(licenseText), 0600)
	if err != nil {
		return fmt.Errorf("error writing to %s: %s", licenseFilePath, err)
	}

	return ShowLicense()
}

func ShowLicense() error {
	if ActiveLicense == nil {
		ui.Println("No active license. Running free version, no support.")
		ui.Println("For more information on licensing, visit https://ruckstack.com/pro")
	} else {
		ui.Printf("Licensed to %s (%s) until %s", ActiveLicense.Name, ActiveLicense.Email, time.Unix(ActiveLicense.Expires, 0).Format("02 Jan 2006"))
		ui.Println("To contact support, visit https://ruckstack.com/support")
	}

	return nil
}

func RemoveLicense() error {
	err := os.Remove(licenseFilePath)
	if err == nil {
		ui.Println("License was removed")
	} else {
		if os.IsNotExist(err) {
			ui.Println("No license file was found, no license to remove")
		} else {
			return fmt.Errorf("error removing license: %s", err)
		}
	}

	ActiveLicense = nil
	ui.Println()

	return ShowLicense()
}

func parse(licenseText string) (*License, error) {
	if licenseText == "" {
		ui.VPrintf("No license specified")
		return nil, nil
	}

	licenseText = strings.TrimSpace(licenseText)

	decodedLicense, err := base64.StdEncoding.DecodeString(licenseText)
	if err != nil {
		return nil, fmt.Errorf("cannot parse license: %s", err)
	}

	splitLicense := strings.SplitN(string(decodedLicense), ":", 3)
	if len(splitLicense) != 3 {
		return nil, fmt.Errorf("invalid license format")
	}

	version := splitLicense[0]
	ui.VPrintf("License version: %s", version)

	signature, err := hex.DecodeString(splitLicense[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse license version: %s", err)
	}
	contentRaw, err := hex.DecodeString(splitLicense[2])
	if err != nil {
		return nil, fmt.Errorf("cannot parse license content: %s", err)
	}
	ui.VPrintf("License content: %s", contentRaw)

	parsedLicense := &License{}
	if err := yaml.Unmarshal(contentRaw, parsedLicense); err != nil {
		return nil, fmt.Errorf("cannot parse license content: %s", err)
	}

	hashed := sha256.Sum256(contentRaw)

	if err := rsa.VerifyPKCS1v15(validateKey, crypto.SHA256, hashed[:], signature); err != nil {
		ui.VPrintf("invalid license: %s", err)
		return nil, fmt.Errorf("invalid license: verification failed")
	}
	ui.VPrintf("License verified")

	expires := time.Unix(parsedLicense.Expires, 0)
	if expires.Before(time.Now()) {
		ui.VPrintf("License expired %s", expires)
		return nil, fmt.Errorf("invalid license: expired")
	}

	return parsedLicense, nil
}
