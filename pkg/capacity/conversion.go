package capacity

import (
	"math"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// CloudUnitsFromResourceUnits converts an amount of RSU to the Cloud unit representation
func CloudUnitsFromResourceUnits(rsu workloads.RSU) (float64, float64) {
	cu := math.Min(float64(rsu.CRU)*4, (rsu.MRU-1)/4)
	su := float64(rsu.HRU)/1000/1.2 + float64(rsu.SRU)/100/1.2

	return cu, su
}
