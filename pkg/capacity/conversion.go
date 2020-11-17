package capacity

import (
	"math"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// CloudUnitsFromResourceUnits converts an amount of RSU to the Cloud unit representation
func CloudUnitsFromResourceUnits(rsu workloads.RSU) (cu float64, su float64, ipu float64) {
	cu = math.Min(float64(rsu.CRU)*2, (rsu.MRU-1)/4)
	su = (float64(rsu.HRU)/1000 + float64(rsu.SRU)/100/2) / 1.2
	ipu = rsu.IPV4U

	return
}
