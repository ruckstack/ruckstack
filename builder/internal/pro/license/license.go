package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/settings"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v3"
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

	if settings.Settings.LicenseKey != "" {
		ActiveLicense, err = parse(settings.Settings.LicenseKey)
		if err != nil {
			ui.VPrintf("Error parsing license: %s", err)
		}
	}

	_ = ShowLicense()
	fmt.Println()

}

func SetLicense(licenseText string) error {
	var err error

	licenseText = strings.TrimSpace(licenseText)

	ActiveLicense, err = parse(licenseText)
	if err != nil {
		return err
	}

	settings.Settings.LicenseKey = licenseText
	if err = settings.Settings.Save(); err != nil {
		return fmt.Errorf("error saving license: %s", err)
	}

	return ShowLicense()
}

func ShowLicense() error {
	if ActiveLicense == nil {
		ui.Println("No active Ruckstack Pro license. Running free version. For more information, visit https://ruckstack.com/pro")
	} else {
		ui.Printf("Licensed to %s (%s) until %s", ActiveLicense.Name, ActiveLicense.Email, time.Unix(ActiveLicense.Expires, 0).Format("02 Jan 2006"))
		ui.Println("To contact support, visit https://ruckstack.com/support")
	}

	return nil
}

func RemoveLicense() error {
	if settings.Settings.LicenseKey == "" {
		ui.Println("No license file was found, no license to remove")
	} else {
		settings.Settings.LicenseKey = ""
		if err := settings.Settings.Save(); err != nil {
			return fmt.Errorf("error saving license: %s", err)
		}

		ui.Println("License was removed")
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

	version := licenseText[0:2]
	licenseText = licenseText[2:]

	splitLicense := strings.SplitN(licenseText, "-", 2)
	if len(splitLicense) != 2 {
		ui.VPrintf("invalid license format")
		return nil, fmt.Errorf("invalid license: data corrupted")
	}

	ui.VPrintf("License version: %s", version)

	signature, err := base64.StdEncoding.DecodeString(splitLicense[0])
	if err != nil {
		ui.VPrintf("cannot parse license validation: %s", err)
		return nil, fmt.Errorf("invalid license: data corrupted")
	}
	contentRaw, err := base64.StdEncoding.DecodeString(splitLicense[1])
	if err != nil {
		ui.VPrintf("cannot parse license content: %s", err)
		return nil, fmt.Errorf("invalid license: data corrupted")
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
