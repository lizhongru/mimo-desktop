import { useSettingsStore } from "../stores/settingsStore";

type TranslationKey =
  | "settings"
  | "tools_mcp"
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
  | "remove_project";

const translations: Record<"zh" | "en", Record<TranslationKey, string>> = {
  zh: {
    settings: "设置",
    tools_mcp: "工具 & MCP",
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
  },
  en: {
    settings: "Settings",
    tools_mcp: "Tools & MCP",
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
