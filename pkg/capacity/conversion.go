package capacity

import "math"

// CloudUnitsFromResourceUnits converts an amount of RSU to the Cloud unit representation
func CloudUnitsFromResourceUnits(cru int64, mru float64, hru float64, sru float64) (float64, float64) {
	cu := math.Min(float64(cru)*4, (mru-1)/4)
	su := hru/1000/1.2 + sru/100/1.2

	return cu, su
}
