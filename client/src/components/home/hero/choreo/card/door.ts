import { fade } from '../engine';
import type { CardChoreo, ChoreoCtx } from '../types';

/**
 * 构造一扇门：preserve-3d 的门扇 = 正面（半幅画面 + 渐暗层）+ 自由边侧面（厚度）+ 背面（暗化）。
 * 门扇不做 opacity / filter / clip（它们会强制拍平 3D），淡出与变暗只作用在子面上。
 */
function buildLeaf(t: ChoreoCtx, card: HTMLElement, sheet: HTMLElement, side: 'l' | 'r') {
  const leaf = t.make(card, `hc-leaf hc-leaf-${side}`);
  const face = document.createElement('div');
  face.className = 'hc-leaf-face';
  const pic = document.createElement('div');
  pic.className = 'hc-leaf-pic';
  for (const node of Array.from(sheet.children)) pic.append(node.cloneNode(true));
  const shade = document.createElement('div');
  shade.className = 'hc-leaf-shade';
  face.append(pic, shade);
  const edge = document.createElement('div');
  edge.className = 'hc-leaf-edge';
  const back = document.createElement('div');
  back.className = 'hc-leaf-back';
  leaf.append(back, edge, face);
  return { leaf, face, shade, edge, back };
}

/**
 * 推门见光（原 08 卡片部分，按定稿改为向外开）：
 * 旧封面从中线分成两扇门，以外缘为轴朝观者方向转出（门扇有厚度、转过去的背面发暗）；
 * 门后是过曝的下一篇，曝光慢慢回落；门后涌光并在地面铺开一道光扇。
 */
export const door: CardChoreo = {
  meta: {
    id: 'door',
    name: '推门见光',
    tag: '叙事',
    desc: '封面像两扇门朝你打开，门后过曝的新封面曝光回落。',
    duration: 1480,
  },
  exit(e, t) {
    const f = e.front;
    // 消失点在卡片中心：card 层 perspective 属性（filter 置空避免拍平）
    t.set(f.el, {
      perspective: `${Math.round(e.album.offsetWidth * 2.4)}px`,
      perspectiveOrigin: '50% 50%',
      filter: 'none',
    });
    t.set(f.sheet, { opacity: '0' });
    (['l', 'r'] as const).forEach((side) => {
      const dir = side === 'l' ? -1 : 1;
      const p = buildLeaf(t, f.el, f.sheet, side);
      // 先裂开一道缝（门后光溢出），再加速朝观者转出，接近 110° 时离开画面
      t.a(p.leaf, [
        { transform: 'rotateY(0deg)' },
        { transform: `rotateY(${dir * 9}deg)`, offset: 0.2 },
        { transform: `rotateY(${dir * 112}deg)` },
      ], { dur: 820, ease: 'quartInOut' });
      t.a(p.shade, [{ opacity: 0 }, { opacity: 0.2, offset: 0.2 }, { opacity: 0.72 }], { dur: 820, ease: 'quartIn' });
      [p.face, p.edge, p.back].forEach((el) => t.a(el, [
        { opacity: 1 },
        { opacity: 1, offset: 0.78 },
        { opacity: 0 },
      ], { dur: 820, ease: 'linear' }));
    });
    e.backs.forEach((c, i) => t.a(c.el, [
      { filter: e.F(c.b), opacity: 1 },
      { filter: e.F(c.b * 0.15), opacity: 0 },
    ], { dur: 420, delay: 60 + i * 40, ease: 'quartIn' }));
    t.a(e.bg, fade(1, 0), { dur: 420, ease: 'quartIn' });
  },
  enter(e, t) {
    const f = e.front;
    // 门开的前半程保持过曝（门缝里透出的是白光），门完全打开后曝光才回落
    t.a(f.el, [
      { filter: e.F(3.4, 0, 0.3), transform: e.T(0, { ds: 0.93 }) },
      { filter: e.F(3, 0, 0.4), transform: e.T(0, { ds: 0.95 }), offset: 0.3 },
      { filter: e.F(1), transform: e.T(0) },
    ], { dur: 1280, delay: 60, ease: 'quartInOut' });
    e.backs.forEach((c, i) => t.a(c.el, [
      { transform: e.T(c.slot, { dz: -80 }), filter: e.F(c.b * 2.4), opacity: 0 },
      { transform: e.T(c.slot), filter: e.F(c.b), opacity: 1 },
    ], { dur: 760, delay: 640 + i * 80 }));
    const r = t.rel(e.album);
    // 门后涌光：门缝打开的瞬间最亮，随曝光回落
    t.set(t.fx.flare, {
      left: `${r.cx - r.w * 0.5}px`,
      top: `${r.cy - r.h * 0.6}px`,
      width: `${r.w}px`,
      height: `${r.h * 1.2}px`,
    });
    t.a(t.fx.flare, [
      { opacity: 0, transform: 'scaleX(.1)' },
      { opacity: 1, transform: 'scaleX(.55)', offset: 0.3 },
      { opacity: 0, transform: 'scaleX(1)' },
    ], { dur: 900, delay: 100, ease: 'brand' });
    // 地面光：从门底铺向文字方向（桌面向左下，移动端向上）
    if (t.mobile) {
      const w = r.w * 1.3;
      t.set(t.fx.floor, {
        left: `${r.cx - w / 2}px`,
        width: `${w}px`,
        top: `${r.y - 340}px`,
        height: '360px',
        clipPath: 'polygon(46% 100%, 54% 100%, 100% 0, 0 0)',
        transformOrigin: '50% 100%',
        background: 'linear-gradient(0deg, rgba(235,244,255,.5), color-mix(in srgb, var(--primary) 20%, transparent) 50%, transparent 92%)',
      });
    } else {
      const w = Math.max(r.cx - 20, 360);
      const top = r.y + r.h * 0.72;
      t.set(t.fx.floor, {
        left: `${r.cx - w * 0.9}px`,
        top: `${top}px`,
        width: `${w}px`,
        height: `${Math.max(t.H - top + 40, 120)}px`,
      });
    }
    t.a(t.fx.floor, [
      { opacity: 0, transform: 'scaleX(.15) scaleY(.4)' },
      { opacity: 1, transform: 'scaleX(1) scaleY(1)', offset: 0.38 },
      { opacity: 0, transform: 'scaleX(1.05) scaleY(1)' },
    ], { dur: 1300, delay: 160, ease: 'brand' });
    t.a(e.bg, fade(0, 1), { dur: 1000, delay: 280 });
  },
};
