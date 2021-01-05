package escrow

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	directorytypes "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	nodeAPIMock struct{}
)

func TestProcessReservation(t *testing.T) {
	data := workloads.ReservationData{
		Containers: []workloads.Container{
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 4096,
				},
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    1,
					Memory: 2096,
				},
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    3,
					Memory: 3096,
				},
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 5000,
				},
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 6589,
				},
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				Capacity: workloads.ContainerCapacity{
					Cpu:    2,
					Memory: 1234,
				},
			},
		},
		Volumes: []workloads.Volume{
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Type:            workloads.VolumeTypeHDD,
				Size:            500,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Type:            workloads.VolumeTypeHDD,
				Size:            500,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Type:            workloads.VolumeTypeSSD,
				Size:            100,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Type:            workloads.VolumeTypeHDD,
				Size:            2500,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				Type:            workloads.VolumeTypeHDD,
				Size:            1000,
			},
		},
		Zdbs: []workloads.ZDB{
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				DiskType:        workloads.DiskTypeSSD,
				Size:            750,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				DiskType:        workloads.DiskTypeSSD,
				Size:            250,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				DiskType:        workloads.DiskTypeHDD,
				Size:            500,
			},
		},
		Kubernetes: []workloads.K8S{
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Size:            1,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Size:            2,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "1"},
				Size:            2,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Size:            2,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "2"},
				Size:            2,
			},
			{
				ReservationInfo: workloads.ReservationInfo{NodeId: "3"},
				Size:            2,
			},
		},
	}

	escrow := Stellar{
		//wallet:             &stellar.StellarWallet{},
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

func TestResourceUnitsToCloudUnits(t *testing.T) {
	tests := []struct {
		rsu rsu
		cu  cloudUnits
	}{
		{
			rsu: rsu{},
			cu:  cloudUnits{},
		},
		{
			rsu: rsu{
				cru: 2,
				mru: 4,
				hru: 0,
				sru: 0,
			},
			cu: cloudUnits{
				cu: 1,
				su: 0,
			},
		},
		{
			rsu: rsu{
				cru: 0,
				mru: 0,
				hru: 1200,
				sru: 0,
			},
			cu: cloudUnits{
				cu: 0,
				su: 1,
			},
		},
		{
			rsu: rsu{
				cru: 0,
				mru: 0,
				hru: 0,
				sru: 300,
			},
			cu: cloudUnits{
				cu: 0,
				su: 1,
			},
		},
		{
			rsu: rsu{
				cru: 2,
				mru: 4,
				hru: 1000,
				sru: 100,
			},
			cu: cloudUnits{
				cu: 1,
				su: 1.167,
			},
		},
		{
			rsu: rsu{
				cru: 1,
				mru: 2,
				hru: 600,
				sru: 0},
			cu: cloudUnits{
				cu: 0.5,
				su: 0.5,
			},
		},
		{
			rsu: rsu{
				cru: 1,
				mru: 8,
				hru: 1024,
				sru: 256,
			},
			cu: cloudUnits{
				cu: 2,
				su: 1.707,
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			assert.Equal(t, test.cu, rsuToCu(test.rsu))
		})
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

func Test_rsuToCu(t *testing.T) {
	type args struct {
		r rsu
	}
	tests := []struct {
		name string
		args args
		want cloudUnits
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rsuToCu(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rsuToCu() = %v, want %v", got, tt.want)
			}
		})
	}
}
