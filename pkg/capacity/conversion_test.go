package capacity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

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
			cu: 0,
			su: 0,
		},
		{
			rsu: workloads.RSU{
				CRU: 2,
				MRU: 4,
			},
			cu: 0.75,
			su: 0,
		},
		{
			rsu: workloads.RSU{
				CRU: 4,
				MRU: 8,
			},
			cu: 1.75,
			su: 0,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.rsu), func(t *testing.T) {
			cu, su := CloudUnitsFromResourceUnits(tt.rsu)
			assert.Equal(t, tt.cu, cu, "wrong number of cu")
			assert.Equal(t, tt.su, su, "wrong number of su")
		})
	}
}
