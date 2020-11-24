package capacity

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/pkg/capacity/types"
)

func Test_usesExpiredResources(t *testing.T) {
	type args struct {
		pool     types.Pool
		workload workloads.Workloader
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "filled pool container",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.Container{
					Capacity: workloads.ContainerCapacity{
						Cpu:      2,
						Memory:   2048,
						DiskSize: 5 * 1024,
						DiskType: workloads.DiskTypeSSD,
					},
				},
			},
			want: false,
		},
		{
			name: "filled pool volume",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.Volume{
					Size: 2048,
					Type: workloads.VolumeTypeSSD,
				},
			},
			want: false,
		},
		{
			name: "filled pool network",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.NetworkResource{},
			},
			want: false,
		},
		{
			name: "filled pool k8s",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.K8S{
					Size: 1,
				},
			},
			want: false,
		},
		{
			name: "filled pool zdb",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.ZDB{
					Size:     2048,
					DiskType: workloads.DiskTypeSSD,
				},
			},
			want: false,
		},
		{
			name: "filled pool gateway 4 to 6",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.Gateway4To6{},
			},
			want: false,
		},
		{
			name: "filled pool gateway delegate",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.GatewayDelegate{},
			},
			want: false,
		},
		{
			name: "filled pool gateway proxy",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.GatewayProxy{},
			},
			want: false,
		},
		{
			name: "filled pool gateway reverse proxy",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.GatewayReverseProxy{},
			},
			want: false,
		},
		{
			name: "filled pool gateway subdmain",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 10_000,
				},
				workload: &workloads.GatewaySubdomain{},
			},
			want: false,
		},
		{
			name: "empty pool container",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.Container{
					Capacity: workloads.ContainerCapacity{
						Cpu:      2,
						Memory:   2048,
						DiskSize: 5 * 1024,
						DiskType: workloads.DiskTypeSSD,
					},
				},
			},
			want: true,
		},
		{
			name: "empty pool volume",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.Volume{
					Size: 2048,
					Type: workloads.VolumeTypeSSD,
				},
			},
			want: true,
		},
		{
			name: "empty pool network",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.NetworkResource{},
			},
			want: false,
		},
		{
			name: "empty pool k8s",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.K8S{
					Size: 1,
				},
			},
			want: true,
		},
		{
			name: "empty pool zdb",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.ZDB{
					Size:     2048,
					DiskType: workloads.DiskTypeSSD,
				},
			},
			want: true,
		},
		{
			name: "empty pool gateway 4 to 6",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.Gateway4To6{},
			},
			want: false,
		},
		{
			name: "empty pool gateway delegate",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.GatewayDelegate{},
			},
			want: false,
		},
		{
			name: "empty pool gateway proxy",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.GatewayProxy{},
			},
			want: false,
		},
		{
			name: "empty pool gateway reverse proxy",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.GatewayReverseProxy{},
			},
			want: false,
		},
		{
			name: "empty pool gateway subdmain",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 0,
				},
				workload: &workloads.GatewaySubdomain{},
			},
			want: false,
		},
		{
			name: "pool empty CU container",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.Container{
					Capacity: workloads.ContainerCapacity{
						Cpu:      2,
						Memory:   2048,
						DiskSize: 5 * 1024,
						DiskType: workloads.DiskTypeSSD,
					},
				},
			},
			want: true,
		},
		{
			name: "pool empty CU volume",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.Volume{
					Size: 2048,
					Type: workloads.VolumeTypeSSD,
				},
			},
			want: false,
		},
		{
			name: "pool empty CU network",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.NetworkResource{},
			},
			want: false,
		},
		{
			name: "pool empty CU k8s",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.K8S{
					Size: 1,
				},
			},
			want: true,
		},
		{
			name: "pool empty CU zdb",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.ZDB{
					Size:     2048,
					DiskType: workloads.DiskTypeSSD,
				},
			},
			want: false,
		},
		{
			name: "pool empty CU gateway 4 to 6",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.Gateway4To6{},
			},
			want: false,
		},
		{
			name: "pool empty CU gateway delegate",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.GatewayDelegate{},
			},
			want: false,
		},
		{
			name: "pool empty CU gateway proxy",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.GatewayProxy{},
			},
			want: false,
		},
		{
			name: "pool empty CU gateway reverse proxy",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.GatewayReverseProxy{},
			},
			want: false,
		},
		{
			name: "pool empty CU gateway subdmain",
			args: args{
				pool: types.Pool{
					Cus: 0,
					Sus: 10_000,
				},
				workload: &workloads.GatewaySubdomain{},
			},
			want: false,
		},
		{
			name: "pool empty SU container free disk",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.Container{
					Capacity: workloads.ContainerCapacity{
						Cpu:      2,
						Memory:   2048,
						DiskSize: 5 * 1024,
						DiskType: workloads.DiskTypeSSD,
					},
				},
			},
			want: false,
		},
		{
			name: "pool empty SU container paying disk",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.Container{
					Capacity: workloads.ContainerCapacity{
						Cpu:      2,
						Memory:   2048,
						DiskSize: 51 * 1024,
						DiskType: workloads.DiskTypeSSD,
					},
				},
			},
			want: true,
		},
		{
			name: "pool empty SU volume",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.Volume{
					Size: 2048,
					Type: workloads.VolumeTypeSSD,
				},
			},
			want: true,
		},
		{
			name: "pool empty SU network",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.NetworkResource{},
			},
			want: false,
		},
		{
			name: "pool empty SU k8s",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.K8S{
					Size: 1,
				},
			},
			want: true,
		},
		{
			name: "pool empty SU zdb",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.ZDB{
					Size:     2048,
					DiskType: workloads.DiskTypeSSD,
				},
			},
			want: true,
		},
		{
			name: "pool empty SU gateway 4 to 6",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.Gateway4To6{},
			},
			want: false,
		},
		{
			name: "pool empty SU gateway delegate",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.GatewayDelegate{},
			},
			want: false,
		},
		{
			name: "pool empty SU gateway proxy",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.GatewayProxy{},
			},
			want: false,
		},
		{
			name: "pool empty SU gateway reverse proxy",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.GatewayReverseProxy{},
			},
			want: false,
		},
		{
			name: "pool empty SU gateway subdmain",
			args: args{
				pool: types.Pool{
					Cus: 10_000,
					Sus: 0,
				},
				workload: &workloads.GatewaySubdomain{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := usesExpiredResources(tt.args.pool, tt.args.workload)
			require.NoError(t, err)
			if got != tt.want {
				t.Errorf("usesExpiredResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
