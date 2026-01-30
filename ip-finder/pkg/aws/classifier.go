package aws

import "strings"

type ResourceType string

const (
	ResourceTypeEC2          ResourceType = "ec2"
	ResourceTypeLambda       ResourceType = "lambda"
	ResourceTypeELB          ResourceType = "elb"
	ResourceTypeNATGateway   ResourceType = "nat_gateway"
	ResourceTypeRDS          ResourceType = "rds"
	ResourceTypeEFS          ResourceType = "efs"
	ResourceTypeVPCEndpoint  ResourceType = "vpc_endpoint"
	ResourceTypeTransitGW    ResourceType = "transit_gateway"
	ResourceTypeEKSControlPlane ResourceType = "eks_control_plane"
	ResourceTypeUnknown      ResourceType = "unknown"
)

type ClassifiedENI struct {
	ENIResult
	ResourceType   ResourceType
	MayBePodIP     bool
}

func ClassifyENI(eni ENIResult) ClassifiedENI {
	classified := ClassifiedENI{
		ENIResult:    eni,
		ResourceType: ResourceTypeUnknown,
		MayBePodIP:   false,
	}

	descLower := strings.ToLower(eni.Description)
	interfaceType := strings.ToLower(eni.InterfaceType)

	switch {
	case interfaceType == "nat_gateway":
		classified.ResourceType = ResourceTypeNATGateway

	case interfaceType == "gateway_load_balancer_endpoint":
		classified.ResourceType = ResourceTypeELB

	case interfaceType == "vpc_endpoint":
		classified.ResourceType = ResourceTypeVPCEndpoint

	case interfaceType == "transit_gateway":
		classified.ResourceType = ResourceTypeTransitGW

	case strings.Contains(descLower, "aws lambda vpc"):
		classified.ResourceType = ResourceTypeLambda

	case strings.Contains(descLower, "elb ") ||
		strings.Contains(descLower, "elbv2") ||
		strings.HasPrefix(descLower, "elb app/") ||
		strings.HasPrefix(descLower, "elb net/"):
		classified.ResourceType = ResourceTypeELB

	case strings.Contains(descLower, "rds"):
		classified.ResourceType = ResourceTypeRDS

	case strings.Contains(descLower, "efs mount target"):
		classified.ResourceType = ResourceTypeEFS

	case strings.Contains(descLower, "amazon eks"):
		classified.ResourceType = ResourceTypeEKSControlPlane

	case eni.InstanceID != "":
		classified.ResourceType = ResourceTypeEC2
		classified.MayBePodIP = true

	default:
		classified.MayBePodIP = true
	}

	return classified
}

func (r ResourceType) ShouldSearchK8s() bool {
	switch r {
	case ResourceTypeEC2, ResourceTypeUnknown:
		return true
	default:
		return false
	}
}

func (r ResourceType) DisplayName() string {
	names := map[ResourceType]string{
		ResourceTypeEC2:          "EC2 Instance",
		ResourceTypeLambda:       "Lambda Function",
		ResourceTypeELB:          "Load Balancer",
		ResourceTypeNATGateway:   "NAT Gateway",
		ResourceTypeRDS:          "RDS Instance",
		ResourceTypeEFS:          "EFS Mount Target",
		ResourceTypeVPCEndpoint:  "VPC Endpoint",
		ResourceTypeTransitGW:    "Transit Gateway",
		ResourceTypeEKSControlPlane: "EKS Control Plane",
		ResourceTypeUnknown:      "Unknown",
	}
	if name, ok := names[r]; ok {
		return name
	}
	return string(r)
}
