package app

import (
	"strings"
	"telego/util"
	"path/filepath"
)

func (i *MenuItem) LoadSubPrjs() {
	util.PrintStep("LoadSubPrjs", "from "+ConfigLoad().ProjectDir)
	if i.Name == "deploy" { // this is mapped to project dir

		// i.Children = []*MenuItem{}
		listPrjDir, err := util.DirLinkTool.List(ConfigLoad().ProjectDir)
		if err == nil {
			for _, entry := range listPrjDir {
				entryPath := filepath.Join(ConfigLoad().ProjectDir, entry.Name())
				if is,err:=util.DirLinkTool.IsDirOrLinkDir(entryPath);is && err==nil {
					{
						//sort mappedRes
						var mi *MenuItem
						if strings.HasPrefix(entry.Name(), "bin_") {
							mi = &MenuItem{
								Name:     entry.Name(),
								Comment:  "安装项目",
								Children: []*MenuItem{},
							}
						} else if strings.HasPrefix(entry.Name(), "k8s_") {
							mi = &MenuItem{
								Name:     entry.Name(),
								Comment:  "部署项目",
								Children: []*MenuItem{},
							}
						} else if strings.HasPrefix(entry.Name(), "dist_") {
							mi = &MenuItem{
								Name:     entry.Name(),
								Comment:  "分布式部署",
								Children: []*MenuItem{},
							}
						} else {
							// mi = &MenuItem{
							// 	Name:     entry.Name(),
							// 	Comment:  "未知类型项目",
							// 	Children: []*MenuItem{},
							// }
						}

						if strings.HasPrefix(entry.Name(), "bin_") ||
							strings.HasPrefix(entry.Name(), "k8s_") ||
							strings.HasPrefix(entry.Name(), "dist_") {

							if util.HasNetwork() {
								mi.Children = append(mi.Children, &MenuItem{
									Name:     "prepare",
									Comment:  "准备资源",
									Children: []*MenuItem{},
								})
							}
							if util.FileServerAccessible() {
								mi.Children = append(mi.Children, &MenuItem{
									Name:     "upload",
									Comment:  "上传文件",
									Children: []*MenuItem{},
								}, &MenuItem{
									Name:     "apply",
									Comment:  "部署、安装",
									Children: []*MenuItem{},
								})
							}
						}

						if mi != nil {
							i.Children = append(i.Children, mi)
						}
					}
				}else{
					is,err:= util.DirLinkTool.IsDirOrLinkDir(entryPath)
					util.Logger.Warnf("skip %s,isDirOrLinkDir:%v,err:%v", entryPath,is,err)
				}
			}
		}
	}
}
