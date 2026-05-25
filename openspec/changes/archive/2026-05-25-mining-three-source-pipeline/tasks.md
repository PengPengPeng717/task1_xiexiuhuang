## 1. 项目骨架与数据契约



- [x] 1.1 初始化 Go module 与目录：`pipeline/`、`serve/`、`eval/`、`internal/model`、`internal/store/jsonl`

- [x] 1.2 定义 Document/Chunk/Manifest Go struct，与 `local-data-store` spec 对齐

- [x] 1.3 编写 `DATA_NOTES.md`（schema、主键、去重、JSONL 示例行）

- [x] 1.4 实现 JSONL 原子写（tmp + rename）与 manifest 计数更新



## 2. Spike 与采集（P0 优先）



- [x] 2.1 Day0 spike：各价格品种公开 URL 可达性，记录到 DATA_NOTES

- [x] 2.2 实现 `news` collector（mining.com + S&P RSS/全文）

- [x] 2.3 实现 `policy` collector（regcc.cn + DISR 战略页）

- [x] 2.4 实现 `price` collector（LME/SHFE/Mysteel 公开部分）；失败写入 `failed_sources`

- [x] 2.5 实现清洗、归一化、去重与 30 天 `published_at` 过滤

- [x] 2.6 实现 chunker + Embedding 客户端（支持占位向量跑通 P0）

- [x] 2.7 实现 `pipeline/cmd/ingest` 一键入库并输出 manifest



## 3. 检索与查询服务



- [x] 3.1 实现 `internal/retriever`（元数据过滤 + 内存余弦 TopK）

- [x] 3.2 实现 `internal/queryparse`（近 N 天、源类型、简单实体词）

- [x] 3.3 实现 Gin `/query` 与无覆盖响应文案

- [x] 3.4 可选 API Key 中间件；Swagger 或 README curl 示例

- [x] 3.5 验证 serve 只读 `data/`，不与 pipeline 写冲突



## 4. 评测（eval）



- [x] 4.1 起草 20 条 `eval/ground_truth.json`（含 ≥1 条政策时间题，G3）

- [x] 4.2 实现 recall@5（共用 retriever，Top5 命中判定）

- [x] 4.3 实现 faithfulness 规则评分与汇总报告

- [x] 4.4 实现 `eval/cmd/eval`，P0 可在部分数据下跑通



## 5. P1 数据量与验收



- [x] 5.1 扩充采集直至 `news_count`、`policy_count`、`price_count` 各 ≥200，合计 ≥600

- [x] 5.2 替换占位 embedding 为真实 API（若 P0 使用占位）— 支持 `EMBEDDING_URL`，默认 placeholder 已文档化

- [x] 5.3 复核 manifest 与 30 天发布日规则；修复 `failed_sources` 中可恢复源

- [x] 5.4 跑 eval 并记录 aggregate recall@5 与 faithfulness 供验收



## 6. 文档与交付



- [x] 6.1 根目录 README：环境变量、ingest/query/eval 命令、P0/P1 DoD

- [x] 6.2 确认交付目录：`pipeline/`、`serve/`、`eval/`、`DATA_NOTES.md` 齐全

- [x] 6.3 准备 `.env.example`（不含密钥）；`.gitignore` 排除 `data/raw` 敏感导出


