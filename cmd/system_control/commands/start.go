package commands

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system_control/server"
	"github.com/ruckstack/ruckstack/internal/system_control/util"
	"github.com/spf13/cobra"
)

func init() {
	packageConfig, err := util.GetPackageConfig()
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
