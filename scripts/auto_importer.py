
import requests
import json
import os
import subprocess
import time
import shutil

# Configuration
API_URL = "http://localhost:8080"
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYmU1NTdmMzQtM2IzMC00ODY4LTk5MjAtMzAwMDBiMmUyNzk5IiwiZW1haWwiOiJpbXBvcnRlckB0ZXN0LmNvbSIsInRpZXIiOiJmcmVlIiwiaXNzIjoieHVwdS1hcGkiLCJleHAiOjE3NzAyNzAzMTAsIm5iZiI6MTc3MDE4MzkxMCwiaWF0IjoxNzcwMTgzOTEwfQ.uNBHhI0ygMsVh5m0nYp08PDIlta4ZXyQ9G1LLXQttPM"
DOWNLOADER_PATH = "./TomatoNovelDownloader-Linux_amd64-v2.1.6"
DOWNLOAD_DIR = "downloaded_novels"

# Fanqie Mobile API
RANKS_URL = "https://api-lf.fanqiesdk.com/api/novel/channel/homepage/rank/rank_list/v2/"

def get_hot_list(limit=5):
    print(f"Fetching hot list (top {limit})...", flush=True)
    params = {
        "aid": "1967",
        "channel": "0",
        "device_platform": "android",
        "device_type": "0",
        "limit": str(limit),
        "offset": "0",
        "side_type": "1", # Male/Main Hot Rank
        "type": "1"
    }
    headers = {
        "User-Agent": "Dalvik/2.1.0 (Linux; U; Android 10; Pixel 4 Build/QD1A.190821.011)",
        "Accept": "application/json"
    }
    
    try:
        print("Sending request to Fanqie API...", flush=True)
        resp = requests.get(RANKS_URL, params=params, headers=headers, timeout=10)
        print(f"Response status: {resp.status_code}", flush=True)
        resp.raise_for_status()
        data = resp.json()
        if data.get("code") != 0:
            print(f"Error fetching rank: {data.get('message')}", flush=True)
            return []
            
        print(f"Raw response data keys: {data.get('data', {}).keys()}", flush=True)
        # Uncomment to see full response if needed, but it might be large
        # print(json.dumps(data, ensure_ascii=False)[:500], flush=True)
        
        results = data.get("data", {}).get("result", [])
        if not results:
             print("No results found in data.data.result", flush=True)
             # Try to print 'data' content to see what's inside
             print(f"Data dump: {json.dumps(data['data'], ensure_ascii=False)[:300]}...", flush=True)
             
        books = []
        for item in results:
            books.append({
                "book_id": item["book_id"],
                "book_name": item["book_name"],
                "author": item["author"],
                "description": item["abstract"],
                "thumb_url": item["thumb_url"],
                "score": item["score"]
            })
        return books
    except Exception as e:
        print(f"Exception fetching rank: {e}", flush=True)
        return []

def download_cover(url, book_id):
    try:
        print(f"Downloading cover for {book_id}...", flush=True)
        resp = requests.get(url, stream=True, timeout=10)
        if resp.status_code == 200:
            path = os.path.join(DOWNLOAD_DIR, f"{book_id}_cover.jpg")
            with open(path, 'wb') as f:
                resp.raw.decode_content = True
                shutil.copyfileobj(resp.raw, f)
            return path
    except Exception as e:
        print(f"Failed to download cover: {e}", flush=True)
    return None

def download_novel_txt(book_id):
    print(f"Downloading novel {book_id} using external tool...", flush=True)
    
    os.chmod(DOWNLOADER_PATH, 0o755)
    
    cmd = [DOWNLOADER_PATH, "--download", book_id]
    
    try:
        # Using timeout to prevent hanging if tool asks for input
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
        print(result.stdout, flush=True)
        
        if result.returncode != 0:
            print(f"Downloader failed with code {result.returncode}", flush=True)
            print(result.stderr, flush=True)
            return False
            
        return True
    except subprocess.TimeoutExpired:
        print("Download timed out.", flush=True)
        return False
    except Exception as e:
        print(f"Download failed: {e}", flush=True)
        return False

def find_downloaded_txt(book_name):
    # Search logic
    search_dirs = ["data/novel", "novel", ".", "data"]
    
    # Normalize name
    clean_name = book_name.replace("/", "_")
    
    print(f"Looking for txt file for {clean_name}...", flush=True)
    
    for d in search_dirs:
        if not os.path.exists(d):
            continue
        for root, dirs, files in os.walk(d):
             for f in files:
                if f.endswith(".txt") and clean_name in f:
                    return os.path.join(root, f)
    return None

def upload_to_server(book, txt_path, cover_path):
    print(f"Uploading {book['book_name']} to server...", flush=True)
    url = f"{API_URL}/api/v1/projects/import"
    
    files = {
        'file': open(txt_path, 'rb')
    }
    
    if cover_path:
        files['cover'] = open(cover_path, 'rb')
        
    data = {
        'author': book['author'],
        'description': book['description'],
        'book_id': book['book_id'] # Pass book_id too if needed later, but server doesn't use it yet
    }
    
    headers = {
        "Authorization": f"Bearer {TOKEN}"
    }
    
    try:
        resp = requests.post(url, headers=headers, files=files, data=data, timeout=300)
        print(f"Upload response: {resp.status_code} - {resp.text}", flush=True)
        if resp.status_code == 200:
            print("Import successful!", flush=True)
        else:
            print("Import failed.", flush=True)
    except Exception as e:
        print(f"Upload exception: {e}", flush=True)
    finally:
        files['file'].close()
        if 'cover' in files:
            files['cover'].close()

def main():
    if not os.path.exists(DOWNLOAD_DIR):
        os.makedirs(DOWNLOAD_DIR)
        
    print("Starting auto importer...", flush=True)
    books = get_hot_list(1) # Limit to 1 for test
    print(f"Found {len(books)} books.", flush=True)
    
    for book in books:
        print(f"\nProcessing {book['book_name']} ({book['book_id']})...", flush=True)
        
        cover_path = download_cover(book['thumb_url'], book['book_id'])
        
        if download_novel_txt(book['book_id']):
            time.sleep(2)
            txt_path = find_downloaded_txt(book['book_name'])
            
            if txt_path:
                print(f"Found novel file: {txt_path}", flush=True)
                upload_to_server(book, txt_path, cover_path)
            else:
                print("Could not find downloaded TXT file.", flush=True)
        else:
            print("Download failed, skipping upload.", flush=True)

if __name__ == "__main__":
    main()
