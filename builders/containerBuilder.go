package builders

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/zos/pkg/container/logger"
	containerstats "github.com/threefoldtech/zos/pkg/container/stats"
)

// ContainerBuilder is a struct that can build containers
type ContainerBuilder struct {
	workloads.Container
}

// NewContainerBuilder creates a new container builder
func NewContainerBuilder(nodeID string, flist string, capacity workloads.ContainerCapacity, networkConnection []workloads.NetworkConnection) *ContainerBuilder {
	return &ContainerBuilder{
		Container: workloads.Container{
			NodeId:            nodeID,
			Flist:             flist,
			Capacity:          capacity,
			NetworkConnection: networkConnection,
		},
	}
}

// LoadContainerBuilder loads a container builder based on a file path
func LoadContainerBuilder(reader io.Reader) (*ContainerBuilder, error) {
	container := workloads.Container{}

	err := json.NewDecoder(reader).Decode(&container)
	if err != nil {
		return &ContainerBuilder{}, err
	}

	return &ContainerBuilder{Container: container}, nil
}

// Save saves the container builder to an IO.Writer
func (c *ContainerBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(c.Container)
	if err != nil {
		return err
	}
	return err
}

// Build validates and encrypts the secret environment of the container
func (c *ContainerBuilder) Build() (workloads.Container, error) {
	// TODO check validity fields
	if c.Container.Flist == "" {
		return workloads.Container{}, fmt.Errorf("flist cannot be empty")
	}

	if c.Container.SecretEnvironment == nil {
		c.Container.SecretEnvironment = make(map[string]string)
	}

	for k, value := range c.Container.Environment {
		secret, err := encryptSecret(value, c.Container.NodeId)
		if err != nil {
			return workloads.Container{}, errors.Wrapf(err, "failed to encrypt env with key '%s'", k)
		}
		c.Container.SecretEnvironment[k] = secret
	}
	c.Container.Environment = make(map[string]string)
	return c.Container, nil
}

// WithNodeID sets the node ID to the container
func (c *ContainerBuilder) WithNodeID(nodeID string) *ContainerBuilder {
	c.Container.NodeId = nodeID
	return c
}

// WithFlist sets the flist to the container
func (c *ContainerBuilder) WithFlist(flist string) *ContainerBuilder {
	c.Container.Flist = flist
	return c
}

// WithNetwork sets the networks to the container
func (c *ContainerBuilder) WithNetwork(connections []workloads.NetworkConnection) *ContainerBuilder {
	c.Container.NetworkConnection = connections
	return c
}

// WithInteractive sets the interactive flag to the container
func (c *ContainerBuilder) WithInteractive(interactive bool) *ContainerBuilder {
	c.Container.Interactive = interactive
	return c
}

// WithHubURL sets the hub url to the container
func (c *ContainerBuilder) WithHubURL(url string) *ContainerBuilder {
	c.Container.HubUrl = url
	return c
}

// WithEnvs sets the environments to the container
func (c *ContainerBuilder) WithEnvs(envs []string) (*ContainerBuilder, error) {
	environments, err := splitEnvs(envs)
	if err != nil {
		return c, errors.Wrap(err, "failed to split envs")
	}
	c.Container.Environment = environments
	return c, nil
}

// WithSecretEnvs sets the secret environments to the container
func (c *ContainerBuilder) WithSecretEnvs(envs map[string]string) *ContainerBuilder {
	c.Container.SecretEnvironment = envs
	return c
}

// WithEntrypoint sets the entrypoint to the container
func (c *ContainerBuilder) WithEntrypoint(entrypoint string) *ContainerBuilder {
	c.Container.Entrypoint = entrypoint
	return c
}

// WithVolumes sets the volumes to the container
func (c *ContainerBuilder) WithVolumes(mounts []string) (*ContainerBuilder, error) {
	containerMounts, err := splitMounts(mounts)
	if err != nil {
		return c, errors.Wrap(err, "failed to split containermounts")
	}
	c.Container.Volumes = containerMounts
	return c, nil
}

// WithConnection sets the conntections to the container
func (c *ContainerBuilder) WithConnection(connections []workloads.NetworkConnection) *ContainerBuilder {
	c.Container.NetworkConnection = connections
	return c
}

// WithStatsAggregator sets the stats aggregators to the container
func (c *ContainerBuilder) WithStatsAggregator(stats string) (*ContainerBuilder, error) {
	aggregators, err := parseStats(stats)
	if err != nil {
		return c, errors.Wrap(err, "failed to parse stats")
	}
	c.Container.StatsAggregator = aggregators
	return c, nil
}

// WithLogs sets the logs to the container
func (c *ContainerBuilder) WithLogs(stdout, stderr string) (*ContainerBuilder, error) {
	logs, err := parseLogs(stdout, stderr)
	if err != nil {
		return c, errors.Wrap(err, "failed to parse logs")
	}
	c.Container.Logs = logs
	return c, nil
}

// WithContainerCapacity sets the container capacity to the container
func (c *ContainerBuilder) WithContainerCapacity(cap workloads.ContainerCapacity) *ContainerBuilder {
	c.Container.Capacity = cap
	return c
}

func splitEnvs(envs []string) (map[string]string, error) {
	out := make(map[string]string, len(envs))

	for _, env := range envs {
		ss := strings.SplitN(env, "=", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("envs flag mal formatted: %v", env)
		}
		out[ss[0]] = ss[1]
	}

	return out, nil
}

func splitMounts(mounts []string) ([]workloads.ContainerMount, error) {
	out := make([]workloads.ContainerMount, 0, len(mounts))

	for _, mount := range mounts {
		ss := strings.SplitN(mount, ":", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("mounts flag mal formatted: %v", mount)
		}

		out = append(out, workloads.ContainerMount{
			VolumeId:   ss[0],
			Mountpoint: ss[1],
		})
	}

	return out, nil
}

func parseLogs(stdout, stderr string) ([]workloads.Logs, error) {
	var logs []workloads.Logs

	// validating stdout argument
	_, _, err := logger.RedisParseURL(stdout)
	if err != nil {
		return []workloads.Logs{}, err
	}

	// copy stdout to stderr
	lr := stdout

	// check if stderr is specified
	if nlr := stderr; nlr != "" {
		// validating stderr argument
		_, _, err := logger.RedisParseURL(nlr)
		if err != nil {
			return []workloads.Logs{}, err
		}

		lr = nlr
	}

	lg := workloads.Logs{
		Type: "redis",
		Data: workloads.LogsRedis{
			Stdout: stdout,
			Stderr: lr,
		},
	}

	logs = append(logs, lg)
	return logs, nil
}

func parseStats(stats string) ([]workloads.StatsAggregator, error) {
	var sts []workloads.StatsAggregator
	if s := stats; s != "" {
		// validating stdout argument
		_, _, err := logger.RedisParseURL(s)
		if err != nil {
			return []workloads.StatsAggregator{}, err
		}

		ss := workloads.StatsAggregator{
			Type: containerstats.RedisType,
			Data: workloads.StatsRedis{
				Endpoint: s,
			},
		}

		sts = append(sts, ss)
	}
	return sts, nil
}
