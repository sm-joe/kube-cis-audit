package audit

import (
 "context"
 "strings"
 "testing"
 corev1 "k8s.io/api/core/v1"
 appsv1 "k8s.io/api/apps/v1"
 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 "k8s.io/client-go/kubernetes/fake"
)

func TestEKSPodAdmissionRequiresEnforcedNamespacePolicy(t *testing.T){
 ctx:=context.Background()
 pod:=&corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"app",Namespace:"apps"},Spec:corev1.PodSpec{Containers:[]corev1.Container{{Name:"app"}}}}
 ns:=&corev1.Namespace{ObjectMeta:metav1.ObjectMeta{Name:"apps"}}
 client:=fake.NewSimpleClientset(ns,pod)
 for _,id:=range []string{"4.2.1","4.2.2","4.2.3","4.2.4","4.2.5"}{
  r,err:=EvaluateEKSPodAdmissionControl(ctx,client,id);if err!=nil{t.Fatal(err)}
  if r.Status!="MANUAL"{t.Fatalf("%s unlabeled namespace falsely marked %s: %+v",id,r.Status,r)}
 }
 ns.Labels=map[string]string{"pod-security.kubernetes.io/enforce":"restricted"}
 _,err:=client.CoreV1().Namespaces().Update(ctx,ns,metav1.UpdateOptions{});if err!=nil{t.Fatal(err)}
 for _,id:=range []string{"4.2.1","4.2.2","4.2.3","4.2.4"}{
  r,err:=EvaluateEKSPodAdmissionControl(ctx,client,id);if err!=nil{t.Fatal(err)}
  if r.Status!="PASS"{t.Fatalf("%s should pass under restricted PSA: %+v",id,r)}
 }
 r,err:=EvaluateEKSPodAdmissionControl(ctx,client,"4.2.5");if err!=nil{t.Fatal(err)}
 if r.Status!="FAIL"{t.Fatalf("missing allowPrivilegeEscalation=false must fail restricted Pod check: %+v",r)}
}

func TestEKSPodAdmissionDetectsPrivilegedExistingPodEvenWithPSA(t *testing.T){
 yes:=true;ctx:=context.Background()
 client:=fake.NewSimpleClientset(
  &corev1.Namespace{ObjectMeta:metav1.ObjectMeta{Name:"apps",Labels:map[string]string{"pod-security.kubernetes.io/enforce":"baseline"}}},
  &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"privileged",Namespace:"apps"},Spec:corev1.PodSpec{Containers:[]corev1.Container{{Name:"app",SecurityContext:&corev1.SecurityContext{Privileged:&yes}}}}},
 )
 r,err:=EvaluateEKSPodAdmissionControl(ctx,client,"4.2.1");if err!=nil{t.Fatal(err)}
 if r.Status!="FAIL"||len(r.Evidence)<2{t.Fatalf("did not detect historical privileged workload: %+v",r)}
}

func TestEKSDefaultServiceAccountChecksAllWorkloadTemplates(t *testing.T){
 ctx:=context.Background()
 client:=fake.NewSimpleClientset(&appsv1.Deployment{
  ObjectMeta:metav1.ObjectMeta{Name:"web",Namespace:"apps"},
  Spec:appsv1.DeploymentSpec{Template:corev1.PodTemplateSpec{Spec:corev1.PodSpec{Containers:[]corev1.Container{{Name:"web"}}}}},
 })
 r,err:=EvaluateEKSDefaultServiceAccountUse(ctx,client);if err!=nil{t.Fatal(err)}
 if r.Status!="FAIL"||len(r.Evidence)!=1||r.Evidence[0].Resource!="Deployment"{
  t.Fatalf("missed default account in deployment template: %+v",r)
 }
}

func TestEKSTokenMountsUsePodOverrideAndNeverAssumeNeed(t *testing.T){
 ctx:=context.Background();no:=false
 client:=fake.NewSimpleClientset(
  &corev1.ServiceAccount{ObjectMeta:metav1.ObjectMeta{Name:"app",Namespace:"apps"}},
  &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"app",Namespace:"apps"},Spec:corev1.PodSpec{ServiceAccountName:"app",AutomountServiceAccountToken:&no,Containers:[]corev1.Container{{Name:"app"}}}},
 )
 r,err:=EvaluateEKSServiceAccountTokenMounts(ctx,client);if err!=nil{t.Fatal(err)}
 if r.Status!="PASS"{t.Fatalf("explicit false token mount should pass with zero projections: %+v",r)}
 pod,err:=client.CoreV1().Pods("apps").Get(ctx,"app",metav1.GetOptions{});if err!=nil{t.Fatal(err)}
 pod.Spec.AutomountServiceAccountToken=nil
 _,err=client.CoreV1().Pods("apps").Update(ctx,pod,metav1.UpdateOptions{});if err!=nil{t.Fatal(err)}
 r,err=EvaluateEKSServiceAccountTokenMounts(ctx,client);if err!=nil{t.Fatal(err)}
 if r.Status!="MANUAL"||!strings.Contains(r.Details,"necessity"){t.Fatalf("token mount must require business-need review: %+v",r)}
}
