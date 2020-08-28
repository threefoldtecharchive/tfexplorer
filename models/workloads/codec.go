package workloads

import (
	"encoding/json"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// Codec is a struc used to encode/decode model.Workloads
type Codec struct {
	Workloader
}

// MarshalBSON implements bson.Marshaller
func (w Codec) MarshalBSON() ([]byte, error) {
	return bson.Marshal(w.Workloader)
}

// UnmarshalBSON implements bson.Unmarshaller
func (w *Codec) UnmarshalBSON(buf []byte) error {
	workload, err := UnmarshalBSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &Codec{}
	}

	*w = Codec{Workloader: workload}

	return nil
}

// MarshalJSON implements JSON.Marshaller
func (w Codec) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Workloader)
}

// UnmarshalJSON implements JSON.Unmarshaller
func (w *Codec) UnmarshalJSON(buf []byte) error {
	workload, err := UnmarshalJSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &Codec{}
	}

	*w = Codec{Workloader: workload}

	return nil
}

// UnmarshalJSON decodes a workload from JSON format
func UnmarshalJSON(buffer []byte) (Workloader, error) {
	var itc ITContract
	if err := json.Unmarshal(buffer, &itc); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch itc.Contract.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		c.ITContract = itc
		err = json.Unmarshal(buffer, &c)
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		g.ITContract = itc
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		g.ITContract = itc
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		k.ITContract = itc
		err = json.Unmarshal(buffer, &k)
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		n.ITContract = itc
		err = json.Unmarshal(buffer, &n)
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		g.ITContract = itc
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		g.ITContract = itc
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		g.ITContract = itc
		err = json.Unmarshal(buffer, &g)
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		v.ITContract = itc
		err = json.Unmarshal(buffer, &v)
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		z.ITContract = itc
		err = json.Unmarshal(buffer, &z)
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}

// UnmarshalBSON decodes a workload from BSON format
func UnmarshalBSON(buffer []byte) (Workloader, error) {
	var itc ITContract
	if err := bson.Unmarshal(buffer, &itc); err != nil {
		return nil, errors.Wrap(err, "could not decode workload type")
	}

	var err error
	var workload Workloader

	switch itc.Contract.WorkloadType {
	case WorkloadTypeContainer:
		var c Container
		c.ITContract = itc
		workload = &c
	case WorkloadTypeDomainDelegate:
		var g GatewayDelegate
		g.ITContract = itc
		workload = &g
	case WorkloadTypeGateway4To6:
		var g Gateway4To6
		g.ITContract = itc
		workload = &g
	case WorkloadTypeKubernetes:
		var k K8S
		k.ITContract = itc
		workload = &k
	case WorkloadTypeNetworkResource:
		var n NetworkResource
		n.ITContract = itc
		workload = &n
	case WorkloadTypeProxy:
		var g GatewayProxy
		g.ITContract = itc
		workload = &g
	case WorkloadTypeReverseProxy:
		var g GatewayReverseProxy
		g.ITContract = itc
		workload = &g
	case WorkloadTypeSubDomain:
		var g GatewaySubdomain
		g.ITContract = itc
		workload = &g
	case WorkloadTypeVolume:
		var v Volume
		v.ITContract = itc
		workload = &v
	case WorkloadTypeZDB:
		var z ZDB
		z.ITContract = itc
		workload = &z
	default:
		return nil, errors.New("unrecognized workload type")
	}

	return workload, err
}
