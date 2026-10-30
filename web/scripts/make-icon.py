# 从 public/favicon.ico 生成 Electron / Windows 所需的多分辨率图标 build/icon.ico
#
# 背景：electron-builder 要求图标源文件至少 256x256（否则报错 ERR_ICON_TOO_SMALL），
# 而 public/favicon.ico 原图只有 180x180 单帧，无法直接作为 exe / 安装包图标。
# 这里沿用同一张图，放大到 256 并补齐 Windows 常用尺寸。
#
# 使用：npm run build:icon（需要 Python + Pillow）
#      或在本目录执行 python scripts/make-icon.py

import os
import sys

try:
    from PIL import Image
except ImportError:
    sys.exit('缺少 Pillow，请先执行：pip install Pillow')

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, 'public', 'favicon.ico')
DST = os.path.join(ROOT, 'build', 'icon.ico')

# Windows 图标常用尺寸，electron-builder 要求最大尺寸 >= 256
SIZES = [(256, 256), (128, 128), (64, 64), (48, 48), (32, 32), (24, 24), (16, 16)]


def main():
    if not os.path.exists(SRC):
        sys.exit('源图标不存在: ' + SRC)

    img = Image.open(SRC).convert('RGBA')
    print('源图标 %s  尺寸 %dx%d' % (SRC, img.size[0], img.size[1]))

    max_size = SIZES[0][0]
    base = img if img.size[0] >= max_size else img.resize((max_size, max_size), Image.LANCZOS)

    os.makedirs(os.path.dirname(DST), exist_ok=True)
    base.save(DST, format='ICO', sizes=SIZES)
    print('已生成 %s  %d bytes  包含尺寸: %s'
          % (DST, os.path.getsize(DST), ' '.join('%dx%d' % s for s in SIZES)))


if __name__ == '__main__':
    main()
