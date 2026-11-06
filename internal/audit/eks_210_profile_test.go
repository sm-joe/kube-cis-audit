package audit

import (
 "path/filepath"
 "testing"
)

// This test keeps the official EKS 2.1.0 catalog and native profile in sync.
// AWS-managed controls are assessed by the validated provider-evidence source,
// not duplicated in the Kubernetes API or EC2 node reports.
func TestEKS210ProfileExactOfficialIDs(t *testing.T) {
 inv,err:=LoadCIS210ControlInventory("eks");if err!=nil{t.Fatal(err)}
 b,err:=readBench(filepath.Join("..","..","benchmarks","cis-eks-2.1.0","benchmark.yaml"));if err!=nil{t.Fatal(err)}
 if b.ID!="cis-eks-2.1.0" {t.Fatalf("wrong EKS benchmark: %s",b.ID)}
 if b.CompatibilityStatus=="verified" {t.Fatal("incomplete EKS controls incorrectly certified")}
 seen:=map[string]bool{"2.1.1":true,"5.4.1":true,"5.4.2":true}
 for _,c:=range b.Checks{
  if seen[c.ID]{t.Fatalf("duplicate official EKS control %s",c.ID)}
  seen[c.ID]=true
  if c.Kind=="cis_not_implemented"&&(c.Scored||c.Manual){
   t.Fatalf("missing evidence control incorrectly scored/marked reviewed: %+v",c)
  }
 }
 if len(seen)!=47||len(b.Checks)!=44{t.Fatalf("EKS v2.1.0 profile incomplete: %d IDs, %d config checks",len(seen),len(b.Checks))}
 for _,c:=range inv.Controls{
  if !seen[c.ID]{t.Fatalf("missing official EKS control %s",c.ID)}
 }
 nativeNode:=map[string]bool{}
 for _,c:=range b.Checks{if c.Kind=="host_file_mode"||c.Kind=="host_kubelet_config_field"{nativeNode[c.ID]=true}}
 for _,id:=range []string{"3.1.3","3.2.1","3.2.2"}{if !nativeNode[id]{t.Fatalf("missing assessed EKS node control %s",id)}}
}
