#!/usr/bin/env python3
"""Fail-closed CIS EKS evaluator verification ledger.

This does not redistribute the CIS PDF and never converts recommendation
identity coverage into an official compliance certification.
"""
import argparse
import datetime
import hashlib
import json
import pathlib
import re
import subprocess
import sys

SOURCE_SHA = "c3c9eb01253643a0176b21aafd19cc7c08af0a19872f869712d2d82be8a967e9"
PATH = pathlib.Path("conformance/eks-2.1.0-evaluator-review.json")
CATALOG = pathlib.Path("internal/audit/catalogs/eks-2.1.0.json")
VALID = {"not_implemented", "candidate", "verified"}
# Factual page numbers from the privately supplied official v2.1.0 table of
# contents, not redistributions of its proprietary audit procedures.
AUDIT_PAGES = dict(
    (entry.split(":")[0], int(entry.split(":")[1]))
    for entry in """
    2.1.1:18 3.1.1:23 3.1.2:27 3.1.3:31 3.1.4:35
    3.2.1:40 3.2.2:44 3.2.3:48 3.2.4:53 3.2.5:57 3.2.6:61
    3.2.7:66 3.2.8:71
    4.1.1:79 4.1.2:83 4.1.3:87 4.1.4:92 4.1.5:96 4.1.6:100
    4.1.7:104 4.1.8:110 4.1.9:116 4.1.10:120 4.1.11:124 4.1.12:128
    4.2.1:133 4.2.2:137 4.2.3:141 4.2.4:145 4.2.5:149
    4.3.1:154 4.3.2:158 4.4.1:162 4.4.2:165
    4.5.1:169 4.5.2:173
    5.1.1:178 5.1.2:182 5.1.3:187 5.1.4:193
    5.2.1:198 5.3.1:203
    5.4.1:208 5.4.2:212 5.4.3:215 5.4.4:219
    5.5.1:223
    """.split()
)

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
        if c.get("auditPage") != AUDIT_PAGES.get(control):
            problems.append(f"{control}: official audit procedure page mismatch")
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
            evidence_root = (repo / "conformance/evidence").resolve()
            for field in ("positiveFixture", "negativeFixture", "runtimeEvidence"):
                val = c.get(field)
                digest = c.get(field + "Sha256")
                if not isinstance(val, str) or not val.startswith("conformance/evidence/"):
                    problems.append(f"{control}: missing approved {field} file")
                    continue
                evidence = repo / val
                if not evidence.resolve().is_relative_to(evidence_root) or evidence.is_symlink() or not evidence.is_file():
                    problems.append(f"{control}: unsafe/missing {field} reference")
                elif not isinstance(digest, str) or not re.fullmatch(r"[a-f0-9]{64}", digest):
                    problems.append(f"{control}: {field} content SHA256 missing")
                elif hashlib.sha256(evidence.read_bytes()).hexdigest() != digest:
                    problems.append(f"{control}: {field} content changed after review")
        elif status == "candidate" and c.get("auditProcedureReviewed") is True:
            problems.append(f"{control}: source-reviewed candidate should not bypass verification")
    missing = set(expected) - seen
    problems += [f"{id}: absent from ledger" for id in sorted(missing)]
    statuses = [c.get("implementationStatus") for c in ledger.get("controls", [])]
    ready = not problems and len(seen) == 47 and all(s == "verified" for s in statuses)
    return problems, ready, {s: statuses.count(s) for s in sorted(VALID)}

def evaluator_digest(repo):
    """Bind a signed review ledger to the actual evaluator and collection sources."""
    digest = hashlib.sha256()
    paths = []
    for directory in ("internal", "cmd", "benchmarks", "scripts", "deploy", "conformance"):
        paths.extend(p for p in (repo / directory).rglob("*")
                     if p.is_file() and p.suffix in {".go", ".yaml", ".yml", ".sh", ".py", ".json"}
                     and not p.relative_to(repo).as_posix().startswith("conformance/evidence/")
                     and p.relative_to(repo).as_posix() != str(PATH))
    paths.extend(repo / name for name in ("go.mod", "go.sum", "Dockerfile"))
    for path in sorted(set(paths)):
        relative = path.relative_to(repo).as_posix()
        digest.update(relative.encode() + b"\\0" + hashlib.sha256(path.read_bytes()).digest())
    return digest.hexdigest()


def verify_approval_signature(ledger_path, public_key_path, key_pin, signature_path):
    """Require independent, explicitly pinned approval of the exact ledger bytes."""
    if not public_key_path or not key_pin or not signature_path:
        return False, "missing independently approved release signing key, pin or ledger signature"
    if not re.fullmatch(r"[a-f0-9]{64}", key_pin):
        return False, "invalid independent approval key fingerprint"
    key = pathlib.Path(public_key_path)
    sig = pathlib.Path(signature_path)
    if (not key.is_file() or key.is_symlink() or key.stat().st_size > 8192 or
            not sig.is_file() or sig.is_symlink() or sig.stat().st_size > 8192):
        return False, "approval key/signature must be bounded regular files"
    if hashlib.sha256(key.read_bytes()).hexdigest() != key_pin:
        return False, "independent reviewer key pin mismatch"
    try:
        proc = subprocess.run(
            ["openssl", "dgst", "-sha256", "-verify", str(key),
             "-signature", str(sig), str(ledger_path)],
            check=False, capture_output=True, timeout=15,
        )
    except (OSError, subprocess.TimeoutExpired):
        return False, "independent signed approval could not be verified"
    return (proc.returncode == 0, "verified" if proc.returncode == 0 else "invalid independent ledger approval signature")


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--ledger", type=pathlib.Path, default=PATH)
    p.add_argument("--catalog", type=pathlib.Path, default=CATALOG)
    p.add_argument("--require-full", action="store_true")
    p.add_argument("--print-evaluator-digest", action="store_true")
    p.add_argument("--approval-public-key", type=pathlib.Path)
    p.add_argument("--approval-key-sha256")
    p.add_argument("--approval-signature", type=pathlib.Path)
    a = p.parse_args()
    if a.print_evaluator_digest:
        print(evaluator_digest(pathlib.Path(".")))
        return 0
    ledger=json.loads(a.ledger.read_text())
    catalog=json.loads(a.catalog.read_text())
    problems, ready, counts=audit(ledger,catalog,pathlib.Path("."))
    if a.require_full:
        if ledger.get("evaluatorSourceSha256") != evaluator_digest(pathlib.Path(".")):
            problems.append("signed ledger is not bound to the current evaluator sources")
            ready = False
        signature_ok, reason = verify_approval_signature(
            a.ledger, a.approval_public_key, a.approval_key_sha256,
            a.approval_signature,
        )
        if not signature_ok:
            problems.append("independent reviewer approval: " + reason)
        ready = ready and signature_ok
    print(json.dumps({"benchmark":"cis-eks-2.1.0","controls":len(ledger.get("controls",[])),
        "statuses":counts,"releaseReady":ready,"structuralProblems":problems},indent=2))
    if problems or (a.require_full and not ready):
        return 1
    return 0

if __name__=="__main__":
    sys.exit(main())
