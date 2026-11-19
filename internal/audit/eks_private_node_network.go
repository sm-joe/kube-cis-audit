package audit

import (
	"encoding/json"
	"fmt"
	"strings"
)

// assessEKSPrivateNodes applies the CIS EKS 2.1.0 5.4.3 node-placement
// requirement to redacted, read-only AWS+Kubernetes evidence. A public
// subnet, public EC2 IP or direct IGW route is non-compliant even when
// intentionally configured. Missing visibility never creates a PASS.
func assessEKSPrivateNodes(review EKSAWSManualEvidence) (string, string, []Evidence) {
	type route struct {
		IPv4    string `json:"ipv4"`
		IPv6    string `json:"ipv6"`
		Gateway string `json:"gateway"`
		State   string `json:"state"`
	}
	type subnet struct {
		SubnetID       string  `json:"SubnetId"`
		VPCID          string  `json:"VpcId"`
		NodeGroup      string  `json:"nodeGroup"`
		PublicOnLaunch *bool   `json:"MapPublicIpOnLaunch"`
		RouteTableID   string  `json:"routeTableId"`
		RouteSource    string  `json:"routeSource"`
		Routes         []route `json:"routes"`
	}
	type node struct {
		Name                      string   `json:"node"`
		InstanceID                string   `json:"InstanceId"`
		SubnetID                  string   `json:"SubnetId"`
		PublicIP                  string   `json:"PublicIpAddress"`
		NetworkInterfacePublicIPs []string `json:"NetworkInterfacePublicIps"`
		State                     string   `json:"State"`
	}
	var evidence []Evidence
	var public []string
	var unknown []string
	subnetIDs := make(map[string]bool)
	if len(review.Items) == 0 {
		unknown = append(unknown, "no managed worker subnets inventoried")
	}
	for _, raw := range review.Items {
		var s subnet
		if err := json.Unmarshal(raw, &s); err != nil || s.SubnetID == "" || s.VPCID == "" {
			unknown = append(unknown, "subnet record unavailable or malformed")
			continue
		}
		subnetIDs[s.SubnetID] = true
		if s.PublicOnLaunch == nil {
			unknown = append(unknown, s.SubnetID+": subnet public IP default not verified")
		} else if *s.PublicOnLaunch {
			// This setting permits future public IP assignment but does not
			// prove that a worker is currently publicly routed. Require
			// verified IGW routing or a public EC2 interface for FAIL.
			unknown = append(unknown, s.SubnetID+": auto-assign public IPv4 enabled; inspect actual routing and node addresses")
		}
		if s.RouteSource != "MAIN" && s.RouteSource != "EXPLICIT" {
			unknown = append(unknown, s.SubnetID+": effective route-table lookup unavailable")
		} else if s.RouteTableID == "" {
			unknown = append(unknown, s.SubnetID+": effective route-table ID missing")
		} else {
			if len(s.Routes) == 0 {
				unknown = append(unknown, s.SubnetID+": effective route table has no route entries")
			}
			for _, r := range s.Routes {
				if r.State != "active" {
					unknown = append(unknown, s.SubnetID+": route state missing or inactive")
					continue
				}
				if strings.HasPrefix(r.Gateway, "igw-") && (r.IPv4 != "" || r.IPv6 != "") {
					public = append(public, fmt.Sprintf("%s: active internet-gateway route via %s (%s)", s.SubnetID, r.Gateway, s.RouteTableID))
				}
			}
		}
		evidence = append(evidence, Evidence{
			Resource: "AWS::EC2::Subnet", Name: s.SubnetID, Field: "effectiveRouteTable",
			Observed: fmt.Sprintf("nodeGroup=%s source=%s table=%s mapPublicIP=%v", s.NodeGroup, s.RouteSource, s.RouteTableID, s.PublicOnLaunch),
			Expected: "no IGW route; MapPublicIpOnLaunch=false; all worker nodes private",
		})
	}
	if len(review.Nodes) == 0 {
		unknown = append(unknown, "no live worker-node EC2 inventory")
	}
	for _, raw := range review.Nodes {
		var n node
		if err := json.Unmarshal(raw, &n); err != nil || n.InstanceID == "" || n.SubnetID == "" {
			unknown = append(unknown, "worker EC2 instance record incomplete")
			continue
		}
		if !subnetIDs[n.SubnetID] {
			unknown = append(unknown, n.InstanceID+": node subnet missing from assessed groups")
		}
		if n.State != "running" {
			unknown = append(unknown, n.InstanceID+": running EC2 instance state unverified")
		}
		if n.PublicIP != "" {
			public = append(public, n.InstanceID+": assigned public IPv4")
		}
		for _, ip := range n.NetworkInterfacePublicIPs {
			if ip != "" {
				public = append(public, n.InstanceID+": secondary network interface public IPv4/EIP")
				break
			}
		}
		evidence = append(evidence, Evidence{
			Resource: "AWS::EC2::Instance", Name: n.InstanceID, Field: "subnet/publicIPv4",
			Observed: fmt.Sprintf("%s/state=%s/primaryPublic=%t/secondaryPublicInterfaces=%d", n.SubnetID, n.State, n.PublicIP != "", len(n.NetworkInterfacePublicIPs)),
			Expected: "node in assessed private subnet without public IPv4",
		})
	}
	if len(public) > 0 {
		return "FAIL", fmt.Sprintf("CIS EKS 5.4.3 public worker-node exposure: %s. Intentional public placement requires a separate approved exception; it cannot become PASS.", strings.Join(public, "; ")), evidence
	}
	if len(unknown) > 0 || review.Status != "COLLECTED" || review.NodeInventoryStatus != "COLLECTED" {
		return "MANUAL", fmt.Sprintf("CIS EKS 5.4.3 private-node assessment incomplete: %s; collection=%s nodes=%s.", strings.Join(unknown, "; "), review.Status, review.NodeInventoryStatus), evidence
	}
	return "PASS", fmt.Sprintf("CIS EKS 5.4.3: %d managed subnet(s) have no IGW routes or public-IP-on-launch; %d discovered EC2 worker node(s) have no public IPv4 and are assigned to assessed subnets.", len(review.Items), len(review.Nodes)), evidence
}
