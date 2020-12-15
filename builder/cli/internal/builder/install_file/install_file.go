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
	"github.com/ruckstack/ruckstack/builder/cli/internal/environment"
	"github.com/ruckstack/ruckstack/builder/cli/internal/util"
	"github.com/ruckstack/ruckstack/builder/internal/docker"
	"github.com/ruckstack/ruckstack/common/config"
	"github.com/ruckstack/ruckstack/common/ui"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"io"
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
		dockerImages: map[string]bool{},
		addedFiles:   map[string]bool{},
	}

	ui.Printf("Building %s...", filepath.Base(installerPath))

	installerBinaryPath, err := environment.ResourcePath("installer")
	if err != nil {
		return nil, err
	}
	installerBytes, err := ioutil.ReadFile(installerBinaryPath)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(installerPath), 755)
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
	//standardize targetPath
	targetPath = regexp.MustCompile("^/").ReplaceAllString(targetPath, "")

	if installFile.addedFiles[targetPath] {
		ui.VPrintf("File %s already added to installer. Not adding %s", targetPath, srcPath)
		return nil
	}
	installFile.addedFiles[targetPath] = true

	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %s", srcPath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat %s: %s", srcPath, err)
	}

	ui.VPrintf("Adding %s to installer", srcPath)

	if strings.HasPrefix(targetPath, "data/agent/images") && strings.HasSuffix(targetPath, ".tar") {
		if err := installFile.saveImagesTar(srcPath, targetPath); err != nil {
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

func (installFile *InstallFile) AddHelmChart(chartFile string, chartId string) error {
	manifest := map[string]interface{}{
		"apiVersion": "helm.cattle.io/v1",
		"kind":       "HelmChart",
		"metadata": map[string]interface{}{
			"name":      chartId,
			"namespace": "kube-system",
		},
		"spec": map[string]interface{}{
			"chart":           "https://%{KUBERNETES_API}%/static/charts/" + chartId + ".tgz",
			"targetNamespace": "default",
		},
	}

	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	if err := installFile.AddFileData(bytes.NewReader(manifestData), "data/server/manifests/"+chartId+".yaml", time.Now()); err != nil {
		return err
	}

	if err := installFile.AddFile(chartFile, "data/server/static/charts/"+chartId+".tgz"); err != nil {
		return err
	}

	loadedChart, err := loader.Load(chartFile)
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

func (installFile *InstallFile) AddDirectory(assetBase string, targetBase string) error {
	err := filepath.Walk(assetBase, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return installFile.AddFile(path, filepath.Join(targetBase, path[len(assetBase):]))
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
		if err := docker.SaveImages(imagesTarPath, allTags...); err != nil {
			return fmt.Errorf("error collecting containers: %s", err)
		}

		if err := installFile.AddFile(imagesTarPath, "data/agent/images/images.tar"); err != nil {
			return err
		}
	}
	return nil
}

func (installFile *InstallFile) saveImagesTar(imagesTarPath string, targetPath string) error {
	monitor := ui.StartProgressf("Compressing " + filepath.Base(imagesTarPath))
	defer monitor.Stop()

	targetPath = strings.Replace(targetPath, ".tar", ".untar", 1)

	ui.VPrintf("Saving images tar to %s", targetPath)

	imagesTarFile, err := os.Open(imagesTarPath)
	if err != nil {
		return err
	}
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
