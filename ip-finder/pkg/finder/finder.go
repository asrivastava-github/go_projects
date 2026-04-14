package finder

import (
	"context"

	awspkg "ip-finder/pkg/aws"
	"ip-finder/pkg/k8s"
	"ip-finder/pkg/logger"
)

type Result struct {
	IP              string
	Region          string
	AWSProfile      string
	ENIs            []awspkg.ClassifiedENI
	Instance        *awspkg.InstanceDetails
	Pods            []PodSearchResult
	K8sSearchErrors []K8sError
	K8sSkipped      bool
	K8sSkipReason   string
}

type PodSearchResult struct {
	Context       string
	Pods          []k8s.PodResult
	AppPodsOnNode []k8s.PodResult // non-hostNetwork pods on the same node (when IP is a node IP)
}

type K8sError struct {
	Context string
	Error   error
}

type Options struct {
	Region      string
	AWSProfile  string
	KubeContext string
	AllContexts bool
	SkipK8s     bool
}

type IPFinder struct {
	awsClient      *awspkg.Client
	eniFinder      *awspkg.ENIFinder
	instanceFinder *awspkg.InstanceFinder
	options        Options
}

func New(ctx context.Context, opts Options) (*IPFinder, error) {
	awsClient, err := awspkg.NewClient(ctx, opts.Region, opts.AWSProfile)
	if err != nil {
		return nil, err
	}

	return &IPFinder{
		awsClient:      awsClient,
		eniFinder:      awspkg.NewENIFinder(awsClient.EC2, opts.Region),
		instanceFinder: awspkg.NewInstanceFinder(awsClient.EC2),
		options:        opts,
	}, nil
}

func (f *IPFinder) Find(ctx context.Context, ip string) (*Result, error) {
	result := &Result{
		IP:         ip,
		Region:     f.options.Region,
		AWSProfile: f.options.AWSProfile,
	}

	logger.Debug("Searching ENIs for IP: %s", ip)
	enis, err := f.eniFinder.FindByIP(ctx, ip)
	if err != nil {
		logger.Error("ENI search failed: %v", err)
		return nil, err
	}

	if len(enis) == 0 {
		logger.Warn("No ENI found for IP: %s", ip)
		result.K8sSkipped = true
		result.K8sSkipReason = "No ENI found. The IP may not exist in this AWS account or region."
		return result, nil
	}
	logger.Debug("Found %d ENI(s) for IP: %s", len(enis), ip)

	classifiedENIs := make([]awspkg.ClassifiedENI, 0, len(enis))
	for _, eni := range enis {
		classifiedENIs = append(classifiedENIs, awspkg.ClassifyENI(eni))
	}
	result.ENIs = classifiedENIs

	for _, eni := range classifiedENIs {
		logger.Debug("ENI %s classified as: %s", eni.NetworkInterfaceID, eni.ResourceType.DisplayName())
		if eni.InstanceID != "" {
			logger.Debug("Fetching EC2 instance details for: %s", eni.InstanceID)
			details, err := f.instanceFinder.GetDetails(ctx, eni.InstanceID)
			if err == nil && details != nil {
				result.Instance = details
				break
			}
		}
	}

	if f.options.SkipK8s {
		logger.Debug("K8s search skipped by user flag")
		result.K8sSkipped = true
		result.K8sSkipReason = "Skipped by user (--skip-k8s flag)."
		return result, nil
	}

	if shouldSkip, reason := f.shouldSkipK8sSearch(classifiedENIs); shouldSkip {
		logger.Debug("K8s search skipped: %s", reason)
		result.K8sSkipped = true
		result.K8sSkipReason = reason
		return result, nil
	}

	logger.Debug("Starting K8s pod search")
	f.searchK8sPods(ctx, ip, result)

	return result, nil
}

func (f *IPFinder) shouldSkipK8sSearch(enis []awspkg.ClassifiedENI) (bool, string) {
	if len(enis) == 0 {
		return false, ""
	}

	for _, eni := range enis {
		if eni.MayBePodIP {
			return false, ""
		}
	}

	resourceType := enis[0].ResourceType
	return true, "IP belongs to " + resourceType.DisplayName() + ", not an EKS pod."
}

func (f *IPFinder) searchK8sPods(ctx context.Context, ip string, result *Result) {
	contexts := f.getK8sContexts()
	logger.Debug("Searching %d K8s context(s)", len(contexts))

	for _, kctx := range contexts {
		ctxName := contextDisplayName(kctx)
		logger.Debug("Connecting to K8s context: %s", ctxName)

		client, err := k8s.NewClient(kctx)
		if err != nil {
			logger.Warn("Failed to connect to context %s: %v", ctxName, err)
			result.K8sSearchErrors = append(result.K8sSearchErrors, K8sError{
				Context: ctxName,
				Error:   err,
			})
			continue
		}

		podFinder := k8s.NewPodFinder(client)
		pods, err := podFinder.FindByIP(ctx, ip)
		if err != nil {
			logger.Warn("Pod search failed in context %s: %v", ctxName, err)
			result.K8sSearchErrors = append(result.K8sSearchErrors, K8sError{
				Context: ctxName,
				Error:   err,
			})
			continue
		}

		if len(pods) > 0 {
			logger.Info("Found %d pod(s) in context %s", len(pods), ctxName)
			psr := PodSearchResult{
				Context: ctxName,
				Pods:    pods,
			}

			if allHostNetwork(pods) {
				logger.Debug("All pods use hostNetwork, searching for application pods on node")
				nodePods, err := podFinder.FindByNodeIP(ctx, ip)
				if err != nil {
					logger.Warn("Failed to find pods by node IP in context %s: %v", ctxName, err)
				} else {
					appPods := filterNonHostNetwork(nodePods)
					if len(appPods) > 0 {
						logger.Info("Found %d application pod(s) on node in context %s", len(appPods), ctxName)
						psr.AppPodsOnNode = appPods
					}
				}
			}

			result.Pods = append(result.Pods, psr)
		} else {
			logger.Debug("No pods found in context %s", ctxName)
		}
	}
}

func (f *IPFinder) getK8sContexts() []string {
	if f.options.AllContexts {
		contexts, err := k8s.GetAvailableContexts()
		if err == nil {
			return contexts
		}
	}

	if f.options.KubeContext != "" {
		return []string{f.options.KubeContext}
	}

	return []string{""}
}

func contextDisplayName(ctx string) string {
	if ctx == "" {
		return "(current context)"
	}
	return ctx
}

func allHostNetwork(pods []k8s.PodResult) bool {
	for _, pod := range pods {
		if !pod.HostNetwork {
			return false
		}
	}
	return len(pods) > 0
}

func filterNonHostNetwork(pods []k8s.PodResult) []k8s.PodResult {
	var result []k8s.PodResult
	for _, pod := range pods {
		if !pod.HostNetwork {
			result = append(result, pod)
		}
	}
	return result
}
