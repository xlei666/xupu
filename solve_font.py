
import freetype
import numpy as np
import sys
import os

# Configuration
OBFUSCATED_FONT_PATH = "font.otf"
REFERENCE_FONT_PATH = "/usr/share/fonts/noto-cjk/NotoSansCJK-Regular.ttc"
FONT_SIZE = 32

# Full CJK Range (0x4E00 - 0x9FFF)
COMMON_CHARS = []
for code in range(0x4E00, 0x9FA6): # Basic CJK block
    COMMON_CHARS.append(chr(code))


def render_glyph(face, char_code):
    """Renders a glyph to a numpy array."""
    try:
        if isinstance(char_code, str):
            char_index = face.get_char_index(ord(char_code))
        else:
            char_index = face.get_char_index(char_code)
            
        if char_index == 0:
            print(f"Glyph not found for {hex(char_code) if isinstance(char_code, int) else char_code}")
            return None # Glyph not found
            
        face.load_glyph(char_index, freetype.FT_LOAD_RENDER | freetype.FT_LOAD_TARGET_LIGHT)
        bitmap = face.glyph.bitmap
        
        # Convert to numpy array
        width = bitmap.width
        rows = bitmap.rows
        
        if rows == 0 or width == 0:
             print(f"Empty bitmap for {char_code}")
             return None
             
        buffer = np.array(bitmap.buffer, dtype=np.uint8).reshape((rows, width))
        return buffer
    except Exception as e:
        print(f"Error rendering {hex(char_code) if isinstance(char_code, int) else char_code}: {e}")
        return None

def mse(imageA, imageB):
    # the 'Mean Squared Error' between the two images is the
    # sum of the squared difference between the two images;
    # NOTE: the two images must have the same dimension
    
    # Resize to match dimensions if needed (simple crop/pad)
    hA, wA = imageA.shape
    hB, wB = imageB.shape
    
    # Target size
    h, w = max(hA, hB), max(wA, wB)
    
    # Pad images
    padA = np.zeros((h, w), dtype=np.uint8)
    padA[:hA, :wA] = imageA
    
    padB = np.zeros((h, w), dtype=np.uint8)
    padB[:hB, :wB] = imageB
    
    err = np.sum((padA.astype("float") - padB.astype("float")) ** 2)
    err /= float(h * w)
    
    return err

def main():
    # Load fonts
    face_obs = freetype.Face(OBFUSCATED_FONT_PATH)
    face_obs.set_char_size(FONT_SIZE * 64)
    
    # Debug Charmaps
    print(f"Num Charmaps: {face_obs.num_charmaps}")
    selected_cmap = None
    for i, charmap in enumerate(face_obs.charmaps):
        print(f"Charmap {i}: {charmap.encoding_name} ID: {charmap.platform_id},{charmap.encoding_id}")
        face_obs.select_charmap(charmap.encoding)
        if face_obs.get_char_index(0xe3e8) != 0:
            print(f"  -> Found PUA support in Charmap {i}")
            selected_cmap = charmap.encoding
            break
            
    if selected_cmap:
        face_obs.select_charmap(selected_cmap)
    else:
        print("Warning: No charmap found for PUA 0xe3e8")
    
    face_ref = None
    # Try finding a face in TTC that supports Chinese
    temp_face = freetype.Face(REFERENCE_FONT_PATH)
    num_faces = temp_face.num_faces
    print(f"Reference font has {num_faces} faces.")
    for i in range(num_faces):
        f = freetype.Face(REFERENCE_FONT_PATH, index=i)
        # Check if it supports '中' (0x4E2D)
        idx = f.get_char_index(0x4E2D)
        if idx != 0:
            print(f"Selected Reference Face Index {i} (Supports '中')")
            face_ref = f
            break
            
    if face_ref is None:
        print("Error: No suitable face found in reference font.")
        face_ref = freetype.Face(REFERENCE_FONT_PATH) # Fallback
        
    face_ref.set_char_size(FONT_SIZE * 64)
    
    # Get PUA codes from obfuscated font
    # We can use the analyze_font.py logic or just iterate known range
    # In file, we saw 0xe3e8 ... 
    
    # Let's read the map file to get codes
    pua_codes = []
    with open("font_map.txt", "r") as f:
        for line in f:
            if "Code: 0x" in line:
                code_hex = line.split("Code: ")[1].split(",")[0]
                pua_codes.append(int(code_hex, 16))
                
    print(f"Loaded {len(pua_codes)} PUA codes.")
    
    # Pre-render reference chars
    print("Pre-rendering reference characters...")
    ref_bitmaps = {}
    print(f"Generating bitmaps for {len(COMMON_CHARS)} reference characters...")
    count = 0
    for char in COMMON_CHARS:
        bmp = render_glyph(face_ref, char)
        if bmp is not None:
            ref_bitmaps[char] = bmp
        count += 1
        if count % 2000 == 0:
             print(f"  Generated {count}/{len(COMMON_CHARS)}...", flush=True)
             
    print(f"Generated {len(ref_bitmaps)} reference bitmaps.", flush=True)
            
    # Solve
    mapping = {}
    
    print("Solving...", flush=True)
    count = 0
    for code in pua_codes:
        count += 1
        target_bmp = render_glyph(face_obs, code)
        if target_bmp is None:
            print(f"Failed to render PUA code {hex(code)}", flush=True)
            continue
            
        best_char = "?"
        min_error = float("inf")
        
        for char, ref_bmp in ref_bitmaps.items():
            error = mse(target_bmp, ref_bmp)
            if error < min_error:
                min_error = error
                best_char = char
                
        # Log all
        print(f"Code {hex(code)} -> {best_char} (Error: {min_error})", flush=True)
        
        if min_error < 100000000: # Always map best guess
            mapping[hex(code)] = best_char
        # else:
        #    mapping[hex(code)] = "?"
            
    # Output result
    print("Decryption Map:", flush=True)
    for code, char in mapping.items():
        print(f"{code}:{char}", flush=True)

if __name__ == "__main__":
    main()
