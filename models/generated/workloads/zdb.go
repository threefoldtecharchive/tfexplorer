package workloads

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

var _ Workloader = (*ZDB)(nil)
var _ Capaciter = (*ZDB)(nil)

type ZDB struct {
	ReservationInfo `bson:",inline"`

	Size            int64             `bson:"size" json:"size"`
	Mode            ZDBModeEnum       `bson:"mode" json:"mode"`
	Password        string            `bson:"password" json:"password"`
	DiskType        DiskTypeEnum      `bson:"disk_type" json:"disk_type"`
	Public          bool              `bson:"public" json:"public"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
}

func (z *ZDB) GetRSU() RSU {
	switch z.DiskType {
	case DiskTypeHDD:
		return RSU{
			HRU: z.Size,
		}
	case DiskTypeSSD:
		return RSU{
			SRU: z.Size,
		}
	}
	return RSU{}
}

func (z *ZDB) SignatureChallenge() ([]byte, error) {
	ric, err := z.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(ric)
	if _, err := fmt.Fprintf(b, "%d", z.Size); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", z.Mode); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%s", z.Password); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%d", z.DiskType); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(b, "%t", z.Public); err != nil {
		return nil, err
	}
	for _, s := range z.StatsAggregator {
		if err := s.SigingEncode(b); err != nil {
			return nil, err
		}
	}

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

type DiskTypeEnum uint8

const (
	DiskTypeHDD DiskTypeEnum = iota
	DiskTypeSSD
)

func (e DiskTypeEnum) String() string {
	switch e {
	case DiskTypeHDD:
		return "hdd"
	case DiskTypeSSD:
		return "ssd"
	}
	return "UNKNOWN"
}

type ZDBModeEnum uint8

const (
	ZDBModeSeq ZDBModeEnum = iota
	ZDBModeUser
)

func (e ZDBModeEnum) String() string {
	switch e {
	case ZDBModeSeq:
		return "seq"
	case ZDBModeUser:
		return "user"
	}
	return "UNKNOWN"
}
