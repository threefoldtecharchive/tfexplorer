package builders

import (
	"encoding/json"
	"io"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// ZDBBuilder is a struct that can build ZDB's
type ZDBBuilder struct {
	workloads.ZDB
}

// NewZdbBuilder creates a new zdb builder and initializes some default values
func NewZdbBuilder(nodeID string, size int64, mode workloads.ZDBModeEnum, diskType workloads.DiskTypeEnum) *ZDBBuilder {
	return &ZDBBuilder{
		ZDB: workloads.ZDB{
			NodeId:   nodeID,
			Size:     size,
			Mode:     mode,
			DiskType: diskType,
		},
	}
}

// LoadZdbBuilder loads a zdb builder based on a file path
func LoadZdbBuilder(reader io.Reader) (*ZDBBuilder, error) {
	zdb := workloads.ZDB{}

	err := json.NewDecoder(reader).Decode(&zdb)
	if err != nil {
		return &ZDBBuilder{}, err
	}

	return &ZDBBuilder{ZDB: zdb}, nil
}

// Save saves the zdb builder to an IO.Writer
func (z *ZDBBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(z.ZDB)
	if err != nil {
		return err
	}
	return err
}

// Build validates and encrypts the zdb secret
func (z *ZDBBuilder) Build() (workloads.ZDB, error) {
	encrypted, err := encryptSecret(z.ZDB.Password, z.ZDB.NodeId)
	if err != nil {
		return workloads.ZDB{}, err
	}

	z.ZDB.Password = encrypted
	return z.ZDB, nil
}

// WithNodeID sets the node ID to the zdb
func (z *ZDBBuilder) WithNodeID(nodeID string) *ZDBBuilder {
	z.ZDB.NodeId = nodeID
	return z
}

// WithSize sets the size on the zdb
func (z *ZDBBuilder) WithSize(size int64) *ZDBBuilder {
	z.ZDB.Size = size
	return z
}

// WithMode sets the mode to the zdb
func (z *ZDBBuilder) WithMode(mode workloads.ZDBModeEnum) *ZDBBuilder {
	z.ZDB.Mode = mode
	return z
}

// WithPassword sets the password to the zdb
func (z *ZDBBuilder) WithPassword(password string) *ZDBBuilder {
	z.ZDB.Password = password
	return z
}

// WithDiskType sets the disktype to the zdb
func (z *ZDBBuilder) WithDiskType(diskType workloads.DiskTypeEnum) *ZDBBuilder {
	z.ZDB.DiskType = diskType
	return z
}

// WithPublic sets if public to the zdb
func (z *ZDBBuilder) WithPublic(public bool) *ZDBBuilder {
	z.ZDB.Public = public
	return z
}

// WithStatsAggregator sets the stats aggregators to the zdb
func (z *ZDBBuilder) WithStatsAggregator(aggregators []workloads.StatsAggregator) *ZDBBuilder {
	z.ZDB.StatsAggregator = aggregators
	return z
}

// WithPoolID sets the poolID to the zdb
func (z *ZDBBuilder) WithPoolID(poolID int64) *ZDBBuilder {
	z.ZDB.PoolId = poolID
	return z
}
