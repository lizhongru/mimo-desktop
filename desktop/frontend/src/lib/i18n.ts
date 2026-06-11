import { useSettingsStore } from "../stores/settingsStore";

type TranslationKey =
  | "settings"
  | "tools_mcp"
  | "tools"
  | "language"
  | "theme"
  | "font_size"
  | "models"
  | "help_feedback"
  | "dark"
  | "light"
  | "chinese"
  | "english"
  | "current_model"
  | "add_model"
  | "model_name"
  | "api_base"
  | "api_key"
  | "model_id"
  | "max_tokens"
  | "temperature"
  | "save"
  | "cancel"
  | "delete"
  | "close"
  | "new_chat"
  | "sessions"
  | "no_sessions"
  | "delete_session"
  | "delete_confirm"
  | "project_dir"
  | "status_ready"
  | "status_thinking"
  | "status_responding"
  | "status_executing"
  | "help_title"
  | "help_shortcuts"
  | "help_docs"
  | "help_report"
  | "general"
  | "safety_level"
  | "planning_mode"
  | "confirm"
  | "deny"
  | "approve_all"
  | "safety_confirm_title"
  | "tools_count"
  | "mcp_servers"
  | "connected"
  | "disconnected"
  | "search_tools"
  | "builtin_tools"
  | "mcp_tools"
  | "about"
  | "version"
  | "activity"
  | "file_changes"
  | "plan_progress"
  // New keys for full i18n coverage
  | "app_subtitle"
  | "app_name"
  | "export_chat"
  | "other_projects"
  | "version_label"
  | "empty_hint"
  | "thinking_label"
  | "input_placeholder"
  | "cancel_esc"
  | "send_enter"
  | "tokens"
  | "running"
  | "done"
  | "arguments"
  | "result"
  | "approve"
  | "unknown"
  | "minimize"
  | "maximize"
  | "close_tooltip"
  | "toggle_left_sidebar"
  | "toggle_right_sidebar"
  | "no_activity"
  | "all_mcp_tools"
  | "compressing_context"
  | "compress_done"
  | "default_badge"
  | "set_default"
  | "links"
  | "send_message"
  | "new_line"
  | "cancel_close"
  | "toggle_left"
  | "toggle_right"
  | "app_description"
  | "tools_suffix"
  | "servers_suffix"
  | "mcp_badge"
  | "status_executing_tool"
  | "error_prefix"
  | "other_workspaces"
  | "manage"
  | "select_all"
  | "deselect_all"
  | "delete_selected"
  | "selected_count"
  | "pin"
  | "unpin"
  | "open_in_explorer"
  | "rename"
  | "remove_project"
  | "export_success"
  | "export_failed"
  | "export_empty"
  | "safety_low"
  | "safety_medium"
  | "safety_high"
  | "safety_critical"
  | "plan_auto"
  | "plan_react"
  | "plan_execute"
  | "perm_readonly"
  | "perm_write"
  | "perm_exec"
  | "permission_label"
  | "attach_file"
  | "attach_folder"
  | "drop_to_add"
  | "reasoning_low"
  | "reasoning_medium"
  | "reasoning_high"
  | "reasoning_label"
  | "safety_lockdown"
  | "safety_confirm"
  | "safety_auto"
  | "help_log"
  | "shortcuts"
  | "language_select"
  | "theme_select"
  | "user_profile"
  | "delete_workspace"
  | "delete_workspace_confirm"
  | "welcome_subtitle"
  | "welcome_input_placeholder"
  | "welcome_hint"
  | "no_project"
  | "select_workspace"
  | "recent_workspaces"
  | "no_recent_workspaces"
| "edit_model"
| "provider"
| "website"
| "api_settings"
| "model_settings"
| "default_model_id"
| "fallback_model_id"
| "available_models"
| "add_model_id"
| "generation_params"
| "supported_features"
| "streaming"
| "vision"
| "function_calling"
| "quick_setup"
| "models_available"
| "streaming_short"
| "tools_short"
| "visit_website"
| "edit"
| "fetch_models"
| "available_from_api"
| "selected_models"
| "no_models_selected"
| "show_api_key"
| "hide_api_key"
| "api_required_for_models"
| "no_model_configured"
| "fetch_models_failed"
| "vision_short"
| "model_management"
| "models_configured"
| "provider_name"
| "api_base_url"
| "search_models"
| "no_models_found"
| "back"
| "conversations"
| "click_to_settings"
| "fetching"
| "selected"
| "click_to_set_default"
| "advanced_settings"
| "fallback_model"
| "optional"
| "active"
| "no_models_configured"
| "click_add_model_to_start"
| "remove"
| "update"
| "fallback_tip"
| "max_tokens_tip"
| "temperature_tip"
| "top_p_tip"
| "features_tip"
| "streaming_tip"
| "vision_tip"
| "tools_tip"
  | "browse_folder"
  | "copy_text"
  | "regenerate"
  | "delete_message"
  | "shortcut_send"
  | "shortcut_new_line"
  | "shortcut_new_chat"
  | "shortcut_toggle_left"
  | "shortcut_toggle_right"
  | "shortcut_compress"
  | "shortcut_cancel"
  | "shortcut_escape"
  | "help_log_title"
  | "about_title"
  | "about_description"
  | "about_github"
  | "about_version_info";

const translations: Record<"zh" | "en", Record<TranslationKey, string>> = {
  zh: {
    settings: "设置",
    tools_mcp: "工具 & MCP",
    tools: "工具",
    language: "语言",
    theme: "主题",
    font_size: "字体大小",
    models: "模型管理",
    help_feedback: "帮助与反馈",
    dark: "深色",
    light: "浅色",
    chinese: "中文",
    english: "English",
    current_model: "当前模型",
    add_model: "添加模型",
    model_name: "模型名称",
    api_base: "API 地址",
    api_key: "API 密钥",
    model_id: "模型 ID",
    max_tokens: "最大 Tokens",
    temperature: "Temperature",
    save: "保存",
    cancel: "取消",
    delete: "删除",
    close: "关闭",
    new_chat: "新建对话",
    sessions: "对话历史",
    no_sessions: "暂无对话",
    delete_session: "删除对话",
    delete_confirm: "确定要删除这个对话吗？此操作不可撤销。",
    project_dir: "项目目录",
    status_ready: "就绪",
    status_thinking: "思考中...",
    status_responding: "响应中...",
    status_executing: "执行中...",
    help_title: "帮助与反馈",
    help_shortcuts: "快捷键",
    help_docs: "文档",
    help_report: "反馈问题",
    general: "通用",
    safety_level: "安全级别",
    planning_mode: "规划模式",
    confirm: "确认",
    deny: "拒绝",
    approve_all: "全部允许",
    safety_confirm_title: "安全确认",
    tools_count: "工具数量",
    mcp_servers: "MCP 服务器",
    connected: "已连接",
    disconnected: "未连接",
    search_tools: "搜索工具...",
    builtin_tools: "内置工具",
    mcp_tools: "MCP 工具",
    about: "关于",
    version: "版本",
    activity: "活动",
    file_changes: "文件变更",
    plan_progress: "计划进度",
    // New keys
    app_subtitle: "AI 驱动的编程助手",
    app_name: "MiMo",
    export_chat: "导出对话",
    other_projects: "其他项目",
    version_label: "0.3.0",
    empty_hint: "输入消息开始对话",
    thinking_label: "思考中",
    input_placeholder: "输入消息... (Shift+Enter 换行)",
    cancel_esc: "取消 (Esc)",
    send_enter: "发送 (Enter)",
    tokens: "Token:",
    running: "运行中",
    done: "完成",
    arguments: "参数",
    result: "结果",
    approve: "允许",
    unknown: "未知",
    minimize: "最小化",
    maximize: "最大化",
    close_tooltip: "关闭",
    toggle_left_sidebar: "切换左侧栏",
    toggle_right_sidebar: "切换右侧栏",
    no_activity: "暂无活动",
    all_mcp_tools: "所有 MCP 工具",
    compressing_context: "正在压缩上下文...",
    compress_done: "上下文压缩完成",
    default_badge: "默认",
    set_default: "设为默认",
    links: "链接",
    send_message: "发送消息",
    new_line: "换行",
    cancel_close: "取消 / 关闭",
    toggle_left: "切换左侧栏",
    toggle_right: "切换右侧栏",
    app_description: "AI 驱动的编程助手，基于 MiMo 大模型",
    tools_suffix: "个工具",
    servers_suffix: "个服务器",
    mcp_badge: "MCP",
    status_executing_tool: "执行",
    error_prefix: "错误",
    other_workspaces: "其他工作区",
    manage: "管理",
    select_all: "全选",
    deselect_all: "取消全选",
    delete_selected: "删除所选",
    selected_count: "已选",
    pin: "置顶项目",
    unpin: "取消置顶",
    open_in_explorer: "在资源管理器中打开",
    rename: "重命名",
    remove_project: "移除项目",
    export_success: "导出完成",
    export_failed: "导出失败",
    export_empty: "该对话没有消息可导出",
    plan_auto: "自动",
    plan_react: "React",
    plan_execute: "规划执行",
    perm_readonly: "只读",
    perm_write: "读写",
    perm_exec: "完全访问",
    permission_label: "MiMo ????",
    attach_file: "添加文件",
    attach_folder: "添加文件夹",
    drop_to_add: "松手即可添加",
    reasoning_low: "低推理",
    reasoning_medium: "中推理",
    reasoning_high: "高推理",
    reasoning_label: "推理",
    safety_lockdown: "锁定",
    safety_confirm: "确认",
    safety_auto: "自动",
    help_log: "帮助日志",
    shortcuts: "快捷键",
    language_select: "语言",
    theme_select: "主题",
    user_profile: "个人信息",
    delete_workspace: "移除项目",
    delete_workspace_confirm: "确定要移除该项目及其所有对话吗？此操作不可撤销。",
    welcome_subtitle: "AI 驱动的编程助手",
    welcome_input_placeholder: "描述你想做的事情...",
    welcome_hint: "按 Enter 发送，Shift+Enter 换行",
    no_project: "不关联项目",
    select_workspace: "选择工作区",
    recent_workspaces: "最近使用",
    no_recent_workspaces: "暂无最近工作区",
    browse_folder: "浏览文件夹",
    edit_model: "编辑模型",
    provider: "供应商",
    website: "官网",
    api_settings: "API 设置",
    model_settings: "模型设置",
    default_model_id: "默认模型 ID",
    fallback_model_id: "兜底模型 ID",
    available_models: "可用模型列表",
    add_model_id: "添加模型 ID...",
    generation_params: "生成参数",
    supported_features: "支持的功能",
    streaming: "流式输出",
    vision: "视觉理解",
    function_calling: "函数调用",
    quick_setup: "快速设置",
    models_available: "个模型可用",
    streaming_short: "流式",
    tools_short: "工具",
    visit_website: "访问官网",
    edit: "编辑",
    fetch_models: "获取模型列表",
    available_from_api: "API 可用模型",
    selected_models: "已选模型",
    no_models_selected: "未选择模型",
    show_api_key: "显示密钥",
    hide_api_key: "隐藏密钥",
    api_required_for_models: "请先填写 API 地址和密钥",
    no_model_configured: "请先配置一个模型",
    fetch_models_failed: "获取失败，请检查 API 地址和密钥",
    vision_short: "视觉",
    model_management: "模型管理",
    models_configured: "个模型已配置",
    provider_name: "供应商名称",
    api_base_url: "API 地址",
    search_models: "搜索模型...",
    no_models_found: "未找到模型",
    back: "返回",
    conversations: "对话",
    click_to_settings: "点击打开设置",
    fetching: "获取中...",
    selected: "已选择",
    click_to_set_default: "点击设为默认模型",
    advanced_settings: "高级设置",
    fallback_model: "兜底模型",
    optional: "可选",
    active: "当前使用",
    no_models_configured: "暂无配置的模型",
    click_add_model_to_start: "点击上方按钮添加模型",
    remove: "移除",
    update: "更新",
    fallback_tip: "主模型不可用时自动切换",
    max_tokens_tip: "单次回复最大长度",
    temperature_tip: "越高越随机(0-2)，精确任务建议0.3",
    top_p_tip: "核采样概率(0-1)，建议0.9",
    features_tip: "勾选模型实际支持的功能",
    streaming_tip: "逐字显示回复，体验更好",
    vision_tip: "支持图片输入分析",
    tools_tip: "函数调用，Agent模式必需",
    safety_low: "低",
    safety_medium: "中",
    safety_high: "高",
    safety_critical: "严重",
    copy_text: "复制文本",
    regenerate: "重新生成",
    delete_message: "删除消息",
    shortcut_send: "发送消息",
    shortcut_new_line: "换行",
    shortcut_new_chat: "新建对话",
    shortcut_toggle_left: "切换左侧栏",
    shortcut_toggle_right: "切换右侧栏",
    shortcut_compress: "压缩上下文",
    shortcut_cancel: "取消操作",
    shortcut_escape: "关闭弹窗 / 拒绝确认",
    help_log_title: "帮助日志",
    about_title: "关于 MiMo Desktop",
    about_description: "基于 Wails 的 AI 聊天桌面客户端，由 MiMo 大模型驱动。",
    about_github: "GitHub 仓库",
    about_version_info: "版本信息",
  },
  en: {
    settings: "Settings",
    tools_mcp: "Tools & MCP",
    tools: "Tools",
    language: "Language",
    theme: "Theme",
    font_size: "Font Size",
    models: "Model Management",
    help_feedback: "Help & Feedback",
    dark: "Dark",
    light: "Light",
    chinese: "中文",
    english: "English",
    current_model: "Current Model",
    add_model: "Add Model",
    model_name: "Model Name",
    api_base: "API Base",
    api_key: "API Key",
    model_id: "Model ID",
    max_tokens: "Max Tokens",
    temperature: "Temperature",
    save: "Save",
    cancel: "Cancel",
    delete: "Delete",
    close: "Close",
    new_chat: "New Chat",
    sessions: "Sessions",
    no_sessions: "No sessions yet",
    delete_session: "Delete Session",
    delete_confirm: "Are you sure you want to delete this session? This action cannot be undone.",
    project_dir: "Project Directory",
    status_ready: "ready",
    status_thinking: "thinking...",
    status_responding: "responding...",
    status_executing: "executing...",
    help_title: "Help & Feedback",
    help_shortcuts: "Keyboard Shortcuts",
    help_docs: "Documentation",
    help_report: "Report Issue",
    general: "General",
    safety_level: "Safety Level",
    planning_mode: "Planning Mode",
    confirm: "Confirm",
    deny: "Deny",
    approve_all: "Approve All",
    safety_confirm_title: "Safety Confirmation",
    tools_count: "Tools Count",
    mcp_servers: "MCP Servers",
    connected: "connected",
    disconnected: "disconnected",
    search_tools: "Search tools...",
    builtin_tools: "Built-in Tools",
    mcp_tools: "MCP Tools",
    about: "About",
    version: "Version",
    activity: "Activity",
    file_changes: "File Changes",
    plan_progress: "Plan Progress",
    // New keys
    app_subtitle: "AI-powered coding assistant",
    app_name: "MiMo",
    export_chat: "Export chat",
    other_projects: "Other projects",
    version_label: "0.3.0",
    empty_hint: "Type a message to get started",
    thinking_label: "Thinking",
    input_placeholder: "Type a message... (Shift+Enter for new line)",
    cancel_esc: "Cancel (Esc)",
    send_enter: "Send (Enter)",
    tokens: "tokens:",
    running: "running",
    done: "done",
    arguments: "Arguments",
    result: "Result",
    approve: "Approve",
    unknown: "UNKNOWN",
    minimize: "Minimize",
    maximize: "Maximize",
    close_tooltip: "Close",
    toggle_left_sidebar: "Toggle sidebar (Ctrl+B)",
    toggle_right_sidebar: "Toggle activity panel (Ctrl+I)",
    no_activity: "No activity yet",
    all_mcp_tools: "All MCP Tools",
    compressing_context: "Compressing context...",
    compress_done: "Context compressed",
    default_badge: "default",
    set_default: "set default",
    links: "Links",
    send_message: "Send message",
    new_line: "New line",
    cancel_close: "Cancel / Close",
    toggle_left: "Toggle left sidebar",
    toggle_right: "Toggle right sidebar",
    app_description: "AI-powered coding assistant powered by MiMo LLM",
    tools_suffix: "tools",
    servers_suffix: "servers",
    mcp_badge: "MCP",
    status_executing_tool: "executing",
    error_prefix: "Error",
    other_workspaces: "Other workspaces",
    manage: "Manage",
    select_all: "Select all",
    deselect_all: "Deselect",
    delete_selected: "Delete selected",
    selected_count: "Selected",
    pin: "Pin project",
    unpin: "Unpin project",
    open_in_explorer: "Open in explorer",
    rename: "Rename",
    remove_project: "Remove project",
    export_success: "Export complete",
    export_failed: "Export failed",
    export_empty: "No messages to export",
    plan_auto: "Auto",
    plan_react: "React",
    plan_execute: "Plan & Execute",
    perm_readonly: "Read Only",
    perm_write: "Read & Write",
    perm_exec: "Full Access",
    permission_label: "MiMo Permission",
    attach_file: "Attach file",
    attach_folder: "Attach folder",
    drop_to_add: "松手即可添加",
    reasoning_low: "Low",
    reasoning_medium: "Medium",
    reasoning_high: "High",
    reasoning_label: "Reasoning",
    safety_lockdown: "Lockdown",
    safety_confirm: "Confirm",
    safety_auto: "Auto",
    help_log: "Help & Logs",
    shortcuts: "Shortcuts",
    language_select: "Language",
    theme_select: "Theme",
    user_profile: "Profile",
    delete_workspace: "Remove Project",
    delete_workspace_confirm: "Are you sure you want to remove this project and all its sessions? This action cannot be undone.",
    welcome_subtitle: "AI-powered coding assistant",
    welcome_input_placeholder: "Describe what you want to do...",
    welcome_hint: "Press Enter to send, Shift+Enter for new line",
    no_project: "No project",
    select_workspace: "Select workspace",
    recent_workspaces: "Recent",
    no_recent_workspaces: "No recent workspaces",
    browse_folder: "Browse folder",
    edit_model: "Edit Model",
    provider: "Provider",
    website: "Website",
    api_settings: "API Settings",
    model_settings: "Model Settings",
    default_model_id: "Default Model ID",
    fallback_model_id: "Fallback Model ID",
    available_models: "Available Models",
    add_model_id: "Add model ID...",
    generation_params: "Generation Parameters",
    supported_features: "Supported Features",
    streaming: "Streaming",
    vision: "Vision",
    function_calling: "Function Calling",
    quick_setup: "Quick Setup",
    models_available: "models available",
    streaming_short: "Stream",
    tools_short: "Tools",
    visit_website: "Visit website",
    edit: "Edit",
    fetch_models: "Fetch Models",
    available_from_api: "Available from API",
    selected_models: "Selected Models",
    no_models_selected: "No models selected",
    show_api_key: "Show API Key",
    hide_api_key: "Hide API Key",
    api_required_for_models: "Please enter API base and key first",
    no_model_configured: "Please configure a model first",
    fetch_models_failed: "Failed, please check API base and key",
    vision_short: "Vision",
    model_management: "Model Management",
    models_configured: "models configured",
    provider_name: "Provider Name",
    api_base_url: "API Base URL",
    search_models: "Search models...",
    no_models_found: "No models found",
    back: "Back",
    conversations: "Conversations",
    click_to_settings: "Click to open settings",
    fetching: "Fetching...",
    selected: "Selected",
    click_to_set_default: "Click to set as default",
    advanced_settings: "Advanced Settings",
    fallback_model: "Fallback Model",
    optional: "Optional",
    active: "Active",
    no_models_configured: "No models configured",
    click_add_model_to_start: "Click the button above to add a model",
    remove: "Remove",
    update: "Update",
    fallback_tip: "Auto-switch when primary model unavailable",
    max_tokens_tip: "Max length per response",
    temperature_tip: "Higher=more random(0-2), 0.3 for precise tasks",
    top_p_tip: "Nucleus sampling(0-1), recommend 0.9",
    features_tip: "Check features the model actually supports",
    streaming_tip: "Show response word by word",
    vision_tip: "Support image input analysis",
    tools_tip: "Function calling, required for Agent mode",
    safety_low: "LOW",
    safety_medium: "MEDIUM",
    safety_high: "HIGH",
    safety_critical: "CRITICAL",
    copy_text: "Copy text",
    regenerate: "Regenerate",
    delete_message: "Delete message",
    shortcut_send: "Send message",
    shortcut_new_line: "New line",
    shortcut_new_chat: "New chat",
    shortcut_toggle_left: "Toggle left sidebar",
    shortcut_toggle_right: "Toggle right sidebar",
    shortcut_compress: "Compress context",
    shortcut_cancel: "Cancel operation",
    shortcut_escape: "Close dialog / Deny confirm",
    help_log_title: "Help & Logs",
    about_title: "About MiMo Desktop",
    about_description: "A Wails-based AI chat desktop client powered by MiMo LLM.",
    about_github: "GitHub repository",
    about_version_info: "Version Info",
  },
};

export function t(key: TranslationKey): string {
  const lang = useSettingsStore.getState().language;
  return translations[lang]?.[key] ?? translations.en[key] ?? key;
}

// Built-in tool description translations
const toolDescriptions: Record<string, { zh: string; en: string }> = {
  clipboard: { zh: "读写系统剪贴板", en: "Read from or write to the system clipboard" },
  dependency: { zh: "管理项目依赖 (自动检测 npm/pip/go/cargo)", en: "Manage project dependencies (auto-detects npm/pip/go/cargo)" },
  dir_list: { zh: "列出指定路径的文件和目录", en: "List files and directories at the given path" },
  env: { zh: "读取或列出环境变量", en: "Read or list environment variables" },
  file_delete: { zh: "删除文件或空目录", en: "Delete a file or empty directory" },
  dir_create: { zh: "创建目录（含父目录）", en: "Create a directory (including parent directories)" },
  file_diff: { zh: "比较两个文件的差异", en: "Compare two files and show differences" },
  file_edit: { zh: "编辑文件内容", en: "Edit file contents" },
  file_write: { zh: "写入文件内容", en: "Write content to a file" },
  git_status: { zh: "查看工作区状态", en: "Show the working tree status" },
  git_diff: { zh: "查看提交或工作区变更", en: "Show changes between commits or working tree" },
  git_log: { zh: "查看提交日志", en: "Show commit logs" },
  git_commit: { zh: "暂存文件并创建提交", en: "Stage files and create a git commit" },
  git_branch: { zh: "列出、创建或删除分支", en: "List, create, or delete git branches" },
  git_checkout: { zh: "切换分支或恢复文件", en: "Switch branches or restore files" },
  git_merge: { zh: "合并分支到当前分支", en: "Merge a branch into the current branch" },
  file_read: { zh: "读取文件内容", en: "Read file contents" },
  glob: { zh: "按模式匹配查找文件", en: "Find files matching a glob pattern" },
  http_request: { zh: "发送 HTTP 请求 (GET/POST/PUT/DELETE)", en: "Send HTTP requests (GET, POST, PUT, DELETE)" },
  json_query: { zh: "读取和查询 JSON/YAML 文件", en: "Read and query JSON/YAML files using dot-notation paths" },
  process: { zh: "列出进程或终止指定进程", en: "List running processes or kill a process by PID" },
  search: { zh: "在文件中搜索文本模式", en: "Search for a text pattern in files" },
  shell: { zh: "执行 Shell 命令", en: "Execute shell commands" },
  web_fetch: { zh: "获取网页内容", en: "Fetch content from a URL" },
  web_search: { zh: "搜索互联网", en: "Search the web" },
};

export function td(toolName: string): string {
  const lang = useSettingsStore.getState().language;
  const entry = toolDescriptions[toolName];
  if (entry) return entry[lang] ?? entry.en;
  // Fallback: replace underscores with spaces
  return toolName.replace(/_/g, " ");
}
