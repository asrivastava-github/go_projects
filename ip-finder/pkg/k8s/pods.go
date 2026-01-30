package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodResult struct {
	Name        string
	Namespace   string
	PodIP       string
	NodeName    string
	NodeIP      string
	Status      string
	HostNetwork bool
	Labels      map[string]string
	Annotations map[string]string
}

type PodFinder struct {
	client *Client
}

func NewPodFinder(client *Client) *PodFinder {
	return &PodFinder{client: client}
}

func (f *PodFinder) FindByIP(ctx context.Context, ip string) ([]PodResult, error) {
	pods, err := f.client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("status.podIP=%s", ip),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	results := make([]PodResult, 0, len(pods.Items))
	for _, pod := range pods.Items {
		results = append(results, podToResult(&pod))
	}

	return results, nil
}

func (f *PodFinder) FindByNodeIP(ctx context.Context, nodeIP string) ([]PodResult, error) {
	nodes, err := f.client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeName := findNodeByIP(nodes.Items, nodeIP)
	if nodeName == "" {
		return nil, nil
	}

	pods, err := f.client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods on node: %w", err)
	}

	results := make([]PodResult, 0, len(pods.Items))
	for _, pod := range pods.Items {
		results = append(results, podToResult(&pod))
	}

	return results, nil
}

func findNodeByIP(nodes []corev1.Node, nodeIP string) string {
	for _, node := range nodes {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP && addr.Address == nodeIP {
				return node.Name
			}
		}
	}
	return ""
}

func podToResult(pod *corev1.Pod) PodResult {
	return PodResult{
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		PodIP:       pod.Status.PodIP,
		NodeName:    pod.Spec.NodeName,
		NodeIP:      pod.Status.HostIP,
		Status:      string(pod.Status.Phase),
		HostNetwork: pod.Spec.HostNetwork,
		Labels:      pod.Labels,
		Annotations: pod.Annotations,
	}
}
