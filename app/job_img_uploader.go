package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"telego/util"
	"telego/util/strext"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
)

type ImgUploaderJob struct {
	WorkDir   string
	ImagePath string
}

type ModJobImgUploaderStruct struct {
	imageLocks sync.Map
}

var ModJobImgUploader *ModJobImgUploaderStruct = &ModJobImgUploaderStruct{}

func (m *ModJobImgUploaderStruct) JobCmdName() string {
	return "img-uploader"
}

func (m *ModJobImgUploaderStruct) uploadImage(imagePath string) {
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(util.GetEntryDir(), imagePath)
	}

	// scan files under imagePath
	files, err := filepath.Glob(filepath.Join(imagePath, "*.tar"))
	if err != nil {
		fmt.Println(color.RedString("Failed to scan files under %s: %v", imagePath, err))
		os.Exit(1)
	}

	// get image uploader url
	img_upload_server, err := util.MainNodeConfReader{}.ReadPubConf(util.PubConfTypeImgUploaderUrl{})
	if err != nil {
		fmt.Println(color.RedString("Failed to read image uploader url: %v", err))
		os.Exit(1)
	}
	img_upload_server = strings.TrimSpace(img_upload_server)

	res, err := util.UploadMultipleFilesInOneConnection(files, img_upload_server)
	if err != nil {
		fmt.Println(color.RedString("Failed to upload files: %v", err))
		os.Exit(1)
	} else {
		fmt.Println(color.GreenString("Uploaded files successfully:"))
		fmt.Println(res)
	}
}

func (m *ModJobImgUploaderStruct) startServer(workdir string) {
	util.PrintStep("ImgUploader", "starting server...")
	_, err := os.Stat(workdir)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(color.RedString("workdir %s access error: %v", workdir, err))
			return
		}
	}

	// 文件上传处理函数
	uploadHandler := func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法是否为 POST
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// 解析 multipart 表单
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 限制上传大小为 10 MB
			http.Error(w, "Error parsing form data", http.StatusBadRequest)
			return
		}

		// 遍历所有上传的文件
		files := r.MultipartForm.File["files"] // 获取字段名为 "files" 的文件
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			return
		}

		tarFiles := []string{}
		tempDir, err := os.MkdirTemp("", "uploads-*")
		defer os.RemoveAll(tempDir)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating temp dir: %v", err), http.StatusInternalServerError)
			return
		}
		for _, fileHeader := range files {
			func() {
				// 打开上传的文件
				if !strings.HasSuffix(fileHeader.Filename, ".tar") {
					return
				}
				file, err := fileHeader.Open()
				if err != nil {
					http.Error(w, fmt.Sprintf("Error opening file %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
					return
				}
				defer file.Close()

				// 创建目标文件

				dstPath := filepath.Join(tempDir, fileHeader.Filename)
				if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
					http.Error(w, fmt.Sprintf("Error creating directory: %v", err), http.StatusInternalServerError)
					return
				}
				dstFile, err := os.Create(dstPath)
				if err != nil {
					http.Error(w, fmt.Sprintf("Error saving file %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
					return
				}
				defer dstFile.Close()

				// 将上传的文件内容写入目标文件
				if _, err := io.Copy(dstFile, file); err != nil {
					http.Error(w, fmt.Sprintf("Error writing file %s: %v", fileHeader.Filename, err), http.StatusInternalServerError)
					return
				}

				tarFiles = append(tarFiles, dstPath)
				fmt.Printf("File uploaded successfully: %s\n", fileHeader.Filename)
			}()
		}

		// 镜像信息结构体
		type ImageInfo struct {
			imagename string // like xxx/xxx
			tag       string
			id        string
			arm       bool
			amd       bool
			file      string
		}

		imageBaseName := func(i ImageInfo) string {
			arr := strings.Split(i.imagename, "/")
			return arr[len(arr)-1]
		}

		imgInfoMap := make(map[string][]ImageInfo)
		processImage := func(tarPath string) error {
			// 1. 加载镜像
			// 加载镜像, 返回 imagename xxx/xxx 和 tag
			loadImage := func(tarPath string) (ImageInfo, error) {
				cmd := exec.Command("docker", "load", "-i", tarPath)
				cmd.CombinedOutput()
				cmd = exec.Command("docker", "load", "-i", tarPath)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return ImageInfo{}, fmt.Errorf("failed to load image: %s", err)
				}
				// 解析镜像名称和标签
				outputStr := string(output)
				if strings.HasPrefix(outputStr, "Loaded image ID: ") {
					id := strings.TrimSpace(strings.Split(outputStr, ": ")[1])
					fmt.Printf("Loaded image from %s with id %s\n", tarPath, id)
					return ImageInfo{
						file: tarPath,
						id:   id,
					}, nil
				}

				if strings.HasPrefix(outputStr, "Loaded image: ") {
					nameTag := strings.TrimSpace(strings.Split(outputStr, ": ")[1])
					fmt.Printf("Loaded image from %s with tag %s\n", tarPath, nameTag)
					// 2. 获取id信息
					getImageID := func(imageName string) ([]string, error) {
						cmd := exec.Command("docker", "image", "list", "-q", imageName)
						output, err := cmd.CombinedOutput()
						if err != nil {
							return nil, fmt.Errorf("failed to list image ids: %s", err)
						}
						ids := strings.Fields(string(output))
						return ids, nil
					}
					ids, err := getImageID(nameTag)
					if err != nil {
						return ImageInfo{}, fmt.Errorf("image '%s' not found", nameTag)
					}
					if len(ids) == 0 {
						return ImageInfo{}, fmt.Errorf("image '%s' not found", nameTag)
					}

					return ImageInfo{
						file:      tarPath,
						imagename: strings.Split(nameTag, ":")[0],
						tag:       strings.Split(nameTag, ":")[1],
						id:        ids[0],
					}, nil
				}

				return ImageInfo{}, fmt.Errorf("unexpected output from docker load: %s", outputStr)
			}
			imageInfo, err := loadImage(tarPath)
			if err != nil {
				return fmt.Errorf("failed to load image: %v", err)
			}

			// 3. 获取架构信息
			loadImageInfoArch := func(info *ImageInfo) error {
				if info == nil {
					return fmt.Errorf("invalid ImageInfo: nil pointer")
				}

				cmd := exec.Command("docker", "inspect", info.id)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("failed to inspect image: %v", err)
				}

				// 解析 JSON 输出
				var inspectResult []map[string]interface{}
				if err := json.Unmarshal(output, &inspectResult); err != nil {
					return fmt.Errorf("failed to parse docker inspect output: %v", err)
				}

				if len(inspectResult) == 0 {
					return fmt.Errorf("no data found for image ID: %s", info.id)
				}

				// 提取架构信息
				arch, ok := inspectResult[0]["Architecture"].(string)
				if !ok {
					return fmt.Errorf("failed to find architecture info in image data")
				}
				fmt.Println("Img Architecture:", arch, " ,id:", info.id)

				// 更新 ImageInfo 结构体
				info.arm = strings.Contains(arch, "arm64")
				info.amd = strings.Contains(arch, "amd64")

				return nil
			}
			err = loadImageInfoArch(&imageInfo)
			if err != nil {
				return fmt.Errorf("failed to get image info: %v", err)
			}

			// 4. 补全镜像信息
			if imageInfo.imagename == "" {
				// use filename
				imageInfo.imagename = strings.Split(tarPath, ".tar")[0]
			}
			if imageInfo.tag == "" {
				imageInfo.tag = strext.SafeSubstring(imageInfo.id, 0, 5)
			}

			// 加入到收集map
			_, ok := imgInfoMap[imageInfo.imagename+":"+imageInfo.tag]
			if !ok {
				imgInfoMap[imageInfo.imagename+":"+imageInfo.tag] = []ImageInfo{}
			}
			imgInfoMap[imageInfo.imagename+":"+imageInfo.tag] = append(
				imgInfoMap[imageInfo.imagename+":"+imageInfo.tag],
				imageInfo,
			)
			return nil
		}
		for _, tarPath := range tarFiles {
			err := processImage(tarPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error processing image: %v", err)))
				return
			}
		}

		if len(imgInfoMap) > 0 {
			registryConfYaml, err := util.MainNodeConfReader{}.ReadSecretConf(util.SecretConfTypeImgRepo{})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error reading image repo config: %v", err)))
				return
			}
			registryConf := util.ContainerRegistryConf{}
			err = yaml.Unmarshal([]byte(registryConfYaml), &registryConf)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Error reading image repo config: %v", err)))
				return
			}

			util.ModDocker.SetUserPwd(registryConf.User, registryConf.Password)
		}

		wait := sync.WaitGroup{}
		failress := make([]error, len(imgInfoMap))
		succress := make([]string, len(imgInfoMap))
		idx := 0
		for _, info := range imgInfoMap {
			wait.Add(1)
			idx_ := idx
			info_ := info
			go func() {
				// 上传镜像
				uploadImage := func(imageInfos []ImageInfo) error {
					var amdInfo *ImageInfo
					var armInfo *ImageInfo
					func() {
						for _, imageInfo := range imageInfos {
							if imageInfo.amd && imageInfo.arm {
								amdInfo = &imageInfo
								armInfo = &imageInfo
							}
						}
						if amdInfo != nil && armInfo != nil {
							return
						}

						for _, imageInfo := range imageInfos {
							if imageInfo.amd {
								amdInfo = &imageInfo
							}
							if imageInfo.arm {
								armInfo = &imageInfo
							}
						}
					}()

					if amdInfo == nil && armInfo == nil {
						return fmt.Errorf("no amd or arm image found")
					}
					fmt.Printf("infos %v %v\n", amdInfo, armInfo)
					{
						tagWithArch := func(info ImageInfo, arch string) (string, error) {
							// 1. tag by info.id
							registryTag := util.UrlJoin(util.ImgRepoAddressNoPrefix, "/teleinfra/"+imageBaseName(info)+":"+info.tag+"_"+arch)
							_, err := util.ModRunCmd.ShowProgress(
								"docker", "tag", info.id, registryTag,
							).WithRoot().BlockRun()
							if err != nil {
								return "", fmt.Errorf("docker tag failed %v", err)
							}
							return registryTag, nil
						}
						amdTag := ""
						if amdInfo != nil {
							amdTag, err = tagWithArch(*amdInfo, "amd64")
							if err != nil {
								return err
							}
						}
						armTag := ""
						if armInfo != nil {
							armTag, err = tagWithArch(*armInfo, "arm64")
							if err != nil {
								return err
							}
						}

						// 2. upload
						push := func(tag string) error {
							cmds, err := util.ModDocker.PushDockerImage(tag)
							for _, cmd := range cmds {
								_, err := cmd.BlockRun()
								if err != nil {
									return fmt.Errorf("docker upload %v failed %v", cmd.Cmds(), err)
								}
							}
							return err
						}
						if amdTag != "" {
							err = push(amdTag)
							if err != nil {
								return err
							}
						}
						if armTag != "" {
							err = push(armTag)
							if err != nil {
								return err
							}
						}

						// 3. manifest, allow fail
						noArchTag := util.UrlJoin(util.ImgRepoAddressNoPrefix, "/teleinfra/"+imageBaseName(*amdInfo)+":"+amdInfo.tag)
						fmt.Println("manifest noArchTag:", noArchTag)
						util.ModRunCmd.ShowProgress(
							"docker", "manifest", "create",
							noArchTag,
							amdTag, armTag, "--insecure", "--amend",
						).WithRoot().BlockRun()
						annotate := func(tag string) error {
							arch := "amd64"
							if strings.Contains(tag, "arm64") {
								arch = "arm64"
							}
							_, err := util.ModRunCmd.ShowProgress(
								"docker", "manifest", "annotate",
								noArchTag,
								"--arch", arch, tag,
							).WithRoot().BlockRun()
							if err != nil {
								return fmt.Errorf("docker manifest annotate failed %v", err)
							}
							return err
						}
						if armTag != "" {
							err := annotate(armTag)
							if err != nil {
								return err
							}
						}
						if amdTag != "" {
							err := annotate(amdTag)
							if err != nil {
								return err
							}
						}
						_, err := util.ModRunCmd.ShowProgress(
							"docker", "manifest", "push",
							noArchTag, "--insecure",
						).WithRoot().BlockRun()
						if err != nil {
							return fmt.Errorf("docker manifest push failed %v", err)
						}
					}
					res := ""
					if amdInfo != nil && armInfo != nil {
						res = fmt.Sprintf("Uploaded image %s:%s with arm in %s & amd in %s",
							armInfo.imagename, armInfo.tag, armInfo.file, amdInfo.file)
					} else if amdInfo != nil {
						res = fmt.Sprintf("Uploaded image %s:%s with amd in %s",
							amdInfo.imagename, amdInfo.tag, amdInfo.file)
					} else if armInfo != nil {
						res = fmt.Sprintf("Uploaded image %s:%s with arm in %s",
							armInfo.imagename, armInfo.tag, armInfo.file)
					}
					succress[idx_] = res
					return nil
				}
				failress[idx_] = uploadImage(info_)

				defer wait.Done()
			}()
			idx += 1
		}
		wait.Wait()

		// 返回成功响应
		if err := funk.Find(failress, func(err error) bool {
			return err != nil
		}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error uploading files: %v", err)))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strings.Join(succress, "\n")))
		}

		fmt.Fprintf(w, "Files uploaded successfully")
	}

	http.HandleFunc("/upload", uploadHandler) // 注册文件上传的处理函数

	// 启动 HTTP 服务器
	port := "8080"
	fmt.Printf("Starting server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func (m *ModJobImgUploaderStruct) ParseJob(ImgUploaderCmd *cobra.Command) *cobra.Command {
	job := &ImgUploaderJob{}

	// 绑定命令行标志到结构体字段
	ImgUploaderCmd.Flags().StringVar(&job.WorkDir, "workdir", "", "workdir for image uploader server")
	ImgUploaderCmd.Flags().StringVar(&job.ImagePath, "image", "", "image path for image uploader client")
	ImgUploaderCmd.Run = func(_ *cobra.Command, _ []string) {
		if job.WorkDir == "" && job.ImagePath == "" {
			fmt.Println(color.RedString("No workdir for server"))
		}
		if job.WorkDir != "" {
			m.startServer(job.WorkDir)
		}
		if job.ImagePath != "" {
			m.uploadImage(job.ImagePath)
		}
	}
	// err := ImgUploaderCmd.Execute()
	// if err != nil {
	// 	return nil,nil
	// }

	return ImgUploaderCmd
}

type ImgUploaderMode interface {
	ImgUploaderModeDummyInterface()
}

type ImgUploaderModeServer struct {
	WorkDir string
}

type ImgUploaderModeClient struct {
	ImagePath string
}

func (m ImgUploaderModeServer) ImgUploaderModeDummyInterface() {}

func (m ImgUploaderModeClient) ImgUploaderModeDummyInterface() {}

func (m *ModJobImgUploaderStruct) NewCmd(
	mode ImgUploaderMode,
) []string {
	switch mode.(type) {
	case ImgUploaderModeServer:
		mode := mode.(ImgUploaderModeServer)
		return []string{"telego", "img-uploader", "--workdir", mode.WorkDir}
	case ImgUploaderModeClient:
		mode := mode.(ImgUploaderModeClient)
		return []string{"telego", "img-uploader", "--image", mode.ImagePath}
	default:
		panic(color.RedString("unknown mode %v", mode))
	}
}

// in app entry
func (m *ModJobImgUploaderStruct) ImgUploaderLocal(mode ImgUploaderMode) {
	// cmds := []string{}
	cmds := ModJobImgUploader.NewCmd(mode)
	util.ModRunCmd.ShowProgress(cmds[0], cmds[1:]...).BlockRun()
}
