package webserver

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"io"
	"os"
)

func ImportKeys(privateKeyFile string, certificateFile string) error {
	if err := copyFileContents(privateKeyFile, sslKeyFilePath); err != nil {
		return err
	}

	if err := copyFileContents(certificateFile, sslCertFilePath); err != nil {
		return err
	}

	ui.Println("Successfully imported key and certificate files.")
	ui.Println()
	ui.Println("Server must be restarted for new certificates to be used.")

	return nil
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot copy %s: %s", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot copy to %s: %s", dst, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("error copying %s: %s", src, err)
	}
	_ = out.Sync()

	return nil
}
