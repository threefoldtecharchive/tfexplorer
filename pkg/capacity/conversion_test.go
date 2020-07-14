package capacity

import (
	"fmt"
	"testing"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

func TestCloudUnitsFromResourceUnits(t *testing.T) {
	tests := []struct {
		rsu workloads.RSU
		cu  float64
		su  float64
	}{
		{
			rsu: workloads.RSU{
				CRU: 1,
				MRU: 1,
			},
			cu: 1,
			su: 0,
		},
		{
			rsu: workloads.RSU{
				CRU: 2,
				MRU: 4,
			},
			cu: 1,
			su: 0,
		},
		{
			rsu: workloads.RSU{
				CRU: 4,
				MRU: 8,
			},
			cu: 1,
			su: 0,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.rsu), func(t *testing.T) {
			got, got1 := CloudUnitsFromResourceUnits(tt.rsu)
			if got != tt.cu {
				t.Errorf("CloudUnitsFromResourceUnits() got = %v, want %v", got, tt.cu)
			}
			if got1 != tt.su {
				t.Errorf("CloudUnitsFromResourceUnits() got1 = %v, want %v", got1, tt.su)
			}
		})
	}
}
