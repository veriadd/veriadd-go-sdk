# Changelog

All notable changes to the `go-veriadd` module are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## 0.1.0 — 2026-10-11

### Added
- `VeriaddClient` covering the full Veriadd REST surface: address verification,
  postcode lookup, autocomplete, nearby and reverse search, assembly,
  wallet and usage, Bachs top-up init/verify, KYB get/submit, API key
  management, and live service status.
- Typed models, `*VeriaddError` parsed from the API error envelope.
- `DoWithRetry` with exponential backoff on transient codes and
  context-cancellation support.
- Functional options (`WithBaseURL`, `WithTimeout`, `WithHTTPClient`),
  per-request timeouts and `X-Request-ID` correlation headers.
