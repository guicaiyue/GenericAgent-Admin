# GenericAgent Admin Go

Go 后端 + Vite React 前端的 GenericAgent 管理端，目标是方便打包成单个 `ga-admin.exe` 分发。

## 功能

- Web 页面管理 GA 服务。
- 自动发现 `reflect/*.py`，按 `python agentmain.py --reflect reflect/<file>.py` 启动。
- 自动发现 `frontends/*app*.py`，其中 `stapp` 使用 `python -m streamlit run`。
- 查看服务 stdout/stderr 日志。
- 页面配置 GA 根目录。
- 页面配置模型，保存独立草稿 `model_profiles.json`。
- 导出 `mykey_admin.generated.py`。
- 安全策略：不读取、不覆盖现有 `mykey.py`。
- Vite 构建产物通过 Go `embed` 打进 exe。

## 构建

```bat
cd /d C:\Users\Fwind43\Desktop\code\GenericAgent-Admin-Go
build.bat
```

产物：

```text
dist\ga-admin.exe
```

## 开发运行

```bat
npm.cmd --prefix web install
npm.cmd --prefix web run build
go run .
```

打开：

```text
http://127.0.0.1:8787
```

## 配置

首次运行会读取默认配置，也可以创建 `config.local.json`：

```json
{
  "ga_root": "E:/Work/GenericAgent",
  "host": "127.0.0.1",
  "port": 8787,
  "log_tail_lines": 200,
  "buffer_lines": 1000
}
```

Go 管理端本身可以单 exe 分发；GA 服务仍依赖 GA 目录中的 Python/venv 环境。
