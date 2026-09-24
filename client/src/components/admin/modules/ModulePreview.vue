<script setup lang="ts">
/** 模块选择器卡片内的迷你预览（按 type 渲染静态示意，外观与前台 about-kit 同构） */
defineProps<{ type: string }>();
const HEAT = [0, 2, 1, 3, 0, 4, 2, 1, 0, 3, 2, 4, 1, 0, 2, 3, 1, 4, 2, 0, 3, 1, 2, 4, 0, 1, 3, 2];
const BARS = [30, 46, 70, 58, 90, 50, 78, 64, 60, 0, 0, 0];
</script>

<template>
  <div class="pv">
    <div v-if="type === 'profile'" class="pv-profile">
      <div>
        <small>你好，我是</small>
        <b>Myself<i>.</i></b>
        <span class="ln" /><span class="ln s" />
      </div>
      <span class="pt" />
    </div>

    <div v-else-if="type === 'chapter'" class="pv-chapter">
      <div><em>01</em><span /></div>
      <b>此刻</b>
    </div>

    <div v-else-if="type === 'status'" class="pv-status">
      <span class="av"><i /></span>
      <div><b>在线</b><small>在写一篇新文章</small></div>
    </div>

    <div v-else-if="type === 'socials'" class="pv-pills">
      <span>GitHub</span><span>邮件</span><span>RSS</span>
    </div>

    <div v-else-if="type === 'contact'" class="pv-contact">
      <div><b>想聊聊？</b><span class="mail">hi@example.com</span></div>
      <i>写邮件</i>
    </div>

    <div v-else-if="type === 'stats'" class="pv-stats">
      <span v-for="s in [['264', '天'], ['42', '文章'], ['128', '随想'], ['9', '标签']]" :key="s[1]"><b>{{ s[0] }}</b><i>{{ s[1] }}</i></span>
    </div>

    <div v-else-if="type === 'github'" class="pv-heat">
      <i v-for="(l, k) in HEAT" :key="k" :data-l="l" />
    </div>

    <div v-else-if="type === 'languages'" class="pv-lang">
      <div><i style="flex: 38; background: #0078ff" /><i style="flex: 27; background: #00c853" /><i style="flex: 18; background: #ffb300" /><i style="flex: 17; background: #ff0032" /></div>
      <small>Go 38% · TS 27% · Vue 18%</small>
    </div>

    <div v-else-if="type === 'skillbars'" class="pv-ruler">
      <div v-for="lv in [16, 12]" :key="lv"><i v-for="k in 20" :key="k" :class="{ f: k <= lv }" /></div>
    </div>

    <div v-else-if="type === 'year'" class="pv-year">
      <i v-for="(h, k) in BARS" :key="k" :class="{ fut: !h, peak: h === 90 }" :style="{ height: `${h || 20}%` }" />
    </div>

    <ul v-else-if="type === 'now'" class="pv-now">
      <li><em>在做</em>关于页组件库<span><u style="width: 72%" /></span></li>
      <li><em>在读</em>设计中的设计</li>
    </ul>

    <div v-else-if="type === 'milestones'" class="pv-ms">
      <div><i /><b>2026.01</b><span>第一行代码</span></div>
      <div class="now"><i /><b>2026.09</b><span>组件库 v2</span></div>
    </div>

    <div v-else-if="type === 'projects'" class="pv-proj">
      <span class="big" />
      <div><span /><span /></div>
    </div>

    <div v-else-if="type === 'principles'" class="pv-pr">
      <div v-for="n in 3" :key="n"><b>0{{ n }}</b><span /></div>
    </div>

    <div v-else-if="type === 'bookshelf'" class="pv-books">
      <i v-for="(b, k) in [['#e9e4d8', 92], ['#1d3a5f', 80], ['#8c2f2a', 96], ['#2f4a3a', 100], ['#c9a14a', 74], ['#23272f', 94]]" :key="k" :style="{ background: b[0], height: `${b[1]}%` }" />
    </div>

    <div v-else-if="type === 'listening'" class="pv-vinyl">
      <span class="disc"><i /></span>
      <div><b>Kid A</b><small>Radiohead</small><span><u /></span></div>
    </div>

    <figure v-else-if="type === 'quotes'" class="pv-quote">
      <span>“</span>
      <p>简单是可靠的先决条件。</p>
      <div><i class="on" /><i /><i /></div>
    </figure>

    <p v-else-if="type === 'motto'" class="pv-motto">记录本身，就是意义。</p>

    <div v-else-if="type === 'gallery'" class="pv-gallery">
      <i class="g0 s1" /><i class="s2" /><i class="s3" /><i class="s4" /><i class="s5" />
    </div>

    <div v-else-if="type === 'places'" class="pv-map">
      <i v-for="k in 40" :key="k" />
      <b style="left: 72%; top: 58%" class="home" /><b style="left: 78%; top: 50%" /><b style="left: 64%; top: 30%" /><b style="left: 40%; top: 62%" />
    </div>

    <div v-else-if="type === 'favorites'" class="pv-fav">
      <div class="tabs"><span class="on">电影</span><span>游戏</span><span>音乐</span></div>
      <p><em>01</em>《一一》<small>2000</small></p>
    </div>

    <div v-else-if="type === 'skills'" class="pv-keys">
      <span class="star">文章</span><span>Go</span><span>Vue</span><span>胶片</span>
    </div>

    <div v-else-if="type === 'stack'" class="pv-stack">
      <span v-for="s in [['Go', '#00add8'], ['Vu', '#42b883'], ['TS', '#3178c6'], ['Vi', '#b58cff']]" :key="s[0]" :style="{ '--c': s[1] }">{{ s[0] }}</span>
    </div>

    <div v-else-if="type === 'uses'" class="pv-uses">
      <div><i /><b>MacBook Pro</b><em>主力</em></div>
      <div><i /><b>VS Code</b><em>每天</em></div>
    </div>

    <div v-else-if="type === 'faq'" class="pv-faq">
      <div><span>为什么叫 Myself？</span><b class="open" /></div>
      <div><span>这个站是用什么搭的？</span><b /></div>
    </div>

    <div v-else-if="type === 'guestbook'" class="pv-gb">
      <div class="in"><span>留下一句话…</span><i>发送</i></div>
      <div class="notes"><span /><span /></div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pv {
  --ln: color-mix(in oklab, var(--text) 10%, transparent);
  --well: color-mix(in oklab, var(--text) 5%, transparent);

  width: 100%;
  font-size: 10.5px;
  color: var(--text);
}

.pv-profile {
  display: flex;
  align-items: center;
  gap: 10px;

  > div { flex: 1; display: flex; flex-direction: column; gap: 3px; }
  small { font-size: 9px; color: var(--text-2); }
  b { font: 700 20px/1 var(--font-serif); letter-spacing: -0.03em; }
  b i { font-style: normal; color: var(--primary); }
  .ln { height: 4px; width: 80%; margin-top: 4px; border-radius: 2px; background: var(--ln); }
  .ln.s { width: 55%; margin-top: 0; }

  .pt {
    width: 46px;
    height: 56px;
    border-radius: 10px;
    background: radial-gradient(60% 50% at 60% 60%, rgba(var(--primary-rgb), 0.5), transparent 70%), #0b1528;
    mask-image: linear-gradient(to right, transparent, #000 40%);
  }
}

.pv-chapter {
  > div { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
  em { font: 500 9px ui-monospace, monospace; font-style: normal; color: var(--primary); }
  span { flex: 1; height: 1px; background: linear-gradient(90deg, var(--ln), transparent); }
  b { font: 700 18px/1.1 var(--font-serif); }
}

.pv-status {
  display: flex;
  align-items: center;
  gap: 10px;

  .av { position: relative; width: 30px; height: 30px; border-radius: 9px; background: #0b1528; }
  .av i { position: absolute; right: -3px; bottom: -3px; width: 10px; height: 10px; border-radius: 50%; background: #00c853; border: 2px solid var(--surface); }
  b { display: block; font: 700 12px var(--font-serif); }
  small { color: var(--text-2); }
}

.pv-pills {
  display: flex;
  gap: 6px;

  span { padding: 4px 10px; border-radius: 8px; border: 1px solid var(--ln); background: var(--well); }
  span:first-child { background: var(--text); color: var(--bg); border-color: transparent; }
}

.pv-contact {
  display: flex;
  align-items: center;
  gap: 10px;

  > div { flex: 1; }
  b { display: block; font: 700 13px var(--font-serif); }
  .mail { display: inline-block; margin-top: 4px; padding: 2px 8px; border-radius: 6px; background: var(--well); font-family: ui-monospace, monospace; font-size: 9.5px; }
  i { padding: 5px 10px; border-radius: 7px; font-style: normal; font-weight: 600; color: #fff; background: var(--primary); }
}

.pv-stats {
  display: flex;

  span { flex: 1; padding-left: 8px; border-left: 1px solid var(--ln); }
  span:first-child { padding-left: 0; border-left: 0; }
  b { display: block; font: 500 16px/1 ui-monospace, monospace; letter-spacing: -0.04em; }
  i { font-style: normal; font-size: 9.5px; color: var(--text-2); }
}

.pv-heat {
  display: grid;
  grid-template-rows: repeat(4, 8px);
  grid-auto-flow: column;
  grid-auto-columns: 8px;
  gap: 2px;

  i { border-radius: 2px; background: var(--well); }
  i[data-l='1'] { background: rgba(var(--primary-rgb), 0.22); }
  i[data-l='2'] { background: rgba(var(--primary-rgb), 0.45); }
  i[data-l='3'] { background: rgba(var(--primary-rgb), 0.72); }
  i[data-l='4'] { background: var(--primary); }
}

.pv-lang {
  > div { display: flex; gap: 2px; height: 9px; border-radius: 5px; overflow: hidden; }
  small { display: block; margin-top: 6px; color: var(--text-2); font-family: ui-monospace, monospace; font-size: 9px; }
}

.pv-ruler {
  display: flex;
  flex-direction: column;
  gap: 6px;

  > div { display: grid; grid-template-columns: repeat(20, 1fr); align-items: end; height: 12px; }
  i { justify-self: center; width: 2px; height: 6px; border-radius: 1px; background: var(--ln); }
  i:nth-child(5n) { height: 9px; }
  i.f { background: var(--primary); }
}

.pv-year {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 3px;
  align-items: end;
  height: 44px;
  border-bottom: 1px solid var(--ln);

  i { border-radius: 3px 3px 1px 1px; background: rgba(var(--primary-rgb), 0.55); }
  i.peak { background: var(--primary); box-shadow: 0 0 8px rgba(var(--primary-rgb), 0.7); }
  i.fut { background: repeating-linear-gradient(135deg, var(--ln) 0 1px, transparent 1px 4px); }
}

.pv-now {
  list-style: none;
  margin: 0;
  padding: 0;

  li { padding: 3px 0; border-top: 1px solid var(--ln); }
  li:first-child { border-top: 0; }
  em { margin-right: 8px; font-style: normal; font-family: ui-monospace, monospace; font-size: 9px; color: var(--primary); }
  span { display: block; width: 60%; height: 2px; margin: 4px 0 0 34px; background: var(--well); }
  u { display: block; height: 100%; background: var(--primary); }
}

.pv-ms {
  position: relative;
  padding-left: 14px;
  border-left: 1px solid var(--ln);

  > div { position: relative; display: flex; gap: 8px; padding: 2px 0; }
  i { position: absolute; left: -18px; top: 6px; width: 7px; height: 7px; border-radius: 50%; background: var(--surface); box-shadow: inset 0 0 0 1.5px var(--text-2); }
  .now i { background: var(--primary); box-shadow: 0 0 6px rgba(var(--primary-rgb), 0.9); }
  b { font-family: ui-monospace, monospace; font-weight: 500; font-size: 9.5px; color: var(--text-2); }
  .now b { color: var(--primary); }
}

.pv-proj {
  display: grid;
  grid-template-columns: 1.3fr 1fr;
  gap: 5px;
  height: 52px;

  .big { border-radius: 7px; background: radial-gradient(60% 50% at 50% 100%, rgba(var(--primary-rgb), 0.5), transparent 70%), linear-gradient(#060b16, #0b1426); position: relative; }
  .big::after { content: ''; position: absolute; left: 46%; width: 8%; top: 18%; height: 50%; background: #fff; box-shadow: 0 0 10px #fff; }
  > div { display: flex; flex-direction: column; gap: 5px; }
  > div span { flex: 1; border-radius: 6px; background: var(--well); }
}

.pv-pr {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;

  b { display: block; font: 700 20px/1 var(--font-serif); background: linear-gradient(180deg, var(--primary), transparent); -webkit-background-clip: text; background-clip: text; color: transparent; }
  span { display: block; height: 4px; margin-top: 4px; border-radius: 2px; background: var(--ln); }
}

.pv-books {
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 50px;
  padding: 0 4px;
  border-bottom: 3px solid var(--ln);

  i { width: 12px; border-radius: 2px 2px 1px 1px; box-shadow: inset 1px 0 0 rgb(255 255 255 / 0.15); }
  i:first-child { position: relative; }
  i:first-child::after { content: ''; position: absolute; top: -4px; right: 2px; width: 4px; height: 8px; background: #ff0032; }
}

.pv-vinyl {
  display: flex;
  align-items: center;
  gap: 10px;

  .disc { display: grid; place-items: center; width: 40px; height: 40px; border-radius: 50%; background: repeating-radial-gradient(circle, #161616 0 1px, #0c0c0c 1.5px 3px); }
  .disc i { width: 14px; height: 14px; border-radius: 50%; background: var(--primary); }
  > div { flex: 1; }
  b { display: block; font: 700 12px var(--font-serif); }
  small { color: var(--text-2); }
  span { display: block; height: 2px; margin-top: 5px; background: var(--well); }
  u { display: block; width: 35%; height: 100%; background: var(--primary); }
}

.pv-quote {
  margin: 0;

  > span { display: block; height: 12px; font: 700 28px/0.8 var(--font-serif); color: rgba(var(--primary-rgb), 0.45); }
  p { margin: 0; font: 700 12px/1.5 var(--font-serif); }
  div { display: flex; gap: 3px; margin-top: 6px; }
  i { width: 16px; height: 2px; border-radius: 2px; background: var(--ln); }
  i.on { background: var(--primary); }
}

.pv-motto { margin: 0; text-align: center; font: 700 14px/1.4 var(--font-serif); letter-spacing: 0.06em; text-shadow: 0 6px 14px rgba(var(--primary-rgb), 0.3); }

.pv-gallery {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  grid-auto-rows: 22px;
  gap: 3px;

  i { border-radius: 4px; }
  .g0 { grid-column: span 2; grid-row: span 2; }
  .s1 { background: linear-gradient(#060b16, #0b1426); box-shadow: inset 0 -10px 16px rgba(var(--primary-rgb), 0.4); }
  .s2 { background: linear-gradient(180deg, #1a2d5a 50%, #f0a45b 52%, #0b1a33 54%); }
  .s3 { background: linear-gradient(135deg, #ffd58a, #e0701a); }
  .s4 { background: linear-gradient(180deg, #4a2c4c, #ffb300); }
  .s5 { background: linear-gradient(180deg, #0b1530, #10203f); }
}

.pv-map {
  position: relative;
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 4px;
  padding: 2px 20px;

  > i { width: 3px; height: 3px; border-radius: 50%; background: var(--text-2); opacity: 0.35; }
  b { position: absolute; width: 6px; height: 6px; border-radius: 50%; background: var(--primary); box-shadow: 0 0 6px var(--primary); }
  b.home { background: #ffb300; box-shadow: 0 0 6px #ffb300; }
}

.pv-fav {
  .tabs { display: inline-flex; gap: 2px; padding: 2px; border-radius: 7px; background: var(--well); }
  .tabs span { padding: 1px 7px; border-radius: 5px; color: var(--text-2); }
  .tabs .on { background: var(--surface); color: var(--text); }
  p { margin: 6px 0 0; font: 700 11.5px var(--font-serif); }
  em { margin-right: 6px; font: 400 9px ui-monospace, monospace; color: var(--text-2); }
  small { float: right; font: 400 9px ui-monospace, monospace; color: var(--text-2); }
}

.pv-keys {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;

  span { padding: 3px 9px; border-radius: 6px; border: 1px solid var(--ln); box-shadow: 0 2px 0 var(--ln); background: var(--surface); }
  .star { color: var(--primary); border-color: rgba(var(--primary-rgb), 0.4); background: rgba(var(--primary-rgb), 0.08); }
}

.pv-stack {
  display: flex;
  gap: 6px;

  span {
    display: grid;
    place-items: center;
    width: 28px;
    height: 28px;
    border-radius: 8px;
    font: 600 10px ui-monospace, monospace;
    color: var(--c);
    background: color-mix(in oklab, var(--c) 13%, transparent);
    box-shadow: inset 0 0 0 1px color-mix(in oklab, var(--c) 30%, transparent);
  }
}

.pv-uses {
  > div { display: flex; align-items: center; gap: 8px; padding: 3px 0; border-top: 1px solid var(--ln); }
  > div:first-child { border-top: 0; }
  i { width: 16px; height: 16px; border-radius: 5px; background: var(--well); box-shadow: inset 0 0 0 1px var(--ln); }
  b { flex: 1; font-weight: 500; }
  em { padding: 1px 5px; border-radius: 4px; font-style: normal; font-size: 9px; color: var(--primary); background: rgba(var(--primary-rgb), 0.1); }
}

.pv-faq {
  > div { display: flex; justify-content: space-between; align-items: center; padding: 4px 0; border-top: 1px solid var(--ln); font: 700 10.5px var(--font-serif); }
  > div:first-child { border-top: 0; }
  b { width: 12px; height: 12px; border-radius: 4px; border: 1px solid var(--ln); }
  b.open { background: var(--primary); border-color: transparent; }
}

.pv-gb {
  .in { display: flex; justify-content: space-between; align-items: center; padding: 4px 4px 4px 8px; border-radius: 8px; background: var(--well); color: var(--text-2); }
  .in i { padding: 2px 8px; border-radius: 5px; font-style: normal; color: #fff; background: var(--primary); }
  .notes { display: grid; grid-template-columns: 1fr 1fr; gap: 4px; margin-top: 5px; }
  .notes span { height: 18px; border-radius: 6px; background: var(--well); }
}
</style>
