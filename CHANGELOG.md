# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.3] - 2026-08-15

Patch release.

### Changed
- Depends on `ai` v1.2.0, which adds the `Capabilities.Images` hint. This driver
  does not generate images, so it reports `Images: false` (the default) - no
  behavior change.

## [1.1.1] - 2026-08-11

Patch release.

### Documentation
- The reference covers the driver's `Capabilities` and the translation of
  model-level 400 refusals into `ai.ErrNoHosted`/`ai.ErrNoFormat`, which until
  now lived only in the godoc.

## [1.1.0] - 2026-08-11

Minor release, on `ai` v1.1.0.

### Added
- `Capabilities` describes what this driver can be asked for. It reports no
  hosted capability, which is the same answer `ai.Request.Hosted` already gets
  from this driver, now available before the call instead of only as an error
  after it. A test pins the two together.
- A provider refusal that means "this model cannot produce that format" is now
  translated into `ai.ErrNoFormat`, so an application degrades with one
  `errors.Is` instead of matching English prose in an error message. The
  provider's own `ai.APIError` is wrapped, not replaced, and stays reachable
  with `errors.As`. The rules are deliberately narrow - only a 400, only a
  format the caller asked for, only an error naming that exact field - because
  mistaking a genuinely bad request for a missing feature would retry it
  forever.

## [1.0.0] - 2026-08-11

First stable release, on `ai` v1.0.0.

### Added
- `ai.Request.Hosted` is answered with `ai.ErrNoHosted` before the request
  leaves, because models here run on the machine that serves them, and that server has no
  search to run.
  That refusal is the documented behavior rather than a gap left in silence:
  an answer produced without the search that was asked for looks exactly like
  one produced with it, so failing loudly is the only way a caller can tell
  them apart. A caller who would rather have the answer anyway asks again
  without `Hosted`.
- Tests pin it, including that nothing is sent to the provider.

## [0.2.2] - 2026-08-05

### Documentation
- The package documentation describes how `ai.Request.Format` reaches this
  provider, so it is on the first page a reader sees rather than only in the
  reference.

## [0.2.1] - 2026-08-05

Released by mistake: it carries this changelog entry and nothing else. The
documentation it describes landed in 0.2.2.

## [0.2.0] - 2026-08-05

### Added
- `ai.Request.Format` is mapped onto the provider's own `format` field, so a request
  for JSON is enforced by the provider rather than merely asked for, and
  `ai.Response.JSON` decodes the reply. `Response.Format` reports
  `ai.FormatNative`. Until now only this package's native request type could
  ask for JSON, so callers going through the provider-agnostic interface had
  to strip code fences from the reply by hand.
  `ai.FormatJSON` sends the bare word `"json"` and `ai.FormatJSONSchema` sends
  the schema directly, which is the shape this server expects.

### Changed
- Requires `github.com/goloop/ai` v0.4.0.

## [0.1.2] - 2026-07-10

### Documentation
- `DOC.md`/`DOC.UK.md` note that a stream cut off before completion yields
  `io.ErrUnexpectedEOF` instead of a silently-completed response.

## [0.1.1] - 2026-07-10

### Fixed
- A stream that ends before a final object marked `done` now surfaces
  `io.ErrUnexpectedEOF` instead of silently emitting a completed chunk.

### Changed
- Require `goloop/ai` v0.2.0 (500 no longer retried; jittered backoff).

## [0.1.0] - 2026-07-09

Initial release, built on the `github.com/goloop/ai` interface.

### Added
- `Client` implementing `ai.Client`: `Generate` and streaming `Stream` over
  `/api/chat`, with tool use and image input. Streaming reads Ollama's
  newline-delimited JSON.
- Native `ChatCompletion` and `ChatStream`.
- Embeddings (`Embed`), installed-model listing (`Models`) and per-model
  details (`Show`: template, parameters, capabilities and context length).
- Functional options: `WithBaseURL`, `WithHTTPClient`, `WithTimeout`,
  `WithMaxRetries`, `WithHeader`. Optional bearer token for authenticated
  proxies.
- Retries on 429 and 5xx with backoff; normalized `*ai.APIError` errors.
