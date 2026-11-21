package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestEKS543PublicPrivateAndIncompleteNetworkEvidence(t *testing.T) {
	data := func(value string) json.RawMessage { return json.RawMessage(value) }
	subnet := func(public bool, source, gateway string) json.RawMessage {
		return data(fmt.Sprintf(
			`{"SubnetId":"subnet-123","VpcId":"vpc-123","nodeGroup":"workers","MapPublicIpOnLaunch":%t,"routeSource":"%s","routeTableId":"rtb-123","routes":[{"ipv4":"0.0.0.0/0","gateway":"%s","state":"active"}]}`,
			public, source, gateway,
		))
	}
	privateNode := data(`{"node":"worker","InstanceId":"i-123","SubnetId":"subnet-123","PublicIpAddress":null,"State":"running"}`)
	base := EKSAWSManualEvidence{Status: "COLLECTED", NodeInventoryStatus: "COLLECTED", Nodes: []json.RawMessage{privateNode}}
	tests := []struct {
		name   string
		record json.RawMessage
		status string
	}{
		{"public via explicit internet gateway", subnet(false, "EXPLICIT", "igw-123"), "FAIL"},
		{"public via inherited main gateway", subnet(false, "MAIN", "igw-456"), "FAIL"},
		{"auto public IP without gateway", subnet(true, "MAIN", "nat-123"), "MANUAL"},
		{"private via NAT", subnet(false, "MAIN", "nat-123"), "PASS"},
		{"private via VPC endpoint", subnet(false, "EXPLICIT", ""), "PASS"},
		{"route unavailable", subnet(false, "UNAVAILABLE", ""), "MANUAL"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			review := base
			review.Items = []json.RawMessage{tc.record}
			status, description, evidence := assessEKSPrivateNodes(review)
			if status != tc.status || len(evidence) == 0 || description == "" {
				t.Fatalf("expected %s, got %s: %s", tc.status, status, description)
			}
		})
	}
	// Inactive gateway routes are not effective internet paths.
	blackhole := base
	blackhole.Items = []json.RawMessage{data(`{"SubnetId":"subnet-123","VpcId":"vpc-123","MapPublicIpOnLaunch":false,"routeSource":"MAIN","routeTableId":"rtb-123","routes":[{"ipv4":"0.0.0.0/0","gateway":"igw-123","state":"blackhole"}]}`)}
	if status, _, _ := assessEKSPrivateNodes(blackhole); status != "MANUAL" {
		t.Fatalf("inactive IGW route must not be treated as active, got %s", status)
	}
	publicNode := data(`{"node":"worker","InstanceId":"i-123","SubnetId":"subnet-123","PublicIpAddress":"203.0.113.5","State":"running"}`)
	private := base
	private.Items = []json.RawMessage{subnet(false, "MAIN", "nat-123")}
	private.Nodes = []json.RawMessage{publicNode}
	if status, _, _ := assessEKSPrivateNodes(private); status != "FAIL" {
		t.Fatalf("public EC2 node must FAIL even if subnet's default route is NAT, got %s", status)
	}
	private.Nodes = []json.RawMessage{data(`{"node":"worker","InstanceId":"i-123","SubnetId":"subnet-123","State":"running","NetworkInterfacePublicIps":["203.0.113.10"]}`)}
	if status, _, _ := assessEKSPrivateNodes(private); status != "FAIL" {
		t.Fatalf("secondary network interface EIP must FAIL: %s", status)
	}
	private.Nodes = nil
	if status, _, _ := assessEKSPrivateNodes(private); status != "MANUAL" {
		t.Fatalf("missing live worker inventory must not PASS: %s", status)
	}
	private.Nodes = []json.RawMessage{privateNode}
	private.NodeInventoryStatus = "PARTIAL"
	if status, _, _ := assessEKSPrivateNodes(private); status != "MANUAL" {
		t.Fatalf("partial worker inventory must not PASS: %s", status)
	}
	private.NodeInventoryStatus = "COLLECTED"
	private.Nodes = []json.RawMessage{data(`{"node":"unmapped","InstanceId":"i-999","SubnetId":"subnet-999"}`)}
	if status, _, _ := assessEKSPrivateNodes(private); status != "MANUAL" {
		t.Fatalf("uninspected EC2 subnet must not PASS: %s", status)
	}
}

func TestEKS543PublicPlacementRemainsFailEvenWithApprovedExceptionText(t *testing.T) {
	review := EKSAWSManualEvidence{Status: "PARTIAL", Items: []json.RawMessage{
		json.RawMessage(`{"SubnetId":"subnet-123","VpcId":"vpc-123","MapPublicIpOnLaunch":true,"routeSource":"EXPLICIT","routeTableId":"rtb-1","routes":[{"ipv4":"0.0.0.0/0","gateway":"igw-1","state":"active"}]}`),
	}, Note: "Approved by owner"}
	status, detail, _ := assessEKSPrivateNodes(review)
	if status != "FAIL" || !strings.Contains(detail, "separate approved exception") {
		t.Fatalf("exception must not override native benchmark failure: %s %s", status, detail)
	}
}

func TestEKS543PrimaryPublicIPDoesNotDoubleCountNetworkInterface(t *testing.T) {
	subnet := json.RawMessage(`{"SubnetId":"subnet-123","VpcId":"vpc-123","MapPublicIpOnLaunch":false,"routeSource":"MAIN","routeTableId":"rtb-123","routes":[{"ipv4":"0.0.0.0/0","gateway":"nat-123","state":"active"}]}`)
	node := func(ips string) json.RawMessage {
		return json.RawMessage(fmt.Sprintf(`{"InstanceId":"i-123","SubnetId":"subnet-123","PublicIpAddress":"203.0.113.5","State":"running","NetworkInterfacePublicIps":%s}`, ips))
	}
	review := EKSAWSManualEvidence{Status: "COLLECTED", NodeInventoryStatus: "COLLECTED", Items: []json.RawMessage{subnet}, Nodes: []json.RawMessage{node(`["203.0.113.5","203.0.113.5"]`)}}
	status, detail, evidence := assessEKSPrivateNodes(review)
	if status != "FAIL" || !strings.Contains(detail, "assigned public IPv4") || strings.Contains(detail, "additional network-interface") {
		t.Fatalf("primary IPv4 must be recorded once: %s %s", status, detail)
	}
	if !strings.Contains(evidence[len(evidence)-1].Observed, "additionalPublicAddresses=0") {
		t.Fatalf("duplicate primary interface must not inflate extra interface count: %+v", evidence)
	}
	review.Nodes = []json.RawMessage{node(`["203.0.113.5","203.0.113.6","203.0.113.6"]`)}
	status, detail, evidence = assessEKSPrivateNodes(review)
	if status != "FAIL" || !strings.Contains(detail, "additional network-interface") || !strings.Contains(evidence[len(evidence)-1].Observed, "additionalPublicAddresses=1") {
		t.Fatalf("distinct additional IPv4 must remain visible: %s %s %+v", status, detail, evidence)
	}
}

func TestEKS543PublicIPDefaultEvidence(t *testing.T) {
	for _, value := range []string{"true", "false", "null"} {
		t.Run(value, func(t *testing.T) {
			review := EKSAWSManualEvidence{Items: []json.RawMessage{
				json.RawMessage(fmt.Sprintf(`{"SubnetId":"subnet-123","VpcId":"vpc-123","MapPublicIpOnLaunch":%s}`, value)),
			}}
			_, _, evidence := assessEKSPrivateNodes(review)
			want := value
			if value == "null" {
				want = "unknown"
			}
			if len(evidence) != 1 || !strings.HasSuffix(evidence[0].Observed, "mapPublicIP="+want) {
				t.Fatalf("expected public-IP default %s, got %+v", want, evidence)
			}
		})
	}
}
