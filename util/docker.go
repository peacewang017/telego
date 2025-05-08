package util

import (
	"errors"
	"fmt"
	"strings"
)

// var ImgRepoAddressNoPrefix = strings.ReplaceAll(
//
//	strings.ReplaceAll(
//		ImgRepoAddressWithPrefix,
//		"https://", ""),
//	"http://", "",
//
// )
func ImgRepoAddressNoPrefix() string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			ImgRepoAddressWithPrefix,
			"https://", ""),
		"http://", "",
	)
}

type ModDockerStruct struct {
}

var ModDocker ModDockerStruct

var dockerUser string
var dockerPassword string

func (m ModDockerStruct) SetUserPwd(user string, pwd string) {
	dockerUser = user
	dockerPassword = pwd
}

func (m ModDockerStruct) DockerLoginCmd() ([]string, error) {
	if dockerUser == "" || dockerPassword == "" {
		return []string{}, errors.New("docker user or password is empty")
	}
	return []string{"docker", "login", ImgRepoAddressNoPrefix(), "-u", dockerUser, "-p", dockerPassword}, nil
}

func (m ModDockerStruct) BuildDockerImage(dockerfilePath string, targetImgName string) *CmdBuilder {
	return ModRunCmd.NewBuilder("docker", "build", "-t", targetImgName, "-f", dockerfilePath, ".")
}

func (m ModDockerStruct) PushDockerImage(targetImgName string) ([]*CmdBuilder, error) {
	logincmd, err := m.DockerLoginCmd()
	if err != nil {
		return nil, err
	}
	return []*CmdBuilder{
		ModRunCmd.NewBuilder(logincmd[0], logincmd[1:]...),
		ModRunCmd.NewBuilder("docker", "push", targetImgName),
	}, nil
}

// DockerComposeService represents a service in the Docker Compose file.
type DockerComposeService struct {
	Name        string
	Image       string
	Ports       []string
	Environment map[string]string
	Volumes     []string
}

// DockerComposeBuilder is the builder for constructing a Docker Compose file.
type DockerComposeBuilder struct {
	services []DockerComposeService
}

// NewDockerComposeBuilder initializes a new builder.
func NewDockerComposeBuilder() *DockerComposeBuilder {
	return &DockerComposeBuilder{
		services: []DockerComposeService{},
	}
}

// AddService adds a service to the Docker Compose file.
func (b *DockerComposeBuilder) AddService(service DockerComposeService) *DockerComposeBuilder {
	b.services = append(b.services, service)
	return b
}

// Build generates the final Docker Compose file content.
func (b *DockerComposeBuilder) Build() string {
	var sb strings.Builder

	sb.WriteString("version: '3.3'\n")
	sb.WriteString("services:\n")
	for _, service := range b.services {
		sb.WriteString(fmt.Sprintf("  %s:\n", service.Name))
		sb.WriteString(fmt.Sprintf("    image: %s\n", service.Image))
		if len(service.Ports) > 0 {
			sb.WriteString("    ports:\n")
			for _, port := range service.Ports {
				sb.WriteString(fmt.Sprintf("      - \"%s\"\n", port))
			}
		}
		if len(service.Environment) > 0 {
			sb.WriteString("    environment:\n")
			for key, value := range service.Environment {
				sb.WriteString(fmt.Sprintf("      %s: \"%s\"\n", key, value))
			}
		}
		if len(service.Volumes) > 0 {
			sb.WriteString("    volumes:\n")
			for _, volume := range service.Volumes {
				sb.WriteString(fmt.Sprintf("      - \"%s\"\n", volume))
			}
		}
	}
	return sb.String()
}
