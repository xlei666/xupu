#!/usr/bin/env python3
"""
番茄小说字体解密脚本
使用方式: python3 solve_font.py <字体文件路径>
输出格式: 0xHHHH:字 (每行一个映射)
"""

import freetype
import numpy as np
import sys
import os

# 参考字体路径
REFERENCE_FONT_PATHS = [
    "/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc",
    "/usr/share/fonts/google-noto-cjk/NotoSansCJK-Regular.ttc",
    "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
    "/usr/share/fonts/noto/NotoSansCJK-Regular.ttc",
]

FONT_SIZE = 32

# CJK范围
CJK_START = 0x4E00
CJK_END = 0x9FA6

# PUA范围
PUA_START = 0xE000
PUA_END = 0xF8FF


def find_reference_font():
    """查找系统中的参考字体"""
    for path in REFERENCE_FONT_PATHS:
        if os.path.exists(path):
            return path
    return None


def render_glyph(face, char_code):
    """渲染字形到numpy数组"""
    try:
        if isinstance(char_code, str):
            char_index = face.get_char_index(ord(char_code))
        else:
            char_index = face.get_char_index(char_code)
            
        if char_index == 0:
            return None
            
        face.load_glyph(char_index, freetype.FT_LOAD_RENDER | freetype.FT_LOAD_TARGET_LIGHT)
        bitmap = face.glyph.bitmap
        
        width = bitmap.width
        rows = bitmap.rows
        
        if rows == 0 or width == 0:
            return None
             
        buffer = np.array(bitmap.buffer, dtype=np.uint8).reshape((rows, width))
        return buffer
    except Exception:
        return None


def mse(imageA, imageB):
    """计算均方误差"""
    if imageA is None or imageB is None:
        return float("inf")
    
    # 调整到相同尺寸
    h1, w1 = imageA.shape
    h2, w2 = imageB.shape
    h = max(h1, h2)
    w = max(w1, w2)
    
    padded_a = np.zeros((h, w), dtype=np.float64)
    padded_b = np.zeros((h, w), dtype=np.float64)
    
    y_off_a = (h - h1) // 2
    x_off_a = (w - w1) // 2
    padded_a[y_off_a:y_off_a+h1, x_off_a:x_off_a+w1] = imageA
    
    y_off_b = (h - h2) // 2
    x_off_b = (w - w2) // 2
    padded_b[y_off_b:y_off_b+h2, x_off_b:x_off_b+w2] = imageB
    
    diff = padded_a - padded_b
    return np.sum(diff ** 2) / float(h * w)


def collect_pua_codes(face):
    """收集字体中的PUA编码"""
    codes = []
    for code in range(PUA_START, PUA_END + 1):
        if face.get_char_index(code) != 0:
            codes.append(code)
    return codes


def collect_cjk_chars(face):
    """收集字体中的CJK字符"""
    chars = []
    for code in range(CJK_START, CJK_END + 1):
        if face.get_char_index(code) != 0:
            chars.append(chr(code))
    return chars


def main():
    if len(sys.argv) < 2:
        print("用法: python3 solve_font.py <字体文件路径>", file=sys.stderr)
        sys.exit(1)
    
    font_path = sys.argv[1]
    if not os.path.exists(font_path):
        print(f"字体文件不存在: {font_path}", file=sys.stderr)
        sys.exit(1)
    
    # 查找参考字体
    ref_font_path = find_reference_font()
    if not ref_font_path:
        print("未找到参考字体(NotoSansCJK)", file=sys.stderr)
        sys.exit(1)
    
    # 加载加密字体
    try:
        face_obs = freetype.Face(font_path)
        face_obs.set_char_size(FONT_SIZE * 64)
    except Exception as e:
        print(f"加载加密字体失败: {e}", file=sys.stderr)
        sys.exit(1)
    
    # 加载参考字体
    try:
        face_ref = None
        for i in range(freetype.Face(ref_font_path).num_faces):
            f = freetype.Face(ref_font_path, i)
            idx = f.get_char_index(0x4E2D)  # 测试'中'
            if idx != 0:
                face_ref = f
                break    
        if face_ref is None:
            face_ref = freetype.Face(ref_font_path)
        face_ref.set_char_size(FONT_SIZE * 64)
    except Exception as e:
        print(f"加载参考字体失败: {e}", file=sys.stderr)
        sys.exit(1)
    
    # 收集字体中的字符
    pua_codes = collect_pua_codes(face_obs)
    cjk_chars = collect_cjk_chars(face_obs)
    
    if not pua_codes:
        print("字体中未找到PUA编码", file=sys.stderr)
        sys.exit(1)
    
    # 优先使用字体自带的CJK字符，否则使用参考字体
    if len(cjk_chars) >= len(pua_codes):
        # 字体自带CJK，直接比对
        ref_face = face_obs
        ref_chars = cjk_chars
    else:
        # 使用参考字体，需要更大范围
        ref_face = face_ref
        ref_chars = [chr(c) for c in range(CJK_START, CJK_END + 1)]
    
    # 预渲染参考字符
    ref_bitmaps = {}
    for char in ref_chars:
        bmp = render_glyph(ref_face, char)
        if bmp is not None:
            ref_bitmaps[char] = bmp
    
    # 匹配
    for code in pua_codes:
        target_bmp = render_glyph(face_obs, code)
        if target_bmp is None:
            continue
            
        best_char = None
        min_error = float("inf")
        
        for char, ref_bmp in ref_bitmaps.items():
            error = mse(target_bmp, ref_bmp)
            if error < min_error:
                min_error = error
                best_char = char
        
        if best_char and min_error < 10000:  # 阈值
            print(f"0x{code:04x}:{best_char}", flush=True)


if __name__ == "__main__":
    main()
