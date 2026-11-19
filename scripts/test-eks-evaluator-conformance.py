#!/usr/bin/env python3
import importlib.util
import hashlib
import json
import pathlib
import subprocess
import tempfile
import unittest
from datetime import datetime, timezone

ROOT = pathlib.Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "scripts/verify-eks-evaluator-conformance.py"
spec = importlib.util.spec_from_file_location("eks_conformance_gate", SCRIPT)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)

class CISConformanceGateTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.catalog = json.loads((ROOT / "internal/audit/catalogs/eks-2.1.0.json").read_text())
        cls.ledger = json.loads((ROOT / "conformance/eks-2.1.0-evaluator-review.json").read_text())

    def test_historical_47_control_catalog_is_truthfully_unverified(self):
        gaps, ready, counts = module.audit(self.ledger, self.catalog, ROOT)
        self.assertEqual(gaps, [])
        self.assertFalse(ready)
        self.assertEqual(sum(counts.values()), 47)
        self.assertEqual(counts["candidate"], 47)
        self.assertEqual(counts["not_implemented"], 0)

    def test_fabricated_verification_is_rejected(self):
        data = json.loads(json.dumps(self.ledger))
        data["controls"][0]["implementationStatus"] = "verified"
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("review" in x or "evidence" in x or "Fixture" in x for x in gaps))

    def test_duplicate_and_removed_controls_fail(self):
        data = json.loads(json.dumps(self.ledger))
        data["controls"][-1] = data["controls"][0]
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("duplicate" in x for x in gaps))
        self.assertTrue(any("absent" in x for x in gaps))

    def test_official_audit_procedure_page_must_match_source(self):
        data = json.loads(json.dumps(self.ledger))
        data["controls"][0]["auditPage"] = 999
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("audit procedure page mismatch" in x for x in gaps))

    def test_verified_entry_requires_untampered_positive_negative_and_live_fixtures(self):
        with tempfile.TemporaryDirectory() as td:
            workspace = pathlib.Path(td)
            evidence = workspace / "conformance/evidence"
            evidence.mkdir(parents=True)
            data = json.loads(json.dumps(self.ledger))
            row = data["controls"][0]
            row["implementationStatus"] = "verified"
            row["auditProcedureReviewed"] = True
            row["reviewer"] = "independent-test-reviewer"
            row["reviewedAt"] = datetime.now(timezone.utc).isoformat()
            paths = []
            for key in ("positiveFixture", "negativeFixture", "runtimeEvidence"):
                relative = "conformance/evidence/" + key + ".json"
                path = workspace / relative
                path.write_text('{"kind":"test-fixture","source":"synthetic"}')
                row[key] = relative
                row[key + "Sha256"] = hashlib.sha256(path.read_bytes()).hexdigest()
                paths.append(path)
            gaps, ready, _ = module.audit(data, self.catalog, workspace)
            self.assertEqual(gaps, [])
            self.assertFalse(ready)  # Another 46 controls still need verification
            paths[2].write_text('{"tampered":true}')
            gaps, ready, _ = module.audit(data, self.catalog, workspace)
            self.assertFalse(ready)
            self.assertTrue(any("content changed" in g for g in gaps))

    def test_independently_pinned_approval_signature_rejects_tampering(self):
        with tempfile.TemporaryDirectory() as td:
            temp = pathlib.Path(td)
            key = temp / "signing-private.pem"
            pub = temp / "approved-public.pem"
            ledger = temp / "ledger.json"
            sig = temp / "ledger.sig"
            ledger.write_text(json.dumps(self.ledger, sort_keys=True))
            for args in (
                ["openssl", "genpkey", "-algorithm", "RSA",
                 "-pkeyopt", "rsa_keygen_bits:2048", "-out", str(key)],
                ["openssl", "pkey", "-in", str(key), "-pubout", "-out", str(pub)],
                ["openssl", "dgst", "-sha256", "-sign", str(key),
                 "-out", str(sig), str(ledger)],
            ):
                subprocess.run(args, check=True, capture_output=True)
            pin = hashlib.sha256(pub.read_bytes()).hexdigest()
            valid, _ = module.verify_approval_signature(ledger, pub, pin, sig)
            self.assertTrue(valid)
            ledger.write_text(ledger.read_text() + " ")
            valid, _ = module.verify_approval_signature(ledger, pub, pin, sig)
            self.assertFalse(valid)

    def test_candidate_cannot_exceed_actual_source_implementation(self):
        catalog = json.loads(json.dumps(self.catalog))
        catalog["controls"][0]["implemented"] = False
        gaps, ready, _ = module.audit(self.ledger, catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("catalog says otherwise" in x for x in gaps))

if __name__ == "__main__":
    unittest.main()
