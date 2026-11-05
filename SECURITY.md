# Security Policy

## Supported versions

Only the latest published release of the provider receives security fixes. We
recommend always running the most recent version.

## Reporting a vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Instead, report them privately using one of the following channels:

1. GitHub's private vulnerability reporting:
   https://github.com/numspot/terraform-provider-numspot/security/advisories
2. Email: [security@numspot.com](mailto:security@numspot.com)

Please include as much of the following information as you can:

- The type of issue (e.g. credential exposure, injection, privilege escalation)
- Full paths of the source file(s) related to the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code, if available
- Impact of the issue, including how an attacker might exploit it

## Response process

The maintainers will acknowledge your report, triage it, and keep you informed
of the progress toward a fix and an announcement. If the issue is confirmed,
we will release a patch as soon as possible depending on complexity, and credit
reporters unless they prefer to remain anonymous.

## Scope

This provider runs on the machine executing Terraform and stores Numspot
service-account credentials in the Terraform state. When reporting, consider
issues such as credential leakage in logs or state, injection through
configuration values, and unsafe handling of API responses.
