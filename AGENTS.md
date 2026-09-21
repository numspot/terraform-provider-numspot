# Project-Specific Agent Guidelines

## Overview

- Purpose: NumSpot Terraform provider (`terraform-provider-numspot`), Go, Terraform Plugin Framework.
- Key files: `internal/services/<service>/` (resource/datasource implementations), `internal/core/` (API call layer), `internal/sdk/api/` (generated SDK), `internal/services/<service>/<service>_spec.json` (schema spec driving codegen).

## Lessons Learned

- **v1.7.0–v1.7.3 kubernetes_nodepool autoscaling bug**: `deserializeCreateNodePool` mapped `autoscaling.Max` from `rootDisk.Iops` and `autoscaling.Min` from `rootDisk.Size` (copy-paste). Fixed on main in `f51c65ef`, but no release tag was cut — released binaries kept sending `autoscaling.max = root_disk.iops`, causing API 400s ("autoscaling.max minimum: got 0, want 1" / "maximum: got 100, want 20") even with valid configs (see Confluence BQBqFQ).
  - When fixing generated-provider code, always add a regression test on the request-body mapping: the acceptance tests (`internal/test/`, VCR cassettes) match request bodies with exact JSON deep-equality, so a cassette replayed against a field-swapped or omitempty-regressed body fails with "requested interaction not found".
  - Verify against released tags, not just main: `git show v1.7.3:<file>` — a fix on main means nothing until released.
- A read-back after "Write file successfully" is NOT required by the tool, but for long files (2000-line read limit) prefer targeted `offset`/`limit` reads.

## Rules

- Do not commit unless explicitly asked.
- `rg` is not installed in this environment — use the grep tool or plain `grep`.
- Terraform workflows follow the terraform-engineer skill (fmt/validate/tflint before plan; approval gates for plan/apply).

## See Also

- `~/.config/opencode/AGENTS.md` (global rules: review loop, security review, English-only docs).
