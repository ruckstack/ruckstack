package commands

import (
	"fmt"
	common2 "github.com/ruckstack/ruckstack/server/internal/environment"
	"github.com/ruckstack/ruckstack/server/system_control/internal/server"
	"github.com/spf13/cobra"
)

func init() {
	packageConfig, err := common2.GetPackageConfig()
	if err != nil {
		fmt.Printf("error loading package config: %s", err)
		return
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts " + packageConfig.Name,
		Annotations: map[string]string{
			REQUIRES_ROOT: "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Start()
		},
	})
}
