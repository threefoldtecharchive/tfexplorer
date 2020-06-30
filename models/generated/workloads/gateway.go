package workloads

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

var _ Workloader = (*GatewayProxy)(nil)
var _ Capaciter = (*GatewayProxy)(nil)

type GatewayProxy struct {
	ReservationInfo

	Domain  string `bson:"domain" json:"domain"`
	Addr    string `bson:"addr" json:"addr"`
	Port    uint32 `bson:"port" json:"port"`
	PortTLS uint32 `bson:"port_tls" json:"port_tls"`
	PoolId  int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayProxy) GetRSU() RSU {
	return RSU{}
}

func (v *GatewayProxy) VerifyJSON() error {
	dup := GatewayProxy{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

var _ Workloader = (*GatewayReverseProxy)(nil)
var _ Capaciter = (*GatewayReverseProxy)(nil)

type GatewayReverseProxy struct {
	ReservationInfo

	Domain string `bson:"domain" json:"domain"`
	Secret string `bson:"secret" json:"secret"`
	PoolId int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayReverseProxy) GetRSU() RSU {
	return RSU{}
}

func (v *GatewayReverseProxy) VerifyJSON() error {
	dup := GatewayReverseProxy{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

var _ Workloader = (*GatewaySubdomain)(nil)
var _ Capaciter = (*GatewaySubdomain)(nil)

type GatewaySubdomain struct {
	ReservationInfo

	Domain string   `bson:"domain" json:"domain"`
	IPs    []string `bson:"ips" json:"ips"`
	PoolId int64    `bson:"pool_id" json:"pool_id"`
}

func (g *GatewaySubdomain) GetRSU() RSU {
	return RSU{}
}

func (v *GatewaySubdomain) VerifyJSON() error {
	dup := GatewaySubdomain{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

var _ Workloader = (*GatewayDelegate)(nil)
var _ Capaciter = (*GatewayDelegate)(nil)

type GatewayDelegate struct {
	ReservationInfo

	Domain string `bson:"domain" json:"domain"`
	PoolId int64  `bson:"pool_id" json:"pool_id"`
}

func (g *GatewayDelegate) GetRSU() RSU {
	return RSU{}
}

func (v *GatewayDelegate) VerifyJSON() error {
	dup := GatewayDelegate{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}

var _ Workloader = (*Gateway4To6)(nil)
var _ Capaciter = (*Gateway4To6)(nil)

type Gateway4To6 struct {
	ReservationInfo

	PublicKey string `bson:"public_key" json:"public_key"`
	PoolId    int64  `bson:"pool_id" json:"pool_id"`
}

func (g *Gateway4To6) GetRSU() RSU {
	return RSU{}
}

func (v *Gateway4To6) VerifyJSON() error {
	dup := Gateway4To6{}

	if err := json.Unmarshal([]byte(v.Json), &dup); err != nil {
		return errors.Wrap(err, "invalid json data")
	}

	// override the fields which are not part of the signature
	dup.ID = v.ID
	dup.Json = v.Json
	dup.CustomerTid = v.CustomerTid
	dup.NextAction = v.NextAction
	dup.SignaturesProvision = v.SignaturesProvision
	dup.SignatureFarmer = v.SignatureFarmer
	dup.SignaturesDelete = v.SignaturesDelete
	dup.Epoch = v.Epoch
	dup.Metadata = v.Metadata
	dup.Result = v.Result
	dup.WorkloadType = v.WorkloadType

	if match := reflect.DeepEqual(v, dup); !match {
		return errors.New("json data does not match actual data")
	}

	return nil
}
