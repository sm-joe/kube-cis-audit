#!/usr/bin/env python3
import importlib.util
import json
import pathlib
import unittest

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

    def test_missing_and_mismatched_runtime_route_fail(self):
        data = json.loads(json.dumps(self.ledger))
        data["controls"][0]["runtimeRoute"] = "fake_evaluator"
        routes = module.load_profile_routes(ROOT / "benchmarks/cis-eks-2.1.0/benchmark.yaml")
        problems, ready, _ = module.audit(data, self.catalog, ROOT, routes)
        self.assertFalse(ready)
        self.assertTrue(any("runtime evaluator route mismatch" in x for x in problems))

    def test_every_profile_control_has_named_runtime_route(self):
        routes = module.load_profile_routes(ROOT / "benchmarks/cis-eks-2.1.0/benchmark.yaml")
        self.assertEqual(len(routes), 41)
        problems, ready, counts = module.audit(self.ledger, self.catalog, ROOT, routes)
        self.assertEqual(problems, [])
        self.assertEqual(sum(counts.values()), 47)
        self.assertFalse(ready)

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

    def test_candidate_cannot_be_labelled_unimplemented_or_verified_without_proof(self):
        data = json.loads(json.dumps(self.ledger))
        data["controls"][0]["implementationStatus"] = "not_implemented"
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("stale implementation" in x for x in gaps))
        data["controls"][0]["implementationStatus"] = "verified"
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("audit procedure not reviewed" in x for x in gaps))

if __name__ == "__main__":
    unittest.main()
