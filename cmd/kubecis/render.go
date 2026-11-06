package main

import (
 "fmt"
 "io"
 "os"
 "path/filepath"
 "sort"
 "strings"

 "github.com/sm-joe/kube-cis-audit/internal/audit"
 "github.com/spf13/cobra"
)

// newRenderCommand formats a previously collected report. It does not re-evaluate
// controls and never promotes missing evidence to a PASS or FAIL.
func newRenderCommand() *cobra.Command {
 var input, output string
 cmd := &cobra.Command{
  Use: "render",
  Short: "Render a KubeCIS JSON report as a kube-bench-style text report",
  RunE: func(cmd *cobra.Command, args []string) error {
   data, err := os.ReadFile(input)
   if err != nil { return fmt.Errorf("read report: %w", err) }
   report, err := audit.DecodeReport(data)
   if err != nil { return fmt.Errorf("decode report: %w", err) }
   var dst io.Writer = os.Stdout
   if output != "" {
    if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil { return err }
    f, err := os.Create(output)
    if err != nil { return err }
    defer f.Close()
    dst = f
   }
   return renderBenchStyle(dst, report)
  },
 }
 cmd.Flags().StringVar(&input,"input","","Input consolidated or node/cluster KubeCIS JSON report")
 cmd.Flags().StringVar(&output,"output-file","","Save human-readable output (default stdout)")
 _ = cmd.MarkFlagRequired("input")
 return cmd
}

type renderCounts struct { pass, fail, warn, manual, unknown, info int }
func (s *renderCounts) add(status string) {
 switch status {
 case "PASS": s.pass++
 case "FAIL": s.fail++
 case "WARN": s.warn++
 case "MANUAL": s.manual++
 case "NOT_ASSESSED": s.unknown++
 default: s.info++
 }
}
func (s renderCounts) write(w io.Writer) {
 fmt.Fprintf(w, "%d checks PASS\n%d checks FAIL\n%d checks WARN\n%d checks MANUAL\n%d checks NOT_ASSESSED\n%d checks INFO\n",s.pass,s.fail,s.warn,s.manual,s.unknown,s.info)
}
func renderSection(id, target string) string {
 parts:=strings.Split(id,".")
 if len(parts)>=2 && parts[0]!="" && parts[1]!="" {
  return parts[0]+"."+parts[1]
 }
 if target!="" {return target}
 return "other"
}
func renderBenchStyle(w io.Writer, report audit.Report) error {
 if _, err:=fmt.Fprintf(w,"KubeCIS Native — benchmark assessment\nBuild: %s\nBenchmark: %s (%s)\nKubernetes: %s\nProvider: %s\nGenerated: %s\n",report.BuildVersion,report.Benchmark,report.CompatibilityStatus,report.ClusterVersion,report.CloudProvider,report.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"));err!=nil{return err}
 if report.BenchmarkProfileCoverageNote!="" {fmt.Fprintf(w,"Coverage note: %s\n",report.BenchmarkProfileCoverageNote)}
 if report.CompatibilityStatus!="verified" {
  fmt.Fprintln(w,"[INFO] RELEASE GATE: official compliance score withheld; profile is not fully verified.")
 }
 // Sort by numeric CIS identifiers, with stable node scope ordering. Never
 // deduplicate different node findings that share the same control ID.
 results:=append([]audit.Result(nil),report.Results...)
 sort.SliceStable(results,func(i,j int) bool{
  ai,aj:=strings.Split(results[i].ID,"."),strings.Split(results[j].ID,".")
  for k:=0;k<len(ai)&&k<len(aj);k++ {
   var ni,nj int
   _,ei:=fmt.Sscanf(ai[k],"%d",&ni)
   _,ej:=fmt.Sscanf(aj[k],"%d",&nj)
   if ei==nil&&ej==nil&&ni!=nj {return ni<nj}
   if ai[k]!=aj[k] {return ai[k]<aj[k]}
  }
  if len(ai)!=len(aj){return len(ai)<len(aj)}
  return results[i].Scope<results[j].Scope
 })
 sections:=[]string{}
 grouped:=map[string][]audit.Result{}
 for _,r:=range results {
  section:=renderSection(r.ID,r.Target)
  if _,ok:=grouped[section];!ok {sections=append(sections,section)}
  grouped[section]=append(grouped[section],r)
 }
 var total renderCounts
 for _,section:=range sections {
  fmt.Fprintf(w,"\n[INFO] %s\n",section)
  var counts renderCounts
  for _,r:=range grouped[section] {
   counts.add(r.Status);total.add(r.Status)
   name:=r.Title
   if name=="" {name="Control "+r.ID}
   scope:=""
   if r.Scope!="" {scope=" [node: "+r.Scope+"]"}
   fmt.Fprintf(w,"[%s] %s %s%s\n",r.Status,r.ID,name,scope)
   if r.Status=="NOT_ASSESSED" || r.Status=="MANUAL" {
    if r.Details!="" {fmt.Fprintf(w,"  Evidence/review: %s\n",strings.ReplaceAll(r.Details,"\n"," "))}
   }
  }
  fmt.Fprintf(w,"\n== Remediations %s ==\n",section)
  remediationCount:=0
  for _,r:=range grouped[section] {
   if r.Status=="PASS"||strings.TrimSpace(r.Remediation)=="" {continue}
   remediationCount++
   fmt.Fprintf(w,"%s",r.ID)
   if r.Scope!="" {fmt.Fprintf(w," [%s]",r.Scope)}
   fmt.Fprintf(w," %s\n",r.Remediation)
  }
  if remediationCount==0 {fmt.Fprintln(w,"No remediation text available for this section.")}
  fmt.Fprintf(w,"\n== Summary %s ==\n",section)
  counts.write(w)
 }
 fmt.Fprintln(w,"\n== Summary total ==")
 total.write(w)
 if total.unknown>0 || total.manual>0 || report.CompatibilityStatus!="verified" {
  fmt.Fprintln(w,"Compliance score: WITHHELD (incomplete assessment and/or review; output is not proof of compliance)")
 }
 return nil
}
