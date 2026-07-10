# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
