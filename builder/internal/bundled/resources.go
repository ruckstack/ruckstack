package bundled

import (
	"embed"
	"io/fs"
)

//go:embed init/* init_common/* install_dir/* installer system-control
var embededFiles embed.FS

func OpenFile(path string) (fs.File, error) {
	return embededFiles.Open(path)
}

func ReadFile(path string) ([]byte, error) {
	return embededFiles.ReadFile(path)
}

func OpenDir(path string) (fs.FS, error) {
	return fs.Sub(embededFiles, path)
}

func ReadDir(path string) ([]fs.DirEntry, error) {
	return embededFiles.ReadDir(path)
}
