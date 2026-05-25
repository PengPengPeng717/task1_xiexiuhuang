## ADDED Requirements

### Requirement: Three source categories

The pipeline SHALL ingest from exactly three categories: mining news, critical minerals policy, and commodity prices, as defined in the project requirements document.

#### Scenario: Categories distinguished in storage

- **WHEN** records from any collector are written
- **THEN** each document has `source_type` of `news`, `policy`, or `price` only

### Requirement: News sources and full-text extraction

The pipeline SHALL collect mining news from mining.com and S&P Global Mining (RSS allowed for discovery). Each news item MUST be one full article (not RSS title-only) with structured fields extracted (e.g. title, published time, body).

#### Scenario: Full article ingested

- **WHEN** a news item is collected from RSS or listing page
- **THEN** the pipeline fetches full article HTML and stores complete `content_text` in the document record

### Requirement: Policy sources

The pipeline SHALL collect policy content from China Rare Earth Group site (regcc.cn) and Australia DISR Critical Minerals Strategy 2023–2030 publication page. Each policy record MUST be one standalone announcement or file (not arbitrary paragraph splitting for counting).

#### Scenario: Policy file counted as one record

- **WHEN** one policy announcement page is ingested
- **THEN** exactly one document record is created for that file

### Requirement: Price sources and public-only access

The pipeline SHALL collect public price data for LME copper/zinc/nickel, SHFE lithium, and Mysteel iron ore. The pipeline MUST NOT bypass login walls or use substitute sources when a configured source is unavailable.

#### Scenario: Login-required page not bypassed

- **WHEN** a price page requires authentication and no public data is visible
- **THEN** the source is listed in `manifest.failed_sources` and no fabricated price rows are written

#### Scenario: One price row per commodity per trade day

- **WHEN** a valid public price for one commodity on one trade date is collected
- **THEN** one document record is written with `commodity`, `exchange`, and `trade_date` set

### Requirement: Sub-source aggregation counts

Counts for news, policy, and price MAY be satisfied by aggregating across sub-sources within each category (e.g. mining.com + S&P combined for news); each sub-source is not required to individually reach 200 records.

#### Scenario: News total from multiple sites

- **WHEN** mining.com contributes 120 articles and S&P contributes 90 within 30 days
- **THEN** `manifest.news_count` is 210 and satisfies the news category minimum

### Requirement: Thirty-day window by publish date

Ingest SHALL include only documents whose `published_at` (or `trade_date` for prices) falls within the last 30 days relative to ingest run time, unless a documented exception is recorded in `DATA_NOTES.md` for sparse sources.

#### Scenario: Old article excluded

- **WHEN** an article published 40 days ago is discovered
- **THEN** it is not written to `documents.jsonl`

### Requirement: Clean and normalize

The pipeline SHALL clean HTML noise, normalize timestamps to RFC3339, trim whitespace, and normalize URLs before deduplication.

#### Scenario: Normalized timestamp

- **WHEN** a page provides a parseable publish date
- **THEN** `published_at` is stored in RFC3339 UTC or with explicit offset

### Requirement: Chunking and embedding

After documents are stored, the pipeline SHALL split text into chunks, compute embeddings via a configured Embedding HTTP API, and append rows to `chunks.jsonl`.

#### Scenario: Chunks created for policy document

- **WHEN** a long policy document is ingested
- **THEN** one or more chunks are written linked by `document_id` for retrieval (without counting each chunk as a separate policy "record" for the 200 policy document threshold)

### Requirement: P0 ingest runnable with partial data

For P0 acceptance, the pipeline MUST be runnable end-to-end and write valid JSONL even if counts are below 600; manifest MUST reflect actual counts.

#### Scenario: P0 demo ingest

- **WHEN** only 50 news documents are available in 30 days
- **THEN** pipeline completes without error and manifest shows `news_count` = 50

### Requirement: Optional one-shot JSON import

For a single failing collector, the pipeline MAY accept a one-shot JSON/JSONL import file produced offline, without requiring a second runtime language in the main path.

#### Scenario: Offline price JSON imported

- **WHEN** operator provides `import/price_batch.jsonl` documented in DATA_NOTES
- **THEN** ingest command merges valid rows into `documents.jsonl` using the same schema and dedup rules
