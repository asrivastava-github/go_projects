package output

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	awspkg "ip-finder/pkg/aws"
	"ip-finder/pkg/finder"
	"ip-finder/pkg/k8s"
)

type Printer struct {
	writer *tabwriter.Writer
}

func NewPrinter() *Printer {
	return &Printer{
		writer: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
	}
}

func (p *Printer) PrintResult(result *finder.Result) {
	profile := result.AWSProfile
	if profile == "" {
		profile = "(default)"
	}
	fmt.Printf("[Search] IP: %s | Region: %s | Profile: %s\n\n", result.IP, result.Region, profile)

	p.printENISection(result.ENIs)

	if result.Instance != nil {
		p.printInstanceSection(result.Instance)
	}

	if result.K8sSkipped {
		p.PrintK8sSkipped(result.K8sSkipReason)
	} else {
		p.printK8sSection(result.Pods, result.K8sSearchErrors)
	}
}

func (p *Printer) printENISection(enis []awspkg.ClassifiedENI) {
	fmt.Println("━━━ AWS ENI Search ━━━")

	if len(enis) == 0 {
		fmt.Println("No ENIs found with this IP address")
		return
	}

	fmt.Fprintln(p.writer, "ENI ID\tResource Type\tStatus\tInstance ID\tDescription")
	fmt.Fprintln(p.writer, "------\t-------------\t------\t-----------\t-----------")

	for _, r := range enis {
		desc := truncate(r.Description, 40)
		instanceID := orDefault(r.InstanceID, "-")
		fmt.Fprintf(p.writer, "%s\t%s\t%s\t%s\t%s\n",
			r.NetworkInterfaceID,
			r.ResourceType.DisplayName(),
			r.Status,
			instanceID,
			desc,
		)
	}
	p.writer.Flush()

	for _, r := range enis {
		fmt.Printf("\n  VPC: %s | Subnet: %s | AZ: %s\n", r.VPCID, r.SubnetID, r.AvailabilityZone)
		if r.FoundViaPrefix {
			fmt.Printf("  [Prefix Delegation] Matched: %s\n", r.MatchedPrefix)
		}
		if len(r.PrivateIPs) > 1 {
			fmt.Printf("  All Private IPs: %s\n", strings.Join(r.PrivateIPs, ", "))
		}
		if len(r.Tags) > 0 {
			fmt.Printf("  Tags: %s\n", formatTags(r.Tags))
		}
	}
}

func (p *Printer) printInstanceSection(d *awspkg.InstanceDetails) {
	fmt.Println("\n━━━ EC2 Instance Details ━━━")

	fmt.Fprintf(p.writer, "Instance ID:\t%s\n", d.InstanceID)
	fmt.Fprintf(p.writer, "Name:\t%s\n", d.Name)
	fmt.Fprintf(p.writer, "Type:\t%s\n", d.InstanceType)
	fmt.Fprintf(p.writer, "State:\t%s\n", d.State)
	fmt.Fprintf(p.writer, "Private IP:\t%s\n", d.PrivateIP)
	if d.PublicIP != "" {
		fmt.Fprintf(p.writer, "Public IP:\t%s\n", d.PublicIP)
	}
	p.writer.Flush()
}

func (p *Printer) printK8sSection(pods []finder.PodSearchResult, errors []finder.K8sError) {
	fmt.Println("\n━━━ Kubernetes Pod Search ━━━")

	for _, e := range errors {
		fmt.Printf("  [Warning] Context %s: %v\n", e.Context, e.Error)
	}

	if len(pods) == 0 {
		fmt.Println("No pods found with this IP address")
		return
	}

	for _, pr := range pods {
		fmt.Printf("\n[Found] Context: %s\n", pr.Context)

		hostNetworkCount := 0
		for _, pod := range pr.Pods {
			if pod.HostNetwork {
				hostNetworkCount++
			}
		}

		if hostNetworkCount == len(pr.Pods) && hostNetworkCount > 0 {
			fmt.Printf("[Note] This is a node IP. All %d pod(s) use hostNetwork and share the node's IP.\n", hostNetworkCount)
		} else if hostNetworkCount > 0 {
			fmt.Printf("[Note] %d of %d pod(s) use hostNetwork (share node IP).\n", hostNetworkCount, len(pr.Pods))
		}

		p.printPods(pr.Pods)

		if len(pr.AppPodsOnNode) > 0 {
			fmt.Printf("\n[App Pods] %d application pod(s) running on this node (potential DB clients via SNAT):\n", len(pr.AppPodsOnNode))
			p.printPods(pr.AppPodsOnNode)
		}
	}
}

func (p *Printer) PrintK8sSkipped(reason string) {
	fmt.Println("\n━━━ Kubernetes Pod Search ━━━")
	fmt.Printf("[Skipped] %s\n", reason)
}

func (p *Printer) printPods(pods []k8s.PodResult) {
	fmt.Fprintln(p.writer, "NAMESPACE\tNAME\tPOD IP\tNODE\tHOSTNET\tSTATUS")
	fmt.Fprintln(p.writer, "---------\t----\t------\t----\t-------\t------")

	for _, pod := range pods {
		hostNet := "no"
		if pod.HostNetwork {
			hostNet = "yes"
		}
		fmt.Fprintf(p.writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			pod.Namespace,
			pod.Name,
			pod.PodIP,
			pod.NodeName,
			hostNet,
			pod.Status,
		)
	}
	p.writer.Flush()
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func formatTags(tags map[string]string) string {
	pairs := make([]string, 0, len(tags))
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}
