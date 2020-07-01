package workloads

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

var _ Workloader = (*Volume)(nil)
var _ Capaciter = (*Volume)(nil)

type Volume struct {
	ReservationInfo `bson:",inline"`

	Size int64          `bson:"size" json:"size"`
	Type VolumeTypeEnum `bson:"type" json:"type"`
}

func (v *Volume) GetRSU() RSU {
	switch v.Type {
	case VolumeTypeHDD:
		return RSU{
			HRU: v.Size,
		}
	case VolumeTypeSSD:
		return RSU{
			SRU: v.Size,
		}
	}
	return RSU{}
}

func (v *Volume) SignatureChallenge() ([]byte, error) {
	ric, err := v.ReservationInfo.SignatureChallenge()
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(ric)
	fmt.Fprintf(b, "%d", v.Size)
	fmt.Fprintf(b, "%d", v.Type)

	h := sha256.New()
	if _, err := h.Write(b.Bytes()); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (v *Volume) BSONMarshal() ([]byte, error) {
	return bson.Marshal(v)
}

type VolumeTypeEnum uint8

const (
	VolumeTypeHDD VolumeTypeEnum = iota
	VolumeTypeSSD
)

func (e VolumeTypeEnum) String() string {
	switch e {
	case VolumeTypeHDD:
		return "HDD"
	case VolumeTypeSSD:
		return "SSD"
	}
	return "UNKNOWN"
}
