package install_file

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InstallFile struct {
	PackageConfig *config.PackageConfig

	dockerImages []string

	file      *os.File
	zipWriter *zip.Writer

	filesToAddToTar map[string]string
}

/**
Begins creating the install file.

The install file is the installer binary followed by a zip of what to install.
This method will create the file with the installer and open a zip write for the remaining contents.
*/
func StartInstallFile(projectConfig *project.ProjectConfig) (*InstallFile, error) {

	installFile := &InstallFile{
		PackageConfig: &config.PackageConfig{
			Id:        projectConfig.Id,
			Name:      projectConfig.Name,
			Version:   projectConfig.Version,
			BuildTime: time.Now().Unix(),

			Files: map[string]string{},
			FilePermissions: map[string]config.PackagedFileConfig{
				".package.config": {
					AdminGroupReadable: true,
				},
				"config/*": {
					AdminGroupReadable: true,
				},
				"bin/*": {
					AdminGroupReadable: true,
					Executable:         true,
				},
				"lib/*": {
					AdminGroupReadable: true,
				},
				"lib/k3s": {
					AdminGroupReadable: true,
					Executable:         true,
				},
				"lib/helm": {
					AdminGroupReadable: true,
					Executable:         true,
				},
			},
		},
		dockerImages:    []string{},
		filesToAddToTar: map[string]string{},
	}

	installerPath := environment.OutPath(projectConfig.Id + "_" + projectConfig.Version + ".installer")
	ui.Printf("Building %s...", filepath.Base(installerPath))

	installerBinaryPath, err := environment.ResourcePath("installer")
	if err != nil {
		return nil, err
	}
	installerBytes, err := ioutil.ReadFile(installerBinaryPath)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(installerPath, installerBytes, 0755); err != nil {
		return nil, err
	}

	installFile.file, err = os.OpenFile(installerPath, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	startOffset, _ := installFile.file.Seek(0, io.SeekEnd)

	installFile.zipWriter = zip.NewWriter(installFile.file)
	installFile.zipWriter.SetOffset(startOffset)

	return installFile, nil
}

func (installFile *InstallFile) Build() error {
	for key, _ := range installFile.filesToAddToTar {
		if err := installFile.AddFile(installFile.filesToAddToTar[key], key); err != nil {
			return err
		}

	}

	if err := installFile.saveDockerImages(); err != nil {
		return err
	}

	if err := installFile.ClearDockerImages(); err != nil {
		return err
	}

	packageConfigFilePath := environment.TempPath("package.config")
	packageConfigFile, err := os.OpenFile(packageConfigFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(packageConfigFile)
	if err := encoder.Encode(installFile.PackageConfig); err != nil {
		return err
	}

	if err := installFile.AddFile(packageConfigFilePath, ".package.config"); err != nil {
		return err
	}

	if err := installFile.zipWriter.Close(); err != nil {
		return err
	}

	if err := installFile.file.Close(); err != nil {
		return err
	}

	ui.Printf("Building %s...DONE", filepath.Base(installFile.file.Name()))

	return nil
}

/**
Downloads the given url and saves it to the installer
*/
func (installFile *InstallFile) AddDownloadedFile(url string, targetPath string) error {
	savedLocation, err := util.DownloadFile(url)
	if err != nil {
		return nil
	}

	return installFile.AddFile(savedLocation, targetPath)
}

/**
Downloads the given url and extracts a specific file out of the archive and saves it to the installer
*/
func (installFile *InstallFile) AddDownloadedNestedFile(url string, wantedFile string, targetPath string) error {
	fileLocation, err := util.DownloadFile(url)
	if err != nil {
		return err
	}

	extractedFile, err := util.ExtractFromGzip(fileLocation, wantedFile)
	if err != nil {
		return err
	}

	return installFile.AddFile(extractedFile, targetPath)

}

/**
Adds the given file to the installer
*/
func (installFile *InstallFile) AddFile(srcPath string, targetPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %s", srcPath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat %s: %s", srcPath, err)
	}

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return fmt.Errorf("cannot compute hash for %s: %s", srcPath, err)
	}

	hashBytes := hash.Sum(nil)[:20]
	hashString := hex.EncodeToString(hashBytes)
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("cannot reset %s: %s", srcPath, err)
	}

	return installFile.addData(file, targetPath, uint64(fileInfo.Size()), fileInfo.ModTime(), hashString)
}

func (installFile *InstallFile) addData(data io.Reader, installerPath string, size uint64, modTime time.Time, dataHash string) error {
	installerPath = strings.Replace(installerPath, "./", "", 1)

	header := &zip.FileHeader{
		Name:               installerPath,
		UncompressedSize64: size,
		Modified:           modTime,
		//Mode:    installFile.installerFileMode(assetPath),
	}

	entryWriter, err := installFile.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not write header for file %s: %s", installFile.file.Name(), err)
	}

	written, err := io.Copy(entryWriter, data)
	if err != nil {
		return fmt.Errorf("could not write to %s: %s", installFile.file.Name(), err)
	}

	if uint64(written) != size {
		return fmt.Errorf("expected %s to be %d bytes but was %d", installerPath, size, written)
	}

	installFile.PackageConfig.Files[installerPath] = dataHash

	return nil
}

func (installFile *InstallFile) installerFileMode(targetPath string) int64 {
	if strings.HasPrefix(targetPath, "lib") || strings.HasPrefix(targetPath, "bin") {
		return int64(0755)
	}
	return int64(0644)
}

func (installFile *InstallFile) AddDirectory(assetBase string, targetBase string) error {
	err := filepath.Walk(assetBase, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return installFile.AddFile(path, filepath.Join(targetBase, path[len(assetBase):]))
		}
		return nil
	})
	return err
}

func (installFile *InstallFile) AddDockerImage(tag string) error {
	installFile.dockerImages = append(installFile.dockerImages, tag)

	return nil
}

func (installFile *InstallFile) saveDockerImages() error {

	for _, tag := range installFile.dockerImages {
		err := docker.ImagePull(tag)
		if err != nil {
			return fmt.Errorf("error pulling %s", tag)
		}
	}

	ui.Printf("Collecting containers...")
	imagesTarPath := environment.TempPath("images.tar")
	if err := docker.SaveImages(imagesTarPath, installFile.dockerImages...); err != nil {
		return fmt.Errorf("error collecting containers: %s", err)
	}

	if err := installFile.AddFile(imagesTarPath, "data/agent/images/images.tar"); err != nil {
		return err
	}

	return nil
}

func (installFile *InstallFile) ClearDockerImages() error {
	ui.Printf("Cleaning up containers...")

	listFilters := filters.NewArgs()
	listFilters.Add("label", "ruckstack.built=true")

	images, err := docker.ImageList(types.ImageListOptions{
		Filters: listFilters,
	})
	if err != nil {
		return err
	}
	for _, image := range images {
		ui.Printf("Removing %s", image.ID)
		if err := docker.ImageRemove(image.ID); err != nil {
			return err
		}
	}

	return nil

}
