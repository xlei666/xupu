from fontTools.ttLib import TTFont
import sys

def analyze_font(font_path):
    try:
        font = TTFont(font_path)
        cmap = font.getBestCmap()
        
        # In Fanqie obfuscation, the 'best' cmap usually maps high-value unicode code points 
        # (like 0xE4C2) to glyph IDs.
        # We need to print these mappings.
        
        print(f"Total mapped characters: {len(cmap)}")
        for code, name in cmap.items():
            print(f"Code: {hex(code)}, Name: {name}")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    analyze_font("font.otf")
