import { computed } from 'vue';
import { useIdentity } from '../../about/useIdentity';
import { displayName } from '../../about/identity';
import { useConfigStore } from '../../stores/config';

/** 随想的身份显示偏好，列表、详情及图片预览共用。 */
export function useThoughtIdentity() {
  const identity = useIdentity();
  const config = useConfigStore();
  const alias = computed(() => config.cfg.thoughts.showAlias !== false ? identity.alias.value : '');
  return {
    ...identity,
    alias,
    handle: computed(() => config.cfg.thoughts.showUsername !== false ? config.cfg.github.username.trim() : ''),
    fullName: computed(() => displayName({ name: identity.name.value, alias: alias.value })),
  };
}
