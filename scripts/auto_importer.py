
import requests
import json
import os
import subprocess
import time
import shutil
import random
import datetime
import argparse

# Configuration
API_URL = "http://localhost:8080"
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYmU1NTdmMzQtM2IzMC00ODY4LTk5MjAtMzAwMDBiMmUyNzk5IiwiZW1haWwiOiJpbXBvcnRlckB0ZXN0LmNvbSIsInRpZXIiOiJmcmVlIiwiaXNzIjoieHVwdS1hcGkiLCJleHAiOjE3NzAyNzAzMTAsIm5iZiI6MTc3MDE4MzkxMCwiaWF0IjoxNzcwMTgzOTEwfQ.uNBHhI0ygMsVh5m0nYp08PDIlta4ZXyQ9G1LLXQttPM"
DOWNLOADER_PATH = "./TomatoNovelDownloader-Linux_amd64-v2.1.6"
DOWNLOAD_DIR = "downloaded_novels"
LOG_FILE = "auto_import.log"
DELAY_MIN = 5
DELAY_MAX = 15

# Fanqie Mobile API
RANKS_URL = "https://api-lf.fanqiesdk.com/api/novel/channel/homepage/rank/rank_list/v2/"

def log(msg, error=False):
    timestamp = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    formatted_msg = f"[{timestamp}] {msg}"
    print(formatted_msg, flush=True)
    with open(LOG_FILE, "a", encoding="utf-8") as f:
        f.write(formatted_msg + "\n")

def get_existing_projects():
    """Fetch list of existing projects from server to prevent duplicates."""
    url = f"{API_URL}/api/v1/projects"
    headers = {"Authorization": f"Bearer {TOKEN}"}
    try:
        resp = requests.get(url, headers=headers, timeout=10)
        if resp.status_code == 200:
            projects = resp.json()
            return {p['name'] for p in projects}
    except Exception as e:
        log(f"Warning: Could not fetch existing projects: {e}", error=True)
    return set()

def get_hot_list(offset=0, limit=20, side_type="1"):
    log(f"Fetching hot list (offset={offset}, limit={limit}, type={side_type})...")
    params = {
        "aid": "1967",
        "channel": "0",
        "device_platform": "android",
        "device_type": "0",
        "limit": str(limit),
        "offset": str(offset),
        "side_type": side_type, 
        "type": "1"
    }
    headers = {
        "User-Agent": "Dalvik/2.1.0 (Linux; U; Android 10; Pixel 4 Build/QD1A.190821.011)",
        "Accept": "application/json"
    }
    
    try:
        resp = requests.get(RANKS_URL, params=params, headers=headers, timeout=20)
        if resp.status_code != 200:
            log(f"API Error: Status {resp.status_code}", error=True)
            return []
            
        data = resp.json()
        if data.get("code") != 0:
            log(f"API Error: {data.get('message')}", error=True)
            return []
            
        results = data.get("data", {}).get("result", [])
        
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
        log(f"Exception fetching rank: {e}", error=True)
        return []

def download_cover(url, book_id):
    if not url: return None
    try:
        resp = requests.get(url, stream=True, timeout=15)
        if resp.status_code == 200:
            path = os.path.join(DOWNLOAD_DIR, f"{book_id}_cover.jpg")
            with open(path, 'wb') as f:
                resp.raw.decode_content = True
                shutil.copyfileobj(resp.raw, f)
            return path
    except Exception as e:
        log(f"Failed to download cover: {e}", error=True)
    return None

def download_novel_txt(book_id, book_name):
    txt_path = find_downloaded_txt(book_name)
    if txt_path:
        log(f"Book {book_name} already exists locally. Skipping download.")
        return True, txt_path

    log(f"Downloading novel {book_name} ({book_id})...")
    
    os.chmod(DOWNLOADER_PATH, 0o755)
    cmd = [DOWNLOADER_PATH, "--download", book_id]
    
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=120) 
        
        if result.returncode != 0:
            # Check for cooldown or specific known errors
            if "Cooldown" in result.stderr:
                log(f"Downloader Cooldown hit.", error=True)
            else:
                log(f"Downloader failed: {result.stderr[:200]}...", error=True)
            return False, None
            
        time.sleep(2)
        txt_path = find_downloaded_txt(book_name)
        return True, txt_path
    except subprocess.TimeoutExpired:
        log(f"Download timed out for {book_name}.", error=True)
        return False, None
    except Exception as e:
        log(f"Download exception: {e}", error=True)
        return False, None

def find_downloaded_txt(book_name):
    # Normalized search to handle potential filename variations
    clean_name = book_name.replace("/", "_").replace(":", "：").replace("?", "？")
    search_dirs = ["data/novel", "novel", ".", "data"]
    
    for d in search_dirs:
        if not os.path.exists(d):
            continue
        for root, dirs, files in os.walk(d):
             for f in files:
                if f.endswith(".txt"):
                    if clean_name in f or book_name in f:
                         return os.path.join(root, f)
    return None

def upload_to_server(book, txt_path, cover_path):
    log(f"Uploading {book['book_name']} to server...")
    url = f"{API_URL}/api/v1/projects/import"
    
    files = {}
    try:
        files['file'] = open(txt_path, 'rb')
        if cover_path:
            files['cover'] = open(cover_path, 'rb')
            
        data = {
            'author': book['author'],
            'description': book['description']
        }
        
        headers = {
            "Authorization": f"Bearer {TOKEN}"
        }
        
        resp = requests.post(url, headers=headers, files=files, data=data, timeout=300)
        
        if resp.status_code == 200:
            log(f"Import success: {book['book_name']}")
            return True
        else:
            log(f"Import failed: {resp.status_code} - {resp.text}", error=True)
            return False
            
    except Exception as e:
        log(f"Upload exception: {e}", error=True)
        return False
    finally:
        if 'file' in files: files['file'].close()
        if 'cover' in files: files['cover'].close()
        
def main():
    parser = argparse.ArgumentParser(description="Automated Fanqie Novel Importer")
    parser.add_argument("--offset", type=int, default=0, help="Start ranking offset (default: 0)")
    parser.add_argument("--count", type=int, default=50, help="Total books to process (default: 50)")
    parser.add_argument("--type", type=str, default="1", help="Rank type (1=Male, 2=Female, etc.)")
    args = parser.parse_args()

    if not os.path.exists(DOWNLOAD_DIR):
        os.makedirs(DOWNLOAD_DIR)
    
    # Fetch existing projects to skip duplicates
    existing_projects = get_existing_projects()
    log(f"Found {len(existing_projects)} existing projects on server.")

    current_offset = args.offset
    max_books = args.count
    processed_count = 0
    page_size = 20
    
    log(f"=== Starting Import: Offset={args.offset}, Count={args.count}, Type={args.type} ===")
    
    while processed_count < max_books:
        books = get_hot_list(offset=current_offset, limit=page_size, side_type=args.type)
        if not books:
            log("No more books found from API.")
            break
            
        log(f"Fetched {len(books)} books (Offset {current_offset})")
        
        for book in books:
            if processed_count >= max_books:
                break
            
            if book['book_name'] in existing_projects:
                log(f"Skipping {book['book_name']} (Already matches project on server).")
                processed_count += 1
                continue
                
            log(f"Processing [{processed_count+1}/{max_books}]: {book['book_name']}")
            
            cover_path = download_cover(book['thumb_url'], book['book_id'])
            success, txt_path = download_novel_txt(book['book_id'], book['book_name'])
            
            if success and txt_path:
                if upload_to_server(book, txt_path, cover_path):
                    existing_projects.add(book['book_name']) # Add to local cache
            else:
                log(f"Skipping upload for {book['book_name']} due to download failure.", error=True)
            
            processed_count += 1
            
            delay = random.uniform(DELAY_MIN, DELAY_MAX)
            log(f"Sleeping for {delay:.2f}s...")
            time.sleep(delay)
            
        current_offset += page_size
        
    log("=== Batch Import Finished ===")

if __name__ == "__main__":
    main()
