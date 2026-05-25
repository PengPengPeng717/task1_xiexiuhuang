## ADDED Requirements

### Requirement: Ground truth dataset

The eval module SHALL include `ground_truth.json` (or equivalent) with exactly 20 question–answer pairs, drafted by implementers and subject to reviewer approval (per confirmed G1=B).

#### Scenario: Dataset size

- **WHEN** eval runs
- **THEN** exactly 20 GT entries are loaded

### Requirement: At least one time-policy composite question

At least one GT question MUST resemble a time-bounded policy change query (e.g. recent days + region + mineral + policy change), per confirmed G3=A.

#### Scenario: Policy time question present

- **WHEN** GT file is validated
- **THEN** at least one entry tags `category` including `policy` and encodes a time window of approximately 7 days

### Requirement: Shared retrieval with serve

Evaluation MUST invoke the same retriever implementation (package or library) used by `/query` so that `recall@5` measures production retrieval behavior.

#### Scenario: Eval uses internal retriever

- **WHEN** eval computes recall for a GT question
- **THEN** it calls `internal/retriever` with the same `data/chunks.jsonl` snapshot as serve

### Requirement: Recall at 5 metric

For each GT question, the system SHALL compute whether the expected evidence appears in the top 5 retrieved chunks. The run SHALL output per-question and aggregate `recall@5`.

#### Scenario: Hit in top 5

- **WHEN** GT expects document `doc-123` and retriever ranks a chunk from `doc-123` at position 3
- **THEN** that question is scored as a recall hit

#### Scenario: Miss outside top 5

- **WHEN** expected evidence appears only at rank 6 or lower
- **THEN** that question is scored as a recall miss

### Requirement: Answer faithfulness metric

For each GT question, the system SHALL score faithfulness of the returned top snippets against GT expected points (rule-based keyword/URL/span overlap minimum; optional LLM judge documented if enabled).

#### Scenario: Faithful when excerpt contains GT point

- **WHEN** top 5 snippets include text overlapping GT `expected_points`
- **THEN** faithfulness score for that question is passing per configured threshold

#### Scenario: Unfaithful when snippets omit GT

- **WHEN** top 5 snippets contain no GT keywords or URLs
- **THEN** faithfulness score for that question is failing

### Requirement: Eval report artifact

Each eval run SHALL write a human-readable report (e.g. `eval/report.json` or stdout summary) with per-question recall, faithfulness, and overall averages.

#### Scenario: CI or local eval run

- **WHEN** operator runs `eval` command after ingest
- **THEN** a report file or console summary is produced with aggregate `recall@5` and faithfulness rate

### Requirement: P0 eval runnable before full 600 records

Eval MUST be executable when corpus size is below 600 records; results MAY show low recall but the command must exit zero on successful run.

#### Scenario: P0 eval on partial corpus

- **WHEN** only 100 total documents exist
- **THEN** eval completes and outputs metrics without crashing

### Requirement: Three-source coverage not mandatory in GT set

The 20 GT questions are NOT required to be evenly split across news, policy, and price (per confirmed G2=B).

#### Scenario: Skewed GT distribution

- **WHEN** 15 questions target policy and 5 target news
- **THEN** eval run is still valid
