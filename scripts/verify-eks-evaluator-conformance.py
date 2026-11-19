#!/usr/bin/env python3
"""Fail-closed CIS EKS evaluator verification ledger.

This does not redistribute the CIS PDF and never converts recommendation
identity coverage into an official compliance certification.
"""
import argparse
import datetime
import json
import pathlib
import re
import sys

SOURCE_SHA = "c3c9eb01253643a0176b21aafd19cc7c08af0a19872f869712d2d82be8a967e9"
PATH = pathlib.Path("conformance/eks-2.1.0-evaluator-review.json")
CATALOG = pathlib.Path("internal/audit/catalogs/eks-2.1.0.json")
VALID = {"not_implemented", "candidate", "verified"}

def audit(ledger, catalog, repo):
    problems = []
    if ledger.get("sourceSha256") != SOURCE_SHA or ledger.get("benchmark") != "cis-eks-2.1.0":
        problems.append("benchmark source identity mismatch")
    if ledger.get("certificationClaim") is not False:
        problems.append("unproven certification claim")
    expected = {x["id"]: x for x in catalog["controls"]}
    seen = set()
    for c in ledger.get("controls", []):
        control = c.get("id")
        if control not in expected or control in seen:
            problems.append(f"{control}: unknown or duplicate control")
            continue
        seen.add(control)
        canonical = expected[control]
        if c.get("assessment") != canonical["assessment"] or c.get("sourcePage") != canonical["pdfPage"]:
            problems.append(f"{control}: source classification/provenance mismatch")
        status = c.get("implementationStatus")
        if status not in VALID:
            problems.append(f"{control}: unsupported state")
            continue
        if not canonical["implemented"] and status != "not_implemented":
            problems.append(f"{control}: marked implemented while catalog says otherwise")
        if canonical["implemented"] and status == "not_implemented":
            problems.append(f"{control}: stale implementation inventory")
        if status == "verified":
            if c.get("auditProcedureReviewed") is not True:
                problems.append(f"{control}: source audit procedure not reviewed")
            if not c.get("reviewer") or not c.get("reviewedAt"):
                problems.append(f"{control}: reviewer signature metadata absent")
            else:
                try:
                    stamp = datetime.datetime.fromisoformat(c["reviewedAt"].replace("Z","+00:00"))
                    if stamp.tzinfo is None or stamp > datetime.datetime.now(datetime.timezone.utc):
                        raise ValueError("invalid review time")
                except (ValueError, TypeError):
                    problems.append(f"{control}: review timestamp invalid")
            for field in ("positiveFixture", "negativeFixture", "runtimeEvidence"):
                val = c.get(field)
                if not isinstance(val, str) or not val.startswith("conformance/evidence/"):
                    problems.append(f"{control}: missing approved {field} file")
                elif not (repo / val).is_file():
                    problems.append(f"{control}: {field} reference not found")
        elif status == "candidate" and c.get("auditProcedureReviewed") is True:
            problems.append(f"{control}: source-reviewed candidate should not bypass verification")
    missing = set(expected) - seen
    problems += [f"{id}: absent from ledger" for id in sorted(missing)]
    statuses = [c.get("implementationStatus") for c in ledger.get("controls", [])]
    ready = not problems and len(seen) == 47 and all(s == "verified" for s in statuses)
    return problems, ready, {s: statuses.count(s) for s in sorted(VALID)}

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--ledger", type=pathlib.Path, default=PATH)
    p.add_argument("--catalog", type=pathlib.Path, default=CATALOG)
    p.add_argument("--require-full", action="store_true")
    a = p.parse_args()
    ledger=json.loads(a.ledger.read_text())
    catalog=json.loads(a.catalog.read_text())
    problems, ready, counts=audit(ledger,catalog,pathlib.Path("."))
    print(json.dumps({"benchmark":"cis-eks-2.1.0","controls":len(ledger.get("controls",[])),
        "statuses":counts,"releaseReady":ready,"structuralProblems":problems},indent=2))
    if problems or (a.require_full and not ready):
        return 1
    return 0

if __name__=="__main__":
    sys.exit(main())
