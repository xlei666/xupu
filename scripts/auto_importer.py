
import requests
import json
import os
import subprocess
import time
import shutil
import random
import datetime

# Configuration
API_URL = "http://localhost:8080"
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYmU1NTdmMzQtM2IzMC00ODY4LTk5MjAtMzAwMDBiMmUyNzk5IiwiZW1haWwiOiJpbXBvcnRlckB0ZXN0LmNvbSIsInRpZXIiOiJmcmVlIiwiaXNzIjoieHVwdS1hcGkiLCJleHAiOjE3NzAyNzAzMTAsIm5iZiI6MTc3MDE4MzkxMCwiaWF0IjoxNzcwMTgzOTEwfQ.uNBHhI0ygMsVh5m0nYp08PDIlta4ZXyQ9G1LLXQttPM"
DOWNLOADER_PATH = "./TomatoNovelDownloader-Linux_amd64-v2.1.6"
DOWNLOAD_DIR = "downloaded_novels"
LOG_FILE = "auto_import.log"
BOOKS_PER_PAGE = 20
MAX_BOOKS = 50 # Total books to fetch
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

def get_hot_list(offset=0, limit=20):
    log(f"Fetching hot list (offset={offset}, limit={limit})...")
    params = {
        "aid": "1967",
        "channel": "0",
        "device_platform": "android",
        "device_type": "0",
        "limit": str(limit),
        "offset": str(offset),
        "side_type": "1", # Male Hot Rank
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
    try:
        # log(f"Downloading cover for {book_id}...")
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
    # Check if TXT already exists to skip
    txt_path = find_downloaded_txt(book_name)
    if txt_path:
        log(f"Book {book_name} already exists at {txt_path}. Skipping download.")
        return True, txt_path

    log(f"Downloading novel {book_name} ({book_id})...")
    
    os.chmod(DOWNLOADER_PATH, 0o755)
    cmd = [DOWNLOADER_PATH, "--download", book_id]
    
    try:
        # 10 minute timeout for large novels
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=1200) 
        
        if result.returncode != 0:
            log(f"Downloader failed: {result.stderr}", error=True)
            return False, None
            
        # Try to find the file again
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
    # Normalized search
    clean_name = book_name.replace("/", "_").replace(":", "：").replace("?", "？")
    search_dirs = ["data/novel", "novel", ".", "data"]
    
    for d in search_dirs:
        if not os.path.exists(d):
            continue
        for root, dirs, files in os.walk(d):
             for f in files:
                if f.endswith(".txt"):
                    # Check if filename contains the book name appropriately
                    # Sometimes downloader uses different punctation
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
            'description': book['description'],
            # 'project_name': book['book_name'] # Server derives from filename for now
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
    if not os.path.exists(DOWNLOAD_DIR):
        os.makedirs(DOWNLOAD_DIR)
        
    offset = 0
    processed_count = 0
    
    log("=== Starting Batch Import Task ===")
    
    while processed_count < MAX_BOOKS:
        books = get_hot_list(offset=offset, limit=BOOKS_PER_PAGE)
        if not books:
            log("No more books found from API.")
            break
            
        log(f"Fetched {len(books)} books from rank (Page offset {offset})")
        
        for book in books:
            if processed_count >= MAX_BOOKS:
                break
                
            log(f"Processing [{processed_count+1}/{MAX_BOOKS}]: {book['book_name']}")
            
            cover_path = download_cover(book['thumb_url'], book['book_id'])
            
            success, txt_path = download_novel_txt(book['book_id'], book['book_name'])
            
            if success and txt_path:
                upload_to_server(book, txt_path, cover_path)
            else:
                log(f"Skipping upload for {book['book_name']} due to download failure.", error=True)
            
            processed_count += 1
            
            # Anti-scraping delay
            delay = random.uniform(DELAY_MIN, DELAY_MAX)
            log(f"Sleeping for {delay:.2f}s...")
            time.sleep(delay)
            
        offset += BOOKS_PER_PAGE
        
    log("=== Batch Import Finished ===")

if __name__ == "__main__":
    main()
