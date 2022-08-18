package container

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	config "noirgate/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kevinburke/twilio-go"
)

var (
	cli, _ = docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
)

func ProvisionNoirgate(uuid string) (containerId string, noirgateID string, IPAddress string, Error error) {

	imageName := config.ImageName
	ctx := context.Background()
	// sha256 the uuid to get the container name
	hasher := sha256.New()
	hasher.Write([]byte(uuid))
	containerHash := fmt.Sprintf("%x", hasher.Sum(nil))
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:       imageName,
		OpenStdin:   true,
		AttachStdin: true,
		Tty:         true,
		Domainname:  "noir",
		Env:         []string{("OTP=" + uuid)},

		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}, &container.HostConfig{
		CapDrop:    []string{"CAP_SYS_ADMIN", "CAP_DAC_OVERRIDE", "CAP_DAC_READ_SEARCH", "CAP_SYS_MODULE", "CAP_SYS_PTRACE"},
		AutoRemove: true,
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			*config.FlagClientDockerNetwork: &network.EndpointSettings{
				NetworkID: *config.FlagClientDockerNetwork,
				// DriverOpts: map[string]string{
				// 	"com.docker.network.bridge.enable_icc": "true",
				// },
			},
		},
	}, nil, ("noirgate-" + containerHash))
	if err != nil {
		Error = err
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", "", "", err
	}

	IPAddress, err = GetContainerPrivateIP(resp.ID)
	noirgateID = fmt.Sprintf("%.32s", resp.ID)
	log.Println("IPAddress:", IPAddress)
	return resp.ID, noirgateID, IPAddress, Error
}

func TerminateNoirgate(containerId string, phoneNumber twilio.PhoneNumber) {
	ctx := context.Background()
	err := cli.ContainerStop(ctx, containerId, nil)
	if err != nil {
		log.Println(err)
	}
}

func IsNoirgateActive(containerId string) (Running bool) {
	ctx := context.Background()
	cli, _ := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	defer cli.Close()
	inspectResult, _, err := cli.ContainerInspectWithRaw(ctx, containerId, true)
	if err != nil {
		Running = false
	} else {
		Running = inspectResult.State.Running
	}
	return Running
}

func GetContainerPrivateIP(containerId string) (IPAddress string, err error) {
	ctx := context.Background()
	cli, _ := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	defer cli.Close()
	// log.Println("Retrieving container with ID", containerId)
	inspectResult, _, err := cli.ContainerInspectWithRaw(ctx, containerId, true)
	if err != nil {
		log.Fatal(err)
	}
	// print container information
	// log.Println(string(inspectBytes))

	// This is only valid for the default network
	// IPAddress = inspectResult.NetworkSettings.IPAddress

	// When network is defined IP addresses live in the NetworkSettings.Networks
	IPAddress = inspectResult.NetworkSettings.Networks[*config.FlagClientDockerNetwork].IPAddress
	return IPAddress, err

}

func SendSandboxContents(containerId string) (string, error) {
	ctx := context.Background()
	cli, _ := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	defer cli.Close()

	// zip -r9 /tmp/files.zip /tmp/.
	containerFiles, pathstat, err := cli.CopyFromContainer(ctx, containerId, "/tmp/files.zip")
	filePath := fmt.Sprintf("/tmp/%s.zip", containerId)

	if err != nil {
		return "", err
	}
	fmt.Println(pathstat)
	tr := tar.NewReader(containerFiles)
	for {

		contentFromSandbox, err := ioutil.ReadAll(tr)
		if err != nil {
			return "", err
		}
		// write tar file
		err = ioutil.WriteFile(filePath, contentFromSandbox, os.FileMode(os.O_CREATE|os.O_APPEND))
		_, err = tr.Next()
		if err != nil {
			break
		}

	}

	// upload tar file from local temp to s3
	// delete tar file  from local temp

	return filePath, err
}
