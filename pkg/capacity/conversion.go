package capacity

import (
	"math"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// CloudUnitsFromResourceUnits converts an amount of RSU to the Cloud unit representation
func CloudUnitsFromResourceUnits(rsu workloads.RSU) (cu float64, su float64, ipu float64) {
	cu = math.Round(math.Min((rsu.MRU)/4, float64(rsu.CRU)/2)*1000) / 1000
	su = math.Round((float64(rsu.HRU)/1200+float64(rsu.SRU)/300)*1000) / 1000
	ipu = rsu.IPV4U

	return
}
