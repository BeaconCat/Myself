<script setup lang="ts">
/** 模块选择器卡片内的真实迷你预览（按模块 type 渲染静态示例） */
defineProps<{ type: string }>();
</script>

<template>
  <div v-if="type === 'stats'" class="pv-stats">
    <span v-for="s in [['365', '天'], ['42', '文'], ['128', '想'], ['9', '签']]" :key="s[1]">
      <strong>{{ s[0] }}</strong><i>{{ s[1] }}</i>
    </span>
  </div>

  <blockquote v-else-if="type === 'motto'" class="pv-motto">记录本身，就是意义。</blockquote>

  <div v-else-if="type === 'skills' || type === 'favorites'" class="pv-chips">
    <b>{{ type === 'skills' ? '创作' : '电影' }}</b>
    <span v-for="c in (type === 'skills' ? ['文章', '摄影'] : ['科幻', '悬疑'])" :key="c">{{ c }}</span>
  </div>

  <div v-else-if="type === 'skillbars'" class="pv-bars">
    <div v-for="b in [['写作', 80], ['摄影', 55]]" :key="b[0]">
      <em>{{ b[0] }}</em>
      <i><u :style="{ width: b[1] + '%' }" /></i>
    </div>
  </div>

  <div v-else-if="type === 'languages'" class="pv-lang">
    <i style="width: 45%; background: #0078ff" />
    <i style="width: 35%; background: #00c853" />
    <i style="width: 20%; background: #ffb300" />
  </div>

  <div v-else-if="type === 'milestones'" class="pv-ms">
    <div><b>2026</b><i /><span>点亮站点</span></div>
    <div><b>2025</b><i /><span>开始记录</span></div>
  </div>

  <div v-else-if="type === 'gallery'" class="pv-gallery">
    <i v-for="n in 4" :key="n" />
  </div>

  <figure v-else-if="type === 'quotes'" class="pv-quote">
    <span>把喜欢的话收藏在这里。</span>
    <em>—— 出处</em>
  </figure>

  <div v-else-if="type === 'devices' || type === 'stack'" class="pv-tiles">
    <span v-for="d in (type === 'devices' ? ['主机', '相机'] : ['Vue', 'Vite'])" :key="d">{{ d }}</span>
  </div>

  <div v-else-if="type === 'faq'" class="pv-faq">
    <div><span>一个常见的问题？</span><b>+</b></div>
    <div><span>另一个问题？</span><b>+</b></div>
  </div>

  <ul v-else-if="type === 'now'" class="pv-now">
    <li><i />在写一篇新文章</li>
    <li><i />在学一门新语言</li>
  </ul>

  <div v-else-if="type === 'github'" class="pv-stats">
    <span v-for="s in [['14', '仓'], ['4', '星'], ['7', '粉'], ['99', '提']]" :key="s[1]">
      <strong>{{ s[0] }}</strong><i>{{ s[1] }}</i>
    </span>
  </div>

  <div v-else-if="type === 'socials'" class="pv-pills">
    <span>GitHub</span><span>Email</span><span>RSS</span>
  </div>

  <div v-else-if="type === 'contact'" class="pv-contact">
    <div><b>想聊聊？</b><span>合作或打个招呼</span></div>
    <i>发邮件</i>
  </div>
</template>

<style scoped lang="scss">
.pv-stats {
  display: flex;
  gap: 8px;

  span {
    flex: 1;
    text-align: center;
    padding: 6px 2px;
    border-radius: 6px;
    background: var(--surface-2);

    strong {
      display: block;
      font-family: var(--font-serif);
      font-size: 14px;
      background: var(--grad-title);
      background-clip: text;
      -webkit-background-clip: text;
      color: transparent;
    }

    i { font-style: normal; font-size: 9.5px; color: var(--text-2); }
  }
}

.pv-motto {
  margin: 0 auto;
  width: fit-content;
  padding: 6px 14px;
  font-family: var(--font-serif);
  font-size: 12px;
  color: var(--text-2);
  border-left: 2px solid rgba(var(--primary-rgb), 0.6);
  border-right: 2px solid rgba(var(--primary-rgb), 0.6);
  background: rgba(var(--primary-rgb), 0.05);
  border-radius: 5px;
}

.pv-chips {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;

  b { font-size: 11px; color: var(--primary); margin-right: 2px; }

  span {
    font-size: 10.5px;
    font-weight: 600;
    padding: 2px 9px;
    background: var(--surface-2);
  }
}

.pv-bars {
  display: flex;
  flex-direction: column;
  gap: 7px;

  > div { display: flex; align-items: center; gap: 8px; }
  em { font-style: normal; font-size: 10.5px; width: 30px; flex-shrink: 0; }

  i {
    flex: 1;
    height: 6px;
    border-radius: 999px;
    background: var(--surface-2);
    overflow: hidden;
    display: block;

    u {
      display: block;
      height: 100%;
      border-radius: inherit;
      background: linear-gradient(90deg, var(--primary), var(--primary-deep));
    }
  }
}

.pv-lang {
  display: flex;
  height: 12px;
  border-radius: 999px;
  overflow: hidden;

  i { display: block; }
}

.pv-ms {
  display: flex;
  flex-direction: column;
  gap: 6px;

  > div { display: flex; align-items: center; gap: 8px; font-size: 10.5px; }
  b { font-family: var(--font-serif); color: var(--primary); font-size: 12px; }

  i {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--primary);
  }

  span { color: var(--text-2); }
}

.pv-gallery {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;

  i {
    aspect-ratio: 1;
    border-radius: 4px;
    background: linear-gradient(135deg, rgba(var(--primary-rgb), 0.35), var(--surface-2));
  }
}

.pv-quote {
  margin: 0;

  span {
    display: block;
    font-family: var(--font-serif);
    font-size: 11.5px;
  }

  em {
    display: block;
    text-align: right;
    font-style: normal;
    font-size: 10px;
    color: var(--text-2);
    margin-top: 4px;
  }
}

.pv-tiles {
  display: flex;
  gap: 6px;

  span {
    flex: 1;
    text-align: center;
    font-family: var(--font-serif);
    font-size: 11px;
    padding: 8px 4px;
    background: var(--surface-2);
  }
}

.pv-faq {
  display: flex;
  flex-direction: column;
  gap: 5px;

  > div {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 5px 9px;
    background: var(--surface-2);
    border-radius: 6px;
    font-size: 10.5px;
  }

  b { color: var(--primary); }
}

.pv-now {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;

  li {
    display: flex;
    align-items: center;
    gap: 7px;
    font-size: 10.5px;
  }

  i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--primary);
    box-shadow: 0 0 5px rgba(var(--primary-rgb), 0.6);
  }
}

.pv-pills {
  display: flex;
  gap: 6px;

  span {
    font-size: 10.5px;
    font-weight: 600;
    padding: 4px 12px;
    border: 1px solid var(--border);
    background: var(--surface-2);
  }
}

.pv-contact {
  display: flex;
  align-items: center;
  gap: 10px;

  > div { flex: 1; }
  b { display: block; font-size: 12px; }
  span { font-size: 10px; color: var(--text-2); }

  i {
    font-style: normal;
    font-size: 10.5px;
    font-weight: 700;
    color: #fff;
    padding: 5px 12px;
    background: linear-gradient(180deg, var(--primary), var(--primary-deep));
    border-radius: 6px;
  }
}
</style>
