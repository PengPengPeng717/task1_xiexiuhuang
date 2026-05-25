## Why

矿业新闻、关键矿产政策与大宗商品价格信息分散在多个公开站点，难以用自然语言快速检索「近 7 天澳洲锂出口政策有何变化」类问题。本项目在 24 小时交付约束下，需要一条可验收的三源聚合管线：采集近 30 天数据、本地可检索存储、片段式问答与自动评测，支撑题 #1 交付。

## What Changes

- 新增 `pipeline/`：从矿业新闻、政策、价格三类源采集；清洗、去重、分块、向量化；写入本地 `data/`（JSONL）。
- 新增 `serve/`：Go + Gin 提供 `/query` REST 接口；自然语言提问；返回 Top5 **原文片段**及来源（非以生成归纳为主）。
- 新增 `eval/`：20 条 ground truth Q&A；自动计算 `recall@5` 与 `answer faithfulness`。
- 新增 `DATA_NOTES.md`：记录 schema、字段、主键、去重策略，与 JSONL 契约一致。
- **BREAKING（相对原始需求文档）**：不使用 FastAPI/MySQL/GORM；改为 Go + Gin + 纯 JSONL + Go struct 本地存储。
- 验收分两阶段：P0 先跑通链路 + 评测可跑（F1）；P1 三类各 ≥200 条、合计 ≥600（A3），未达标不算完成。

## Capabilities

### New Capabilities

- `local-data-store`：本地 `data/` 目录、JSONL 主数据/chunk/manifest 契约、主键与去重规则。
- `data-pipeline`：三源采集（新闻全文+结构化、政策、公开价格）、清洗、去重、分块、嵌入、入库。
- `query-api`：自然语言 `/query`；时间/源类型/主题过滤；Top5 片段响应；无数据明确说明。
- `rag-evaluation`：20 条 GT Q&A、recall@5、faithfulness 自动评测；与 query 共用检索逻辑。

### Modified Capabilities

（无。`openspec/specs/` 下尚无既有能力规格。）

## Impact

- 新建仓库模块：`pipeline/`、`serve/`、`eval/`、`internal/`（共享 model/store/retriever）、`data/`（运行时数据，不入库敏感凭据）。
- 外部依赖：公开网站/RSS、可选 Embedding HTTP API；无数据库服务。
- 业务确认来源：`需求文档.md`、`最小确认文档.md`（已填 P0/P1/P2）。
- 主要风险：价格源仅公开（E1/E2）；部分源 30 天内可能不足 200 条；F1 与 A3 的分阶段验收须在实现中显式区分。
