<script setup lang="ts">
/** 骨架屏：按版式渲染占位块，光带扫过（reduced-motion 下静止）。 */
withDefaults(defineProps<{ variant?: 'rows' | 'hero' | 'grid' | 'list'; count?: number }>(), {
  variant: 'rows',
  count: 5,
});
</script>

<template>
  <div class="ma-skel" :class="variant" aria-busy="true">
    <template v-if="variant === 'hero'">
      <div class="card">
        <i class="l w30" />
        <i class="l big w45" />
        <div class="bars"><i v-for="k in 7" :key="k" :style="{ height: `${24 + ((k * 37) % 60)}%` }" /></div>
      </div>
      <div class="trio"><i /><i /><i /></div>
    </template>
    <div v-else-if="variant === 'grid'" class="grid">
      <i v-for="k in count" :key="k" />
    </div>
    <div v-else-if="variant === 'list'" class="list">
      <div v-for="k in count" :key="k" class="li"><i class="ic" /><i class="l" :style="{ width: `${40 + ((k * 23) % 40)}%` }" /></div>
    </div>
    <template v-else>
      <div v-for="k in count" :key="k" class="row">
        <i class="thumb" />
        <div class="lines">
          <i class="l" :style="{ width: `${60 + ((k * 17) % 30)}%` }" />
          <i class="l sm w40" />
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped lang="scss">
.ma-skel i {
  display: block;
  border-radius: 8px;
  background: linear-gradient(100deg, var(--fill) 30%, var(--fill-2) 50%, var(--fill) 70%) 0 0 / 300% 100%;
  animation: ma-shimmer 1.4s linear infinite;
}

@keyframes ma-shimmer {
  from { background-position: 100% 0; }
  to { background-position: -50% 0; }
}

.row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;

  .thumb { width: 52px; height: 52px; border-radius: 13px; flex: none; }
  .lines { flex: 1; display: flex; flex-direction: column; gap: 8px; }
  .l { height: 14px; }
  .sm { height: 10px; }
}

.w30 { width: 30%; }
.w40 { width: 40%; }
.w45 { width: 45%; }

.card {
  margin: 4px 16px 0;
  padding: 20px;
  border-radius: 26px;
  background: var(--elev);
  box-shadow: inset 0 0 0 0.5px var(--line);
  display: flex;
  flex-direction: column;
  gap: 12px;

  .l { height: 12px; }
  .big { height: 40px; border-radius: 10px; }

  .bars {
    display: flex;
    align-items: flex-end;
    gap: 10px;
    height: 110px;

    i { flex: 1; border-radius: 6px; }
  }
}

.trio {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin: 14px 16px 0;

  i { height: 64px; border-radius: 18px; }
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 3px;

  i { aspect-ratio: 1; border-radius: 0; }
}

.list {
  margin: 0 16px;
  border-radius: 18px;
  background: var(--elev);
  overflow: hidden;

  .li {
    display: flex;
    align-items: center;
    gap: 13px;
    height: 52px;
    padding: 0 14px;
  }

  .ic { width: 29px; height: 29px; border-radius: 8px; flex: none; }
  .l { height: 12px; }
}
</style>
