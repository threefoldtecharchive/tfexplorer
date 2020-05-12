package escrow

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	directorytypes "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	nodeAPIMock struct{}
)

const precision = 1e7

func TestProcessReservation(t *testing.T) {
	data := workloads.ReservationData{
		Containers: []workloads.Container{
			{
				NodeId: "1",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 4096,
				},
			},
			{
				NodeId: "1",
				Capacity: workloads.ContainerCapacity{
					Cpu:    1,
					Memory: 2096,
				},
			},
			{
				NodeId: "2",
				Capacity: workloads.ContainerCapacity{
					Cpu:    3,
					Memory: 3096,
				},
			},
			{
				NodeId: "2",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 5000,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 6589,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 1234,
				},
			},
		},
		Volumes: []workloads.Volume{
			{
				NodeId: "1",
				Type:   workloads.VolumeTypeHDD,
				Size:   500,
			},
			{
				NodeId: "1",
				Type:   workloads.VolumeTypeHDD,
				Size:   500,
			},
			{
				NodeId: "2",
				Type:   workloads.VolumeTypeSSD,
				Size:   100,
			},
			{
				NodeId: "2",
				Type:   workloads.VolumeTypeHDD,
				Size:   2500,
			},
			{
				NodeId: "3",
				Type:   workloads.VolumeTypeHDD,
				Size:   1000,
			},
		},
		Zdbs: []workloads.ZDB{
			{
				NodeId:   "1",
				DiskType: workloads.DiskTypeSSD,
				Size:     750,
			},
			{
				NodeId:   "3",
				DiskType: workloads.DiskTypeSSD,
				Size:     250,
			},
			{
				NodeId:   "3",
				DiskType: workloads.DiskTypeHDD,
				Size:     500,
			},
		},
		Kubernetes: []workloads.K8S{
			{
				NodeId: "1",
				Size:   1,
			},
			{
				NodeId: "1",
				Size:   2,
			},
			{
				NodeId: "1",
				Size:   2,
			},
			{
				NodeId: "2",
				Size:   2,
			},
			{
				NodeId: "2",
				Size:   2,
			},
			{
				NodeId: "3",
				Size:   2,
			},
		},
	}

	escrow := Stellar{
		wallet:             &stellar.Wallet{},
		db:                 nil,
		reservationChannel: nil,
		nodeAPI:            &nodeAPIMock{},
	}

	farmRsu, err := escrow.processReservationResources(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(farmRsu) != 3 {
		t.Errorf("Found %d farmers, expected to find 3", len(farmRsu))
	}

	// check farm tid 1
	rsu := farmRsu[1]
	if rsu.cru != 8 {
		t.Errorf("Farmer 1 total cru is %d, expected 8", rsu.cru)
	}
	if rsu.mru != 16.0469 {
		t.Errorf("Farmer 1 total mru is %f, expected 10", rsu.mru)
	}
	if rsu.sru != 1000 {
		t.Errorf("Farmer 1 total sru is %d, expected 1000", rsu.sru)
	}
	if rsu.hru != 1000 {
		t.Errorf("Farmer 1 total hru is %d, expected 1000", rsu.hru)
	}

	// check farm tid 2
	rsu = farmRsu[2]
	if rsu.cru != 9 {
		t.Errorf("Farmer 2 total cru is %d, expected 9", rsu.cru)
	}
	if rsu.mru != 15.9062 {
		t.Errorf("Farmer 2 total mru is %f, expected 8", rsu.mru)
	}
	if rsu.sru != 300 {
		t.Errorf("Farmer 2 total sru is %d, expected 300", rsu.sru)
	}
	if rsu.hru != 2500 {
		t.Errorf("Farmer 2 total hru is %d, expected 2500", rsu.hru)
	}

	// check farm tid 3
	rsu = farmRsu[3]
	if rsu.cru != 6 {
		t.Errorf("Farmer 3 total cru is %d, expected 6", rsu.cru)
	}
	if rsu.mru != 11.6397 {
		t.Errorf("Farmer 3 total mru is %f, expected 4", rsu.mru)
	}
	if rsu.sru != 350 {
		t.Errorf("Farmer 3 total sru is %d, expected 350", rsu.sru)
	}
	if rsu.hru != 1500 {
		t.Errorf("Farmer 3 total hru is %d, expected 1500", rsu.hru)
	}
}

func TestProcessContainer(t *testing.T) {
	cont := workloads.Container{
		Capacity: workloads.ContainerCapacity{
			Cpu:    1,
			Memory: 1024,
		},
	}
	rsu := processContainer(cont)

	if rsu.cru != 1 {
		t.Errorf("Processed volume cru is %d, expected 1", rsu.cru)
	}
	if rsu.mru != 1 {
		t.Errorf("Processed volume mru is %f, expected 1", rsu.mru)
	}
	if rsu.sru != 0 {
		t.Errorf("Processed volume sru is %d, expected 0", rsu.sru)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed volume hru is %d, expected 0", rsu.hru)
	}

	cont = workloads.Container{
		Capacity: workloads.ContainerCapacity{
			Cpu:    4,
			Memory: 2024,
		},
	}
	rsu = processContainer(cont)

	if rsu.cru != 4 {
		t.Errorf("Processed volume cru is %d, expected 1", rsu.cru)
	}
	if rsu.mru != 1.9766 {
		t.Errorf("Processed volume mru is %f, expected 1", rsu.mru)
	}
	if rsu.sru != 0 {
		t.Errorf("Processed volume sru is %d, expected 0", rsu.sru)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed volume hru is %d, expected 0", rsu.hru)
	}
}

func TestProcessVolume(t *testing.T) {
	testSize := int64(27) // can be random number

	vol := workloads.Volume{
		Size: testSize,
		Type: workloads.VolumeTypeHDD,
	}
	rsu := processVolume(vol)

	if rsu.cru != 0 {
		t.Errorf("Processed volume cru is %d, expected 0", rsu.cru)
	}
	if rsu.mru != 0 {
		t.Errorf("Processed volume mru is %f, expected 0", rsu.mru)
	}
	if rsu.sru != 0 {
		t.Errorf("Processed volume sru is %d, expected 0", rsu.sru)
	}
	if rsu.hru != testSize {
		t.Errorf("Processed volume hru is %d, expected %d", rsu.hru, testSize)
	}

	vol = workloads.Volume{
		Size: testSize,
		Type: workloads.VolumeTypeSSD,
	}
	rsu = processVolume(vol)

	if rsu.cru != 0 {
		t.Errorf("Processed volume cru is %d, expected 0", rsu.cru)
	}
	if rsu.mru != 0 {
		t.Errorf("Processed volume mru is %f, expected 0", rsu.mru)
	}
	if rsu.sru != testSize {
		t.Errorf("Processed volume sru is %d, expected %d", rsu.sru, testSize)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed volume hru is %d, expected 0", rsu.hru)
	}
}

func TestProcessZdb(t *testing.T) {
	testSize := int64(43) // can be random number

	zdb := workloads.ZDB{
		DiskType: workloads.DiskTypeHDD,
		Size:     testSize,
	}
	rsu := processZdb(zdb)

	if rsu.cru != 0 {
		t.Errorf("Processed zdb cru is %d, expected 0", rsu.cru)
	}
	if rsu.mru != 0 {
		t.Errorf("Processed zdb mru is %f, expected 0", rsu.mru)
	}
	if rsu.sru != 0 {
		t.Errorf("Processed zdb sru is %d, expected 0", rsu.sru)
	}
	if rsu.hru != testSize {
		t.Errorf("Processed zdb hru is %d, expected %d", rsu.hru, testSize)
	}

	zdb = workloads.ZDB{
		DiskType: workloads.DiskTypeSSD,
		Size:     testSize,
	}
	rsu = processZdb(zdb)

	if rsu.cru != 0 {
		t.Errorf("Processed zdb cru is %d, expected 0", rsu.cru)
	}
	if rsu.mru != 0 {
		t.Errorf("Processed zdb mru is %f, expected 0", rsu.mru)
	}
	if rsu.sru != testSize {
		t.Errorf("Processed zdb sru is %d, expected %d", rsu.sru, testSize)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed zdb hru is %d, expected 0", rsu.hru)
	}
}

func TestProcessKubernetes(t *testing.T) {
	k8s := workloads.K8S{
		Size: 1,
	}
	rsu := processKubernetes(k8s)

	if rsu.cru != 1 {
		t.Errorf("Processed zdb cru is %d, expected 1", rsu.cru)
	}
	if rsu.mru != 2 {
		t.Errorf("Processed zdb mru is %f, expected 2", rsu.mru)
	}
	if rsu.sru != 50 {
		t.Errorf("Processed zdb sru is %d, expected 50", rsu.sru)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed zdb hru is %d, expected 0", rsu.hru)
	}

	k8s = workloads.K8S{
		Size: 2,
	}
	rsu = processKubernetes(k8s)

	if rsu.cru != 2 {
		t.Errorf("Processed zdb cru is %d, expected 2", rsu.cru)
	}
	if rsu.mru != 4 {
		t.Errorf("Processed zdb mru is %f, expected 4", rsu.mru)
	}
	if rsu.sru != 100 {
		t.Errorf("Processed zdb sru is %d, expected 100", rsu.sru)
	}
	if rsu.hru != 0 {
		t.Errorf("Processed zdb hru is %d, expected 0", rsu.hru)
	}
}

func TestRsuAdd(t *testing.T) {
	first := rsu{cru: 1, mru: 2, sru: 3, hru: 4}
	second := rsu{cru: 8, mru: 6, sru: 4, hru: 2}
	result := first.add(second)

	if result.cru != 9 {
		t.Errorf("Result cru is %d, expected 9", result.cru)
	}
	if result.mru != 8 {
		t.Errorf("Result mru is %f, expected 8", result.mru)
	}
	if result.sru != 7 {
		t.Errorf("Result sru is %d, expected 7", result.sru)
	}
	if result.hru != 6 {
		t.Errorf("Result hru is %d, expected 6", result.hru)
	}
}

func TestCalculateReservationCost(t *testing.T) {
	data := workloads.ReservationData{
		Containers: []workloads.Container{
			{
				NodeId: "1",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 2048,
				},
			},
			{
				NodeId: "1",
				Capacity: workloads.ContainerCapacity{
					Cpu:    4,
					Memory: 5120,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 1000,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    4,
					Memory: 4000,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    4,
					Memory: 4096,
				},
			},
			{
				NodeId: "3",
				Capacity: workloads.ContainerCapacity{
					Cpu:    1,
					Memory: 1024,
				},
			},
		},
		Volumes: []workloads.Volume{
			{
				NodeId: "1",
				Type:   workloads.VolumeTypeHDD,
				Size:   500,
			},
			{
				NodeId: "1",
				Type:   workloads.VolumeTypeHDD,
				Size:   500,
			},
			{
				NodeId: "3",
				Type:   workloads.VolumeTypeSSD,
				Size:   100,
			},
			{
				NodeId: "3",
				Type:   workloads.VolumeTypeHDD,
				Size:   2500,
			},
			{
				NodeId: "3",
				Type:   workloads.VolumeTypeHDD,
				Size:   1000,
			},
		},
		Zdbs: []workloads.ZDB{
			{
				NodeId:   "1",
				DiskType: workloads.DiskTypeSSD,
				Size:     750,
			},
			{
				NodeId:   "3",
				DiskType: workloads.DiskTypeSSD,
				Size:     250,
			},
			{
				NodeId:   "3",
				DiskType: workloads.DiskTypeHDD,
				Size:     500,
			},
		},
		Kubernetes: []workloads.K8S{
			{
				NodeId: "1",
				Size:   1,
			},
			{
				NodeId: "1",
				Size:   2,
			},
			{
				NodeId: "1",
				Size:   2,
			},
			{
				NodeId: "3",
				Size:   2,
			},
			{
				NodeId: "3",
				Size:   2,
			},
			{
				NodeId: "3",
				Size:   2,
			},
		},
	}

	escrow := Stellar{
		wallet:             &stellar.Wallet{},
		db:                 nil,
		reservationChannel: nil,
		nodeAPI:            &nodeAPIMock{},
	}

	farmRsu, err := escrow.processReservationResources(data)
	assert.NoError(t, err)

	t.Run("1_hour", func(t *testing.T) {
		duration := time.Hour
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * 1
		assert.Equal(t, xdr.Int64(3.454642*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * 1
		assert.Equal(t, xdr.Int64(3.543002*precision), res[3])
	})

	t.Run("1_month", func(t *testing.T) {
		duration := month
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * 720 * (1 - 0.50)
		assert.Equal(t, xdr.Int64(1243.67112*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * 720 * (1 - 0.50)
		assert.Equal(t, xdr.Int64(1275.48072*precision), res[3])
	})

	t.Run("2_month", func(t *testing.T) {
		duration := 2 * month
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * (720*2) * (1-0.50)
		assert.Equal(t, xdr.Int64(2487.34224*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * (720*2) * (1-0.50)
		assert.Equal(t, xdr.Int64(2550.96144*precision), res[3])
	})

	t.Run("17_days", func(t *testing.T) {
		duration := 17 * day
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * (17*24) * (1-0.25)
		assert.Equal(t, xdr.Int64(1057.120452*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * (17*24) * (1-0.25)
		assert.Equal(t, xdr.Int64(1084.158612*precision), res[3])
	})

	t.Run("week", func(t *testing.T) {
		duration := week
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * (7*24) * (1-0.25)
		assert.Equal(t, xdr.Int64(435.284892*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * (7*24) * (1-0.25)
		assert.Equal(t, xdr.Int64(446.418252*precision), res[3])
	})

	t.Run("6_month", func(t *testing.T) {
		duration := 6 * month
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * (6*30*24) * (1-0.60)
		assert.Equal(t, xdr.Int64(5969.621376*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * (6*30*24) * (1-0.60)
		assert.Equal(t, xdr.Int64(6122.307456*precision), res[3])
	})

	t.Run("12_month", func(t *testing.T) {
		duration := 12 * month
		res, err := escrow.calculateReservationCost(farmRsu, duration)
		if ok := assert.NoError(t, err); !ok {
			t.Fatal()
		}

		assert.True(t, len(res) == 2)
		// cru: 11, sru: 1000, hru: 1000, mru: 17
		// (4.037 * 0.266 + 11.904 * 0.200) * (12*30*24) * (1-0.70)
		assert.Equal(t, xdr.Int64(8954.432064*precision), res[1])
		// cru: 17, sru: 650, hru: 4000, mru: 21.8829
		// (5.197 * 0.266 + 10.803 * 0.200) * (12*30*24) * (1-0.70)
		assert.Equal(t, xdr.Int64(9183.461184*precision), res[3])
	})

}

func TestResourceUnitsToCloudUnits(t *testing.T) {
	rsus := []rsu{
		{},
		{
			cru: 2,
			mru: 4,
			hru: 1093,
			sru: 91,
		},
		{
			cru: 1,
			mru: 2,
			hru: 1093,
			sru: 0,
		},
		{
			cru: 1,
			mru: 8,
			hru: 0,
			sru: 91,
		},
		{
			cru: 1,
			mru: 12,
			hru: 1000,
			sru: 250,
		},
	}

	expectedCUs := []cloudUnits{
		{},
		{
			cu: 0.95,
			su: 2,
		},
		{
			cu: 0.475,
			su: 1,
		},
		{
			cu: 1.9,
			su: 1,
		},
		{
			cu: 2,
			su: 3.662,
		},
	}

	for i := range rsus {
		assert.Equal(t, rsuToCu(rsus[i]), expectedCUs[i])
	}
}

func (napim *nodeAPIMock) Get(_ context.Context, _ *mongo.Database, nodeID string, _ bool) (directorytypes.Node, error) {
	idInt, err := strconv.Atoi(nodeID)
	if err != nil {
		return directorytypes.Node{}, errors.New("node not found")
	}
	return directorytypes.Node{
		ID:     schema.ID(idInt),
		NodeId: nodeID,
		FarmId: int64(idInt),
	}, nil
}

func Test_getDiscount(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want float64
	}{
		{
			d:    time.Hour,
			want: 1 - 0,
		},
		{
			d:    5 * time.Hour,
			want: 1 - 0,
		},
		{
			d:    day,
			want: 1 - 0,
		},
		{
			d:    week,
			want: 1 - 0.25,
		},
		{
			d:    week + 2*day,
			want: 1 - 0.25,
		},
		{
			d:    month,
			want: 1 - 0.50,
		},
		{
			d:    2 * month,
			want: 1 - 0.50,
		},
		{
			d:    6 * month,
			want: 1 - 0.60,
		},
		{
			d:    12 * month,
			want: 1 - 0.70,
		},
		{
			d:    2*month + 2*week,
			want: 1 - 0.50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.d.String(), func(t *testing.T) {
			discount := getDiscount(tt.d)
			assert.Equal(t, tt.want, discount)
		})
	}
}
