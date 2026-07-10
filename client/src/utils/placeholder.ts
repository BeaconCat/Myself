/** 本地生成 SVG 渐变占位封面（data URI，无任何外部资源） */
export function placeholderCover(from: string, to: string, label: string): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="800" height="600">
  <defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
    <stop offset="0" stop-color="${from}"/><stop offset="1" stop-color="${to}"/>
  </linearGradient></defs>
  <rect width="800" height="600" fill="url(#g)"/>
  <text x="50%" y="52%" text-anchor="middle" font-family="sans-serif" font-size="42"
    fill="rgba(255,255,255,.85)" font-weight="700">${label}</text>
</svg>`;
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}
