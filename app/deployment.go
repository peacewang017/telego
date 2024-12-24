package app

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"telego/app/config"
	"telego/util"

	"github.com/fatih/color"
	"github.com/mholt/archiver/v3"
	"github.com/mitchellh/mapstructure"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
)

func StrPtr(s string) *string {
	return &s
}

type LocalValue interface {
	LocalValueInterfaceDummy()
}

type LocalValueStr struct {
	Value string
}

func (s LocalValueStr) LocalValueInterfaceDummy() {}

// replace:
//
//	match: "\n"
//	with: "    \n"
type LocalValueReadFileReplace struct {
	Match string `mapstructure:"match"`
	With  string `mapstructure:"with"`
}

// func (l *LocalValueReadFileReplace) TransUnquote() {

// 	new := util.Unquote(l.Match)
// 	util.Logger.Debugf("replace: %s -> %s , %v, %v", l.Match, new, strings.Contains(l.Match, "\n"), strings.Contains(new, "\n"))
// 	l.Match = new

// 	new = util.Unquote(l.With)
// 	util.Logger.Debugf("replace: %s -> %s", l.With, new)
// 	l.With = new
// }

type LocalValueReadFile struct {
	ReadFromFile string                     `mapstructure:"read_from_file"`
	Replace      *LocalValueReadFileReplace `mapstructure:"replace,omitempty"`
}

func (l LocalValueReadFile) LocalValueInterfaceDummy() {}

// DeploymentYaml represents the structure of yml.
type DeploymentYaml struct {
	Comment     string                          `yaml:"comment"`
	LocalValues map[string]interface{}          `yaml:"local_values"`
	Secrets     []string                        `yaml:"secret,omitempty"`
	Prepare     []DeploymentPrepareItemYaml     `yaml:"prepare"`
	Helms       map[string]DeploymentHelm       `yaml:"helms,omitempty"`
	K8s         map[string]DeploymentK8s        `yaml:"k8s,omitempty"`
	Bin         map[string]DeploymentBinDetails `yaml:"bin,omitempty"`
}

type Deployment struct {
	Comment     string
	LocalValues map[string]LocalValue
	Secrets     []string
	Prepare     []DeploymentPrepareItem
	Helms       map[string]DeploymentHelm
	K8s         map[string]DeploymentK8s
	Bin         map[string]DeploymentBinDetails
}

func (d *DeploymentYaml) Verify(prjDir string, yml []byte, skipReadFromFile bool) (*Deployment, error) {

	dply := &Deployment{
		Comment: d.Comment,
		// default values
		LocalValues: map[string]LocalValue{
			"MAIN_NODE_IP": LocalValueStr{Value: util.MainNodeIp},
			"IMG_REPO":     LocalValueStr{Value: util.ImgRepoAddressNoPrefix},
		},
		Secrets: d.Secrets,
		Prepare: []DeploymentPrepareItem{},
		Helms:   d.Helms,
		K8s:     d.K8s,
		Bin:     d.Bin,
	}
	// trans to specific interface
	for idx, v := range d.LocalValues {
		_, ok := v.(string)
		if ok {
			dply.LocalValues[idx] = LocalValueStr{Value: v.(string)}
		} else {
			target := LocalValueReadFile{}
			err := mapstructure.Decode(v, &target)
			if err != nil {
				return nil, fmt.Errorf("invalid local_values item: only string or read_file is allowed")
			}
			dply.LocalValues[idx] = target
		}
	}

	// resolve local values self
	foreachLocalValue := func(cb func(k string, v LocalValue)) {
		for k, v := range dply.LocalValues {
			cb(k, v)
		}
	}
	// first iter replace ${key} in value or file path
	for k, _ := range dply.LocalValues {
		switch dply.LocalValues[k].(type) {
		case LocalValueStr:
			foreachLocalValue(func(otherK string, otherV LocalValue) {
				otherV_, ok := otherV.(LocalValueStr)
				if ok {
					switch dply.LocalValues[k].(type) {
					case LocalValueStr:
						v_ := dply.LocalValues[k].(LocalValueStr)
						v_.Value = strings.ReplaceAll(v_.Value, fmt.Sprintf("${%s}", otherK), otherV_.Value)
						dply.LocalValues[k] = v_
					case LocalValueReadFile:
						v_ := dply.LocalValues[k].(LocalValueReadFile)
						v_.ReadFromFile = strings.ReplaceAll(v_.ReadFromFile, fmt.Sprintf("${%s}", otherK), otherV_.Value)
						dply.LocalValues[k] = v_
					}
				}
			})
		}
	}

	if !skipReadFromFile {
		// second iter read files
		for k, v := range dply.LocalValues {
			switch v.(type) {
			case LocalValueReadFile:
				v_ := v.(LocalValueReadFile)
				fileContent, err := os.ReadFile(path.Join(prjDir, v_.ReadFromFile))
				if err != nil {
					return nil, fmt.Errorf("invalid local_values item: read_from_file %s not found", v_.ReadFromFile)
				}
				v_.ReadFromFile = string(fileContent)
				dply.LocalValues[k] = v_
			}
		}

		// third iter again replace all strs
		for k, _ := range dply.LocalValues {
			// switch v.(type) {
			// case LocalValueStr:
			foreachLocalValue(func(otherK string, otherV LocalValue) {
				// fmt.Printf("%s,%s:%s,%s\n", k, v, otherK, otherV)
				otherStr := ""
				switch otherV.(type) {
				case LocalValueStr:
					otherStr = otherV.(LocalValueStr).Value
				case LocalValueReadFile:
					otherStr = otherV.(LocalValueReadFile).ReadFromFile
				}
				switch dply.LocalValues[k].(type) {
				case LocalValueStr:
					v_ := dply.LocalValues[k].(LocalValueStr)
					newValue := strings.ReplaceAll(v_.Value, fmt.Sprintf("${%s}", otherK), otherStr)
					v_.Value = newValue
					dply.LocalValues[k] = v_
				case LocalValueReadFile:
					v_ := dply.LocalValues[k].(LocalValueReadFile)
					// fmt.Printf("replace %s in %s with %s\n", fmt.Sprintf("${%s}", otherK), v_.ReadFromFile, otherStr)
					newValue := strings.ReplaceAll(v_.ReadFromFile, fmt.Sprintf("${%s}", otherK), otherStr)
					// if newValue != v_.ReadFromFile {
					// 	fmt.Printf("%s -> %s\n", v_.ReadFromFile, newValue)
					// }
					v_.ReadFromFile = newValue
					replace := v_.Replace
					if replace != nil {
						v_.Replace = nil // consume it, so that it will not be replaced again
						// replace.TransUnquote()
						v_.ReadFromFile = strings.ReplaceAll(
							v_.ReadFromFile,
							replace.Match, replace.With)
						// util.Logger.Debugf("replace in value %s, (%s->%s)", k, replace.Match, replace.With)
					}
					dply.LocalValues[k] = v_
				}
			})
			// }
		}
	}

	// forth iter checks ${...} remains
	type ErrorOwnerError struct {
		err error
	}
	errOwner := &ErrorOwnerError{}
	foreachLocalValue(func(k string, v LocalValue) {
		switch v.(type) {
		case LocalValueStr:
			v_ := v.(LocalValueStr)
			if strings.Contains(v_.Value, "${") {
				errOwner.err = fmt.Errorf("invalid local_values item: ${...} not replaced: %s\n  values: %v", v_.Value, dply.LocalValues)
			}
		case LocalValueReadFile:
			v_ := v.(LocalValueReadFile)
			if strings.Contains(v_.ReadFromFile, "${") {
				errOwner.err = fmt.Errorf("invalid local_values item: ${...} not replaced: %s\n  values: %v", v_.ReadFromFile, dply.LocalValues)
			}
		}
	})
	if errOwner.err != nil {
		return nil, errOwner.err
	}

	// transform other field
	for _, item := range d.Prepare {
		itemTyped, err := item.Verify()
		if err != nil {
			return nil, fmt.Errorf("invalid prepare item: %w", err)
		}
		dply.Prepare = append(dply.Prepare, *itemTyped)
	}

	//replace all string with value
	replaceWithValue := func(s *string) {
		if s == nil {
			return
		}
		foreachLocalValue(func(k string, v LocalValue) {
			vStr := ""
			switch v := v.(type) {
			case LocalValueStr:
				vStr = v.Value
			case LocalValueReadFile:
				vStr = v.ReadFromFile
			}

			*s = strings.ReplaceAll(*s, fmt.Sprintf("${%s}", k), vStr)
		})
	}
	for _, b := range dply.Bin {
		replaceWithValue(b.PyInstaller)
	}
	for _, s := range dply.Helms {
		replaceWithValue(s.HelmDir)
		replaceWithValue(s.Namespace)
	}
	for _, s := range dply.Prepare {
		replaceWithValue(s.As)
		replaceWithValue(s.Image)
		replaceWithValue(s.URL)
		if s.FileMap != nil {
			replaceWithValue(s.FileMap.Content)
			replaceWithValue(s.FileMap.Path)
		}
		for _, t := range s.Trans {
			switch t.(type) {
			case DeploymentTransformCopy:
				for _, v := range t.(DeploymentTransformCopy).Copy {
					replaceWithValue(v.from)
					replaceWithValue(v.to)
				}
			case DeploymentTransformExtract:
			}
		}
	}

	return dply, nil
}

// DeploymentBinDetails represents the details for a binary in the bin field.
type DeploymentBinDetails struct {
	NoDefaultInstaller bool    `yaml:"no_default_installer,omitempty"`
	WinInstaller       string  `yaml:"win_installer,omitempty"`
	Appimage           string  `yaml:"appimage,omitempty"`
	PyInstaller        *string `yaml:"py_installer,omitempty"`
}

// DeploymentPrepareItemYaml represents an item in the "prepare" list.
type DeploymentPrepareItemYaml struct {
	Image   string             `yaml:"image,omitempty"`
	URL     string             `yaml:"url,omitempty"`
	As      string             `yaml:"as,omitempty"`
	FileMap *DeploymentFileMap `yaml:"filemap,omitempty"` // todo file map verify
	Trans   []interface{}      `yaml:"trans,omitempty"`
}

type DeploymentPrepareItem struct {
	Image   *string
	URL     *string
	As      *string
	FileMap *DeploymentFileMap
	Trans   []DeploymentTransform
}

func (i *DeploymentPrepareItemYaml) Verify() (*DeploymentPrepareItem, error) {
	count := 0
	if i.Image != "" {
		count++
	}
	if i.URL != "" {
		count++
	}
	if i.FileMap != nil {
		count++
	}
	if count != 1 {
		return nil, fmt.Errorf("only one of image/url/filemap can be specified")
	}
	item := &DeploymentPrepareItem{
		Image:   StrPtr(i.Image),
		URL:     StrPtr(i.URL),
		As:      StrPtr(i.As),
		FileMap: i.FileMap,
		Trans:   []DeploymentTransform{},
	}
	for _, trans := range i.Trans {
		switch trans.(type) {
		case string:
			if trans.(string) == "extract" {
				item.Trans = append(item.Trans, DeploymentTransformExtract{})
				continue
			} else {
				return nil, fmt.Errorf("invalid prepare/trans item: only 'extract' is allowed")
			}
		case map[string]interface{}:
			transMap := trans.(map[string]interface{})
			target := map[string][]map[string]string{}
			err := mapstructure.Decode(transMap, &target)
			copy, exist := target["copy"]
			if !exist {
				err = fmt.Errorf("could not find copy in one trans")
			}
			for _, oneCopy := range copy {
				if len(oneCopy) != 1 {
					err = fmt.Errorf("one copy should be single kv as {src: dest}")
					break
				}
			}
			if err != nil {
				err := fmt.Errorf("invalid prepare/trans item: only 'extract' or {'copy':[{'from':'to'}]} is allowed, \n  err: %v, \n  current is %v %v", err, trans, reflect.TypeOf(trans))
				return nil, err
			}
			// i.Trans[idx] = DeploymentTransformCopy{
			// 	Copy: copy,
			// }
			item.Trans = append(item.Trans, DeploymentTransformCopy{
				Copy: funk.Map(copy, func(v map[string]string) DeploymentTransformCopyOne {
					firstK := ""
					firstV := ""
					for k, v := range v {
						firstK = k
						firstV = v
					}
					return DeploymentTransformCopyOne{
						from: &firstK,
						to:   &firstV,
					}
				}).([]DeploymentTransformCopyOne),
			})
		default:
			err := fmt.Errorf("invalid prepare/trans item: only 'extract' or {'copy':[{'from':'to'}]} is allowed, \n  current is %v %v", trans, reflect.TypeOf(trans))
			return nil, err
		}
	}
	return item, nil
}

type DeploymentTransform interface {
	DeploymentTransformDummy()
}

type DeploymentTransformCopyOne struct {
	from *string
	to   *string
}

// Example
//
//	trans:
//	- copy:
//	  - rclone-v1.68.0-windows-amd64/rclone.exe: rclone_windows.exe
type DeploymentTransformCopy struct {
	Copy []DeploymentTransformCopyOne
}

func (c DeploymentTransformCopy) DeploymentTransformDummy() {}

type DeploymentTransformExtract struct{}

func (e DeploymentTransformExtract) DeploymentTransformDummy() {}

// DeploymentFileMap represents the file mapping details.
type DeploymentFileMap struct {
	Content *string `yaml:"content"`
	Path    *string `yaml:"path"`
	Mode    *string `yaml:"mode"`
}

// WriteToFile 打开文件并写入内容
func (dfm *DeploymentFileMap) WriteToFile() error {
	// 检查 Path 和 Content 是否有效
	if dfm.Path == nil || dfm.Content == nil {
		return fmt.Errorf("path or content is nil")
	}

	if len(*dfm.Mode) == 3 {
		*dfm.Mode = "0" + *dfm.Mode
	}

	mode, err := strconv.ParseInt(*dfm.Mode, 8, 64)
	if err != nil {
		return fmt.Errorf("Error parsing octal string:", err)
	}

	dir := path.Dir(*dfm.Path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("Error creating directory:", err)
	}

	// 创建或打开文件
	file, err := os.OpenFile(*dfm.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(mode))
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", *dfm.Path, err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(*dfm.Content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", *dfm.Path, err)
	}

	return nil
}

// DeploymentHelm represents a helm chart configuration.
type DeploymentHelm struct {
	HelmDir   *string `yaml:"helm-dir"`
	Namespace *string `yaml:"namespace"`
}

// DeploymentK8s represents a Kubernetes configuration.
type DeploymentK8s struct {
	K8sDir    *string `yaml:"k8s-dir"`
	Namespace *string `yaml:"namespace"`
}

func LoadDeploymentYml(subPrjDir string) (*Deployment, error) {
	ymlFile := path.Join(subPrjDir, "deployment.yml")
	if path.Base(ymlFile) != "deployment.yml" {
		util.Logger.Fatal("yml not found")
		os.Exit(1)
	}
	if !util.PathIsAbsolute(ymlFile) {
		util.Logger.Fatalf("path should be absolute %s", ymlFile)
		os.Exit(1)
	}

	data, err := os.ReadFile(ymlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	return LoadDeploymentYmlByContent(subPrjDir, data)
}

func LoadDeploymentYmlByContent(projectDir string, data []byte) (*Deployment, error) {
	var deployment_ DeploymentYaml

	err := yaml.Unmarshal(data, &deployment_)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML file at projectDir , %w", err)
	}

	skipReadFile := false
	if projectDir == "" {
		skipReadFile = true
	}
	d, err := deployment_.Verify(projectDir, data, skipReadFile)
	if err != nil {
		return nil, fmt.Errorf("invalid yml format: %w", err)
	}
	return d, nil
}

func DeploymentPrepare(project string, deployment *Deployment) error {
	curDir0 := util.CurDir()
	defer os.Chdir(curDir0)

	// projectDir := path.Join(LoadConfig().ProjectDir, project)
	os.Chdir(config.Load().ProjectDir)

	fmt.Println("deploymentPrepare", project)
	// os.MkdirAll("prepare", 0755)

	// // Step1: Load deployment
	// deployment, err := LoadDeploymentYml(project)
	// if err != nil {
	// 	return err
	// }

	os.Chdir(project)

	// Step2: Process prepare items
	// fmt.Printf("yml mapped as %v", deployment)
	// os.Chdir("prepare")
	for _, item := range deployment.Prepare {
		fmt.Println()
		if *item.Image != "" {
			// prepare image
			fmt.Printf("Preparing image: %s\n", *item.Image)
			ModJobImgPrepare.PrepareImages([]string{*item.Image})

		} else if *item.URL != "" {
			fmt.Println(color.BlueString("Downloading file from URL: %s", *item.URL))
			// download to download_cache
			downloadCachePath := path.Join("download_cache", filepath.Base(*item.URL))
			_, err := os.Stat(downloadCachePath)
			if err != nil {
				err := util.DownloadFile(*item.URL, downloadCachePath)
				if err != nil {
					fmt.Println(color.RedString("Failed to download file from %s\n   err: %v", *item.URL, err))
					// return fmt.Errorf("failed to download file from %s: %w", item.URL, err)
				}
			}

			_, err = os.Stat(downloadCachePath)
			if err == nil {
				// Execute "trans" command if provided
				if len(item.Trans) > 0 {
					defer os.RemoveAll("extract")
					for _, step := range item.Trans {
						switch step.(type) {
						case DeploymentTransformExtract:
							os.RemoveAll("extract")
							fmt.Println(color.BlueString("Extracting file: %s", downloadCachePath))
							err := os.MkdirAll("extract", 0755)
							if err != nil {
								return fmt.Errorf("failed to create directory: %w", err)
							}
							err = archiver.Unarchive(downloadCachePath, "extract")
							if err != nil {
								return fmt.Errorf("failed to extract file: %w", err)
							}
						case DeploymentTransformCopy:
							for _, copyStep := range step.(DeploymentTransformCopy).Copy {
								src := path.Join("extract", *copyStep.from)
								dest := *copyStep.to

								fmt.Println(color.BlueString("Copying %s to %s", src, dest))
								newSrc := path.Join(path.Dir(src), path.Base(dest))
								err := os.Rename(src, newSrc)
								if err != nil {
									return fmt.Errorf("failed to rename %s to %s: %w", src, dest, err)
								}

								util.ModRunCmd.CopyDirContentOrFileTo(newSrc, path.Dir(dest))
							}
						}

					}

					// fmt.Println(color.BlueString("Executing user defined transformation: %s at %s\n", item.Trans, CurDir()))
					// err := ModRunCmd.RunCommandShowProgress(item.Trans)
					// if err != nil {
					// 	return fmt.Errorf("failed to execute transformation %s: %w", item.Trans, err)
					// }
				} else if *item.As != "" {
					util.ModRunCmd.CopyDirContentOrFileTo(downloadCachePath, path.Dir(*item.As))
					src := path.Join(path.Dir(*item.As), path.Base(downloadCachePath))
					dest := *item.As
					err := os.Rename(src, dest)
					if err != nil {
						return fmt.Errorf("failed to rename %s to %s: %w", src, dest, err)
					}
				}
			}
		} else if item.FileMap != nil {
			err := item.FileMap.WriteToFile()
			if err != nil {
				fmt.Println(color.RedString("filemap write failed %v", err))
			}
		}

		// if item.FileMap != nil {
		// 	fmt.Printf("Creating file from map: %s\n", item.FileMap.Path)
		// 	err := createFileFromMap(item.FileMap)
		// 	if err != nil {
		// 		return fmt.Errorf("failed to create file %s: %w", item.FileMap.Path, err)
		// 	}
		// }
	}

	return nil
}

func DeploymentUpload(project string, deploymentConf *Deployment) error {
	if deploymentConf == nil {
		util.Logger.Warnf("deploymentConf is nil")
		return fmt.Errorf("deploymentConf is nil")
	}
	curDir0 := util.CurDir()
	defer os.Chdir(curDir0)

	// cd into projectDir
	projectDir := path.Join(config.Load().ProjectDir, project)
	fmt.Println(color.BlueString("Uploading project %s contents in %s", project, projectDir))
	os.Chdir(projectDir)

	// teledeploy mapped to /teledeploy on main node, which is a public file server on port 8003
	util.PrintStep("DeploymentUpload", "uploading public content (teledeploy) to main_node.fileserver")
	_, err := os.Stat("teledeploy")
	if err != nil {
		fmt.Println(color.YellowString("teledeploy is not exist, \n  please make sure if there are things to upload to public fileserver and prepare is executed"))
	} else {
		util.UploadToMainNode("teledeploy", path.Join("/teledeploy", project))
	}

	// teledeploy_secret mapped to /teledeploy_secret on main node, which is a private position, only can be accessed by rclone
	util.PrintStep("DeploymentUpload", "uploading secret content (teledeploy_secret) to main_node")
	_, err = os.Stat("teledeploy_secret")
	if err != nil {
		fmt.Println(color.YellowString("teledeploy_secret is not exist, \n  please make sure if there are things to upload to secret fileserver and prepare is executed"))
	} else {
		util.UploadToMainNode("teledeploy_secret", path.Join("/teledeploy_secret", project))
	}

	if strings.HasPrefix(project, "bin_") {
		// upload k8s things
		// - deployment.yml
		util.UploadToMainNode("deployment.yml", path.Join("/teledeploy", project))

	} else if strings.HasPrefix(project, "k8s_") {
		// upload k8s things
		// - images

		// upload images
		imageNames := funk.Map(
			funk.Filter(deploymentConf.Prepare, func(item DeploymentPrepareItem) bool {
				return item.Image != nil && *item.Image != ""
			}).([]DeploymentPrepareItem),
			func(item DeploymentPrepareItem) string {
				// fetch image name between last / and :
				splitTail := strings.Split(*item.Image, "/")
				return strings.Split(splitTail[len(splitTail)-1], ":")[0]
			},
		).([]string)
		if len(imageNames) > 0 {
			for _, imageName := range imageNames {
				localPath := path.Join(config.Load().ProjectDir, "container_image", fmt.Sprintf("image_%s", imageName))
				remotePath := fmt.Sprintf("/teledeploy_secret/container_image/image_%s", imageName)
				fmt.Println(color.BlueString("Uploading img %s to %s", localPath, remotePath))
				util.UploadToMainNode(localPath, remotePath)
			}
			imageNamesWithQuotes := []string{}
			for _, imageName := range imageNames {
				imageNamesWithQuotes = append(imageNamesWithQuotes, fmt.Sprintf("\"%s\"", imageName))
			}

			uploadScript := fmt.Sprintf(`
import os, subprocess

images=[%s]
REPO_HOST="192.168.31.96:5000"
REPO_NAMESPACE="teleinfra"
REPO_USER="admin"
REPOPW=f"74123"

os.chdir("/teledeploy_secret/container_image")
def run_command(command,allow_fail=False):
	"""
	Run a shell command and ensure it completes successfully.
	"""
	result = subprocess.run(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
	if result.returncode != 0:
		print(f"Error running command: {command}")
		print(result.stderr)
		if not allow_fail:
			raise Exception(f"Command failed: {command}")
	else:
		print(result.stdout)
		return result.stdout.decode("utf-8")
run_command(f"docker login -u {REPO_USER} -p {REPOPW} {REPO_HOST}")
for image in images:
	os.chdir(f"image_{image}")
	list_dir=os.listdir()
	arm64=[f for f in list_dir if f.find("_arm64_")!=-1]
	amd64=[f for f in list_dir if f.find("_amd64_")!=-1]

	def upload_arch(arch,arch_tar_pkg):
		img_name=run_command(f"docker load -i {arch_tar_pkg}").split("Loaded image: ")[-1].strip()
		img_key_no_prefix=img_name.split("/")[-1]
		id=run_command(f"docker image list -q {img_name}").strip()
		run_command(f"docker tag {id} {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_{arch}")
		run_command(f"docker push {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_{arch} -q")
		return img_key_no_prefix
	img_key_no_prefix=""
	mani_create_arch=""
	if len(arm64)>0:
		img_key_no_prefix=upload_arch("arm64",arm64[0])
		mani_create_arch+=f"{REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_amd64 "
	if len(amd64)>0:
		img_key_no_prefix=upload_arch("amd64",amd64[0])
		mani_create_arch+=f"{REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_arm64 "
	run_command(f"docker manifest create {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix} {mani_create_arch} --insecure", allow_fail=True)
	if len(arm64)>0:
		run_command(f"docker manifest annotate {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix} "
			f"--arch arm64 {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_arm64")
	if len(amd64)>0:
		run_command(f"docker manifest annotate {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix} "
			f"--arch amd64 {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix}_amd64")
	run_command(f"docker manifest push {REPO_HOST}/{REPO_NAMESPACE}/{img_key_no_prefix} --insecure")
	os.chdir("..")
	`, strings.Join(imageNamesWithQuotes, ","))

			util.StartRemoteCmds([]string{
				fmt.Sprintf("%s@%s", util.MainNodeUser, util.MainNodeIp),
			}, "sudo "+util.EncodeRemoteRunPy(uploadScript), "")
		}
	}

	return nil
}

func deploymentApply(project string) error {
	return nil
}
