# Milestone 2.3 — Admission Policy Assessment

## Step 2.3.1 — Admission mechanism discovery

This increment improves the existing `api_policy_control` evaluator. It does not add a new benchmark check or claim that discovered configuration is effective enforcement.

### Inventory sources

- Pod Security Admission namespace labels: `pod-security.kubernetes.io/enforce` values `baseline` and `restricted`.
- ValidatingAdmissionPolicy objects and their bindings.
- ValidatingWebhookConfiguration objects and individual validating webhooks.

The existing ClusterRole already grants read-only `get`/`list` access to these admissionregistration resources; no RBAC expansion is required for this increment.

### Matching semantics

- A webhook is considered a candidate only when its admission rules can match core/v1 Pod `CREATE` or `UPDATE` requests. The configuration name is not used to identify Kyverno, Gatekeeper, or another vendor.
- A ValidatingAdmissionPolicy candidate requires at least one validation, a policy match rule that can match Pod `CREATE`/`UPDATE`, an existing binding, and a binding scope that does not rule out Pods.
- `Deny` is treated as a potentially blocking ValidatingAdmissionPolicy action. `Warn` and `Audit` are recorded as advisory/non-blocking actions.
- Resource exclusions are considered when determining whether a VAP policy/binding can match Pods.
- Namespace selectors, object selectors, match conditions, exclusions, and webhook endpoint behaviour are recorded as items requiring further review; this inventory does not evaluate CEL expressions or prove live endpoint health.
- Webhook `failurePolicy` is recorded, but it is not used to infer that the webhook cannot deny a request. It controls handling of webhook call errors, not an ordinary denial response. A webhook configured with `Ignore` can still reject a request when it is called and returns a denial.
- VAP `failurePolicy` is recorded separately from binding `validationActions`: the former controls policy evaluation/configuration errors, while `Deny`, `Warn`, and `Audit` determine how a failed validation is enforced.
- API inventory failures produce `NOT_ASSESSED` rather than silently ignoring the missing source.

### Result semantics

- `MANUAL`: full PSA enforce-label coverage or at least one Pod-matching potentially blocking VAP/webhook candidate was found. Effective enforcement still requires manual review.
- `WARN`: partial PSA coverage or VAP bindings with only `Warn`/`Audit` actions were found, without a discovered blocking candidate.
- `NOT_ASSESSED`: no candidate was discovered, or required API inventory was incomplete. This is not proof that no provider-managed or external policy mechanism exists.

### Tests added

- VAP with Pod CREATE/UPDATE rules and a Deny binding is identified as a manual-review candidate.
- Webhook name alone does not qualify a webhook; actual Pod rules are required.
- A validating webhook with Pod CREATE/UPDATE rules is identified regardless of vendor naming.
- VAP Warn/Audit actions are not classified as blocking enforcement.
- VAP Pod exclusions take precedence over matching include rules.

### Known limitations for subsequent steps

This increment does not evaluate CEL expression meaning, VAP status/type-check conditions, selector truth against all namespace/Pod labels, parameter resolution, webhook service reachability, TLS validity, webhook response behaviour, or API-server admission plugin configuration/exemptions. Those belong to subsequent Milestone 2.3 steps. Results remain `MANUAL` where a candidate exists; candidate discovery is not a compliance pass.

## Validation

Run with the repository's required Go toolchain (Go 1.26.0 or later):

```powershell
gofmt -w .\internal\audit\admission.go .\internal\audit\runner.go .\internal\audit\runner_test.go
go test ./internal/audit
go test ./...
```
