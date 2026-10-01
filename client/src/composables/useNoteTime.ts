import { useI18n } from 'vue-i18n';
import { useNow } from './useNow';
import { formatDate } from '../utils/date';
import { useConfigStore } from '../stores/config';

/** Minutes/hours for recent notes; a compact site-local date for older notes. */
export function useNoteTime() {
  const now = useNow();
  const { t } = useI18n();
  const config = useConfigStore();
  return (value: string): string => {
    const elapsed = Math.max(0, now.value - new Date(`${value.replace(' ', 'T')}Z`).getTime());
    if (elapsed < 60_000) return t('thoughts.justNow');
    if (elapsed < 3_600_000) return t('thoughts.minutesAgo', { n: Math.floor(elapsed / 60_000) });
    if (elapsed < 86_400_000) return t('thoughts.hoursAgo', { n: Math.floor(elapsed / 3_600_000) });
    const [year, month, day] = formatDate(value).split('-');
    const currentYear = new Intl.DateTimeFormat('en', { timeZone: config.cfg.timezone, year: 'numeric' }).format(now.value);
    return t(year === currentYear ? 'thoughts.shortDate' : 'thoughts.fullDate', { y: year, m: Number(month), d: Number(day) });
  };
}
