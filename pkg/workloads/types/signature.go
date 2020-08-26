package types

import (
	generated "github.com/threefoldtech/tfexplorer/models/workloads"
)

func countSignatures(signatures []generated.SigningSignature, req generated.SigningRequest) int {
	in := func(i int64, l []int64) bool {
		for _, x := range l {
			if x == i {
				return true
			}
		}
		return false
	}

	signers := map[int64]interface{}{}
	for _, signature := range signatures {
		if !in(signature.Tid, req.Signers) {
			continue
		}
		_, exists := signers[signature.Tid]
		if !exists {
			signers[signature.Tid] = struct{}{}
		}
	}

	return len(signers)
}
