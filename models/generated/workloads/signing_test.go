package workloads

import (
	"crypto/sha256"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/crypto"
	"github.com/threefoldtech/zos/pkg/identity"
)

func TestVolumeSigningChalenge(t *testing.T) {
	kp, err := identity.GenerateKeyPair()
	require.NoError(t, err)

	v := &Volume{
		ReservationInfo: ReservationInfo{
			CustomerTid:            1,
			ID:                     1,
			WorkloadId:             1,
			PoolId:                 1,
			Description:            "this is a volume",
			Metadata:               "this is metadata",
			Epoch:                  schema.Date{Time: time.Now()},
			WorkloadType:           WorkloadTypeVolume,
			NodeId:                 "node1",
			ExpirationProvisioning: schema.Date{Time: time.Now().Add(time.Minute * 10)},
		},
		Size: 1,
		Type: VolumeTypeSSD,
	}
	sc, err := v.SignatureChallenge()
	require.NoError(t, err)

	msg := sha256.Sum256(sc)

	signature, err := crypto.Sign(kp.PrivateKey, msg[:])
	require.NoError(t, err)

	err = crypto.Verify(kp.PublicKey, msg[:], signature)
	assert.NoError(t, err)

	v.NodeId = "node2"
	sc, err = v.SignatureChallenge()
	require.NoError(t, err)

	msg = sha256.Sum256(sc)

	err = crypto.Verify(kp.PublicKey, sc, signature)
	assert.Error(t, err)
}

func TestContainerSigningChalenge(t *testing.T) {
	kp, err := identity.GenerateKeyPair()
	require.NoError(t, err)

	v := &Container{
		ReservationInfo: ReservationInfo{
			CustomerTid:            1,
			ID:                     1,
			WorkloadId:             1,
			PoolId:                 1,
			Description:            "this is a volume",
			Metadata:               "this is metadata",
			Epoch:                  schema.Date{Time: time.Now()},
			WorkloadType:           WorkloadTypeVolume,
			NodeId:                 "node1",
			ExpirationProvisioning: schema.Date{Time: time.Now().Add(time.Minute * 10)},
		},
		Flist:      "https://flist.com",
		Entrypoint: "/sbin/my_init",
		Capacity: ContainerCapacity{
			Cpu:    2,
			Memory: 1024,
		},
		Environment: map[string]string{
			"hello": "world",
		},
		Interactive: false,
		NetworkConnection: []NetworkConnection{
			{
				NetworkId: "network1",
				Ipaddress: net.ParseIP("192.1268.0.1"),
				PublicIp6: true,
			},
		},
		Volumes: []ContainerMount{
			{
				Mountpoint: "/data",
				VolumeId:   "12-3",
			},
		},
	}
	sc, err := v.SignatureChallenge()
	require.NoError(t, err)

	msg := sha256.Sum256(sc)
	signature, err := crypto.Sign(kp.PrivateKey, msg[:])
	require.NoError(t, err)

	fmt.Printf("pubkey: %x\n", kp.PublicKey)
	fmt.Printf("signarure: %x\n", signature)
	err = crypto.Verify(kp.PublicKey, msg[:], signature)
	assert.NoError(t, err)

	v.NodeId = "node2"
	sc, err = v.SignatureChallenge()
	require.NoError(t, err)
	msg = sha256.Sum256(sc)

	err = crypto.Verify(kp.PublicKey, msg[:], signature)
	assert.Error(t, err)
}
