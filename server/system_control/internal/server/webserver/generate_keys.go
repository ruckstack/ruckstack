package webserver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func GenerateKeys(outputDir string) error {

	ui.Printf("Generating SSL keys to %s", outputDir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output directory %s: %s", outputDir, err)
	}
	certificateFile := filepath.Join(outputDir, environment.PackageConfig.Id+".crt")
	privateKeyFile := filepath.Join(outputDir, environment.PackageConfig.Id+"-private.key")
	csrFile := filepath.Join(outputDir, environment.PackageConfig.Id+"-request.csr")

	hostname := ui.PromptForString("Hostname", "",
		func(input string) error {
			if input == "" {
				return fmt.Errorf("hostname is required")
			}
			return nil
		},
	)

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("Cannot generate RSA key: %s", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: hostname,
		},
		NotBefore: time.Now().Add(-1 * time.Hour * 48),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 10),

		BasicConstraintsValid: true,
	}

	certificate, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %s", err)
	}

	certFileObj, err := os.Create(certificateFile)
	if err != nil {
		return fmt.Errorf("cannot write %s: %s", certificateFile, err)
	}
	defer certFileObj.Close()

	if err := pem.Encode(certFileObj, &pem.Block{Type: "CERTIFICATE", Bytes: certificate}); err != nil {
		return err
	}

	privateKeyFileObj, err := os.OpenFile(privateKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot write %s: %s", privateKeyFile, err)
	}
	defer privateKeyFileObj.Close()

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(privateKeyFileObj, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes}); err != nil {
		return err
	}

	csrOut, err := os.OpenFile(csrFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("cannot write %s: %s", csrFile, err)
	}
	defer csrOut.Close()

	requestTemplate := x509.CertificateRequest{
		Subject:            certTemplate.Subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, &requestTemplate, privateKey)
	if err := pem.Encode(csrOut, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		return err
	}

	ui.Println("Created the following files:")
	ui.Println("   Private key: " + privateKeyFile)
	ui.Println("   Self-signed certificate: " + certificateFile)
	ui.Println("   Certificate Signing request: " + csrFile)
	ui.Println()
	ui.Println("Next steps:")
	ui.Println("  1. Use the csr file with your certificate authority to create a valid certificate")
	ui.Printf("  2. Import the private key and signed certificate using `%s https import`\n", environment.PackageConfig.ManagerFilename)

	return nil
}
