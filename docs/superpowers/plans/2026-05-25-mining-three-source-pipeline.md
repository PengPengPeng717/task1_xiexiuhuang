# 矿业三源聚合管线 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 `pipeline/` + `serve/` + `eval/` + `DATA_NOTES.md`：采集新闻/政策/价格 → JSONL 本地库 → Gin `/query` 返回 Top5 原文片段 → 20 条 GT 自动评测 recall@5 与 faithfulness。

**Architecture:** Go 单体仓库三 cmd（ingest / api / eval）；`internal/` 共享 model、jsonl store、retriever、queryparse；`data/` 存 documents/chunks/manifest JSONL；向量 TopK 内存余弦；无 SQL/ORM。

**Tech Stack:** Go 1.22+、Gin、colly/goquery、标准库 `encoding/json`、可选 OpenAI 兼容 Embedding HTTP API。

**Spec source:** `openspec/changes/mining-three-source-pipeline/`（proposal、design、4×spec、tasks）

**验收阶段:**
- **P0（24h）：** 链路跑通 + eval 可跑（F1=B）
- **P1（最终完成）：** news/policy/price 各 ≥200、合计 ≥600、未达标不算完成（A3=A）

**计划内默认（二次澄清 R1–R8 未回复前）：**
- 时区：`Asia/Shanghai` 计算「近 7/30 天」
- `/query`：Top5 片段 + 固定免责声明一句（非 LLM 归纳）
- P0：每类至少 1 条真实采集（其余可少）
- eval：输出分数，**不设自动及格线**（验收人工看报告）
- DISR 战略页 = **1 条** policy document；regcc 爬公告列表凑政策条数
- 价格：5 品种各用公开日行情；`failed_sources` 非空且导致某类 &lt;200 → P1 失败（E2=A）
- 离线 import：仅允许「与配置官方 URL 同源的公开数据」手工 JSONL，禁止替代站

---

## File Structure（实施前锁定）

| 路径 | 职责 |
|------|------|
| `go.mod` | module `mining-pipeline` |
| `internal/model/document.go` | Document、Chunk、Manifest struct + JSON 标签 |
| `internal/model/types.go` | SourceType 常量、Commodity 枚举 |
| `internal/store/jsonl/reader.go` | 加载 documents/chunks |
| `internal/store/jsonl/writer.go` | tmp+rename 原子写、追加行 |
| `internal/store/jsonl/manifest.go` | 计数、failed_sources 更新 |
| `internal/dedup/fingerprint.go` | canonical_url + sha256 去重 |
| `internal/chunker/splitter.go` | 按字符/段落分块 |
| `internal/embedder/client.go` | HTTP embedding + 占位向量 |
| `internal/retriever/retriever.go` | 过滤 + 余弦 TopK |
| `internal/queryparse/parse.go` | 近 N 天、源类型、关键词 |
| `pipeline/collector/news/news.go` | mining.com + S&P RSS 全文 |
| `pipeline/collector/policy/policy.go` | regcc + DISR |
| `pipeline/collector/price/price.go` | LME/SHFE/Mysteel 公开页 |
| `pipeline/processor/normalize.go` | 清洗、RFC3339、30 天过滤 |
| `pipeline/cmd/ingest/main.go` | 采集编排入口 |
| `serve/cmd/api/main.go` | Gin 启动 |
| `serve/handler/query.go` | `/query` |
| `serve/middleware/apikey.go` | 可选 API_KEY |
| `eval/ground_truth.json` | 20 条 GT（draft → approved） |
| `eval/cmd/eval/main.go` | 评测入口 |
| `eval/metrics/recall.go` | recall@5 |
| `eval/metrics/faithfulness.go` | 规则 faithfulness |
| `data/.gitkeep` | 运行时数据目录 |
| `DATA_NOTES.md` | schema 文档 |
| `README.md` | 命令与 P0/P1 DoD |
| `.env.example` | EMBEDDING_URL、API_KEY 等 |

测试文件与实现同目录 `_test.go` 或使用 `internal/.../*_test.go`。

---

## Task 1: Go module 与 model 契约

**Files:**
- Create: `go.mod`, `internal/model/types.go`, `internal/model/document.go`
- Test: `internal/model/document_test.go`

- [ ] **Step 1: 写失败测试 — JSON 往返**

```go
// internal/model/document_test.go
package model_test

import (
	"encoding/json"
	"testing"
	"time"
	"mining-pipeline/internal/model"
)

func TestDocumentJSONRoundTrip(t *testing.T) {
	doc := model.Document{
		DocumentID:   "news-1",
		SourceType:   model.SourceNews,
		CanonicalURL: "https://www.mining.com/article/1",
		Title:        "Test",
		PublishedAt:  time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		ContentText:  "body",
		ContentSHA256: "abc",
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var got model.Document
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.DocumentID != doc.DocumentID {
		t.Fatalf("got %s", got.DocumentID)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd f:\xxh_task1 && go test ./internal/model/... -v`  
Expected: FAIL（package 不存在）

- [ ] **Step 3: 实现 model**

```go
// go.mod
module mining-pipeline

go 1.22
```

```go
// internal/model/types.go
package model

type SourceType string

const (
	SourceNews   SourceType = "news"
	SourcePolicy SourceType = "policy"
	SourcePrice  SourceType = "price"
)
```

```go
// internal/model/document.go
package model

import "time"

type Document struct {
	DocumentID    string     `json:"document_id"`
	SourceType    SourceType `json:"source_type"`
	CanonicalURL  string     `json:"canonical_url"`
	Title         string     `json:"title"`
	PublishedAt   time.Time  `json:"published_at"`
	ContentText   string     `json:"content_text"`
	ContentSHA256 string     `json:"content_sha256"`
	Region        string     `json:"region,omitempty"`
	Commodity     string     `json:"commodity,omitempty"`
	Exchange      string     `json:"exchange,omitempty"`
	TradeDate     string     `json:"trade_date,omitempty"` // YYYY-MM-DD for price
}

type Chunk struct {
	ChunkID     string     `json:"chunk_id"`
	DocumentID  string     `json:"document_id"`
	ChunkIndex  int        `json:"chunk_index"`
	Text        string     `json:"text"`
	PublishedAt time.Time  `json:"published_at"`
	SourceType  SourceType `json:"source_type"`
	SourceURL   string     `json:"source_url"`
	Embedding   []float32  `json:"embedding"`
}

type Manifest struct {
	NewsCount      int       `json:"news_count"`
	PolicyCount    int       `json:"policy_count"`
	PriceCount     int       `json:"price_count"`
	TotalCount     int       `json:"total_count"`
	LastIngestAt   time.Time `json:"last_ingest_at"`
	FailedSources  []string  `json:"failed_sources"`
	EmbeddingMode  string    `json:"embedding_mode"` // "real" | "placeholder"
}
```

- [ ] **Step 4: 测试通过**

Run: `go test ./internal/model/... -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod internal/model/
git commit -m "feat: add core document chunk manifest models"
```

---

## Task 2: JSONL store 原子写与 manifest

**Files:**
- Create: `internal/store/jsonl/writer.go`, `internal/store/jsonl/reader.go`, `internal/store/jsonl/manifest.go`
- Test: `internal/store/jsonl/writer_test.go`

- [ ] **Step 1: 写失败测试 — 原子替换**

```go
func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	err := jsonl.WriteFileAtomic(path, []byte("{\"a\":1}\n"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "{\"a\":1}\n" {
		t.Fatalf("got %q", b)
	}
}
```

- [ ] **Step 2: 运行确认 FAIL**

Run: `go test ./internal/store/jsonl/... -v`

- [ ] **Step 3: 实现 WriteFileAtomic、LoadDocuments、LoadChunks、SaveManifest**

`WriteFileAtomic`: 写 `path.tmp` → `os.Rename(path.tmp, path)`  
`AppendDocuments`: 读现有行 + 新行 → 全量 tmp 写（600 规模可接受）

- [ ] **Step 4: PASS**

- [ ] **Step 5: Commit** `feat: jsonl atomic store and manifest`

---

## Task 3: DATA_NOTES.md 与样例数据

**Files:**
- Create: `DATA_NOTES.md`, `data/documents.jsonl`（空或 .gitkeep）, `data/chunks.jsonl`, `data/manifest.json`

- [ ] **Step 1: 根据 `local-data-store` spec 写 DATA_NOTES**（字段表、主键、去重、示例一行 JSON）

- [ ] **Step 2: 添加 `data/manifest.json` 初值**

```json
{"news_count":0,"policy_count":0,"price_count":0,"total_count":0,"last_ingest_at":"1970-01-01T00:00:00Z","failed_sources":[],"embedding_mode":"placeholder"}
```

- [ ] **Step 3: `.gitignore` 增加 `data/raw/`、`.env`**

- [ ] **Step 4: 人工核对与 OpenSpec spec 一致**

- [ ] **Step 5: Commit** `docs: add DATA_NOTES and data directory scaffold`

---

## Task 4: Day0 价格源 spike（阻塞采集）

**Files:**
- Modify: `DATA_NOTES.md`（价格公开 URL 表）
- Create: `docs/spike/price-sources.md`

- [ ] **Step 1: 手工访问 LME/SHFE/Mysteel 各品种公开页，记录 URL、是否登录、可见字段**

- [ ] **Step 2: 估算 30 天内可达条数（5 品种 × 交易日）**

- [ ] **Step 3: 若某源不可用，写入 `failed_sources` 命名约定（如 `price:lme:copper`）**

- [ ] **Step 4: 与产品确认是否触发 P1 失败（E2=A）— 记录到 spike 文档**

- [ ] **Step 5: Commit** `docs: price source spike results`

---

## Task 5: 去重、清洗、30 天过滤

**Files:**
- Create: `internal/dedup/fingerprint.go`, `pipeline/processor/normalize.go`
- Test: `internal/dedup/fingerprint_test.go`, `pipeline/processor/normalize_test.go`

- [ ] **Step 1: 测试 `WithinLast30Days`（Asia/Shanghai）与 `DedupKey(sourceType, url, sha)`**

- [ ] **Step 2: FAIL**

- [ ] **Step 3: 实现 `NormalizeURL`、`ParsePublishedAt`、`FilterLast30Days`**

- [ ] **Step 4: PASS**

- [ ] **Step 5: Commit** `feat: dedup and 30-day publish filter`

---

## Task 6: News collector（全文）

**Files:**
- Create: `pipeline/collector/news/news.go`
- Test: `pipeline/collector/news/news_test.go`（用 `httptest` 假 HTML）

- [ ] **Step 1: 测试从固定 HTML 抽 title/body/published_at**

- [ ] **Step 2: FAIL**

- [ ] **Step 3: RSS 发现 + colly 抓全文；source 标记 `mining.com` / `spglobal`**

- [ ] **Step 4: 集成测试可选 `-short` 跳过网络**

- [ ] **Step 5: Commit** `feat: news collector with full text`

---

## Task 7: Policy collector

**Files:**
- Create: `pipeline/collector/policy/policy.go`

- [ ] **Step 1: regcc 列表页 → 公告链接 → 单 document**

- [ ] **Step 2: DISR 战略 URL → **单条** document（ID 固定 `policy-disr-strategy-2023-2030`）**

- [ ] **Step 3: 反爬：User-Agent + 请求间隔 1–2s（写入 DATA_NOTES）**

- [ ] **Step 4: 本地试跑记录 policy_count**

- [ ] **Step 5: Commit** `feat: policy collector regcc and DISR`

---

## Task 8: Price collector（仅公开）

**Files:**
- Create: `pipeline/collector/price/price.go`

- [ ] **Step 1: 按 spike 文档实现各品种 parser**

- [ ] **Step 2: 登录墙页 → 返回 error，上层记入 `failed_sources`，**不**写假数据**

- [ ] **Step 3: 每条 `commodity+trade_date+exchange` 一条 document**

- [ ] **Step 4: 试跑 `price_count`**

- [ ] **Step 5: Commit** `feat: public-only price collector`

---

## Task 9: Chunker + Embedder + ingest 编排

**Files:**
- Create: `internal/chunker/splitter.go`, `internal/embedder/client.go`, `pipeline/cmd/ingest/main.go`

- [ ] **Step 1: chunker：政策/新闻按 ~800 字切分；价格 document 单 chunk**

- [ ] **Step 2: embedder：`EMBEDDING_URL` 未设置时用 `PlaceholderEmbed(text)` 确定性向量**

- [ ] **Step 3: ingest：顺序 news → policy → price → dedup → chunk → embed → 写 jsonl → 更新 manifest**

Run: `go run ./pipeline/cmd/ingest -data-dir ./data`  
Expected: manifest 计数增加；embedding_mode=placeholder

- [ ] **Step 4: 验证 documents/chunks 可被 store 加载**

- [ ] **Step 5: Commit** `feat: ingest pipeline end-to-end`

---

## Task 10: Retriever + Queryparse

**Files:**
- Create: `internal/retriever/retriever.go`, `internal/queryparse/parse.go`
- Test: `internal/retriever/retriever_test.go`, `internal/queryparse/parse_test.go`

- [ ] **Step 1: queryparse 测试 `"近 7 天澳洲锂政策"` → days=7, region=澳洲, commodity=锂, sourceType=policy**

- [ ] **Step 2: retriever 测试：给定 10 chunks，过滤后 Top5 含目标 document_id**

- [ ] **Step 3: 实现余弦相似度；无候选时返回空**

- [ ] **Step 4: PASS**

- [ ] **Step 5: Commit** `feat: retriever and query parser`

---

## Task 11: Gin `/query` API

**Files:**
- Create: `serve/cmd/api/main.go`, `serve/handler/query.go`, `serve/middleware/apikey.go`
- Test: `serve/handler/query_test.go`

- [ ] **Step 1: httptest `GET /query?question=近7天锂价格` → 200 + JSON hits[]**

- [ ] **Step 2: FAIL**

- [ ] **Step 3: 响应结构**

```json
{
  "disclaimer": "以下为知识库原文摘录，不构成投资或政策分析建议。",
  "hits": [{"text":"...","source_url":"...","published_at":"...","source_type":"price","score":0.82}],
  "coverage": "ok"
}
```

`coverage`: `ok` | `no_data` | `no_recent`（7 天内有库无近期）

- [ ] **Step 4: 启动服务 `go run ./serve/cmd/api -data-dir ./data -port 8080`**

- [ ] **Step 5: Commit** `feat: gin query API with snippet response`

---

## Task 12: Ground truth + eval

**Files:**
- Create: `eval/ground_truth.json`, `eval/metrics/recall.go`, `eval/metrics/faithfulness.go`, `eval/cmd/eval/main.go`

- [ ] **Step 1: 起草 20 条 GT（≥1 条政策+7天题 G3）；字段：`id`, `question`, `expected_document_ids`, `expected_points`, `category`**

- [ ] **Step 2: recall@5：每条 question 调 retriever.TopK(5)，命中任一 expected id 即 1**

- [ ] **Step 3: faithfulness：Top5 文本包含 expected_points 任一关键词即 pass**

- [ ] **Step 4: 输出 `eval/report.json`**

Run: `go run ./eval/cmd/eval -data-dir ./data`  
Expected: exit 0；报告含 `recall_at_5`、`faithfulness_rate`

- [ ] **Step 5: Commit** `feat: eval recall@5 and faithfulness report`

---

## Task 13: P1 数据补全与真实 embedding

**Files:**
- Modify: `pipeline/cmd/ingest`, `internal/embedder/client.go`, `data/*`

- [ ] **Step 1: 配置真实 `EMBEDDING_URL`/`EMBEDDING_MODEL`，manifest.embedding_mode=real**

- [ ] **Step 2: 扩大 regcc 历史公告抓取（30 天发布日过滤仍生效）**

- [ ] **Step 3: 循环 ingest 直至 manifest 三类各 ≥200**

- [ ] **Step 4: 全量重嵌所有 chunks（ingest -reembed 标志）**

- [ ] **Step 5: 跑 eval 保存报告供验收；Commit** `chore: P1 corpus 600+ and real embeddings`

---

## Task 14: README 与交付核对

**Files:**
- Create: `README.md`, `.env.example`

- [ ] **Step 1: README 含 P0/P1 DoD 表格、三条命令、curl 示例**

- [ ] **Step 2: 对照 `需求文档.md` 交付清单：pipeline serve eval DATA_NOTES**

- [ ] **Step 3: `openspec/changes/mining-three-source-pipeline/tasks.md` 勾选已完成项**

- [ ] **Step 4: 最终 manifest 截图或粘贴到 README 附录**

- [ ] **Step 5: Commit** `docs: README and delivery checklist`

---

## Self-Review（计划 vs OpenSpec）

| Spec 要求 | 对应 Task |
|-----------|-----------|
| local-data-store JSONL/manifest/原子写 | 2, 3 |
| data-pipeline 三源/全文/公开价/30天 | 4–9 |
| query-api /query Top5/时间过滤/无覆盖 | 10–11 |
| rag-evaluation 20 GT/recall/faithfulness | 12 |
| DATA_NOTES | 3 |
| P0 部分数据可跑 | 9–12 |
| P1 600 条 | 13 |

**Gap（需在实现前确认 R1–R8 或更新 spec）:**
- 政策 200 条与 DISR=1 条 — Task 7+13 依赖 regcc 历史量
- 价格 200 条数学 — Task 4 spike 可能判定 P1 不可达
- eval 及格线 — 本计划采用「仅报告、无自动 gate」
- 新闻结构化字段最小集 — Task 3 DATA_NOTES 需列全

**Placeholder 扫描:** 无 TBD 步骤；Task 4/7/8 依赖 spike 结果但步骤已定义动作。

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-05-25-mining-three-source-pipeline.md`.**

**Two execution options:**

1. **Subagent-Driven (recommended)** — 每 Task 派生子代理 + 任务间审查；SUB-SKILL: `superpowers:subagent-driven-development`
2. **Inline Execution** — 本会话按 Task 顺序实现；SUB-SKILL: `superpowers:executing-plans` 或 **`/opsx:apply`** 对齐 OpenSpec tasks

**OpenSpec 对齐:** 实现时同步勾选 `openspec/changes/mining-three-source-pipeline/tasks.md`；若 R1–R8 有答复，先更新 spec 再改 Task 7/8/13。

**Which approach?**
