package commands

import (
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server/webserver"
	"github.com/spf13/cobra"
)

func init() {
	var httpsCmd = &cobra.Command{
		Use:   "https",
		Short: "Configures https",
	}

	initHttpsImport(httpsCmd)
	initHttpsGenerate(httpsCmd)

	rootCmd.AddCommand(httpsCmd)

}

func initHttpsImport(parent *cobra.Command) {

	var privateKeyFile string
	var certificateFile string

	var cmd = &cobra.Command{
		Use: "import",
		Annotations: map[string]string{
			RequiresRoot: "true",
		},
		Short: "Loads private keys and certificates to use with HTTPS",
		Long: `Commands to create and sign your keys will depend on your IT infrastructure.
A self-signed key can be created with the "https generate-key" command.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return webserver.ImportKeys(privateKeyFile, certificateFile)
		},
	}

	cmd.Flags().StringVar(&privateKeyFile, "private-key", "", "File containing the private key.")
	cmd.Flags().StringVar(&certificateFile, "certificate", "", "File containing the certificate.")

	ui.MarkFlagsRequired(cmd, "private-key", "certificate")
	ui.MarkFlagsFilename(cmd, "private-key", "certificate")

	parent.AddCommand(cmd)
}

func initHttpsGenerate(parent *cobra.Command) {

	var outputDir string

	var cmd = &cobra.Command{
		Use:   "generate-key",
		Short: "Generates a self-signed certificate to use with HTTPS",
		RunE: func(cmd *cobra.Command, args []string) error {
			return webserver.GenerateKeys(outputDir)
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", ".", "Directory to create keys and certificates in")

	ui.MarkFlagsDirname(cmd, "output")

	parent.AddCommand(cmd)
}
