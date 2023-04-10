package GoLib

import (
	"github.com/docker/docker/api/types/swarm"
)

func createServiceSpec(
	imageName string,
	imageVersion string,
	envVars map[string]string,
	networks []string,
	secrets []string,
	volumes map[string]string,
	ports []swarm.PortConfig,
) swarm.ServiceSpec {
	if imageVersion == "" {
		imageVersion = "latest"
	}

	env := []string{}
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}

	networkConfigs := []swarm.NetworkAttachmentConfig{}
	for _, network := range networks {
		networkConfigs = append(networkConfigs, swarm.NetworkAttachmentConfig{Target: network})
	}

	secretRefs := []*swarm.SecretReference{}
	for _, secret := range secrets {
		secretRefs = append(secretRefs, &swarm.SecretReference{
			SecretName: secret,
			File: &swarm.SecretReferenceFile{
				Name: "/run/secrets/" + secret,
				UID:  "0",
				GID:  "0",
				Mode: 0444,
			},
		})
	}

	mounts := []swarm.Mount{}
	for src, target := range volumes {
		mounts = append(mounts, swarm.Mount{
			Type:   swarm.MountTypeBind,
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
	}
}
