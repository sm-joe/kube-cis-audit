package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sm-joe/kubecis-b1-upstream/upstream"
	"os"
)

func main() {
	config := flag.String("config", "", "path to kube-bench cfg/config.yaml")
	file := flag.String("file", "", "path to one target benchmark YAML")
	version := flag.String("k8s-version", "", "Kubernetes version for explicit profile resolution")
	distribution := flag.String("distribution", "", "optional distribution selector")
	flag.Parse()
	failed := false
	if *config != "" {
		c, e := upstream.LoadConfig(*config)
		if e != nil {
			fatal(e)
		}
		if *version != "" {
			s, e := upstream.ResolveProfile(c, *version, *distribution)
			if e != nil {
				fmt.Fprintln(os.Stderr, "PROFILE ERROR:", e)
				failed = true
			} else {
				fmt.Printf("Resolved profile: %s (%s)\n", s.Profile, s.MatchedBy)
			}
		} else {
			fmt.Println("Config YAML parsed successfully; pass -k8s-version to resolve a profile.")
		}
	}
	if *file != "" {
		b, e := upstream.LoadBenchmark(*file)
		if e != nil {
			fatal(e)
		}
		issues := upstream.ValidateBenchmark(b)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(issues)
		if len(issues) > 0 {
			failed = true
		} else {
			fmt.Printf("Validated %d top-level groups.\n", len(b.Groups))
		}
	}
	if *config == "" && *file == "" {
		flag.Usage()
		os.Exit(2)
	}
	if failed {
		os.Exit(1)
	}
}
func fatal(e error) { fmt.Fprintln(os.Stderr, "ERROR:", e); os.Exit(2) }
