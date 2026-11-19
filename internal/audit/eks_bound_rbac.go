package audit

import (
 "context"
 "fmt"
 "sort"
 "strings"

 rbacv1 "k8s.io/api/rbac/v1"
 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 "k8s.io/client-go/kubernetes"
)

// evaluateEKSBoundRBAC collects effective, bound RBAC rule evidence for the
// human-review CIS EKS 4.1 controls. A role without a binding is not a grant.
// Absence of candidates is NOT an automatic CIS PASS.
func evaluateEKSBoundRBAC(ctx context.Context, client kubernetes.Interface, c Check) Result {
 r := Result{ID:c.ID, Title:c.Title, Target:c.Target, Status:"MANUAL", Scored:false,
  Method:"kubernetes_bound_rbac_inventory", EvidenceSource:"kubernetes-rbac",
  Remediation:c.Remediation}
 cr,err:=client.RbacV1().ClusterRoles().List(ctx,metav1.ListOptions{})
 if err!=nil { r.Status="NOT_ASSESSED";r.Details="Cannot inventory ClusterRoles: "+err.Error();return r }
 roles,err:=client.RbacV1().Roles("").List(ctx,metav1.ListOptions{})
 if err!=nil { r.Status="NOT_ASSESSED";r.Details="Cannot inventory Roles across all namespaces: "+err.Error();return r }
 cb,err:=client.RbacV1().ClusterRoleBindings().List(ctx,metav1.ListOptions{})
 if err!=nil { r.Status="NOT_ASSESSED";r.Details="Cannot inventory ClusterRoleBindings: "+err.Error();return r }
 rb,err:=client.RbacV1().RoleBindings("").List(ctx,metav1.ListOptions{})
 if err!=nil { r.Status="NOT_ASSESSED";r.Details="Cannot inventory RoleBindings across all namespaces: "+err.Error();return r }
 byCluster:=map[string]rbacv1.ClusterRole{}
 byRole:=map[string]rbacv1.Role{}
 for _,v:=range cr.Items { byCluster[v.Name]=v }
 for _,v:=range roles.Items { byRole[v.Namespace+"/"+v.Name]=v }
 candidates,unresolved,emptySubjects:=0,0,0
 // Each binding is separately evaluated. A RoleBinding referencing a
 // ClusterRole grants only within its namespace, not cluster-wide.
 inspect:=func(kind,namespace,name string, ref rbacv1.RoleRef,subjects []rbacv1.Subject) {
  var rules []rbacv1.PolicyRule
  ok:=false
  switch ref.Kind {
  case "ClusterRole":
   var role rbacv1.ClusterRole
   role,ok=byCluster[ref.Name]
   if ok {rules=role.Rules}
  case "Role":
   if kind=="RoleBinding" {
    var role rbacv1.Role
    role,ok=byRole[namespace+"/"+ref.Name]
    if ok {rules=role.Rules}
   }
  }
  if !ok {
   unresolved++
   if len(r.Evidence)<200 {
    r.Evidence=append(r.Evidence,Evidence{Resource:kind,Namespace:namespace,Name:name,
     Field:"roleRef",Observed:ref.Kind+"/"+ref.Name+" (unresolved)",
     Expected:"binding references an inventoried Role/ClusterRole"})
   }
   return
  }
  if len(subjects)==0 {emptySubjects++;return}
  subjectNames:=make([]string,0,len(subjects))
  for _,s:=range subjects {
   subjectNames=append(subjectNames,fmt.Sprintf("%s:%s:%s",s.Kind,s.Namespace,s.Name))
  }
  sort.Strings(subjectNames)
  scope:="cluster"
  if kind=="RoleBinding" {scope="namespace:"+namespace}
  for i,rule:=range rules {
   if !eksRBACSensitiveRule(c.ID,rule) {continue}
   candidates++
   if len(r.Evidence)>=200 {continue}
   r.Evidence=append(r.Evidence,Evidence{Resource:kind,Namespace:namespace,Name:name,
    Field:fmt.Sprintf("roleRef=%s/%s.rules[%d]",ref.Kind,ref.Name,i),
    Observed:fmt.Sprintf("scope=%s; subjects=%s; apiGroups=%s; resources=%s; verbs=%s; resourceNames=%s; nonResourceURLs=%s",
     scope,strings.Join(subjectNames,","),strings.Join(rule.APIGroups,","),strings.Join(rule.Resources,","),
     strings.Join(rule.Verbs,","),strings.Join(rule.ResourceNames,","),strings.Join(rule.NonResourceURLs,",")),
    Expected:"verify each bound subject, effective permission and documented least-privilege approval"})
  }
 }
 for _,b:=range cb.Items {inspect("ClusterRoleBinding","",b.Name,b.RoleRef,b.Subjects)}
 for _,b:=range rb.Items {inspect("RoleBinding",b.Namespace,b.Name,b.RoleRef,b.Subjects)}
 if unresolved>0 {r.Status="NOT_ASSESSED"}
 r.Details=fmt.Sprintf("Bound RBAC inventory: %d ClusterRoles, %d Roles, %d ClusterRoleBindings and %d RoleBindings; %d sensitive grant(s), %d unresolved role reference(s), %d binding(s) without subjects; evidence retained for %d of %d candidate(s). Effective namespace scope and bound subjects are recorded. Authorization necessity requires human review; no automatic PASS.",
  len(cr.Items),len(roles.Items),len(cb.Items),len(rb.Items),candidates,unresolved,emptySubjects,min(candidates,len(r.Evidence)),candidates)
 if len(r.Evidence)>=200 && candidates>200 {
  r.Details+=" Evidence sample truncated at 200; do not approve without full inventory."
 }
 return r
}

// Match the actual Kubernetes API group and verb for sensitive resources.
// Wildcards cover resource/subresource names; a literal "nodes" does not
// imply nodes/proxy, and read-only access is not a token creation grant.
func eksRBACSensitiveRule(id string, rule rbacv1.PolicyRule) bool {
 has:=func(xs []string,value string)bool {
  for _,x:=range xs {if x==value||x=="*" {return true}}
  return false
 }
 verb:=func(values ...string)bool {
  for _,v:=range values {if has(rule.Verbs,v){return true}}
  return false
 }
 resource:=func(value string)bool {
  if has(rule.Resources,value){return true}
  parts:=strings.SplitN(value,"/",2)
  if len(parts)!=2{return false}
  return hasExact(rule.Resources,parts[0]+"/*")
 }
 core:=has(rule.APIGroups,"")
 rbac:=has(rule.APIGroups,"rbac.authorization.k8s.io")
 admission:=has(rule.APIGroups,"admissionregistration.k8s.io")
 switch id {
 case "4.1.2":return core&&resource("secrets")&&verb("get","list","watch")
 case "4.1.3":
  return hasExact(rule.Verbs,"*")||hasExact(rule.APIGroups,"*")||hasExact(rule.Resources,"*")||
   hasRBACResourceWildcard(rule.Resources)
 case "4.1.8":
  return verb("bind","escalate")&&rbac || verb("impersonate")&&len(rule.Resources)>0
 case "4.1.10":return core&&resource("nodes/proxy")&&verb("get","list","watch","create","update","patch","delete","deletecollection")
 case "4.1.11":return admission&&(resource("validatingwebhookconfigurations")||resource("mutatingwebhookconfigurations"))&&verb("create","update","patch","delete","deletecollection")
 case "4.1.12":return core&&resource("serviceaccounts/token")&&verb("create")
 }
 return false
}
func hasExact(items []string,value string)bool {
 for _,x:=range items {if x==value{return true}}
 return false
}
func hasRBACResourceWildcard(items []string)bool {
 for _,x:=range items {if strings.HasSuffix(x,"/*"){return true}}
 return false
}
