package escrow

import (
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

type (
	rsuPerFarmer map[int64]rsu

	rsuPerNode map[string]rsu

	rsu struct {
		cru int64
		sru int64
		hru int64
		mru float64
	}

	cloudUnits struct {
		cu float64
		su float64
	}
)

// Price for 1 CU or SU for 1 month is 10$
// TFT price is fixed at 0.05
// CU and SU price per second is
// CU -> (10 / 0.15) / (3600 *24*30) = 257.2
// SU -> (8 / 0.15) / (3600 *24*30) = 257.2
const (
	// CuPriceDollarMonth CU price per month in dollar
	CuPriceDollarMonth = 10
	// SuPriceDollarMonth SU price per month in dollar
	SuPriceDollarMonth = 8
	// IP4uPriceDollarMonth IPv4U price per month in dollar
	IP4uPriceDollarMonth = 6

	// TftPriceMill tft price in millies
	TftPriceMill = 100 // 0.01 * 1000 (1mill = 1/1000 of a dollar)

	// express as stropes, to simplify things a bit
	// TODO: check if the rounding errors here matter
	computeUnitSecondTFTStropesCost = (CuPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30)
	storageUnitSecondTFTStropesCost = (SuPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30)
	ipv4UnitSecondTFTStropesCost    = (IP4uPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30)
)

const (
	// durations
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
)

func getDiscount(d time.Duration) float64 {

	switch {
	case d >= 12*month:
		return 1 - 0.70
	case d >= 6*month:
		return 1 - 0.60
	case d >= month:
		return 1 - 0.50
	case d >= week:
		return 1 - 0.25
	default:
		return 1
	}

}

func getComputeUnitSecondTFTStropesCost(cuPriceDollarMonth float64) int64 {
	return int64((cuPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30))
}

func getStorageUnitSecondTFTStropesCost(suPriceDollarMonth float64) int64 {
	return int64((suPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30))
}

func getIPv4UnitSecondTFTStropesCost(ip4uPriceDollarMonth float64) int64 {
	return int64((ip4uPriceDollarMonth * 10_000_000_000 / TftPriceMill) / (3600 * 24 * 30))
}

// calculateCapacityReservationCost calculates the cost of a capacity reservation
func (e Stellar) calculateCustomCapacityReservationCost(CUs, SUs, IPv4Us uint64, cuDollarPerMonth, suDollarPerMonth, ip4uDollarPerMonth float64) (xdr.Int64, error) {
	total := big.NewInt(0)
	cuCost := big.NewInt(0)
	suCost := big.NewInt(0)
	ipuCost := big.NewInt(0)

	cuSecondTFTStropesCost := getComputeUnitSecondTFTStropesCost(cuDollarPerMonth)
	suSecondTFTStropesCost := getStorageUnitSecondTFTStropesCost(suDollarPerMonth)
	ip4uSecondTFTStropesCost := getIPv4UnitSecondTFTStropesCost(ip4uDollarPerMonth)

	cuCost = cuCost.Mul(big.NewInt(cuSecondTFTStropesCost), big.NewInt(int64(CUs)))
	suCost = suCost.Mul(big.NewInt(suSecondTFTStropesCost), big.NewInt(int64(SUs)))
	ipuCost = ipuCost.Mul(big.NewInt(ip4uSecondTFTStropesCost), big.NewInt(int64(IPv4Us)))
	// TODO: Discount??
	total = total.Add(total.Add(cuCost, suCost), ipuCost)

	total = total.Div(total, big.NewInt(e.getNetworkDivisor()))

	return xdr.Int64(total.Int64()), nil
}

// calculateCapacityReservationCost calculates the cost of a capacity reservation
func (e Stellar) calculateCapacityReservationCost(CUs, SUs, IPv4Us uint64) (xdr.Int64, error) {
	total := big.NewInt(0)
	cuCost := big.NewInt(0)
	suCost := big.NewInt(0)
	ipuCost := big.NewInt(0)

	cuCost = cuCost.Mul(big.NewInt(computeUnitSecondTFTStropesCost), big.NewInt(int64(CUs)))
	suCost = suCost.Mul(big.NewInt(storageUnitSecondTFTStropesCost), big.NewInt(int64(SUs)))
	ipuCost = ipuCost.Mul(big.NewInt(ipv4UnitSecondTFTStropesCost), big.NewInt(int64(IPv4Us)))
	// TODO: Discount??
	total = total.Add(total.Add(cuCost, suCost), ipuCost)

	total = total.Div(total, big.NewInt(e.getNetworkDivisor()))

	return xdr.Int64(total.Int64()), nil
}

func (e Stellar) processReservationResources(resData workloads.ReservationData) (rsuPerFarmer, error) {
	rsuPerNodeMap := make(rsuPerNode)
	for _, cont := range resData.Containers {
		rsuPerNodeMap[cont.NodeId] = rsuPerNodeMap[cont.NodeId].add(processContainer(cont))
	}
	for _, vol := range resData.Volumes {
		rsuPerNodeMap[vol.NodeId] = rsuPerNodeMap[vol.NodeId].add(processVolume(vol))
	}
	for _, zdb := range resData.Zdbs {
		rsuPerNodeMap[zdb.NodeId] = rsuPerNodeMap[zdb.NodeId].add(processZdb(zdb))
	}
	for _, k8s := range resData.Kubernetes {
		rsuPerNodeMap[k8s.NodeId] = rsuPerNodeMap[k8s.NodeId].add(processKubernetes(k8s))
	}
	rsuPerFarmerMap := make(rsuPerFarmer)
	for nodeID, rsu := range rsuPerNodeMap {
		node, err := e.nodeAPI.Get(e.ctx, e.db, nodeID, false)
		if err != nil {
			return nil, errors.Wrap(err, "could not get node")
		}
		rsuPerFarmerMap[node.FarmId] = rsuPerFarmerMap[node.FarmId].add(rsu)
	}
	return rsuPerFarmerMap, nil
}

func processContainer(cont workloads.Container) rsu {
	rsu := rsu{
		cru: cont.Capacity.Cpu,
		// round mru to 4 digits precision
		mru: math.Round(float64(cont.Capacity.Memory)/1024*10000) / 10000,
	}
	switch cont.Capacity.DiskType {
	case workloads.DiskTypeHDD:
		hru := math.Round(float64(cont.Capacity.DiskSize)/1024*10000) / 10000
		rsu.hru = int64(hru)
	case workloads.DiskTypeSSD:
		sru := math.Round(float64(cont.Capacity.DiskSize)/1024*10000) / 10000
		rsu.sru = int64(sru)
	}

	return rsu
}

func processVolume(vol workloads.Volume) rsu {
	switch vol.Type {
	case workloads.VolumeTypeHDD:
		return rsu{
			hru: vol.Size,
		}
	case workloads.VolumeTypeSSD:
		return rsu{
			sru: vol.Size,
		}
	}
	return rsu{}
}

func processZdb(zdb workloads.ZDB) rsu {
	switch zdb.DiskType {
	case workloads.DiskTypeHDD:
		return rsu{
			hru: zdb.Size,
		}
	case workloads.DiskTypeSSD:
		return rsu{
			sru: zdb.Size,
		}
	}
	return rsu{}

}

func processKubernetes(k8s workloads.K8S) rsu {
	switch k8s.Size {
	case 1:
		return rsu{
			cru: 1,
			mru: 2,
			sru: 50,
		}
	case 2:
		return rsu{
			cru: 2,
			mru: 4,
			sru: 100,
		}
	}
	return rsu{}

}

func (r rsu) add(other rsu) rsu {
	return rsu{
		cru: r.cru + other.cru,
		sru: r.sru + other.sru,
		hru: r.hru + other.hru,
		mru: r.mru + other.mru,
	}
}

// rsuToCu converts resource units to cloud units. Cloud units are rounded to 3 decimal places
// Values taken from https://github.com/threefoldfoundation/info_foundation/commit/517855d2b6dd673567f8494462326ca0d24fb818#diff-9daca117ae814b405e7ecde62e8c83ac5108bd64fe4ed6de0eb592724fb0421aR41
func rsuToCu(r rsu) cloudUnits {
	cloudUnits := cloudUnits{
		cu: math.Round(math.Min(r.mru/4, float64(r.cru)*2)*1000) / 1000,
		su: math.Round((float64(r.hru)/1200+float64(r.sru)/300)*1000) / 1000,
	}
	return cloudUnits
}
