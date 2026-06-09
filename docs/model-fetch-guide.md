# CodexPlusPlus 模型获取方法

## 总览

CodexPlusPlus 获取模型列表有 **2 个入口**、**3 种来源**。

---

## 入口

| 入口 | 触发方 | 调用函数 |
|------|--------|---------|
| **A. Bridge** | Codex 应用启动，渲染进程请求 `/codex-model-catalog` | `read_codex_model_catalog()` |
| **B. Manager UI** | 用户点击"从上游获取"按钮 | `fetch_relay_profile_model_ids(profile)` |

---

## 来源 ①：Relay Profile 静态列表

**入口**：A（优先级最高，命中后直接返回，不走后续来源）

**触发条件**：`~/.config/Codex++/settings.json` 中存在激活的 Relay Profile。

**数据**：从 profile 的 `model` + `model_list` 字段读取，**不发 HTTP**。

```rust
// model_catalog.rs:37-43
pub async fn read_codex_model_catalog() -> Value {
    if settings_path.exists() {
        if let Ok(settings) = SettingsStore::new(settings_path).load() {
            return relay_profile_model_catalog_value(&home, &settings.active_relay_profile());
        }
    }
    // 无活跃 profile → 走来源 ②③ ...
}
```

模型列表组装：

```rust
fn relay_profile_model_ids(profile: &RelayProfile) -> Vec<String> {
    unique_strings(
        std::iter::once(profile.model.as_str())                    // 当前选中模型
            .chain(profile.model_list.split(['\r', '\n', ',']))    // model_list 按换行/逗号分割
            .map(str::trim)
            .filter(|v| !v.is_empty())
            ...
    )
}
```

**说明**：`model_list` 中的模型 ID 来源于用户之前通过入口 B 从上游获取并保存的结果。中继模式下这是正常路径——用户只需获取一次，之后 Codex 每次启动都直接读取。

---

## 来源 ②：环境变量 → HTTP 获取

**入口**：A（来源 ① 未命中时的回退路径）

**触发条件**：系统环境变量中有 API 端点地址。

**Base URL 环境变量**（按优先级）：

| 优先级 | 变量名 |
|--------|--------|
| 1 | `CODEX_PLUS_OPENAI_BASE_URL` |
| 2 | `CODEX_PLUS_BASE_URL` |
| 3 | `OPENAI_BASE_URL` |
| 4 | `OPENAI_API_BASE_URL` |
| 5 | `OPENAI_API_BASE` |
| 6 | `OPENAI_API_URL` |

**API Key 环境变量**（按优先级）：

| 优先级 | 变量名 |
|--------|--------|
| 1 | `CODEX_PLUS_OPENAI_API_KEY` |
| 2 | `CODEX_PLUS_API_KEY` |
| 3 | `OPENAI_API_KEY` |

API Key 为空时回退读取 `~/.codex/auth.json`。

```rust
fn model_sources_from_environment(env, auth_api_key) -> Vec<ModelSource> {
    let base_url = first_env_value(env, BASE_URL_ENV_KEYS);
    if base_url.is_empty() { return Vec::new(); }
    let api_key = first_env_value(env, API_KEY_ENV_KEYS);
    vec![ModelSource { source_id: "env:openai-compatible", base_url, api_key, ... }]
}
```

发现 source 后调用 `fetch_models_from_source()` 发 HTTP GET。

---

## 来源 ③：model_catalog_json 本地文件

**入口**：A（与来源 ② 并列，不发 HTTP）

**触发条件**：config.toml 中有 `model_catalog_json = "path/to/file.json"`。

CodexPlusPlus 自带一份嵌入二进制的 `codex-models.json`，包含 OpenAI 官方完整模型目录。读取后过滤：

```rust
fn catalog_model_visible_in_api(model: &Value) -> bool {
    // supported_in_api == true（默认 true）
    // visibility == "list"（默认 "list"，不区分大小写）
}
// 提取 slug 字段作为模型 ID
```

---

## 入口 B：从上游获取 — `fetch_relay_profile_model_ids(profile)`

**这是用户在 Manager UI 中输入 baseUrl + API Key 后点击"从上游获取"的唯一路径。**

```rust
pub async fn fetch_relay_profile_model_ids(profile: &RelayProfile) -> Result<(Vec<String>, String)> {
    let source = ModelSource {
        base_url: profile.upstream_base_url,   // 优先用 upstream_base_url
                                             // 为空则回退到 profile.base_url
        api_key: profile.api_key,
        ...
    };
    let endpoint = models_endpoint(&source.base_url);
    let client = crate::http_client::proxied_client(&profile.user_agent)?;
    let (models, status) = fetch_models_from_source(&client, &source).await;
    if models.is_empty() → bail!(错误信息)
    Ok((models, endpoint))
}
```

**完整数据流**：

```
用户在 Manager UI 填写 upstream_base_url + api_key
  → 点击"从上游获取"
    → fetch_relay_profile_model_ids(profile)
      → ModelSource { base_url: upstream_base_url, api_key }
        → models_endpoint(base_url) → 拼接 /v1/models
          → HTTP GET {endpoint}
            → parse_model_payload() 解析响应
              → 返回模型 ID 列表
                → 前端展示，用户选择后保存到 profile.model_list
```

---

## HTTP 请求函数 — `fetch_models_from_source()`

入口 B 和入口 A 的来源 ② 共用此函数。

### 端点构造 — `models_endpoint()`

```rust
fn models_endpoint(base_url: &str) -> String {
    let cleaned = base_url.trim_end_matches('/');
    if cleaned.ends_with("/models")  → 原样返回
    if cleaned.ends_with("/v1")      → 拼接 "/models"
    否则                              → 拼接 "/v1/models"
}
```

| 输入 | 输出 |
|------|------|
| `https://api.xiaomimimo.com/v1` | `https://api.xiaomimimo.com/v1/models` |
| `https://api.example.com` | `https://api.example.com/v1/models` |
| `https://api.example.com/v1/models` | `https://api.example.com/v1/models` |
| `https://api.example.com/anthropic` | `https://api.example.com/anthropic/v1/models`（可能 404） |

### HTTP 请求

```rust
client
    .get(endpoint)
    .header("Accept", "application/json")
    .bearer_auth(api_key)    // api_key 非空时
    .send()
    .await
```

- User-Agent：入口 A 为 `CodexPlusPlus/1.0`，入口 B 使用 profile 自定义 `user_agent`
- 支持系统代理（`HTTP_PROXY` / `HTTPS_PROXY`）
- 无硬编码超时

### 响应解析 — `parse_model_payload()`

支持 6 种格式，递归解析：

| 格式 | 示例 |
|------|------|
| 字符串数组 | `["mimo-v2.5-pro", "mimo-v2-flash"]` |
| 对象数组 | `[{"id": "mimo-v2.5-pro"}]` |
| `data` 嵌套 | `{"data": [{"id": "mimo-v2.5-pro"}]}` |
| `models` 嵌套 | `{"models": [{"id": "mimo-v2.5-pro"}]}` |
| `items` 嵌套 | `{"items": [{"id": "mimo-v2.5-pro"}]}` |
| 单个对象 | `{"id": "mimo-v2.5-pro"}` |

对象中提取 model ID 的键名优先级：`id` → `model` → `name`

---

## 完整流程图

```
┌─────────────────────────────────────────────────────────────┐
│  入口 A：Bridge（Codex 启动时自动触发）                        │
│  read_codex_model_catalog()                                 │
│                                                             │
│  ① 有活跃 Relay Profile?                                    │
│     └─ YES → 读 profile.model + model_list → 直接返回       │
│                                                             │
│  ② NO → read_codex_model_catalog_from_home()                │
│     │                                                       │
│     ├─ 环境变量有 URL? → HTTP GET /v1/models                │
│     ├─ config.toml 有 model_catalog_json? → 读本地文件       │
│     │                                                       │
│     └─ 所有来源合并去重 → 返回                                │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  入口 B：Manager UI "从上游获取" 按钮                          │
│  fetch_relay_profile_model_ids(profile)                     │
│                                                             │
│  读取 profile.upstream_base_url + api_key                   │
│  → HTTP GET /v1/models                                      │
│  → 返回模型列表 → 用户选择后保存到 profile.model_list          │
└─────────────────────────────────────────────────────────────┘
```

---

## 来源速查表

| | 来源 ① | 来源 ② | 来源 ③ |
|--|--------|--------|--------|
| **数据来自** | profile.model_list | 环境变量 | 本地 JSON 文件 |
| **发 HTTP** | ❌ | ✅ | ❌ |
| **入口** | A（优先级最高） | A（回退） | A（回退） |
| **API Key 来源** | — | 环境变量 → auth.json | — |
| **典型场景** | 中继模式正常路径 | 设了 OPENAI_BASE_URL | OpenAI 官方目录 |

---

## 关键文件

```
crates/codex-plus-core/
├── assets/codex-models.json   # OpenAI 静态模型目录（来源 ③）
├── src/
│   ├── model_catalog.rs       # 核心：所有获取逻辑（784 行）
│   ├── settings.rs            # RelayProfile 定义 + SettingsStore
│   ├── relay_config.rs        # 写入 ~/.codex/config.toml（中继配置）
│   ├── bridge.rs              # CDP Bridge → /codex-model-catalog
│   └── http_client.rs         # HTTP 客户端（支持代理）
└── tests/model_catalog.rs     # 集成测试（384 行）

apps/codex-plus-manager/src-tauri/src/
└── commands.rs                # fetch_relay_profile_models Tauri 命令
```
