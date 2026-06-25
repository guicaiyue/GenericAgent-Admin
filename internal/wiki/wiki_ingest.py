#!/usr/bin/env python3
"""
wiki_ingest.py — llm-wiki ingest entrypoint for GenericAgent.

Reflect protocol entry: INTERVAL/ONCE/init/check/on_done/on_error.
Can also run as standalone: python3 reflect/wiki_ingest.py

Receives WIKI_DIR, RAW_DIR, GA_ROOT, MODEL_NO, INGEST_STATE env vars.
Maps raw/ source files into cross-linked wiki pages under wiki/.
"""
import json
import logging
import os
import sys
from datetime import datetime

# Reflect protocol: poll interval for scheduler
INTERVAL = 315360000   # 10 years = effectively one-shot via scheduler
ONCE = True            # only trigger once per deploy

GA_ROOT = os.environ.get("GA_ROOT", "/vol1/1000/开发/NAS/GenericAgent")
sys.path.insert(0, os.path.join(GA_ROOT, "memory"))

WIKI_DIR = ""
RAW_DIR = ""
MODEL_NO = 0
INGEST_STATE = ""

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("wiki_ingest")


def init(args):
    """Reflect protocol: receive CLI args from agentmain --reflect."""
    global WIKI_DIR, RAW_DIR, MODEL_NO, INGEST_STATE
    WIKI_DIR = args.get("wiki_dir", os.environ.get("WIKI_DIR", "")).strip()
    RAW_DIR = args.get("raw_dir", os.environ.get("RAW_DIR", "")).strip()
    raw = args.get("model_no", os.environ.get("MODEL_NO", "0"))
    try:
        MODEL_NO = int(raw) if raw else 0
    except ValueError:
        MODEL_NO = 0
    INGEST_STATE = args.get("ingest_state", os.environ.get("INGEST_STATE", "")).strip()
    log.info("init: WIKI_DIR=%s RAW_DIR=%s MODEL_NO=%s", WIKI_DIR, RAW_DIR, MODEL_NO)


def write_log(wiki_dir, op, desc):
    """Append ingest log to wikiDir/log/"""
    log_dir = os.path.join(wiki_dir, "log")
    os.makedirs(log_dir, exist_ok=True)
    today = datetime.now().strftime("%Y-%m-%d")
    now = datetime.now().strftime("%H:%M")
    log_file = os.path.join(log_dir, f"{today}.md")
    entry = f"## [{now}] {op} | {desc}\n"
    with open(log_file, "a", encoding="utf-8") as f:
        f.write(entry)


def check():
    """Reflect protocol: return prompt string or None."""
    d = WIKI_DIR or os.environ.get("WIKI_DIR", "")
    if not d:
        return None
    raw_dir = RAW_DIR or os.path.join(d, "raw")
    if not os.path.isdir(raw_dir):
        log.warning("raw directory not found: %s", raw_dir)
        return None
    # Collect source files
    files = []
    for root, dirs, filenames in os.walk(raw_dir):
        for fn in filenames:
            if fn.endswith(".md"):
                files.append(os.path.join(root, fn))
    if not files:
        log.info("no .md files in raw/, skipping ingest")
        return None
    return (
        f"请对 {raw_dir} 执行 llm-wiki ingest 操作。\n"
        f"读取该目录下的所有源文件，使用 llm_wiki_skill SOP "
        f"将其编译为交叉链接的 wiki 知识页面，写入 {d}/wiki/。\n"
        f"必须生成 {d}/index.md 作为入口页。"
    )


def _write_state(status, **extra):
    if not INGEST_STATE:
        return
    state = {"status": status, "ts": datetime.now().isoformat()}
    state.update(extra)
    with open(INGEST_STATE, "w", encoding="utf-8") as f:
        json.dump(state, f, indent=2, ensure_ascii=False)


def on_done(result):
    """Reflect protocol: GA completed callback."""
    _write_state("done", result_summary=str(result)[:500])
    log.info("ingest done: %s", result)


def on_error(err):
    """Reflect protocol: GA error callback."""
    _write_state("error", error=str(err))
    log.error("ingest error: %s", err)


def ensure_dirs(wiki_dir):
    os.makedirs(os.path.join(wiki_dir, "raw"), exist_ok=True)
    os.makedirs(os.path.join(wiki_dir, "wiki"), exist_ok=True)
    os.makedirs(os.path.join(wiki_dir, "log"), exist_ok=True)


def main():
    """Standalone entry (called by ingest.go or direct)."""
    d = WIKI_DIR or os.environ.get("WIKI_DIR", "")
    if not d:
        log.error("WIKI_DIR not set")
        sys.exit(1)
    ensure_dirs(d)
    prompt = check()
    if not prompt:
        log.info("nothing to ingest")
        return
    log.info("prompt ready (%d chars)", len(prompt))
    _write_state("running", prompt=prompt)
    # Write prompt to file for external agent
    prompt_file = os.path.join(d, ".ingest_prompt.txt")
    with open(prompt_file, "w", encoding="utf-8") as f:
        f.write(prompt)
    log.info("prompt written to %s", prompt_file)
    write_log(d, "ingest", "启动 ingest，raw 文件数待确认")


if __name__ == "__main__":
    main()
