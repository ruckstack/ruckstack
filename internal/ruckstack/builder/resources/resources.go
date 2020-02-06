package resources

import (
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/resources/bindata"
	"os"
)

//Because the bindata.go file is so large, IDEs may refuse to parse it so this exists as an intermediary to isolate "unknown method" warnings
//Reference this module in the rest of the code, not bindir directly
func AssetDir(name string) ([]string, error) {
	return bindata.AssetDir(name)
}

func AssetInfo(name string) (os.FileInfo, error) {
	return bindata.AssetInfo(name)
}

func Asset(name string) ([]byte, error) {
	return bindata.Asset(name)
}

func AssetNames() []string {
	return bindata.AssetNames()
}
