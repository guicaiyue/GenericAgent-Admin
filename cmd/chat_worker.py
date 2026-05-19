import json, os, sys, time, traceback, threading, queue
from pathlib import Path


def _force_utf8_stdio():
    # Windows pipes otherwise may inherit the active ANSI code page and corrupt CJK text.
    for stream in (sys.stdin, sys.stdout, sys.stderr):
        try:
            stream.reconfigure(encoding='utf-8', errors='replace')
        except Exception:
            pass


_force_utf8_stdio()


def emit(ev):
    print(json.dumps(ev, ensure_ascii=False), flush=True)


def main():
    req = json.load(sys.stdin)
    root = Path(req.get('ga_root') or os.environ.get('GA_ROOT') or '.').resolve()
    if str(root) not in sys.path:
        sys.path.insert(0, str(root))
    os.chdir(root)
    from agentmain import GeneraticAgent

    prompt = req.get('prompt') or ''
    llm_no = int(req.get('llm_no') or 0)
    agent = GeneraticAgent()
    try:
        agent.next_llm(llm_no)
    except Exception:
        pass
    # GA core supports incremental display through inc_out + put_task display queue.
    agent.verbose = True
    agent.inc_out = True

    try:
        worker = threading.Thread(target=agent.run, name='ga-admin-chat-worker', daemon=True)
        worker.start()
        display_queue = agent.put_task(prompt, source='admin_chat')
        chunks = []
        while True:
            try:
                item = display_queue.get(timeout=1.0)
            except queue.Empty:
                if not worker.is_alive():
                    raise RuntimeError('GA core worker exited unexpectedly')
                continue
            if 'next' in item:
                delta = str(item.get('next') or '')
                if delta:
                    chunks.append(delta)
                    emit({'type': 'delta', 'delta': delta})
            if 'done' in item:
                text = str(item.get('done') or ''.join(chunks))
                msg = {'id': new_id(), 'role': 'assistant', 'content': text, 'created_at': int(time.time())}
                emit({'type': 'done', 'message': msg})
                return
    except Exception as e:
        try:
            agent.abort()
        except Exception:
            pass
        msg = {'id': new_id(), 'role': 'assistant', 'content': '执行失败：%s\n%s' % (e, traceback.format_exc()), 'created_at': int(time.time()), 'error': True}
        emit({'type': 'error', 'message': msg})


def new_id():
    import uuid
    return str(uuid.uuid4())

if __name__ == '__main__':
    main()
