
import json
import re

def decrypt_content():
    # Load mapping
    mapping = {}
    with open("decryption_map_final.txt", "r") as f:
        for line in f:
            if "Code 0x" in line: # Debug lines
                continue
            if ":" in line and line.strip().startswith("0x"):
                parts = line.strip().split(":")
                if len(parts) >= 2: # Handle cases like 0xE4C2: (empty char?)
                    code_hex = parts[0].strip()
                    char = parts[1].strip()
                    try:
                        mapping[int(code_hex, 16)] = char
                    except ValueError:
                        continue

    # Read HTML
    with open("reader_page.html", "r") as f:
        html = f.read()

    # Extract JSON
    # Look for window.__INITIAL_STATE__={
    match = re.search(r'window\.__INITIAL_STATE__=(.*?);', html, re.DOTALL)
    if not match:
        print("Could not find initial state in HTML")
        return

    json_str = match.group(1)
    
    # Fix JSON if needed (the regex might capture extra)
    # Usually it ends with }; before the script tag.
    # We might need to be careful with parsing.
    
    try:
        data = json.loads(json_str)
    except json.JSONDecodeError as e:
        print(f"JSON Parse Error: {e}")
        # print(f"JSON Str snippet: {json_str[:100]} ... {json_str[-100:]}")
        pass

    content = ""
    if 'data' in locals() and 'reader' in data and 'chapterData' in data['reader']:
        content = data['reader']['chapterData']['content']
    else:
        # Fallback regex for content
        match_content = re.search(r'"content":"(.*?)"', html)
        if match_content:
            try:
                # Treat the extracted string as a JSON string literal
                content = json.loads(f'"{match_content.group(1)}"')
            except:
                # If that fails, just use raw string (might have \u escapes but readable)
                content = match_content.group(1)
        else:
            print("Could not extract content")
            return

    # Decrypt
    decrypted_chars = []
    for char in content:
        code = ord(char)
        if code in mapping:
            decrypted_chars.append(mapping[code])
        else:
            decrypted_chars.append(char)

    decrypted_text = "".join(decrypted_chars)
    
    # Clean up HTML tags for readability
    decrypted_text = decrypted_text.replace("<p>", "\n").replace("</p>", "").replace("\\\"", "\"")
    
    print(decrypted_text)
    
    with open("decrypted_chapter.txt", "w") as f:
        f.write(decrypted_text)

if __name__ == "__main__":
    decrypt_content()
