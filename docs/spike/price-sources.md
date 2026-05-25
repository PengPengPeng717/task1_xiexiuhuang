# Price Source Spike (Day0)

| Commodity | Exchange | URL | Result |
|-----------|----------|-----|--------|
| 铜 | LME | https://www.lme.com/en/Metals/Non-ferrous/Copper | Bot 访问常受限；collector 记 `price:LME:铜` failed |
| 锌 | LME | https://www.lme.com/en/Metals/Non-ferrous/Zinc | 同上 |
| 镍 | LME | https://www.lme.com/en/Metals/Non-ferrous/Nickel | 同上 |
| 锂 | SHFE | https://www.shfe.com.cn/eng/reports/StatisticalData/ | 需进一步页面解析 |
| 铁矿石 | Mysteel | https://www.mysteel.com/ | 登录墙 |

**结论：** P0 使用 `-seed` 跑通；P1 真实价格需合法公开数据导入或页面解析成功后写入，禁止替代源（E2=A）。
