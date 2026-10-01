/** Scheduling inputs use site-local wall time; API responses use explicit UTC. */
export function scheduleInput(value: string, timezone: string): string {
  if (!value) return '';
  if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2})?$/.test(value)) return value.length === 16 ? `${value}:00` : value;
  const date = new Date(/[zZ]|[+-]\d\d:?\d\d$/.test(value) ? value : `${value.replace(' ', 'T')}Z`);
  if (!Number.isFinite(date.getTime())) return '';
  const fields = Object.fromEntries(new Intl.DateTimeFormat('en-CA', {
    timeZone: timezone, year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
  }).formatToParts(date).map(part => [part.type, part.value]));
  return `${fields.year}-${fields.month}-${fields.day}T${fields.hour}:${fields.minute}:${fields.second}`;
}

export function scheduleValid(value: string | null | undefined, timezone: string, now = Date.now()): boolean {
  if (value == null) return true;
  if (/[zZ]|[+-]\d\d:?\d\d$/.test(value)) return new Date(value).getTime() > now;
  const local = scheduleInput(value, timezone);
  return !!local && local > scheduleInput(new Date(now).toISOString(), timezone);
}

export function scheduleLabel(value: string, timezone: string): string {
  return scheduleInput(value, timezone).replace('T', ' ');
}
