package server

import (
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
)

func Start() error {
	if config.LocalConfig == nil {
		ui.Println("Configure server!")
	}
	ui.Println("Starting server!")

	//createCmd := cluster.NewCmdClusterDelete()
	//createCmd.SetArgs([]string{"start"})
	////ctx := context.WithValue(context.Background(), "value", nil)
	////
	//err := createCmd.Execute()
	//if err != nil {
	//	ui.Fatalf("error!", err)
	//}

	//k3dCmd.NewCmdK3d().
	ui.Println("Started!")
	return nil
}
