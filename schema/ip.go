package schema

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// IPRange type (deprecated, please check IPCidr type)
// this is kept for backward compatibility for objects
// that are already in the database
type IPRange struct{ net.IPNet }

// ParseIPRange parse iprange
func ParseIPRange(txt string) (r IPRange, err error) {
	if len(txt) == 0 {
		//empty ip net value
		return r, nil
	}
	//fmt.Println("parsing: ", string(text))
	ip, net, err := net.ParseCIDR(txt)
	if err != nil {
		return r, err
	}

	net.IP = ip
	r.IPNet = *net
	return
}

// MustParseIPRange prases iprange, panics if invalid
func MustParseIPRange(txt string) IPRange {
	r, err := ParseIPRange(txt)
	if err != nil {
		panic(err)
	}
	return r
}

// UnmarshalText loads IPRange from string
func (i *IPRange) UnmarshalText(text []byte) error {
	v, err := ParseIPRange(string(text))
	if err != nil {
		return err
	}

	i.IPNet = v.IPNet
	return nil
}

// MarshalJSON dumps iprange as a string
func (i IPRange) MarshalJSON() ([]byte, error) {
	if len(i.IPNet.IP) == 0 {
		return []byte(`""`), nil
	}
	v := fmt.Sprint("\"", i.String(), "\"")
	return []byte(v), nil
}

func (i IPRange) String() string {
	return i.IPNet.String()
}

// IPCidr type is improved version of IPRange which can
// marshal/unmarshal itself to BSON for database readability
type IPCidr struct{ net.IPNet }

// ParseIPCidr parse iprange
func ParseIPCidr(txt string) (r IPCidr, err error) {
	if len(txt) == 0 {
		//empty ip net value
		return r, nil
	}
	//fmt.Println("parsing: ", string(text))
	ip, net, err := net.ParseCIDR(txt)
	if err != nil {
		return r, err
	}

	net.IP = ip
	r.IPNet = *net
	return
}

// MustParseIPCidr prases ipcidr, panics if invalid
func MustParseIPCidr(txt string) IPCidr {
	r, err := ParseIPCidr(txt)
	if err != nil {
		panic(err)
	}
	return r
}

// UnmarshalText loads IPRange from string
func (i *IPCidr) UnmarshalText(text []byte) error {
	v, err := ParseIPCidr(string(text))
	if err != nil {
		return err
	}

	i.IPNet = v.IPNet
	return nil
}

// MarshalBSONValue dumps ip as a bson
func (i IPCidr) MarshalBSONValue() (bsontype.Type, []byte, error) {
	value := ""
	if len(i.IPNet.IP) != 0 {
		value = i.String()
	}

	return bson.MarshalValue(value)
}

// UnmarshalBSONValue loads IP from bson
func (i *IPCidr) UnmarshalBSONValue(t bsontype.Type, bytes []byte) error {
	if t != bsontype.String {
		return fmt.Errorf("invalid ip bson type '%s'", t.String())
	}
	ip, _, ok := bsoncore.ReadString(bytes)
	if !ok {
		return fmt.Errorf("invalid bson ip format input")
	}

	v, err := ParseIPCidr(ip)
	if err != nil {
		return errors.Wrap(err, "failed to parse ip address")
	}

	i.IPNet = v.IPNet
	return nil
}

// MarshalJSON dumps iprange as a string
func (i IPCidr) MarshalJSON() ([]byte, error) {
	if len(i.IPNet.IP) == 0 {
		return []byte(`""`), nil
	}
	v := fmt.Sprint("\"", i.String(), "\"")
	return []byte(v), nil
}

func (i IPCidr) String() string {
	return i.IPNet.String()
}

// IP schema type
type IP struct{ net.IP }

// MarshalBSONValue dumps ip as a bson
func (i IP) MarshalBSONValue() (bsontype.Type, []byte, error) {
	value := ""
	if len(i.IP) != 0 {
		value = i.String()
	}

	return bson.MarshalValue(value)
}

// UnmarshalBSONValue loads IP from bson
func (i *IP) UnmarshalBSONValue(t bsontype.Type, bytes []byte) error {
	if t != bsontype.String {
		return fmt.Errorf("invalid ip bson type '%s'", t.String())
	}
	ip, _, ok := bsoncore.ReadString(bytes)
	if !ok {
		return fmt.Errorf("invalid bson ip format input")
	}

	v := net.ParseIP(ip)

	i.IP = v
	return nil
}
