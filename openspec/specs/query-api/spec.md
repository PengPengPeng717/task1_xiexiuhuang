## ADDED Requirements

### Requirement: Query HTTP endpoint

The serve module SHALL expose a REST endpoint `/query` (GET or POST) accepting a natural language `question` string and returning a JSON response.

#### Scenario: Successful query request

- **WHEN** client sends `question` = "近 7 天澳洲锂出口政策有何变化？"
- **THEN** the API responds with HTTP 200 and a JSON body containing ranked source snippets

### Requirement: Response dominated by source snippets

The response SHALL prioritize verbatim or near-verbatim text spans from ingested chunks (not a free-form summary as the primary payload). Each hit MUST include `text`, `source_url`, `published_at`, and `source_type`.

#### Scenario: Top hits include citations

- **WHEN** relevant chunks exist in the knowledge base
- **THEN** the response includes up to 5 hits ordered by relevance score
- **THEN** each hit contains extractable source citation fields

### Requirement: Time and source-type filtering

The query handler SHALL parse time expressions (e.g. "近 7 天", "last 7 days") and apply `published_at` filtering across news, policy, and price sources when applicable.

#### Scenario: Seven-day policy filter

- **WHEN** question implies last 7 days and policy content exists in that window
- **THEN** returned hits only include chunks whose `published_at` is within 7 days of query time

#### Scenario: Seven-day price filter

- **WHEN** question asks for recent lithium price movement in last 7 days
- **THEN** price chunks outside the 7-day window are excluded from hits

### Requirement: No coverage response

When no chunks match filters or similarity threshold, the API MUST return an explicit message that the knowledge base has no coverage (MUST NOT invent policy or price facts).

#### Scenario: Empty knowledge base topic

- **WHEN** question asks about a commodity with zero ingested documents
- **THEN** response states that no in-corpus data was found and does not present fabricated excerpts

### Requirement: Read-only access to data files

The serve process MUST NOT write to `data/*.jsonl` during query handling.

#### Scenario: Concurrent pipeline ingest

- **WHEN** pipeline replaces `chunks.jsonl` via atomic rename after ingest
- **THEN** subsequent queries load the updated file without server restart if documented reload behavior is implemented, or require documented restart

### Requirement: Optional API key

If environment variable `API_KEY` is set, the server SHALL require matching credentials on `/query`; if unset, `/query` is open on the bind address.

#### Scenario: API key enforced

- **WHEN** `API_KEY` is configured and client omits the key
- **THEN** the API responds with HTTP 401

### Requirement: Swagger or documented curl examples

The serve module SHALL provide machine- or human-readable API documentation sufficient to call `/query` without a custom frontend.

#### Scenario: Reviewer tests without UI

- **WHEN** reviewer follows README or Swagger
- **THEN** they can execute a sample `/query` and receive a valid JSON response
