## ADDED Requirements

### Requirement: Local data directory layout

The system SHALL persist all ingested knowledge under a `data/` directory at the project root with at least: `documents.jsonl`, `chunks.jsonl`, `manifest.json`, and optional `raw/` for source artifacts.

#### Scenario: Pipeline completes ingest

- **WHEN** a pipeline ingest run finishes successfully
- **THEN** `data/documents.jsonl` and `data/chunks.jsonl` exist and are valid JSONL (one JSON object per line)
- **THEN** `data/manifest.json` exists and is valid JSON

### Requirement: Document record schema

Each line in `documents.jsonl` SHALL represent one business record with fields at minimum: `document_id`, `source_type` (`news` | `policy` | `price`), `canonical_url`, `title`, `published_at` (RFC3339), `content_text`, `content_sha256`, and source-specific metadata (`region`, `commodity`, `exchange`, `trade_date` as applicable).

#### Scenario: News document stored

- **WHEN** one full news article is ingested
- **THEN** exactly one document record is appended with `source_type` = `news` and non-empty `content_text`

#### Scenario: Price document stored

- **WHEN** one commodity trade-day price row is ingested
- **THEN** exactly one document record is appended with `source_type` = `price` and `trade_date` set to that trading day

### Requirement: Chunk record schema

Each line in `chunks.jsonl` SHALL represent one retrieval unit with fields at minimum: `chunk_id`, `document_id`, `chunk_index`, `text`, `published_at`, `source_type`, `source_url`, and `embedding` (array of floats) or a documented sidecar reference with the same `chunk_id`.

#### Scenario: Chunk linked to document

- **WHEN** a document is chunked for retrieval
- **THEN** every chunk line includes the parent `document_id` and inherits filter fields needed for time and source-type queries

### Requirement: Manifest counters for acceptance

`manifest.json` SHALL track `news_count`, `policy_count`, `price_count`, `total_count`, `last_ingest_at`, and `failed_sources` (array of strings describing sources that could not be ingested).

#### Scenario: Verify P1 data volume

- **WHEN** a reviewer checks P1 acceptance
- **THEN** manifest shows `news_count` >= 200, `policy_count` >= 200, `price_count` >= 200, and `total_count` >= 600
- **THEN** counts reflect records whose `published_at` falls within the last 30 days

### Requirement: Document primary key and deduplication

The system SHALL treat `source_type` + `canonical_url` as the primary key for documents; if `canonical_url` is unstable, `content_sha256` SHALL be used as the deduplication fingerprint. Duplicate documents MUST NOT be appended twice.

#### Scenario: Duplicate URL skipped

- **WHEN** ingest encounters a document with the same `source_type` and `canonical_url` as an existing record
- **THEN** the duplicate is skipped and manifest counters are not double-incremented

### Requirement: Atomic writes

The pipeline SHALL write JSONL updates to temporary files and atomically rename them into place so that `serve` and `eval` never read partially written files.

#### Scenario: Ingest during query

- **WHEN** pipeline writes new data while serve is running
- **THEN** serve continues to read the last complete JSONL snapshot without parse errors

### Requirement: DATA_NOTES documentation

The repository SHALL include `DATA_NOTES.md` describing the same schemas, primary keys, deduplication rules, and field semantics as this spec.

#### Scenario: New developer onboarding

- **WHEN** a developer reads `DATA_NOTES.md`
- **THEN** they can map each JSONL field to business meaning without reading application code
