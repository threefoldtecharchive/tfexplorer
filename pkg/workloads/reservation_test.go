package workloads

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

func Test_userCanSign(t *testing.T) {
	type args struct {
		userTid    int64
		req        workloads.SigningRequest
		signatures []workloads.SigningSignature
	}
	tests := []struct {
		name string
		args args
		err  bool
	}{
		{
			name: "user_required_no_signature",
			args: args{
				userTid: 1,
				req: workloads.SigningRequest{
					QuorumMin: 1,
					Signers:   []int64{1},
				},
				signatures: []workloads.SigningSignature{},
			},
			err: false,
		},
		{
			name: "user_required_already_signed",
			args: args{
				userTid: 1,
				req: workloads.SigningRequest{
					QuorumMin: 1,
					Signers:   []int64{1},
				},
				signatures: []workloads.SigningSignature{
					{
						Tid:       2,
						Signature: "foobar",
					},
					{
						Tid:       1,
						Signature: "foobar",
					},
				},
			},
			err: true,
		},
		{
			name: "user_not_required",
			args: args{
				userTid: 1,
				req: workloads.SigningRequest{
					QuorumMin: 1,
					Signers:   []int64{2},
				},
				signatures: []workloads.SigningSignature{},
			},
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userCanSign(tt.args.userTid, tt.args.req, tt.args.signatures)
			assert.Equal(t, tt.err, err != nil)
		})
	}
}
