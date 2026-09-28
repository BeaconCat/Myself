"""生成默认封面（client/public/covers/*.webp）：12 幅程序化抽象构图，编辑风格、低饱和、带纸面颗粒。

用途：无封面文章 / 随想配图的兜底，与 Demo 数据的封面。构图随机但固定种子，可重复生成：
    python scripts/gen_covers.py
每幅输出两档：NN.webp（1600×1000）与 NN-s.webp（640×400，列表缩略图）。
依赖：Pillow、numpy（仅开发时运行，产物入库）。
"""
import math
import os
import random

import numpy as np
from PIL import Image, ImageDraw, ImageFilter

W, H = 1600, 1000
SS = 2  # 形状超采样倍数，抗锯齿
OUT = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'client', 'public', 'covers')


def hexrgb(h):
    h = h.lstrip('#')
    return np.array([int(h[i:i + 2], 16) for i in (0, 2, 4)], dtype=np.float32)


def vgrad(top, bottom, curve=1.0):
    t = np.linspace(0, 1, H, dtype=np.float32)[:, None, None] ** curve
    return np.broadcast_to(hexrgb(top) * (1 - t) + hexrgb(bottom) * t, (H, W, 3)).copy()


def radial(img, cx, cy, r, color, alpha):
    """在 img 上叠一团径向柔光（屏幕坐标，r 为半径像素）"""
    y, x = np.mgrid[0:H, 0:W].astype(np.float32)
    d = np.sqrt((x - cx) ** 2 + (y - cy) ** 2) / r
    a = np.clip(1 - d, 0, 1) ** 2 * alpha
    img[:] = img * (1 - a[..., None]) + hexrgb(color) * a[..., None]


def grain(img, amount=6.0, seed=0):
    rng = np.random.default_rng(seed)
    n = rng.normal(0, amount, (H, W, 1)).astype(np.float32)
    return np.clip(img + n, 0, 255)


def vignette(img, strength=0.18):
    y, x = np.mgrid[0:H, 0:W].astype(np.float32)
    d = np.sqrt(((x - W / 2) / (W / 2)) ** 2 + ((y - H / 2) / (H / 2)) ** 2)
    k = 1 - strength * np.clip(d - 0.35, 0, 1) ** 1.6
    return img * k[..., None]


class Layer:
    """超采样形状层：draw 在 2 倍画布上画，composite 时缩回并按 alpha 叠加"""

    def __init__(self):
        self.im = Image.new('RGBA', (W * SS, H * SS), (0, 0, 0, 0))
        self.d = ImageDraw.Draw(self.im)

    def s(self, *v):
        return [c * SS for c in v]

    def over(self, base, blur=0):
        im = self.im.resize((W, H), Image.LANCZOS)
        if blur:
            im = im.filter(ImageFilter.GaussianBlur(blur))
        a = np.asarray(im, dtype=np.float32)
        al = a[..., 3:4] / 255
        return base * (1 - al) + a[..., :3] * al


def rgba(h, a=255):
    return tuple(int(c) for c in hexrgb(h)) + (a,)


# ---------------------------------------------------------------- 构图

def c01_dusk():
    """地平线：暖奶油到杏色，半落的日轮与一道地平线"""
    img = vgrad('#f3e7d7', '#e9c7a8', 1.2)
    radial(img, W * 0.62, H * 0.66, 520, '#f7d9b8', 0.55)
    L = Layer()
    L.d.ellipse(L.s(W * 0.62 - 190, H * 0.66 - 190, W * 0.62 + 190, H * 0.66 + 190), fill=rgba('#d9663f'))
    img = L.over(img)
    sea = vgrad('#d8b294', '#c99a7c', 1.0)
    img[int(H * 0.66):] = sea[int(H * 0.66):]
    L = Layer()
    for i in range(7):
        y = H * 0.70 + i * 26
        w = 260 - i * 30
        L.d.rounded_rectangle(L.s(W * 0.62 - w / 2, y, W * 0.62 + w / 2, y + 5), radius=6, fill=rgba('#e98a5f', 150 - i * 18))
    img = L.over(img)
    L = Layer()
    L.d.rectangle(L.s(0, H * 0.66 - 1, W, H * 0.66 + 1), fill=rgba('#8c5a44', 90))
    return L.over(img)


def c02_moon():
    """月相：深墨蓝上一轮米白月与一道影"""
    img = vgrad('#1b2436', '#0e1422', 1.0)
    radial(img, W * 0.35, H * 0.4, 700, '#2a3854', 0.6)
    L = Layer()
    cx, cy, r = W * 0.36, H * 0.46, 230
    L.d.ellipse(L.s(cx - r, cy - r, cx + r, cy + r), fill=rgba('#ece4d4'))
    img = L.over(img)
    L = Layer()
    L.d.ellipse(L.s(cx - r + 120, cy - r - 30, cx + r + 120, cy + r - 30), fill=rgba('#161e2e', 235))
    img = L.over(img, blur=2)
    rng = random.Random(2)
    L = Layer()
    for _ in range(70):
        x, y = rng.uniform(0, W), rng.uniform(0, H * 0.9)
        if math.hypot(x - cx, y - cy) < r + 40:
            continue
        s = rng.choice([1.2, 1.6, 2.2])
        L.d.ellipse(L.s(x - s, y - s, x + s, y + s), fill=rgba('#e8e2d6', rng.randint(60, 170)))
    img = L.over(img)
    L = Layer()
    L.d.rectangle(L.s(W * 0.58, H * 0.46, W * 0.9, H * 0.46 + 2), fill=rgba('#c9bfae', 110))
    return L.over(img)


def c03_sheets():
    """叠纸：鼠尾草绿底上三张错落纸页与柔影"""
    img = vgrad('#b9c3ad', '#a3af96', 1.0)
    for i, (x, y, rot, col) in enumerate([(0.30, 0.54, -7, '#eef0e6'), (0.47, 0.48, 3, '#f6f5ee'), (0.64, 0.55, 9, '#e3e7d8')]):
        w, h = 430, 560
        sh = Image.new('RGBA', (W, H), (0, 0, 0, 0))
        page = Image.new('RGBA', (w, h), rgba(col))
        pd = ImageDraw.Draw(page)
        for k in range(9):
            yy = 90 + k * 44
            pd.rectangle((60, yy, w - 60 - (k % 3) * 50, yy + 3), fill=rgba('#9aa38e', 120))
        shadow = Image.new('RGBA', (w, h), (40, 50, 30, 90))
        shadow = shadow.rotate(rot, expand=True, resample=Image.BICUBIC)
        page = page.rotate(rot, expand=True, resample=Image.BICUBIC)
        px, py = int(W * x - page.width / 2), int(H * y - page.height / 2)
        sh.alpha_composite(shadow, (px + 18, py + 26))
        sh = sh.filter(ImageFilter.GaussianBlur(22))
        sh.alpha_composite(page, (px, py))
        a = np.asarray(sh, dtype=np.float32)
        al = a[..., 3:4] / 255
        img = img * (1 - al) + a[..., :3] * al
    return img


def c04_arch():
    """拱窗：沙色墙面上一扇陶土色拱门与投下的光"""
    img = vgrad('#ead9c3', '#dcc4a8', 1.0)
    L = Layer()
    x0, x1, top, bot = W * 0.40, W * 0.60, H * 0.18, H * 0.86
    r = (x1 - x0) / 2
    L.d.rectangle(L.s(x0, top + r, x1, bot), fill=rgba('#b8603f'))
    L.d.ellipse(L.s(x0, top, x1, top + 2 * r), fill=rgba('#b8603f'))
    img = L.over(img)
    L = Layer()
    L.d.polygon(L.s(x0, bot, x1, bot, x1 + 340, H, x0 + 120, H), fill=rgba('#f5e6d0', 150))
    img = L.over(img, blur=6)
    L = Layer()
    L.d.rectangle(L.s(0, bot, W, bot + 2), fill=rgba('#a88a6c', 120))
    return L.over(img)


def c05_tide():
    """潮汐：层叠的正弦色带，由浅到深的蓝"""
    img = vgrad('#e4ebf0', '#cfdbe4', 1.0)
    cols = ['#b8cad8', '#8fabc4', '#6b8fb0', '#4a6f95', '#2f5178']
    for i, c in enumerate(cols):
        L = Layer()
        base = H * (0.42 + i * 0.11)
        pts = [(0, H)]
        for x in range(0, W + 20, 20):
            y = base + math.sin(x / (230 + i * 40) + i * 1.3) * (34 - i * 3) + math.sin(x / 90 + i) * 6
            pts.append((x, y))
        pts.append((W, H))
        L.d.polygon([(px * SS, py * SS) for px, py in pts], fill=rgba(c))
        img = L.over(img)
    return img


def c06_grid():
    """格物：米白纸上的发丝网格，一格填以品牌黄"""
    img = vgrad('#f4f1ea', '#ece7dc', 1.0)
    L = Layer()
    step = 100
    for x in range(100, W, step):
        L.d.rectangle(L.s(x, 0, x + 1, H), fill=rgba('#b9b1a2', 110))
    for y in range(100, H, step):
        L.d.rectangle(L.s(0, y, W, y + 1), fill=rgba('#b9b1a2', 110))
    L.d.rectangle(L.s(1000, 400, 1100, 500), fill=rgba('#ffb300'))
    L.d.rectangle(L.s(500, 600, 600, 700), fill=rgba('#1f2430'))
    L.d.ellipse(L.s(700, 200, 900, 400), outline=rgba('#1f2430', 200), width=3 * SS)
    return L.over(img)


def c07_mesh():
    """光斑：低饱和的红黄蓝色场交融（品牌三色的柔化）"""
    img = vgrad('#efe9e1', '#e6ddd2', 1.0)
    radial(img, W * 0.25, H * 0.35, 760, '#e98b86', 0.7)
    radial(img, W * 0.75, H * 0.30, 700, '#f2c46b', 0.7)
    radial(img, W * 0.55, H * 0.85, 820, '#7fa6d9', 0.75)
    return img


def c08_ridges():
    """山影：大气透视的层层山脊"""
    img = vgrad('#e6e3e0', '#cfd3d8', 1.0)
    rng = random.Random(8)
    cols = ['#c3c8cf', '#a5adb8', '#838e9d', '#5f6b7c', '#3d4757']
    for i, c in enumerate(cols):
        L = Layer()
        base = H * (0.40 + i * 0.1)
        pts = [(0, H)]
        y = base
        for x in range(0, W + 40, 40):
            y += rng.uniform(-22, 22)
            y = min(max(y, base - 90), base + 60)
            pts.append((x, y))
        pts.append((W, H))
        L.d.polygon([(px * SS, py * SS) for px, py in pts], fill=rgba(c))
        img = L.over(img, blur=0.6)
    radial(img, W * 0.7, H * 0.22, 220, '#f6efe6', 0.9)
    return img


def c09_spines():
    """书脊：不等宽竖条，一组安静的书"""
    img = vgrad('#efe8dd', '#e4dccd', 1.0)
    rng = random.Random(9)
    pal = ['#2f3b52', '#8a3b2e', '#c9a25a', '#5c7160', '#d8cbb3', '#44506a', '#a5654b', '#e9dfcc']
    x = 180
    L = Layer()
    floor = H * 0.8
    while x < W - 200:
        w = rng.choice([46, 58, 70, 84, 38])
        h = rng.uniform(330, 520)
        col = rng.choice(pal)
        lean = rng.random() < 0.08
        if lean:
            x += 30
        L.d.rectangle(L.s(x, floor - h, x + w, floor), fill=rgba(col))
        L.d.rectangle(L.s(x + 8, floor - h + 40, x + w - 8, floor - h + 44), fill=rgba('#f3ead8', 140))
        x += w + 4
    img = L.over(img)
    L = Layer()
    L.d.rectangle(L.s(120, floor, W - 120, floor + 10), fill=rgba('#8f7f68'))
    return L.over(img)


def c10_circles():
    """圆舞：三枚半透明圆相交（红 / 黄 / 蓝）"""
    img = vgrad('#f2eee7', '#e9e3d9', 1.0)
    out = img.copy()
    for cx, cy, col in [(0.40, 0.44, '#ff5a5f'), (0.56, 0.44, '#ffb300'), (0.48, 0.60, '#2f7bff')]:
        L = Layer()
        r = 230
        L.d.ellipse(L.s(W * cx - r, H * cy - r, W * cx + r, H * cy + r), fill=rgba(col))
        im = np.asarray(L.im.resize((W, H), Image.LANCZOS), dtype=np.float32)
        al = im[..., 3:4] / 255 * 0.78
        out = out * (1 - al) + (out * im[..., :3] / 255) * al  # multiply
    return out


def c11_rain():
    """雨线：石板灰上的斜向细线，一点暖光"""
    img = vgrad('#5d6673', '#3f4652', 1.0)
    radial(img, W * 0.72, H * 0.3, 380, '#c9a27a', 0.35)
    rng = random.Random(11)
    L = Layer()
    for _ in range(260):
        x, y = rng.uniform(-200, W), rng.uniform(-100, H)
        ln = rng.uniform(40, 120)
        L.d.line(L.s(x, y, x + ln * 0.35, y + ln), fill=rgba('#d7dde6', rng.randint(40, 110)), width=SS)
    return L.over(img)


def c12_window():
    """纸窗：米白墙上斜落的窗格光影"""
    img = vgrad('#efe7da', '#e4d8c6', 1.0)
    L = Layer()
    ox, oy = W * 0.28, H * 0.12
    sk = 0.42
    for r in range(2):
        for c in range(3):
            x = ox + c * 250 + r * 380 * sk
            y = oy + r * 380
            pts = [(x, y), (x + 230, y), (x + 230 + 360 * sk, y + 360), (x + 360 * sk, y + 360)]
            L.d.polygon([(px * SS, py * SS) for px, py in pts], fill=rgba('#fbf5ea', 235))
    img = L.over(img, blur=5)
    radial(img, W * 0.2, H * 0.9, 600, '#d6c3a8', 0.35)
    return img


COVERS = [c01_dusk, c02_moon, c03_sheets, c04_arch, c05_tide, c06_grid, c07_mesh, c08_ridges, c09_spines, c10_circles, c11_rain, c12_window]


def main():
    os.makedirs(OUT, exist_ok=True)
    for i, fn in enumerate(COVERS, 1):
        img = fn()
        img = vignette(img, 0.14)
        img = grain(img, 5.0, seed=i)
        im = Image.fromarray(np.clip(img, 0, 255).astype(np.uint8), 'RGB')
        im.save(os.path.join(OUT, f'{i:02d}.webp'), 'WEBP', quality=84, method=6)
        im.resize((640, 400), Image.LANCZOS).save(os.path.join(OUT, f'{i:02d}-s.webp'), 'WEBP', quality=82, method=6)
        print(f'{i:02d}', fn.__doc__.split('：')[0])


if __name__ == '__main__':
    main()
