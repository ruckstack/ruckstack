package install_file

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/ruckstack/ruckstack/builder/internal/bundled"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"github.com/ruckstack/ruckstack/builder/internal/util"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/global_util"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"io"
	"io/fs"
	"io/ioutil"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type InstallFile struct {
	PackageConfig    *config.PackageConfig
	SystemConfig     *config.SystemConfig
	CompressionLevel int

	dockerImages map[string]bool
	addedFiles   map[string]bool

	file      *os.File
	zipWriter *zip.Writer
}

/**
Begins creating the install file.

The install file is the installer binary followed by a zip of what to install.
This method will create the file with the installer and open a zip write for the remaining contents.
*/
func StartCreation(installerPath string, compressionLevel int) (*InstallFile, error) {

	installFile := &InstallFile{
		CompressionLevel: compressionLevel,
		PackageConfig: &config.PackageConfig{
			BuildTime: time.Now().Unix(),

			Files: map[string]string{},
			FilePermissions: map[string]config.PackagedFileConfig{
				".package.config": {
					AdminGroupReadable: true,
				},
				"config/*": {
					AdminGroupReadable: true,
				},
				"logs/*": {
					AdminGroupReadable: true,
				},
				"data/kubectl": {
					AdminGroupReadable: true,
					AdminGroupWritable: true,
				},
				"data/web": {
					AdminGroupReadable: true,
					AdminGroupWritable: false,
				},
				"data/**": {
					PreservePermissions: true,
				},
				"tmp/*": {
					AdminGroupReadable: true,
					AdminGroupWritable: true,
				},
				"bin/*": {
					AdminGroupReadable: true,
					Executable:         true,
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
		SystemConfig: &config.SystemConfig{},
		dockerImages: map[string]bool{},
		addedFiles:   map[string]bool{},
	}

	ui.Printf("Building %s...", filepath.Base(installerPath))

	installerBytes, err := bundled.ReadFile("installer")
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(installerPath), 0755)
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
	installFile.zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, installFile.CompressionLevel)
	})
	installFile.zipWriter.SetOffset(startOffset)

	return installFile, nil
}

func (installFile *InstallFile) CompleteCreation() error {
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

	packageConfigFile.Close()
	packageConfigFile, err = os.Open(packageConfigFilePath)
	if err != nil {
		return err
	}

	if err := installFile.AddFile(packageConfigFile, ".package.config"); err != nil {
		return err
	}

	systemConfigFilePath := environment.TempPath("config/system.config")
	_ = os.MkdirAll(filepath.Dir(systemConfigFilePath), 0755)
	systemConfigFile, err := os.OpenFile(systemConfigFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	encoder = yaml.NewEncoder(systemConfigFile)
	if err := encoder.Encode(installFile.SystemConfig); err != nil {
		return err
	}

	_ = systemConfigFile.Close()
	systemConfigFile, err = os.Open(systemConfigFilePath)
	if err != nil {
		return err
	}

	if err := installFile.AddFile(systemConfigFile, "config/system.config"); err != nil {
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

	downloadedFile, err := os.Open(savedLocation)
	if err != nil {
		return err
	}
	defer downloadedFile.Close()
	return installFile.AddFile(downloadedFile, targetPath)
}

/**
Downloads the given url and extracts a specific file out of the archive and saves it to the installer
*/
func (installFile *InstallFile) AddDownloadedNestedFile(url string, wantedFile string, targetPath string) error {
	fileLocation, err := util.DownloadFile(url)
	if err != nil {
		return err
	}

	extractedFilePath, err := util.ExtractFromGzip(fileLocation, wantedFile)
	if err != nil {
		return err
	}

	extractedFile, err := os.Open(extractedFilePath)
	if err != nil {
		return err
	}

	return installFile.AddFile(extractedFile, targetPath)

}

func (installFile *InstallFile) AddFileByPath(filePath string, targetPath string) error {
	fileObj, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fileObj.Close()

	if err := installFile.AddFile(fileObj, targetPath); err != nil {
		return err
	}

	return nil

}

/**
Adds the given file to the installer
*/
func (installFile *InstallFile) AddFile(file fs.File, targetPath string) error {
	//standardize targetPath
	targetPath = regexp.MustCompile("^/").ReplaceAllString(targetPath, "")

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file for %s: %s", targetPath, err)
	}

	if installFile.addedFiles[targetPath] {
		ui.VPrintf("File %s already added to installer", targetPath)
		return nil
	}
	installFile.addedFiles[targetPath] = true

	ui.VPrintf("Adding %s to installer", targetPath)

	if strings.HasPrefix(targetPath, "data/agent/images") && strings.HasSuffix(targetPath, ".tar") {
		if err := installFile.saveImagesTar(file, targetPath); err != nil {
			return err
		}
	} else {
		return installFile.AddFileData(file, targetPath, fileInfo.ModTime())
	}

	return nil

}

func (installFile *InstallFile) AddFileData(data io.Reader, installerPath string, modTime time.Time) error {
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}

	hash := sha1.New()
	_, err = io.Copy(hash, bytes.NewReader(dataBytes))
	if err != nil {
		return fmt.Errorf("cannot compute hash for %s: %s", installerPath, err)
	}

	hashBytes := hash.Sum(nil)[:20]
	dataHash := hex.EncodeToString(hashBytes)

	installerPath = regexp.MustCompile("^.?/").ReplaceAllString(installerPath, "")

	uncompressedSize := uint64(len(dataBytes))
	header := &zip.FileHeader{
		Name:               installerPath,
		UncompressedSize64: uncompressedSize,
		Modified:           modTime,
	}

	entryWriter, err := installFile.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not write header for file %s: %s", installFile.file.Name(), err)
	}

	written, err := io.Copy(entryWriter, bytes.NewReader(dataBytes))
	if err != nil {
		return fmt.Errorf("could not write to %s: %s", installFile.file.Name(), err)
	}

	if uint64(written) != uncompressedSize {
		return fmt.Errorf("expected %s to be %d bytes but was %d", installerPath, uncompressedSize, written)
	}

	installFile.PackageConfig.Files[installerPath] = dataHash

	return nil
}

func (installFile *InstallFile) AddHelmChart(chartFilePath string, chartId string, overrideParameters map[string]interface{}) error {
	chartFileHash, err := global_util.HashFile(chartFilePath)
	if err != nil {
		return err
	}

	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      chartId,
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           "https://%{KUBERNETES_API}%/static/charts/" + chartId + "-" + chartFileHash + ".tgz",
			"targetNamespace": "default",
		},
	}

	if overrideParameters != nil && len(overrideParameters) > 0 {
		valuesString, err := yaml.Marshal(overrideParameters)
		if err != nil {
			return err
		}
		manifest["spec"].(map[string]interface{})["valuesContent"] = string(valuesString)
	}

	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	manifestData = []byte(strings.ReplaceAll(string(manifestData), "valuesContent: |\n", "valuesContent: |-\n"))

	if err := installFile.AddFileData(bytes.NewReader(manifestData), "data/server/manifests/"+chartId+".yaml", time.Now()); err != nil {
		return err
	}

	chartFile, err := os.Open(chartFilePath)
	if err != nil {
		return err
	}
	if err := installFile.AddFile(chartFile, "data/server/static/charts/"+chartId+"-"+chartFileHash+".tgz"); err != nil {
		return err
	}

	loadedChart, err := loader.Load(chartFilePath)
	if err != nil {
		return err
	}

	return installFile.processManifests(loadedChart)

}

func (installFile *InstallFile) installerFileMode(targetPath string) int64 {
	if strings.HasPrefix(targetPath, "lib") || strings.HasPrefix(targetPath, "bin") {
		return int64(0755)
	}
	return int64(0644)
}

func (installFile *InstallFile) AddDirectory(assetBase fs.FS, targetBase string) error {
	err := fs.WalkDir(assetBase, ".", func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() {
			pathFile, err := assetBase.Open(path)
			if err != nil {
				return err
			}
			return installFile.AddFile(pathFile, filepath.Join(targetBase, path))
		}
		return nil
	})
	return err
}

func (installFile *InstallFile) AddImage(tag string) error {
	if !installFile.dockerImages[tag] {
		ui.Printf("Including image %s", tag)
		installFile.dockerImages[tag] = true
	}

	return nil
}

func (installFile *InstallFile) saveDockerImages() error {

	var allTags []string
	for tag, _ := range installFile.dockerImages {
		allTags = append(allTags, tag)
		err := docker.ImagePull(tag)
		if err != nil {
			return fmt.Errorf("error pulling %s: %s", tag, err)
		}
	}

	ui.Printf("Collecting containers...")
	if len(installFile.dockerImages) > 0 {
		imagesTarPath := environment.TempPath("images-*.tar")
		ui.VPrintf("Including %s in %s", strings.Join(allTags, ", "), imagesTarPath)
		if err := docker.SaveImages(imagesTarPath, allTags...); err != nil {
			return fmt.Errorf("error collecting containers: %s", err)
		}

		imagesTarFile, err := os.Open(imagesTarPath)
		if err != nil {
			return fmt.Errorf("error opening %s: %s", imagesTarPath, err)
		}

		if err := installFile.AddFile(imagesTarFile, "data/agent/images/images.tar"); err != nil {
			return err
		}
	}
	return nil
}

func (installFile *InstallFile) saveImagesTar(imagesTarFile fs.File, targetPath string) error {
	defer ui.StartProgressf("Compressing " + targetPath).Stop()

	targetPath = strings.Replace(targetPath, ".tar", ".untar", 1)

	ui.VPrintf("Saving images tar to %s", targetPath)

	tarReader := tar.NewReader(imagesTarFile)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if header == nil {
			continue
		}

		target := targetPath + "/" + header.Name

		switch header.Typeflag {

		case tar.TypeReg:
			if strings.HasSuffix(target, ".tar") {
				cachePath := environment.CachePath("files/" + target + ".gz")

				_, err := os.Stat(cachePath)
				if os.IsNotExist(err) {
					err := os.MkdirAll(filepath.Dir(cachePath), 0755)
					if err != nil {
						return fmt.Errorf("cannot create cache directory %s: %s", cachePath, err)
					}

					ui.VPrintf("caching compressed layer at %s", cachePath)
					cacheFile, err := os.Create(cachePath)
					if err != nil {
						return fmt.Errorf("cannot create cache file: %s", err)
					}

					gzipWriter, err := gzip.NewWriterLevel(cacheFile, installFile.CompressionLevel)
					if err != nil {
						return err
					}

					buffTarReader := bufio.NewReader(tarReader)

					_, err = buffTarReader.WriteTo(gzipWriter)
					if err != nil {
						log.Fatal(err)
					}
					_ = gzipWriter.Close()
					_ = cacheFile.Close()
				}

				cacheFile, err := os.Open(cachePath)
				if err != nil {
					return fmt.Errorf("error opening cache file %s: %s", cachePath, err)
				}

				if err := installFile.AddFileData(cacheFile, target+".gz", header.ModTime); err != nil {
					return err
				}
			} else {
				if err := installFile.AddFileData(tarReader, target, header.ModTime); err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func (installFile *InstallFile) ClearDockerImages() error {
	ui.Printf("Cleaning up containers...")

	//listFilters := filters.NewArgs()
	//listFilters.Add("label", "ruckstack.built=true")
	//
	//images, err := docker.ImageList(types.ImageListOptions{
	//	Filters: listFilters,
	//})
	//if err != nil {
	//	return err
	//}
	//for _, image := range images {
	//	ui.Printf("Removing %s", image.ID)
	//	if err := docker.ImageRemove(image.ID); err != nil {
	//		return err
	//	}
	//}

	return nil

}

/**
Parses the descriptorContent as a kubernetes descriptor and saves any referenced containers to the install file
*/
func (installFile *InstallFile) AddImagesInManifest(descriptorContent []byte) error {
	decoder := yaml.NewDecoder(bytes.NewReader(descriptorContent))

	for {
		var value interface{}
		err := decoder.Decode(&value)

		if err == io.EOF {
			break
		}
		if err != nil {
			errMessage := strings.Replace(err.Error(), "yaml: ", "", 1) //remove extra "yaml: "
			return fmt.Errorf("yaml syntax error: %s", errMessage)
		}
		if value == nil {
			continue
		}

		var output bytes.Buffer
		outputWriter := bufio.NewWriter(&output)
		encoder := yaml.NewEncoder(outputWriter)
		if err := encoder.Encode(value); err != nil {
			return err
		}
		if err := outputWriter.Flush(); err != nil {
			return nil
		}

		obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(output.Bytes(), nil, nil)
		if err != nil {
			return err
		}

		var podSpec *coreV1.PodSpec

		switch groupVersionKind.Kind {
		case "StatefulSet":
			podSpec = &obj.(*appV1.StatefulSet).Spec.Template.Spec
		case "Deployment":
			podSpec = &obj.(*appV1.Deployment).Spec.Template.Spec
		case "DaemonSet":
			podSpec = &obj.(*appV1.DaemonSet).Spec.Template.Spec
		case "ReplicaSet":
			podSpec = &obj.(*appV1.ReplicaSet).Spec.Template.Spec
		case "Pod":
			podSpec = &obj.(*coreV1.Pod).Spec
		}

		if podSpec != nil {
			for _, container := range podSpec.Containers {
				if err := installFile.AddImage(container.Image); err != nil {
					return err
				}
			}
		}

	}
	return nil

}

func (installFile *InstallFile) processManifests(loadedChart *chart.Chart) error {
	options := chartutil.ReleaseOptions{
		Name:      "testRelease",
		Namespace: "default",
	}

	cvals, err := chartutil.CoalesceValues(loadedChart, map[string]interface{}{})
	if err != nil {
		return err
	}
	valuesToRender, err := chartutil.ToRenderValues(loadedChart, cvals, options, nil)
	if err != nil {
		return err
	}

	render, err := engine.Render(loadedChart, valuesToRender)
	if err != nil {
		return err
	}

	for filename, data := range render {
		data = strings.TrimSpace(data)
		if len(data) == 0 {
			continue
		}
		if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			if err := installFile.AddImagesInManifest([]byte(data)); err != nil {
				return fmt.Errorf("Error parsing %s: %s", filename, err)
			}
		}
	}

	return nil
}
