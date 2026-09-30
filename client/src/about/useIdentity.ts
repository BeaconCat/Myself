import { computed } from 'vue';
import { useConfigStore } from '../stores/config';
import { displayName, plainText } from './identity';
import { BUILTIN_LOGO } from '../utils/siteLogo';

/**
 * 站点身份的展示值（全站统一取这里）：头像（未上传时用站点 logo）、名字（别名）、签名（去掉高亮标记）、格言。
 * 作者栏、页脚、抽屉、随想信息流等处都用它，避免各处自行拼接出现不一致。
 */
export function useIdentity() {
  const config = useConfigStore();
  const about = computed(() => config.cfg.about);
  return {
    avatar: computed(() => about.value.avatar || config.cfg.site.logo || BUILTIN_LOGO),
    hasAvatar: computed(() => !!about.value.avatar),
    name: computed(() => about.value.name || config.cfg.site.title),
    fullName: computed(() => displayName(about.value, config.cfg.site.title)),
    sign: computed(() => plainText(about.value.tagline)),
    motto: computed(() => about.value.motto),
  };
}
