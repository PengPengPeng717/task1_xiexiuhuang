## Context

- **项目**：题 #1 — 24h 矿业新闻 + 政策 + 价格三源聚合管线。
- **现状**：绿地仓库，无既有应用代码；业务基线见 `需求文档.md` 与 `最小确认文档.md`。
- **约束**：后端 Go + Gin；**无 SQL/ORM**；本地 JSONL + Go struct；pipeline/serve/eval 分目录交付；价格仅公开、不可替代源。

## Goals / Non-Goals

**Goals:**

- P0（24h）：`pipeline` 可写入 `data/`；`serve` `/query` 返回 Top5 片段；`eval` 对 20 条 GT 输出 recall@5 与 faithfulness 报告。
- P1：新闻/政策/价格各 ≥200 条（近 30 天、按发布日），合计 ≥600；`manifest.json` 可验收。
- 问答以原文片段为主；三源支持「近 7 天」类时间过滤；库内无覆盖时明确说明。
- `DATA_NOTES.md` 与 JSONL 字段、主键、去重策略一致。

**Non-Goals:**

- MySQL/PostgreSQL/SQLite、GORM 或任意 ORM。
- 多租户、OAuth、细粒度 RBAC。
- 绕过登录墙、非授权替代价格源。
- 实时流行情、移动端 App、大规模分布式爬虫。
- Phase1 独立 SPA（可选 htmx/Swagger 演示，非必须）。

## Decisions

| ID | 决策 | 理由 | 备选（未选） |
|----|------|------|----------------|
| D-01 | 语言与 HTTP：Go + Gin | 用户指定；单二进制易部署 | FastAPI（原需求） |
| D-02 | 存储：`data/*.jsonl` + `manifest.json` | 无 DB 运维；600 条规模足够 | MySQL + GORM |
| D-03 | 向量：embedding 存 chunk 记录；TopK 内存余弦 | 规模小，免 Qdrant | Qdrant 旁路 |
| D-04 | 模块：`pipeline` 写、`serve`/`eval` 读；共享 `internal/retriever` | 保证 eval 与线上一致 | eval 重复实现检索 |
| D-05 | 采集：Go colly/goquery + RSS；单源失败可一次性 JSON 导入 | 24h 主路径简单 | 全量 Python 双栈 |
| D-06 | 写入：pipeline 写 `*.tmp` 后 rename；serve/eval 只读 | 避免 JSONL 写坏 | 并发读写同一文件 |
| D-07 | 权限：默认无登录；可选 API Key 中间件 | 需求未要求多用户 | 完整 RBAC |
| D-08 | 验收：P0 链路 + eval；P1 条数达标（调和 F1 与 A3） | 用户确认 F1=B、A3=A | 仅 P0 即算完成 |

### 目录结构（目标）

```
pipeline/cmd/ingest
pipeline/collector/{news,policy,price}
pipeline/processor, dedup, chunker, embedder
serve/cmd/api, handler/query
eval/cmd/eval, ground_truth.json
internal/model, internal/store/jsonl, internal/retriever, internal/queryparse
data/documents.jsonl, data/chunks.jsonl, data/manifest.json, data/raw/
DATA_NOTES.md
```

### 数据契约（摘要，详见 `local-data-store` spec）

- **Document**：一条业务记录（1 篇新闻 / 1 份政策文件 / 1 条价格日行情）。
- **Chunk**：检索单元，含 `chunk_id`, `document_id`, `text`, `embedding`, 过滤元数据。
- **Manifest**：`news_count`, `policy_count`, `price_count`, `failed_sources`, `last_ingest_at`。

### 主键与去重

| 实体 | 主键 | 规则 |
|------|------|------|
| document | `source_type` + `canonical_url`（冲突时用 `content_sha256`） | 写入前查重，重复跳过 |
| chunk | `document_id` + `chunk_index` | 同文档重建 chunk 时先删旧 chunk |
| 价格 | `commodity` + `trade_date` + `exchange` | 与 B3 一致 |

### `/query` 流程

1. 解析问句：时间窗（如 7 天）、`source_type`、实体（锂、澳洲等）。
2. 加载 `chunks.jsonl`（可启动时缓存）→ 元数据预过滤（`published_at` 等）。
3. 向量 TopK（默认 K=5 用于响应，eval 用 recall@5）。
4. 组装响应：片段文本、`source_url`、`published_at`；无命中返回固定「库内无覆盖」文案（D3=A）。

### Embedding

- 调用 OpenAI 兼容 Embedding HTTP API；模型名与环境变量写入 `DATA_NOTES.md`。
- 无密钥时可 Phase1 用确定性占位向量仅跑通链路（须在 manifest 标注，P1 前替换）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 价格公开页拿不到，类 200 条不达标 | Day0 spike；manifest 记录 `failed_sources`；E2 不允许替代源 |
| 政策/新闻 30 天内不足 200 条 | 扩大抓取窗口或合法历史列表页；验收以 manifest 计数为准 |
| JSONL 全量加载内存 | 当前规模可接受；chunk >10k 再考虑索引文件 |
| faithfulness 规则过粗 | 先关键词/URL 覆盖；可选 LLM judge 作 Phase2 |
| F1 vs A3 表述冲突 | proposal/tasks 明确 P0/P1 DoD |
| 嵌入 API 不可用 | 占位向量 + 文档标注；阻塞 P1 质量验收 |

## Migration Plan

- 不适用（绿地项目）。部署：构建 `pipeline`、`serve`、`eval` 三个 cmd；准备 `data/` 目录；配置 `.env`（API Key、Embedding URL）。
- 回滚：保留上一版 `data/` 快照目录。

## Open Questions

| ID | 问题 | 默认 |
|----|------|------|
| OQ-01 | Embedding/LLM 供应商与密钥 | 环境变量配置，文档说明 |
| OQ-02 | Phase1 是否要 htmx 单页演示 | 否，Swagger/curl 即可 |
| OQ-03 | 期内无新稿 vs 完全无数据 的文案区分 | 实现时在 query-api spec 场景补一条 |
| OQ-04 | faithfulness 阈值 | 先规则：GT 要点出现在 Top5 片段或 URL 匹配 |
| OQ-05 | 各价格品种公开 URL 清单 | spike 任务 0.1 产出写入 DATA_NOTES |
