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
        self.assertEqual(counts["candidate"], 19)
        self.assertEqual(counts["not_implemented"], 28)

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

    def test_unimplemented_mapping_cannot_be_asserted_implemented(self):
        data = json.loads(json.dumps(self.ledger))
        item = next(x for x in data["controls"] if x["implementationStatus"] == "not_implemented")
        item["implementationStatus"] = "candidate"
        gaps, ready, _ = module.audit(data, self.catalog, ROOT)
        self.assertFalse(ready)
        self.assertTrue(any("catalog says otherwise" in x for x in gaps))

if __name__ == "__main__":
    unittest.main()
