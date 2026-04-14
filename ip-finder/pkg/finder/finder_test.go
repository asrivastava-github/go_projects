package finder

import (
	"testing"

	"ip-finder/pkg/k8s"
)

func TestAllHostNetwork(t *testing.T) {
	tests := []struct {
		name string
		pods []k8s.PodResult
		want bool
	}{
		{
			name: "empty list",
			pods: []k8s.PodResult{},
			want: false,
		},
		{
			name: "all hostNetwork",
			pods: []k8s.PodResult{
				{Name: "calico-node-abc", HostNetwork: true},
				{Name: "kube-proxy-xyz", HostNetwork: true},
			},
			want: true,
		},
		{
			name: "mixed",
			pods: []k8s.PodResult{
				{Name: "calico-node-abc", HostNetwork: true},
				{Name: "my-app-xyz", HostNetwork: false},
			},
			want: false,
		},
		{
			name: "none hostNetwork",
			pods: []k8s.PodResult{
				{Name: "my-app-xyz", HostNetwork: false},
				{Name: "my-api-abc", HostNetwork: false},
			},
			want: false,
		},
		{
			name: "single hostNetwork",
			pods: []k8s.PodResult{
				{Name: "kube-proxy-xyz", HostNetwork: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := allHostNetwork(tt.pods)
			if got != tt.want {
				t.Errorf("allHostNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterNonHostNetwork(t *testing.T) {
	tests := []struct {
		name      string
		pods      []k8s.PodResult
		wantCount int
		wantNames []string
	}{
		{
			name:      "empty list",
			pods:      []k8s.PodResult{},
			wantCount: 0,
		},
		{
			name: "all hostNetwork returns empty",
			pods: []k8s.PodResult{
				{Name: "calico-node-abc", HostNetwork: true},
				{Name: "kube-proxy-xyz", HostNetwork: true},
			},
			wantCount: 0,
		},
		{
			name: "filters out hostNetwork pods",
			pods: []k8s.PodResult{
				{Name: "calico-node-abc", HostNetwork: true},
				{Name: "my-app-xyz", Namespace: "default", HostNetwork: false},
				{Name: "kube-proxy-xyz", HostNetwork: true},
				{Name: "my-api-abc", Namespace: "default", HostNetwork: false},
			},
			wantCount: 2,
			wantNames: []string{"my-app-xyz", "my-api-abc"},
		},
		{
			name: "all non-hostNetwork returns all",
			pods: []k8s.PodResult{
				{Name: "my-app-xyz", HostNetwork: false},
				{Name: "my-api-abc", HostNetwork: false},
			},
			wantCount: 2,
			wantNames: []string{"my-app-xyz", "my-api-abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterNonHostNetwork(tt.pods)
			if len(got) != tt.wantCount {
				t.Errorf("filterNonHostNetwork() returned %d pods, want %d", len(got), tt.wantCount)
			}
			for i, name := range tt.wantNames {
				if i < len(got) && got[i].Name != name {
					t.Errorf("filterNonHostNetwork()[%d].Name = %q, want %q", i, got[i].Name, name)
				}
			}
		})
	}
}
