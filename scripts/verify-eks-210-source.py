#!/usr/bin/env python3
"""Validate the factual EKS 2.1.0 control inventory against the licensed PDF.
The official document must remain on the operator's machine and is NOT
redistributed. This gate does NOT certify evaluator semantic correctness.
"""
import argparse
import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

PINNED_SHA256 = "c3c9eb01253643a0176b21aafd19cc7c08af0a19872f869712d2d82be8a967e9"
MANUAL = set("4.1.1 4.1.2 4.1.3 4.1.4 4.1.8 4.1.9 4.1.10 4.1.11 4.1.12 4.3.2 4.4.2 4.5.1 5.1.2 5.1.3 5.1.4 5.3.1 5.4.3 5.4.4".split())
PROVIDER = set("2.1.1 4.1.7 5.1.1 5.4.1 5.4.2 5.5.1".split())

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--pdf", required=True, type=Path, help="privately held official CIS EKS v2.1.0 PDF")
    p.add_argument("--catalog", default=Path("internal/audit/catalogs/eks-2.1.0.json"), type=Path)
    p.add_argument("--benchmark", default=Path("benchmarks/cis-eks-2.1.0/benchmark.yaml"), type=Path)
    a = p.parse_args()
    if hashlib.sha256(a.pdf.read_bytes()).hexdigest() != PINNED_SHA256:
        sys.exit("FAIL: unknown PDF revision; pin and verify the source independently")
    text = subprocess.check_output(["pdftotext", "-layout", str(a.pdf), "-"], text=True)
    marker = "\fAppendix: Summary Table"
    if marker not in text:
        sys.exit("FAIL: official summary appendix not found")
    appendix = text.split(marker, 1)[1].split("\fAppendix: CIS Controls", 1)[0]
    ids = re.findall(r"(?m)^\s*([2-5]\.\d+\.\d+)\s{2,}", appendix)
    if len(ids) != 47 or len(set(ids)) != 47:
        sys.exit("FAIL: official appendix must have 47 unique recommendation IDs")
    catalog = json.loads(a.catalog.read_text(encoding="utf-8"))
    control_ids = [row["id"] for row in catalog["controls"]]
    if len(control_ids) != 47 or len(set(control_ids)) != 47 or set(control_ids) != set(ids):
        sys.exit("FAIL: repository catalog mismatches independently pinned official appendix")
    for row in catalog["controls"]:
        expected = "manual" if row["id"] in MANUAL else "automated"
        if row["assessment"] != expected:
            sys.exit("FAIL: recommendation type mismatch for " + row["id"])
    profile = a.benchmark.read_text(encoding="utf-8")
    local_ids = re.findall(r'^\s+- id: "([2-5]\.\d+\.\d+)"$', profile, re.M)
    if len(local_ids) != 41 or len(set(local_ids)) != 41 or set(local_ids) != set(ids) - PROVIDER:
        sys.exit("FAIL: benchmark profile does not include exactly 41 non-provider controls")
    if "compatibilityStatus: provisional" not in profile:
        sys.exit("FAIL: unverified evaluator coverage must not be certified")
    print("PASS: licensed source inventory: 47 exact EKS IDs/types, 41 local + 6 provider mappings")
    print("NOT VERIFIED: evaluator semantics, manual reviews, supported runtime matrix or CIS certification")

if __name__ == "__main__":
    main()
