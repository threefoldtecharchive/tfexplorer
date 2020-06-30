package client

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

func TestIntermediateWLWorkloadsUnknownType(t *testing.T) {
	for _, tc := range []struct {
		name string
		iwl  intermediateWL
		err  error
	}{
		{
			name: "container",
			iwl: intermediateWL{
				ReservationWorkload: workloads.ReservationWorkload{
					Type: workloads.WorkloadTypeContainer,
				},
				Content: func(i workloads.Container) []byte {
					b, err := json.Marshal(i)
					require.NoError(t, err)
					return b
				}(workloads.Container{ReservationInfo: workloads.ReservationInfo{WorkloadId: 1}}),
			},
			err: nil,
		},
		{
			name: "unknown",
			iwl: intermediateWL{
				ReservationWorkload: workloads.ReservationWorkload{
					Type: 12,
				},
			},
			err: errUnknownWorkload,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {

			_, err := tc.iwl.Workload()
			if tc.err != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, errUnknownWorkload))
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
