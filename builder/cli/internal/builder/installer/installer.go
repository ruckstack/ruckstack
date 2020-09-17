package installer

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/cli/internal/builder/global"
	"github.com/ruckstack/ruckstack/builder/cli/internal/project"
	"github.com/ruckstack/ruckstack/builder/cli/internal/resources"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var (
	filesToAddToTar map[string]string
)

type Installer struct {
	filename string

	PackageConfig *config.PackageConfig

	dockerImages []string

	file      *os.File
	zipWriter *zip.Writer
}

func NewInstaller(projectConfig *project.ProjectConfig) (*Installer, error) {

	installer := &Installer{
		filename:     projectConfig.Id + "_" + projectConfig.Version + ".installer",
		dockerImages: []string{},
		PackageConfig: &config.PackageConfig{
			Id:        projectConfig.Id,
			Name:      projectConfig.Name,
			Version:   projectConfig.Version,
			BuildTime: time.Now().Unix(),

			Files: map[string]string{},
			FilePermissions: map[string]config.InstalledFileConfig{
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
	}

	filesToAddToTar = map[string]string{}

	installerPath := path.Join(global.BuildEnvironment.OutDir, installer.filename)
	ui.Printf("Creating %s...", installerPath)

	resourcePath, err := resources.ResourcePath("system/installer")
	if err != nil {
		return nil, err
	}
	installerBytes, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(installerPath, installerBytes, 0755); err != nil {
		return nil, err
	}

	installer.file, err = os.OpenFile(installerPath, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	startOffset, _ := installer.file.Seek(0, io.SeekEnd)

	installer.zipWriter = zip.NewWriter(installer.file)
	installer.zipWriter.SetOffset(startOffset)

	return installer, nil
}

func (installer *Installer) Build() error {
	for key, _ := range filesToAddToTar {
		installer.AddFile(filesToAddToTar[key], key)
	}

	packageConfigFilePath := global.BuildEnvironment.WorkDir + string(filepath.Separator) + "package.config"
	packageConfigFile, err := os.OpenFile(packageConfigFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(packageConfigFile)
	if err := encoder.Encode(installer.PackageConfig); err != nil {
		return err
	}

	if err := installer.AddFile(packageConfigFilePath, ".package.config"); err != nil {
		return err
	}

	if err := installer.zipWriter.Close(); err != nil {
		return err
	}

	if err := installer.file.Close(); err != nil {
		return err
	}

	ui.Printf("Built to %s", installer.file.Name())

	return nil
}

/**
Downloads the given url and saves it to the installer
*/
func (installer *Installer) AddDownloadedFile(url string, targetPath string) error {
	savedLocation, err := util.DownloadFile(url)
	if err != nil {
		return nil
	}

	return installer.AddFile(savedLocation, targetPath)
}

/**
Downloads the given url and extracts a specific file out of the archive and saves it to the installer
*/
func (installer *Installer) AddDownloadedNestedFile(url string, wantedFile string, targetPath string) error {
	fileLocation, err := util.DownloadFile(url)
	if err != nil {
		return err
	}

	extractedFile, err := util.ExtractFromGzip(fileLocation, wantedFile)
	if err != nil {
		return err
	}

	return installer.AddFile(extractedFile, targetPath)

}

/**
Adds the given file to the installer
*/
func (installer *Installer) AddFile(srcPath string, targetPath string) error {
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

	return installer.addData(file, targetPath, uint64(fileInfo.Size()), fileInfo.ModTime(), hashString)
}

func (installer *Installer) addData(data io.Reader, installerPath string, size uint64, modTime time.Time, dataHash string) error {
	installerPath = strings.Replace(installerPath, "./", "", 1)

	header := &zip.FileHeader{
		Name:               installerPath,
		UncompressedSize64: size,
		Modified:           modTime,
		//Mode:    installer.installerFileMode(assetPath),
	}

	entryWriter, err := installer.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not write header for file %s: %s", installer.filename, err)
	}

	written, err := io.Copy(entryWriter, data)
	if err != nil {
		return fmt.Errorf("could not write to %s: %s", installer.filename, err)
	}

	if uint64(written) != size {
		return fmt.Errorf("expected %s to be %d bytes but was %d", installerPath, size, written)
	}

	installer.PackageConfig.Files[installerPath] = dataHash

	return nil
}

func (installer *Installer) installerFileMode(targetPath string) int64 {
	if strings.HasPrefix(targetPath, "lib") || strings.HasPrefix(targetPath, "bin") {
		return int64(0755)
	}
	return int64(0644)
}

func (installer *Installer) AddAsset(assetPath string, targetPath string) error {
	resourcePath, err := resources.ResourcePath(assetPath)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(resourcePath)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		return err
	}

	hash := sha1.New()
	hash.Write(data)
	hashBytes := hash.Sum(nil)[:20]
	hashString := hex.EncodeToString(hashBytes)

	return installer.addData(bytes.NewReader(data), targetPath, uint64(fileInfo.Size()), fileInfo.ModTime(), hashString)
}

func (installer *Installer) AddResourceDir(assetBase string, targetBase string) error {
	//TODO: restore with new resources logic
	//files, err := resources.AssetDir(assetBase)
	//if err != nil {
	//	return err
	//}
	//
	//for _, assetPath := range files {
	//	_, err := resources.AssetInfo(assetBase + "/" + assetPath)
	//	if err != nil {
	//		//probably a directory
	//		//TODO: check error to be sure
	//		if assertErr := installer.AddResourceDir(assetBase+"/"+assetPath, targetBase+"/"+assetPath); assertErr != nil {
	//			return assertErr
	//		}
	//	} else {
	//		if assetErr := installer.AddAsset(assetBase+"/"+assetPath, targetBase+"/"+assetPath); assetErr != nil {
	//			return assetErr
	//		}
	//	}
	//}

	return nil
}

func (installer *Installer) IncludeDockerImage(tag string) error {
	installer.dockerImages = append(installer.dockerImages, tag)

	dockerSaveCmd := exec.Command("docker", "image", "inspect", tag)
	dockerSaveCmd.Stdout = ui.GetOutput()
	dockerSaveCmd.Stderr = ui.GetOutput()
	err := dockerSaveCmd.Run()
	if err == nil {
		ui.Printf("Already have image %s", tag)
		return nil
	} else {
		dockerSaveCmd := exec.Command("docker", "pull", tag)

		dockerSaveCmd.Stdout = os.Stdout
		dockerSaveCmd.Stderr = os.Stderr
		return dockerSaveCmd.Run()
	}
}

func (installer *Installer) SaveDockerImages() error {
	ui.Printf("Collecting containers...")
	appImagePath := global.BuildEnvironment.WorkDir + string(filepath.Separator) + "images.tar"
	dockerSaveCmd := exec.Command("docker", append([]string{"save", "--output", appImagePath}, installer.dockerImages...)...)
	dockerSaveCmd.Stdout = os.Stdout
	dockerSaveCmd.Stderr = os.Stderr
	if err := dockerSaveCmd.Run(); err != nil {
		return err
	}
	return installer.AddFile(appImagePath, "data/agent/images/images.tar")
}

func (installer *Installer) ClearDockerImages() error {
	ui.Printf("Cleaning up containers...")

	output, err := exec.Command("docker", "image", "ls",
		"--format", "'{{.Repository}}:{{.Tag}}'",
		"--filter", "label=ruckstack.built=true").Output()
	if err != nil {
		return err
	}

	for _, builtTag := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		builtTag = strings.Trim(builtTag, "'")
		ui.Printf("Removing %s", builtTag)
		output, err := exec.Command("docker", "image", "rm", builtTag).Output()
		ui.Println(string(output))

		if err != nil {
			return err
		}

	}

	return nil

}
