<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import type { AboutModule } from '../stores/config';
import { metaOf, spanOf, titleOf, variantOf } from './registry';
import { injectIdentity, type Identity } from './identity';
import { MODULE_COMPONENTS } from './components';
import './kit.scss';

/**
 * 单个模块的高保真预览（后台「关于」积木卡片用）：真实模块组件按「卡片宽 ÷ ZOOM」的宽度排版
 * （模块内部的容器查询按这个宽度生效），再缩小 ZOOM 倍铺满卡片，看起来是前台的等比缩略；
 * 按真实高度完整显示（不截断）。只读：不接收指针事件。
 */
const props = withDefaults(defineProps<{
  mod: AboutModule;
  about: Partial<Identity>;
  /** 章节编号（01、02…），与前台按顺序自动编号一致 */
  no?: string;
}>(), { no: '01' });

/** 缩略比例：文字约为前台的八成，既看得清又不喧宾夺主 */
const ZOOM = 0.8;

const item = computed(() => {
  const mod = injectIdentity(props.mod, props.about);
  const variant = variantOf(mod);
  return {
    mod,
    variant,
    span: spanOf(mod),
    title: titleOf(mod),
    chrome: metaOf(mod.type)?.chrome(variant) ?? true,
  };
});

const box = ref<HTMLElement | null>(null);
const cell = ref<HTMLElement | null>(null);
const width = ref(0);
const natural = ref(0);
let ro: ResizeObserver | null = null;

function measure(): void {
  if (!box.value || !cell.value) return;
  width.value = box.value.clientWidth / ZOOM;
  natural.value = cell.value.offsetHeight * ZOOM;
}

onMounted(() => {
  ro = new ResizeObserver(measure);
  if (box.value) ro.observe(box.value);
  if (cell.value) ro.observe(cell.value);
  measure();
});
onBeforeUnmount(() => ro?.disconnect());

</script>

<template>
  <div
    ref="box"
    class="mp"
    :style="{ height: natural ? `${natural}px` : undefined }"
    aria-hidden="true"
    inert
  >
    <div class="mp-stage ak" :style="{ width: width ? `${width}px` : undefined, transform: `scale(${ZOOM})` }">
      <section
        ref="cell"
        class="ak-m rv in"
        :class="[`m-${item.mod.type}`, { card: item.chrome }]"
        :data-span="item.span"
        :data-type="item.mod.type"
      >
        <div class="ak-body">
          <component
            :is="MODULE_COMPONENTS[item.mod.type]"
            v-if="MODULE_COMPONENTS[item.mod.type]"
            :mod="item.mod"
            :variant="item.variant"
            :span="item.span"
            :title="item.title"
            :no="no"
          />
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped lang="scss">
.mp {
  position: relative;
  min-height: 48px;
  overflow: hidden;
  pointer-events: none;
  user-select: none;
  transition: height var(--dur) var(--ease-out);
}

/* 排版宽度 = 卡片宽 ÷ ZOOM，左上角为原点缩小回卡片宽 */
.mp-stage {
  position: absolute;
  top: 0;
  left: 0;
  transform-origin: 0 0;
}

/* 身份区在前台顶部留白让出悬浮导航；预览里不需要，贴顶显示 */
.mp-stage :deep(.m-profile .pf) { padding-top: 0; }
</style>
