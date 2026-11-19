package audit

import (
 "context"
 "strings"
 "testing"

 appsv1 "k8s.io/api/apps/v1"
 corev1 "k8s.io/api/core/v1"
 metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
 "k8s.io/client-go/kubernetes/fake"
)

func TestEKSRecentScanDefaultServiceAccountPlatformRequiresReview(t *testing.T) {
 ctx:=context.Background()
 platform:=fake.NewSimpleClientset(
  &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"eks-pod-identity-agent-4qr74",Namespace:"kube-system"},Spec:corev1.PodSpec{ServiceAccountName:"default"}},
  &corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"eks-pod-identity-agent-6wtfw",Namespace:"kube-system"},Spec:corev1.PodSpec{ServiceAccountName:"default"}},
  &appsv1.DaemonSet{ObjectMeta:metav1.ObjectMeta{Name:"eks-pod-identity-agent",Namespace:"kube-system"},Spec:appsv1.DaemonSetSpec{Template:corev1.PodTemplateSpec{Spec:corev1.PodSpec{}}}},
 )
 r,err:=EvaluateEKSDefaultServiceAccountUse(ctx,platform)
 if err!=nil {t.Fatal(err)}
 if r.Status!="MANUAL" || len(r.Evidence)!=3 || !strings.Contains(r.Details,"duplicates") {
  t.Fatalf("platform references should be evidence for review, not auto-exempt or auto-FAIL: %+v",r)
 }
 // An application in another namespace remains a concrete deviation.
 application:=fake.NewSimpleClientset(&corev1.Pod{ObjectMeta:metav1.ObjectMeta{Name:"user-app",Namespace:"apps"},Spec:corev1.PodSpec{ServiceAccountName:"default"}})
 r,err=EvaluateEKSDefaultServiceAccountUse(ctx,application)
 if err!=nil {t.Fatal(err)}
 if r.Status!="FAIL" {t.Fatalf("app default SA must remain FAIL: %+v",r)}
}

func TestEKSRecentScanSecretEnvEvidencePreserved(t *testing.T) {
 ctx:=context.Background()
 newPod:=func(ns,name string)*corev1.Pod{return &corev1.Pod{
  ObjectMeta:metav1.ObjectMeta{Name:name,Namespace:ns},
  Spec:corev1.PodSpec{Containers:[]corev1.Container{{Name:"ebs-plugin",Env:[]corev1.EnvVar{
   {Name:"AWS_ACCESS_KEY_ID",ValueFrom:&corev1.EnvVarSource{SecretKeyRef:&corev1.SecretKeySelector{LocalObjectReference:corev1.LocalObjectReference{Name:"aws-creds"},Key:"access"}}},
   {Name:"AWS_SECRET_ACCESS_KEY",ValueFrom:&corev1.EnvVarSource{SecretKeyRef:&corev1.SecretKeySelector{LocalObjectReference:corev1.LocalObjectReference{Name:"aws-creds"},Key:"secret"}}},
  }}}}}
 }
 check:=Check{ID:"4.4.1",Title:"Review Secret environment references",Kind:"eks_secret_env",Target:"policies",Scored:true}
 platform:=fake.NewSimpleClientset(newPod("kube-system","ebs-csi-controller-a"),newPod("kube-system","ebs-csi-controller-b"))
 r:=EvaluateEKSAdditionalPolicy(ctx,platform,check)
 if r.Status!="MANUAL" || len(r.Evidence)!=4 || !strings.Contains(r.Details,"4 active Pod") {
  t.Fatalf("platform evidence must not become an unsupported compromise claim: %+v",r)
 }
 app:=fake.NewSimpleClientset(newPod("apps","user-app"))
 r=EvaluateEKSAdditionalPolicy(ctx,app,check)
 if r.Status!="FAIL" || len(r.Evidence)!=2 {t.Fatalf("application secret environment references must remain FAIL: %+v",r)}
 empty:=fake.NewSimpleClientset()
 r=EvaluateEKSAdditionalPolicy(ctx,empty,check)
 if r.Status!="PASS" {t.Fatalf("no observed references must PASS: %+v",r)}
}
