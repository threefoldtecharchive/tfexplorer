package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	model "github.com/threefoldtech/tfexplorer/models/workloads"
)

func Test_countSignatures(t *testing.T) {
	type args struct {
		signatures []model.SigningSignature
		req        model.SigningRequest
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "included",
			args: args{
				signatures: []model.SigningSignature{
					{
						Tid: 1,
					},
				},
				req: model.SigningRequest{
					Signers: []int64{1},
				},
			},
			want: 1,
		},
		{
			name: "not_required",
			args: args{
				signatures: []model.SigningSignature{
					{
						Tid: 1,
					},
				},
				req: model.SigningRequest{
					Signers: []int64{2},
				},
			},
			want: 0,
		},
		{
			name: "multiple",
			args: args{
				signatures: []model.SigningSignature{
					{Tid: 1},
					{Tid: 2},
					{Tid: 3},
				},
				req: model.SigningRequest{
					Signers: []int64{1, 3},
				},
			},
			want: 2,
		},
		{
			name: "empty",
			args: args{
				signatures: []model.SigningSignature{},
				req: model.SigningRequest{
					Signers: []int64{1},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSignatures(tt.args.signatures, tt.args.req)
			assert.Equal(t, got, tt.want)
		})
	}
}
