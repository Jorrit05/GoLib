package GoLib

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"

	"github.com/docker/docker/client"
)

func GetDockerClient() *client.Client {
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	// Check if Swarm is active
	info, err := cli.Info(context.Background())
	if err != nil {
		log.Fatalf("Error getting Docker info: %v", err)
	}
	if !info.Swarm.ControlAvailable {
		log.Fatal("This node is not a swarm manager. The agent can only be run on a swarm manager.")
	}
	return cli
}

func CreateServiceSpec(
	imageName string,
	imageVersion string,
	envVars map[string]string,
	networks []string,
	secrets []string,
	volumes map[string]string,
	ports []swarm.PortConfig,
) (swarm.ServiceSpec, *client.Client) {

	cli := GetDockerClient()

	if imageVersion == "" {
		imageVersion = "latest"
	}

	env := []string{}
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}

	networkConfigs := []swarm.NetworkAttachmentConfig{}
	for _, network := range networks {
		networkConfigs = append(networkConfigs, swarm.NetworkAttachmentConfig{
			Target:  network,
			Aliases: []string{imageName},
		})
	}

	secretRefs := []*swarm.SecretReference{}
	for _, secret := range secrets {
		id, err := GetSecretIDByName(cli, secret)
		if err != nil {
			log.Fatalf("Secret does not exist, %s", err)
		}

		secretRefs = append(secretRefs, &swarm.SecretReference{
			SecretName: secret,
			SecretID:   id,
			File: &swarm.SecretReferenceFileTarget{
				Name: fmt.Sprintf("/run/secrets/%s", secret), // This should be just the filename, not the full path
				UID:  "0",
				GID:  "0",
				Mode: 0444,
			},
		})
	}

	mounts := []mount.Mount{}
	for src, target := range volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: src,
			Target: target,
		})
	}

	return swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: imageName,
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   imageName + ":" + imageVersion,
				Env:     env,
				Secrets: secretRefs,
				Mounts:  mounts,
			},
			Networks: networkConfigs,
		},
		EndpointSpec: &swarm.EndpointSpec{
			Mode:  swarm.ResolutionModeVIP,
			Ports: ports,
		},
	}, cli
}

func GetSecretIDByName(cli *client.Client, secretName string) (string, error) {
	secrets, err := cli.SecretList(context.Background(), types.SecretListOptions{})
	if err != nil {
		return "", err
	}

	for _, secret := range secrets {
		if secret.Spec.Name == secretName {
			return secret.ID, nil
		}
	}

	return "", fmt.Errorf("secret not found: %s", secretName)
}
