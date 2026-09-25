# 待修问题（随全站翻新一并处理）

## [已修复 2026-09-24] Hero 卡片着色阴影在编舞结束时硬切闪现

> 修复：阴影拆为独立中性层 `.shade`，品牌溢光为 `.deck-spill`，五种卡组编舞经 `choreo/card/lights.ts` 只动画 opacity（离场 200ms 淡出、入场落定前 300ms 淡入）；`::after` 改 opacity 过渡。逐帧像素比对阴影区 0 变化。

- 记录日期：2026-09-24
- 现象：首页 Hero 卡组切换动画播放完毕的瞬间，卡片背后的主色着色阴影（蓝色外晕）突然出现，没有过渡。
- 位置：`client/src/components/home/hero/HeroDeck.vue`
  - `.sheet` 的静态阴影 `box-shadow: var(--shadow), 0 30px 70px -20px rgba(var(--primary-rgb), .35)` 不在任何编舞关键帧里。编舞期间卡片层被 WAAPI 接管，结束时动画 cancel、交回静态样式，阴影一次性生效。
  - `.album-card.leaving::after { display: none }` 同样按类名开关，没有淡入淡出。
- 修复方向：
  1. 把着色阴影从 `.sheet` 挪到独立的伪元素或光晕层，只动画 `opacity`（不要动画 box-shadow 本身）。
  2. 各卡组编舞（`choreo/card/*.ts`）入场结尾统一加一段约 300ms 的阴影层淡入；离场开头淡出。
  3. `::after` 光晕改为 `opacity` + `transition`，去掉 `display: none` 切换。
  4. 结合第三轮强调系统：着色发光只保留给封面本身的门缝光，控件不再用彩色外发光；评估卡片外晕是否改为中性阴影。
- 验收：五种卡组动效 × 深浅模式，在动画结束前后 100ms 连续截图，阴影无跳变。
