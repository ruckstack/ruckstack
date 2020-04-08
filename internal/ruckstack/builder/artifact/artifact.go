package artifact

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/ruckstack/ruckstack/internal"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/resources"
	"github.com/ruckstack/ruckstack/internal/ruckstack/builder/shared"
	"github.com/ruckstack/ruckstack/internal/ruckstack/project"
	"github.com/ruckstack/ruckstack/internal/ruckstack/util"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	filesToAddToTar map[string]string
)

type Artifact struct {
	artifactPath string

	PackageConfig *internal.PackageConfig

	outDir string

	dockerImages []string

	artifactFile *os.File
	zipWriter    *zip.Writer
}

func NewArtifact(filename string, outDir string) *Artifact {
	artifact := &Artifact{
		outDir:       outDir,
		dockerImages: []string{},

		artifactPath: outDir + string(filepath.Separator) + filename,
	}

	artifact.PackageConfig = &internal.PackageConfig{
		Files:           map[string]string{},
		FilePermissions: map[string]internal.InstalledFileConfig{},
	}

	filesToAddToTar = map[string]string{}
	_, err := os.Stat(artifact.artifactPath)
	if !os.IsNotExist(err) {
		err = os.Remove(artifact.artifactPath)
		util.Check(err)
	}

	installerBytes, err := resources.Asset("out/installer")
	util.Check(err)
	err = ioutil.WriteFile(artifact.outDir+string(filepath.Separator)+filename, installerBytes, 0755)
	util.Check(err)

	artifact.artifactFile, err = os.OpenFile(artifact.artifactPath, os.O_RDWR|os.O_APPEND, 0755)
	util.Check(err)

	startOffset, err := artifact.artifactFile.Seek(0, os.SEEK_END)

	absoluteArtifactPath, err := filepath.Abs(artifact.artifactFile.Name())
	util.Check(err)
	log.Printf("Creating %s...", absoluteArtifactPath)
	artifact.zipWriter = zip.NewWriter(artifact.artifactFile)
	artifact.zipWriter.SetOffset(startOffset)

	return artifact
}

func (artifact *Artifact) Build(projectConfig *project.ProjectConfig, buildEnv *shared.BuildEnvironment) {
	for key, _ := range filesToAddToTar {
		artifact.AddFile(filesToAddToTar[key], key)
	}

	packageConfigFilePath := buildEnv.WorkDir + string(filepath.Separator) + "package.config"
	packageConfigFile, err := os.OpenFile(packageConfigFilePath, os.O_CREATE|os.O_RDWR, 0644)
	util.Check(err)
	encoder := yaml.NewEncoder(packageConfigFile)
	encoder.Encode(artifact.PackageConfig)

	artifact.AddFile(packageConfigFilePath, ".package.config")
	artifact.zipWriter.Close()
	artifact.artifactFile.Close()

	log.Printf("Built to %s", artifact.artifactFile.Name())
}

func (artifact *Artifact) AddFile(srcPath string, targetPath string) {
	file, err := os.Open(srcPath)
	util.Check(err)

	fileInfo, err := file.Stat()
	util.Check(err)

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	util.Check(err)
	hashBytes := hash.Sum(nil)[:20]
	hashString := hex.EncodeToString(hashBytes)
	_, err = file.Seek(0, 0)
	util.Check(err)

	artifact.addData(file, targetPath, uint64(fileInfo.Size()), fileInfo.ModTime(), hashString)
}

func (artifact *Artifact) addData(data io.Reader, artifactPath string, size uint64, modTime time.Time, dataHash string) {
	artifactPath = strings.Replace(artifactPath, "./", "", 1)

	header := &zip.FileHeader{
		Name:               artifactPath,
		UncompressedSize64: size,
		Modified:           modTime,
		//Mode:    artifact.artifactFileMode(assetPath),
	}

	entryWriter, err := artifact.zipWriter.CreateHeader(header)
	if err != nil {
		log.Printf("Could not write header for file '%s', got error '%s'", artifact.artifactPath, err.Error())
		util.Check(err)
	}

	written, err := io.Copy(entryWriter, data)
	util.Check(err)
	if uint64(written) != size {
		panic(fmt.Sprintf("Expected %s to be %d bytes but was %d", artifactPath, size, written))
	}

	artifact.PackageConfig.Files[artifactPath] = dataHash
}

func (artifact *Artifact) artifactFileMode(targetPath string) int64 {
	if strings.HasPrefix(targetPath, "lib") || strings.HasPrefix(targetPath, "bin") {
		return int64(0755)
	}
	return int64(0644)
}

func (artifact *Artifact) AddAsset(assetPath string, targetPath string) {
	fileInfo, err := resources.AssetInfo(assetPath)

	data, err := resources.Asset(assetPath)
	util.Check(err)

	hash := sha1.New()
	hash.Write(data)
	hashBytes := hash.Sum(nil)[:20]
	hashString := hex.EncodeToString(hashBytes)

	artifact.addData(bytes.NewReader(data), targetPath, uint64(fileInfo.Size()), fileInfo.ModTime(), hashString)
}

func (artifact *Artifact) AddAssetDir(assetBase string, targetBase string) {
	files, err := resources.AssetDir(assetBase)
	util.Check(err)
	for _, assetPath := range files {
		_, err := resources.AssetInfo(assetBase + "/" + assetPath)
		if err != nil {
			//probably a directory
			//TODO: check error to be sure
			artifact.AddAssetDir(assetBase+"/"+assetPath, targetBase+"/"+assetPath)
		} else {
			artifact.AddAsset(assetBase+"/"+assetPath, targetBase+"/"+assetPath)
		}

	}
}

//func (artifact *Artifact) handleTarFile(data io.Reader, path string, size uint64) (io.Reader, uint64) {
//	pathDir, pathFile := filepath.Split(path)
//	if pathDir == "data/server/manifests/" {
//		if strings.HasSuffix(strings.ToLower(pathFile), "yaml") || strings.HasSuffix(strings.ToLower(pathFile), "yml") {
//			outData := new(bytes.Buffer)
//
//			var parsedYaml map[interface{}]interface{}
//
//			encoder := yaml.NewEncoder(outData)
//
//			decoder := yaml.NewDecoder(data)
//			err := decoder.Decode(&parsedYaml)
//
//			for true {
//				if err != nil {
//					if err.Error() == "EOF" {
//						break
//					}
//					log.Printf("Error parsing %s", path)
//					util.Check(err)
//				}
//
//				kind := parsedYaml["kind"].(string)
//				log.Printf("Found kind %s in %s", kind, path)
//
//				//handler := yamlHandlers[strings.ToLower(kind)]
//				//if handler == nil {
//				//	log.Printf("No handlers for %s", kind)
//				//} else {
//				//	handler(parsedYaml)
//				//}
//
//				encoder.Encode(parsedYaml)
//
//				err = decoder.Decode(&parsedYaml)
//			}
//
//			return outData, uint64(outData.Len())
//
//		}
//	}
//
//	return data, size
//
//}

func (artifact *Artifact) IncludeDockerImage(tag string) {
	artifact.dockerImages = append(artifact.dockerImages, tag)

	dockerSaveCmd := exec.Command("docker", "image", "inspect", tag)
	dockerSaveCmd.Stdout = log.Writer()
	dockerSaveCmd.Stderr = log.Writer()
	err := dockerSaveCmd.Run()
	if err == nil {
		log.Printf("Already have image %s", tag)
	} else {
		dockerSaveCmd := exec.Command("docker", "pull", tag)

		dockerSaveCmd.Stdout = os.Stdout
		dockerSaveCmd.Stderr = os.Stderr
		err := dockerSaveCmd.Run()
		util.Check(err)
	}
}

func (artifact *Artifact) SaveDockerImages(buildEnv *shared.BuildEnvironment) {
	log.Printf("Collecting containers...")
	appImagePath := buildEnv.WorkDir + string(filepath.Separator) + "images.tar"
	dockerSaveCmd := exec.Command("docker", append([]string{"save", "--output", appImagePath}, artifact.dockerImages...)...)
	dockerSaveCmd.Stdout = os.Stdout
	dockerSaveCmd.Stderr = os.Stderr
	err := dockerSaveCmd.Run()
	util.Check(err)
	artifact.AddFile(appImagePath, "data/agent/images/images.tar")
}

func (artifact *Artifact) ClearDockerImages(env *shared.BuildEnvironment) {
	log.Printf("Cleaning up containers...")

	output, err := exec.Command("docker", "image", "ls",
		"--format", "'{{.Repository}}:{{.Tag}}'",
		"--filter", "label=ruckstack.built=true").Output()
	util.Check(err)

	for _, builtTag := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		builtTag = strings.Trim(builtTag, "'")
		log.Printf("Removing %s", builtTag)
		output, err := exec.Command("docker", "image", "rm", builtTag).Output()
		log.Println(string(output))
		util.Check(err)

	}

}

//func pullDockerImages(dockerImages []string) {
//	for _, image := range dockerImages {
//		dockerSaveCmd := exec.Command("docker", "pull", image)
//		//dockerSaveCmd.Dir = outDir
//		dockerSaveCmd.Stdout = os.Stdout
//		dockerSaveCmd.Stderr = os.Stderr
//		err := dockerSaveCmd.Run()
//		Check(err)
//
//	}
//}
//
//var yamlHandlers = map[interface{}]handleFunction{
//	"helmchart": handleHelmChart,
//}
//
//func handleHelmChart(yaml map[interface{}]interface{}) {
//	chart := getYamlValue(yaml, "spec", "chart").(string)
//	version := getYamlValue(yaml, "spec", "version").(string)
//	repo := getYamlValue(yaml, "spec", "repo").(string)
//
//	if strings.Contains(chart, "KUBERNETES_API") {
//		//allready set up correctly
//		return
//	}
//
//	if repo == "" {
//		repo = "https://kubernetes-charts.storage.googleapis.com"
//	} else {
//		repo = fmt.Sprintf("%s/helm", repo)
//	}
//
//	if chart == "" {
//		panic("chart is required")
//	}
//	if version == "" {
//		panic("version is required")
//	}
//
//	log.Printf("Parsing helm chart %s %s %s", repo, chart, version)
//	chartLocation, _ := downloadFile(cacheDir, fmt.Sprintf("%s/%s-%s.tgz", repo, chart, version))
//	log.Printf("Saved to %s", chartLocation)
//
//	chartName := regexp.MustCompile(`.+/`).ReplaceAllString(chart, "")
//
//	filesToAddToTar[fmt.Sprintf("data/server/static/charts/%s.tgz", chartName)] = chartLocation
//	setYamlValue(yaml, fmt.Sprintf("https://%%{KUBERNETES_API}%%/static/charts/%s.tgz", chartName), "spec", "chart")
//	deleteYamlValue(yaml, "spec", "repo")
//	deleteYamlValue(yaml, "spec", "version")
//}
//
//func setYamlValue(yaml map[interface{}]interface{}, value string, keys ...string) {
//	parentMap := getYamlValue(yaml, keys[0:len(keys)-1]...).(map[interface{}]interface{})
//	if parentMap == nil {
//		panic("Unknown parent")
//	}
//	parentMap[keys[len(keys)-1]] = value
//}
//
//func deleteYamlValue(yaml map[interface{}]interface{}, keys ...string) {
//	parentMap := getYamlValue(yaml, keys[0:len(keys)-1]...).(map[interface{}]interface{})
//	if parentMap == nil {
//		panic("Unknown parent")
//	}
//	delete(parentMap, keys[len(keys)-1])
//}
//
//type handleFunction func(map[interface{}]interface{})
//
//func getYamlValue(yaml map[interface{}]interface{}, keys ...string) interface{} {
//	value := yaml[keys[0]]
//	if value == nil {
//		if len(keys) == 1 {
//			return ""
//		}
//		return nil
//	}
//
//	if len(keys) == 1 {
//		return value
//	}
//
//	nestedMap, ok := value.(map[interface{}]interface{})
//	if !ok {
//		panic("Not a map")
//	}
//	return getYamlValue(nestedMap, keys[1:]...)
//
//}
