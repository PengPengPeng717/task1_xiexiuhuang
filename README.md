# 矿业三源聚合管线（题 #1）

新闻 + 政策 + 价格 → 本地 JSONL → `/query` 片段检索 → `eval` 自动评测。

## 要求

- Go 1.22+
- 可选：`EMBEDDING_URL`、`EMBEDDING_API_KEY`

## 快速开始（P0）

```bash
# 1. 入库（合成种子数据演示；真实采集去掉 -seed）
go run ./pipeline/cmd/ingest -data-dir ./data -seed

# 2. 启动查询 API
go run ./serve/cmd/api -data-dir ./data -port 8080

# 3. 查询示例
curl "http://localhost:8080/query?question=近%207%20天澳洲锂出口政策有何变化？"

# 4. 评测（需先 -seed 或 -seed-p1 使 GT 文档 ID 对齐）
go run ./eval/cmd/eval -data-dir ./data
```

## P1 数据量验收

```bash
go run ./pipeline/cmd/ingest -data-dir ./data -seed-p1
# 检查 data/manifest.json: news/policy/price 各 >= 200, total >= 600
```

真实采集（无 seed）：

```bash
go run ./pipeline/cmd/ingest -data-dir ./data
```

价格源失败会出现在 `manifest.failed_sources`；按业务规则（E2=A）需修复公开源后重跑，不可用替代站凑数。

## 环境变量

见 `.env.example`。

| 变量 | 说明 |
|------|------|
| EMBEDDING_URL | OpenAI 兼容 embedding 端点 |
| EMBEDDING_MODEL | 模型名 |
| EMBEDDING_API_KEY | API 密钥 |
| API_KEY | 若设置，则 `/query` 需 `X-API-Key` |

## 交付目录

- `pipeline/` — 采集入库
- `serve/` — Gin `/query`
- `eval/` — 20 条 GT + recall@5 + faithfulness
- `DATA_NOTES.md` — schema 说明

## P0 / P1 Definition of Done

| 阶段 | 标准 |
|------|------|
| P0 | ingest 写 JSONL；`/query` 返回 Top5 片段；eval 出 `eval/report.json` |
| P1 | manifest 三类各 ≥200、合计 ≥600；近 30 天发布日 |

OpenSpec 变更：`openspec/changes/mining-three-source-pipeline/`
