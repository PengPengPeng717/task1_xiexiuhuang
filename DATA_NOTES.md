# DATA_NOTES

## 目录

- `data/documents.jsonl` — 业务主记录（一行一条 JSON）
- `data/chunks.jsonl` — 检索分块（含 embedding）
- `data/manifest.json` — 采集统计与失败源
- `data/raw/` — 可选原始 HTML（不入库凭据）

## Document 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| document_id | string | 唯一 ID |
| source_type | news \| policy \| price | 源类型 |
| canonical_url | string | 规范 URL（主键组成部分） |
| title | string | 标题 |
| published_at | RFC3339 | 发布/交易日 |
| content_text | string | 正文或行情描述 |
| content_sha256 | string | 正文哈希（去重备用） |
| region | string? | 政策地域 |
| commodity | string? | 品种 |
| exchange | string? | 交易所 |
| trade_date | YYYY-MM-DD? | 价格交易日 |

## Chunk 字段

| 字段 | 说明 |
|------|------|
| chunk_id | `{document_id}-{index}` |
| document_id | 父文档 |
| chunk_index | 分块序号 |
| text | 检索文本 |
| published_at / source_type / source_url | 过滤与引用 |
| embedding | float32 数组 |

## 主键与去重

- 主键：`source_type + canonical_url`
- 冲突：使用 `content_sha256` 跳过重复行
- manifest 计数不因重复递增

## 计条规则（业务）

- news：1 篇完整报道
- policy：1 篇公告/独立文件（DISR 战略页 = 1 条）
- price：1 品种 × 1 交易日 1 条

## 时间

- 采集窗口：近 30 天（`published_at`，Asia/Shanghai）
- 查询：「近 N 天」按同一时区过滤

## 价格源（Day0 spike）

| 品种 | 交易所 | URL | 公开可达备注 |
|------|--------|-----|----------------|
| 铜 | LME | https://www.lme.com/en/Metals/Non-ferrous/Copper | 常需登录/反爬，记 failed_sources |
| 锌 | LME | https://www.lme.com/en/Metals/Non-ferrous/Zinc | 同上 |
| 镍 | LME | https://www.lme.com/en/Metals/Non-ferrous/Nickel | 同上 |
| 锂 | SHFE | https://www.shfe.com.cn/eng/reports/StatisticalData/ | 部分公开统计 |
| 铁矿石 | Mysteel | https://www.mysteel.com/ | 登录墙常见 |

拿不到公开数据时 **不写假行情**（E2=A），写入 `manifest.failed_sources`。

## 示例行

```json
{"document_id":"seed-policy-0","source_type":"policy","canonical_url":"https://example.com/policy/0","title":"政策公告样本 0","published_at":"2026-05-20T00:00:00Z","content_text":"稀土与锂出口政策公告样本...","content_sha256":"..."}
```

## Embedding

- `EMBEDDING_URL` 未设置 → `manifest.embedding_mode=placeholder`（确定性向量，仅 P0 链路）
- 设置 OpenAI 兼容接口 → `embedding_mode=real`
