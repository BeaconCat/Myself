/**
 * 外链白名单：配置 / 内容里的 URL 绑定到 href 前统一经过这里。
 * 只放行 http(s)、mailto、tel 与站内路径（/ 开头、非 //），其余（javascript:、data:、vbscript: …）返回 undefined，
 * 链接渲染为不可点击。CSP 也会拦截 javascript: 链接，这里是第二道防线。
 */
const ALLOWED = /^(https?:|mailto:|tel:)/i;

export function safeHref(url: string | undefined | null): string | undefined {
  const u = (url ?? '').trim();
  if (!u) return undefined;
  if (u.startsWith('/') && !u.startsWith('//')) return u;
  if (u.startsWith('#')) return u;
  // 去掉控制字符与空白后再判断协议，防止 "java\tscript:" 之类的绕过
  const probe = u.replace(/[\u0000- ]/g, '');
  return ALLOWED.test(probe) ? u : undefined;
}
