[English](README.md) | [中文](README.zh-CN.md)

# mrr-cli

终端 MRR（月经常性收入）追踪器，为独立开发者打造。追踪来自 Stripe、Gumroad、Paddle 或手动录入的收入 — 全在命令行完成。

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## 特性

- 📊 **追踪 MRR**，支持多来源（Stripe、Gumroad、Paddle、手动录入）
- 💰 **货币格式化**，默认 USD（可配置）
- 📈 **增长率计算**，与上月对比
- 💵 **ARR 和估值**，可配置倍数
- 🔮 **预测** — 预估未来 MRR 和里程碑
- 🎯 **目标追踪** — 设定目标并追踪进度
- 🌐 **公开仪表盘** — 漂亮的网页，适合 Open Startup 风格展示
- 📤 **CSV 导入/导出**，数据可移植
- 🤖 **Agent 友好** JSON 输出，便于自动化
- 🏷️ **状态徽章**，可嵌入 README
- 🎨 **彩色美观输出**，表格格式化
- 🖥️ **交互式 TUI**，vim 风格快捷键
- 🗄️ **SQLite 存储** — 单文件 `~/.mrr-cli/data.db`
- ⚡ **单二进制** — 无依赖

## 安装

### 从源码

```bash
git clone https://github.com/indiekitai/mrr-cli.git
cd mrr-cli
go build -o mrr .
sudo mv mrr /usr/local/bin/
```

### Go Install

```bash
go install github.com/indiekitai/mrr-cli@latest
```

## 用法

### 添加收入记录

```bash
# 基本添加（默认来源 manual，类型 recurring）
mrr add 29.99

# 带选项
mrr add 99.00 --source stripe --type recurring
mrr add 49.99 --source gumroad --note "Lifetime license" --type one-time
mrr add 100 --date 2024-01-15
```

**选项：**
- `--source, -s`：收入来源（`stripe`、`gumroad`、`paddle`、`manual`）
- `--type, -t`：收入类型（`recurring`、`one-time`）
- `--note, -n`：备注
- `--date, -d`：日期，YYYY-MM-DD 格式（默认今天）

### 列出记录

```bash
# 列出所有记录
mrr list

# 按月过滤
mrr list --month 2024-01

# 按来源或类型过滤
mrr list --source stripe
mrr list --type recurring

# JSON 输出，便于自动化
mrr list --json
```

### 编辑记录

```bash
mrr edit 1 --amount 49.99
mrr edit 1 --source stripe
mrr edit 1 --note "Updated note"
mrr edit 1 --amount 99 --source gumroad
```

### 删除记录

```bash
mrr delete 1        # 需要确认
mrr delete 1 -f     # 强制删除（跳过确认）
```

### 生成报告

```bash
# 当月报告
mrr report

# 指定月份
mrr report --month 2024-01

# 自定义估值倍数（默认 3x）
mrr report --multiplier 5

# JSON 输出
mrr report --json

# 安静模式 - 只输出 MRR 数字
mrr report --quiet
```

报告显示：
- **MRR**（月经常性收入）
- **ARR**（年经常性收入 = MRR × 12）
- **增长率** vs 上月
- **估值**（ARR × 倍数）
- 一次性收入
- 按来源分布

示例输出：
```
Monthly Report: February 2026
───────────────────────────────────

MRR:        $1,234.00
ARR:        $14,808.00
Growth:     +15.2% vs last month
Valuation:  $44,424.00 (at 3x ARR)

By Source:
    stripe:   $800.00  (64.8%)
    gumroad:  $434.00  (35.2%)
```

### CSV 导出

```bash
# 导出所有记录到 stdout
mrr export

# 导出指定月份
mrr export --month 2024-01

# 导出到文件
mrr export --output entries.csv

# 导出为 JSON
mrr export --json
```

### CSV 导入

```bash
mrr import entries.csv
```

### MRR 预测

```bash
# 显示预估和里程碑
mrr forecast

# JSON 输出
mrr forecast --json
```

基于当前增长率预估 3、6、12 个月的 MRR，以及到达各收入里程碑（$1k、$5k、$10k、$50k、$100k MRR）的预计时间。

示例输出：
```
MRR Forecast (based on 15.2% monthly growth)
────────────────────────────────────────────

Current:      $1,234.00
In 3 months:  $1,890.00
In 6 months:  $2,895.00
In 12 months: $6,786.00

Milestones:
  $5,000 MRR: ~8 months (Oct 2026)
  $10,000 MRR: ~13 months (Mar 2027)
  $50,000 MRR: ~25 months (Mar 2028)
```

### 目标追踪

设定 MRR 目标并追踪进度：

```bash
# 设定目标
mrr goal set 10000                  # 设定 $10,000 MRR 目标
mrr goal set 10000 --by 2026-06     # 带截止时间

# 查看进度
mrr goal status

# 清除目标
mrr goal clear
```

**目标状态输出：**
```
🎯 Goal: $10,000 MRR by June 2026
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Current:  $1,234 (12.3%)
████████░░░░░░░░░░░░░░░░░░░░░░░░░

Progress: $1,234 / $10,000
Remaining: $8,766
Time left: 4 months

📈 At current growth rate (15%/mo):
   Projected to reach goal in: 3.2 months ✅
   Expected date: May 2026
```

### 公开仪表盘

启动漂亮的 Web 仪表盘，适合 Open Startup 风格的透明展示：

```bash
# 启动仪表盘服务器
mrr serve                   # 在端口 8080 启动
mrr serve --port 3000       # 自定义端口
mrr serve --public          # 公开模式（隐藏详细记录）
```

### 生成徽章

```bash
# 输出 SVG 到 stdout
mrr badge

# 保存到文件
mrr badge --output mrr.svg
```

生成 shields.io 风格的 SVG 徽章，显示当前 MRR。适合放在 README 里！

### 交互式 TUI

```bash
mrr tui
```

**快捷键：**
| 键 | 操作 |
|----|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `g` | 跳到第一条 |
| `G` | 跳到最后一条 |
| `a` | 添加新记录 |
| `e` | 编辑选中记录 |
| `d` | 删除选中记录 |
| `r` | 刷新 |
| `q` / `Esc` | 退出 |

## Agent 友好输出

所有列表和报告命令支持 `--json` / `-j` 获取机器可读输出：

```bash
# 获取 MRR 数字
mrr report --quiet
# 输出: 1234.00

# 完整报告 JSON
mrr report --json

# 记录列表 JSON
mrr list --json

# 导出 JSON
mrr export --json
```

方便与脚本、自动化工具或 AI agent 集成。

## 数据存储

所有数据本地存储在 SQLite `~/.mrr-cli/data.db`。

### 备份

```bash
cp ~/.mrr-cli/data.db ~/backup/mrr-backup.db
```

## License

MIT
