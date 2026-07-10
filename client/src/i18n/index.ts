import { createI18n } from 'vue-i18n';
import zhCN from './locales/zh-CN';

/** 默认中文；后续语言在 locales/ 下新增字典并注册即可 */
export const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: { 'zh-CN': zhCN },
});
