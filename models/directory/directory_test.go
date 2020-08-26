package directory

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

func TestPublicIface_Validate(t *testing.T) {
	type fields struct {
		Master  string
		Type    IfaceTypeEnum
		Ipv4    schema.IPRange
		Ipv6    schema.IPRange
		Gw4     net.IP
		Gw6     net.IP
		Version int64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: false,
		},
		{
			name: "missing master",
			fields: fields{
				Master: "",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "invalid master",
			fields: fields{
				Master: "thisisatoolongnameforanetworkinterface",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},

		{
			name: "wrong iface type",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeVlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "wrong ip4",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeVlan,
				Ipv4:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "wrong gw4",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeVlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("2a02:1802:5e::1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "wrong ip6",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeVlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "wrong gw6",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("192.168.0.1"),
			},
			wantErr: true,
		},
		{
			name: "missing ip4",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.IPRange{},
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.MustParseIPRange("2a02:1802:5e:0:1000:0:ff:1/64"),
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
		{
			name: "missing ip6",
			fields: fields{
				Master: "eth0",
				Type:   IfaceTypeMacvlan,
				Ipv4:   schema.MustParseIPRange("192.168.0.0/24"),
				Gw4:    net.ParseIP("192.168.0.1"),
				Ipv6:   schema.IPRange{},
				Gw6:    net.ParseIP("2a02:1802:5e::1"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PublicIface{
				Master:  tt.fields.Master,
				Type:    tt.fields.Type,
				Ipv4:    tt.fields.Ipv4,
				Ipv6:    tt.fields.Ipv6,
				Gw4:     tt.fields.Gw4,
				Gw6:     tt.fields.Gw6,
				Version: tt.fields.Version,
			}
			err := p.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
