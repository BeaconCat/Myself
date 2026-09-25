# 第三轮 · 高密度设计语言（全站铺开规则）

用户已认可的打样：首页底部「关于 + GitHub」两张卡
（`client/src/components/home/AboutMeCard.vue`、`GithubStatusCard.vue`，`views/HomeView.vue` 的 `.me` 5:7 并排网格）。
**动手前先打开这两个文件，照它们的尺寸、结构与写法来。** 强调系统（抬升 + 轻染、实底主按钮、无发光、圆角 token）继续遵守 `design/round3/IMPLEMENT.md`。

用户原话：信息量太低，更倾向图表 / 图标 / 文字大一些，减少不必要的留白，布局不要零散。

## 规则

1. **收紧留白**
   - 页面分区之间 56–64px（原 96px）；分区标题到内容 18–20px。
   - 卡片内边距 24px（原 28–32px），卡片内部分段间距 18–22px。
   - 页面顶部大标题区高度压缩：标题 + 副标题 + 操作同一行对齐，不单独占一大块。
2. **数字与图表优先**
   - 凡是统计：用「统计条」—— `background: var(--fill)`、`border-radius: var(--r-md)`，等分格子，格内 `padding: 16px 16px 14px`，数字 `font-family: var(--font-mono); font-size: 30px; font-weight: 600; font-variant-numeric: tabular-nums;`，标签 12.5px `--text-3`，格间用 `box-shadow: -1px 0 0 var(--line)` 分隔。窄处 4 → 2 列。
   - 图表（热力图、柱状、进度、占比）**撑满容器宽度**，格子/柱子随宽度等比放大，不缩在角落。
3. **字号与图标放大一档**
   - 分区标题：思源宋体 24–28px 700；卡片标题 20–24px；正文与列表摘要 15px；元信息 13px；12px 以下只留给脚注。
   - 线性图标 18–20px（原 14–16px）；圆形图标按钮 38px，`background: var(--fill)`，hover `--fill-2`。
   - 胶囊标签 13px，`padding: 5px 12px`，描边 `inset 0 0 0 1px var(--line-2)`。
4. **结构集中，不零散**
   - 零散的小卡合并成少数几个信息完整的大块；同一行卡片 `align-items: stretch` 等高，内部 `display:flex; flex-direction:column` + `margin-top:auto` 把操作压到底部。
   - 用 12 栏思路做 5/7、4/8、6/6、4/4/4 的并排；避免「一张卡独占一行、右侧留白」。
   - 头部三段式：左身份/标题，中主体，右操作；操作按钮与标题基线对齐。
5. **列表与卡片**
   - 列表行：缩略 + 标题 + 摘要一行截断 + 元信息并排，行高紧凑（行内上下 12–14px）。
   - 封面卡：封面高度收一档（16:9 → 约 2:1），标题、摘要、状态、日期更醒目地排在封面下方同一信息块内。
6. **自检**：1440 与 1024 宽、深浅两种模式；首屏信息量应明显多于改前；不出现孤立的大片空白。
