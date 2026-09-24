import { computed, watch, type ComputedRef } from 'vue';
import type { AboutModule } from '../../../stores/config';
import { migrateInPlace } from '../../../about/migrate';

/**
 * 编辑器入口：先把模块原地迁移为 about-kit v2 schema（旧 devices / 字符串数组等），
 * 再以目标类型返回 mod.data（computed，模板里直接 d.xxx 使用、v-model 编辑）。
 * mod.data 被整体替换（如重新加载配置）时会再次迁移。
 */
export function useModuleData<T>(getMod: () => AboutModule): ComputedRef<T> {
  migrateInPlace(getMod());
  watch(() => getMod().data, () => migrateInPlace(getMod()));
  return computed(() => getMod().data as T);
}
