#!/usr/bin/env bash
# Verify independently produced third-party scan evidence against an operator-
# pinned public key. No unsigned JSON may be promoted to CIS evidence.
set -Eeuo pipefail
die(){ echo "ATTESTATION: $*" >&2; exit 1; }
dir='' key='' pinned='' registry='' repository='' digest='' outfile=''
while (($#)); do
 case "$1" in
 --dir|--public-key|--key-sha256|--registry|--repository|--digest|--output-file)
  (($#>=2)) || die "Missing $1"
  case "$1" in
   --dir) dir=$2;; --public-key) key=$2;; --key-sha256) pinned=$2;;
   --registry) registry=$2;; --repository) repository=$2;; --digest) digest=$2;;
   --output-file) outfile=$2;;
  esac; shift 2;;
 *) die "Unknown option $1";;
 esac
done
[[ -n "$dir" && -n "$key" && -n "$pinned" && -n "$registry" && -n "$repository" && -n "$digest" && -n "$outfile" ]] || die 'Missing required input'
for bin in jq openssl sha256sum date stat; do command -v "$bin" >/dev/null || die "$bin is required"; done
[[ "$digest" =~ ^sha256:([a-f0-9]{64})$ ]] || die 'Invalid deployed digest'
hex="${BASH_REMATCH[1]}"
[[ "$pinned" =~ ^[a-f0-9]{64}$ ]] || die 'Invalid signer key pin'
[[ -f "$key" && ! -L "$key" && -d "$dir" && ! -L "$dir" ]] || die 'Trusted key or attestation directory missing/unsafe'
actual="$(sha256sum -- "$key" | cut -d' ' -f1)"
[[ "$actual" == "$pinned" ]] || die 'Public key pin mismatch'
payload="$dir/$hex.json"
sig="$dir/$hex.sig"
[[ -f "$payload" && ! -L "$payload" && -f "$sig" && ! -L "$sig" ]] || die 'No signed claim for deployed digest'
(( $(stat -c %s -- "$payload") <= 65536 && $(stat -c %s -- "$sig") <= 8192 )) || die 'Signed claim exceeds limits'
openssl dgst -sha256 -verify "$key" -signature "$sig" "$payload" >/dev/null 2>&1 || die 'Invalid detached signature'
jq -e --arg registry "$registry" --arg repository "$repository" --arg digest "$digest" '
 type == "object" and
 (.schemaVersion == 1) and
 (.subject.registry == $registry) and
 (.subject.repository == $repository) and
 (.subject.digest == $digest) and
 (.scanner | type == "string" and length > 0 and length <= 128) and
 (.scanId | type == "string" and length > 0 and length <= 256) and
 (.status == "COMPLETE") and
 (.findingsAvailable == true) and
 (.scannedAt | type == "string") and
 (.scannedDigests | type == "array" and length > 0 and length <= 8) and
 all(.scannedDigests[]; type == "string" and test("^sha256:[a-f0-9]{64}$")) and
 ([.scannedDigests[]] | unique | length) == (.scannedDigests | length)
' "$payload" >/dev/null || die 'Signed claim missing required matching scan details'
scanned="$(jq -r '.scannedAt' "$payload")"
timestamp="$(date -u -d "$scanned" +%s 2>/dev/null)" || die 'Invalid scan date'
now="$(date -u +%s)"
(( timestamp <= now + 120 && timestamp >= now - 30*86400 )) || die 'Scan older than 30 days or future dated'
jq -n --arg digest "$digest" --arg signer "$actual" --arg scannedAt "$scanned" \
 --arg scanner "$(jq -r '.scanner' "$payload")" \
 --argjson scanDigests "$(jq -c '.scannedDigests' "$payload")" \
 '{digest:$digest,signerKeySha256:$signer,scanner:$scanner,scannedAt:$scannedAt,scanDigests:$scanDigests,verified:true}' > "$outfile"
