export namespace desktop {
	
	export class ActorInfo {
	    id: string;
	    type: string;
	    session_id: string;
	    parent_id?: string;
	    status: string;
	    prompt: string;
	    result?: string;
	    error?: string;
	    created_at: number;
	    started_at?: number;
	    completed_at?: number;
	
	    static createFrom(source: any = {}) {
	        return new ActorInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.session_id = source["session_id"];
	        this.parent_id = source["parent_id"];
	        this.status = source["status"];
	        this.prompt = source["prompt"];
	        this.result = source["result"];
	        this.error = source["error"];
	        this.created_at = source["created_at"];
	        this.started_at = source["started_at"];
	        this.completed_at = source["completed_at"];
	    }
	}
	export class ActorResult {
	    success: boolean;
	    message: string;
	    actor?: ActorInfo;
	
	    static createFrom(source: any = {}) {
	        return new ActorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.actor = this.convertValues(source["actor"], ActorInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PermissionRuleDTO {
	    permission: string;
	    action: string;
	    pattern?: string;
	
	    static createFrom(source: any = {}) {
	        return new PermissionRuleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.permission = source["permission"];
	        this.action = source["action"];
	        this.pattern = source["pattern"];
	    }
	}
	export class PermissionSettingsDTO {
	    rules: PermissionRuleDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PermissionSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rules = this.convertValues(source["rules"], PermissionRuleDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CheckpointSettingsDTO {
	    autoCheckpoint: boolean;
	    tokenThreshold: number;
	    maxCheckpoints: number;
	    reconstructOnResume: boolean;
	    contextBudget: number;
	
	    static createFrom(source: any = {}) {
	        return new CheckpointSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoCheckpoint = source["autoCheckpoint"];
	        this.tokenThreshold = source["tokenThreshold"];
	        this.maxCheckpoints = source["maxCheckpoints"];
	        this.reconstructOnResume = source["reconstructOnResume"];
	        this.contextBudget = source["contextBudget"];
	    }
	}
	export class MemorySettingsDTO {
	    ccIndex: boolean;
	    searchScoreFloor: number;
	
	    static createFrom(source: any = {}) {
	        return new MemorySettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ccIndex = source["ccIndex"];
	        this.searchScoreFloor = source["searchScoreFloor"];
	    }
	}
	export class AdvancedSettingsDTO {
	    memory: MemorySettingsDTO;
	    checkpoint: CheckpointSettingsDTO;
	    permission: PermissionSettingsDTO;
	
	    static createFrom(source: any = {}) {
	        return new AdvancedSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memory = this.convertValues(source["memory"], MemorySettingsDTO);
	        this.checkpoint = this.convertValues(source["checkpoint"], CheckpointSettingsDTO);
	        this.permission = this.convertValues(source["permission"], PermissionSettingsDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AgentConfigInfo {
	    name: string;
	    mode: string;
	    color: string;
	    description: string;
	    prompt: string;
	    tool_allowlist?: string[];
	
	    static createFrom(source: any = {}) {
	        return new AgentConfigInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.mode = source["mode"];
	        this.color = source["color"];
	        this.description = source["description"];
	        this.prompt = source["prompt"];
	        this.tool_allowlist = source["tool_allowlist"];
	    }
	}
	export class AgentDTO {
	    maxIterations: number;
	    planningMode: string;
	    permission: string;
	    reasoningLevel: string;
	    showTokenUsage: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxIterations = source["maxIterations"];
	        this.planningMode = source["planningMode"];
	        this.permission = source["permission"];
	        this.reasoningLevel = source["reasoningLevel"];
	        this.showTokenUsage = source["showTokenUsage"];
	    }
	}
	export class AgentSwitchResult {
	    success: boolean;
	    message: string;
	    agent?: AgentConfigInfo;
	
	    static createFrom(source: any = {}) {
	        return new AgentSwitchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.agent = this.convertValues(source["agent"], AgentConfigInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SafetyDTO {
	    level: string;
	    permission: string;
	
	    static createFrom(source: any = {}) {
	        return new SafetyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.permission = source["permission"];
	    }
	}
	export class ModelDTO {
	    provider: string;
	    website: string;
	    apiBase: string;
	    apiKey: string;
	    model: string;
	    models: string[];
	    fallback: string;
	    maxTokens: number;
	    temperature: number;
	    topP: number;
	    streaming: boolean;
	    vision: boolean;
	    tools: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.website = source["website"];
	        this.apiBase = source["apiBase"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.models = source["models"];
	        this.fallback = source["fallback"];
	        this.maxTokens = source["maxTokens"];
	        this.temperature = source["temperature"];
	        this.topP = source["topP"];
	        this.streaming = source["streaming"];
	        this.vision = source["vision"];
	        this.tools = source["tools"];
	    }
	}
	export class AppConfigDTO {
	    defaultModel: string;
	    language: string;
	    theme: string;
	    userName: string;
	    models: Record<string, ModelDTO>;
	    safety: SafetyDTO;
	    agent: AgentDTO;
	    memory: MemorySettingsDTO;
	    checkpoint: CheckpointSettingsDTO;
	    permission: PermissionSettingsDTO;
	
	    static createFrom(source: any = {}) {
	        return new AppConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultModel = source["defaultModel"];
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.userName = source["userName"];
	        this.models = this.convertValues(source["models"], ModelDTO, true);
	        this.safety = this.convertValues(source["safety"], SafetyDTO);
	        this.agent = this.convertValues(source["agent"], AgentDTO);
	        this.memory = this.convertValues(source["memory"], MemorySettingsDTO);
	        this.checkpoint = this.convertValues(source["checkpoint"], CheckpointSettingsDTO);
	        this.permission = this.convertValues(source["permission"], PermissionSettingsDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatMessageDTO {
	    role: string;
	    content: string;
	    thinking?: string;
	    toolLines?: string[];
	    tokens: number;
	    toolCalls: number;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.thinking = source["thinking"];
	        this.toolLines = source["toolLines"];
	        this.tokens = source["tokens"];
	        this.toolCalls = source["toolCalls"];
	        this.durationMs = source["durationMs"];
	    }
	}
	export class CheckpointInfo {
	    id: string;
	    summary: string;
	    token_count: number;
	    message_offset: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new CheckpointInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.summary = source["summary"];
	        this.token_count = source["token_count"];
	        this.message_offset = source["message_offset"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CheckpointResult {
	    success: boolean;
	    message: string;
	    id?: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckpointResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.id = source["id"];
	    }
	}
	
	export class DreamResult {
	    success: boolean;
	    message: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new DreamResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.count = source["count"];
	    }
	}
	export class ExportMessage {
	    role: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	    }
	}
	export class MCPServerInfo {
	    name: string;
	    connected: boolean;
	    toolCount: number;
	    tools: string[];
	
	    static createFrom(source: any = {}) {
	        return new MCPServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.connected = source["connected"];
	        this.toolCount = source["toolCount"];
	        this.tools = source["tools"];
	    }
	}
	export class MemoryFileInfo {
	    path: string;
	    name: string;
	    size: number;
	    // Go type: time
	    updatedAt: any;
	    scope: string;
	
	    static createFrom(source: any = {}) {
	        return new MemoryFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.scope = source["scope"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MemorySearchResult {
	    path: string;
	    snippet: string;
	    score: number;
	    scope: string;
	    scope_id: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new MemorySearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.snippet = source["snippet"];
	        this.score = source["score"];
	        this.scope = source["scope"];
	        this.scope_id = source["scope_id"];
	        this.type = source["type"];
	    }
	}
	
	
	
	
	
	export class SessionDTO {
	    id: string;
	    workspaceId: string;
	    modelName: string;
	    userName: string;
	    lastMessage: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.modelName = source["modelName"];
	        this.userName = source["userName"];
	        this.lastMessage = source["lastMessage"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SessionData {
	    id: string;
	    workspaceId: string;
	    modelName: string;
	    messages: ChatMessageDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SessionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.modelName = source["modelName"];
	        this.messages = this.convertValues(source["messages"], ChatMessageDTO);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SkillCandidateInfo {
	    name: string;
	    description: string;
	    confidence: number;
	    pattern?: string;
	    commands?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillCandidateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.confidence = source["confidence"];
	        this.pattern = source["pattern"];
	        this.commands = source["commands"];
	    }
	}
	export class TaskEventInfo {
	    id: number;
	    task_id: string;
	    at: number;
	    kind: string;
	    summary?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskEventInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.task_id = source["task_id"];
	        this.at = source["at"];
	        this.kind = source["kind"];
	        this.summary = source["summary"];
	    }
	}
	export class TaskInfo {
	    id: string;
	    session_id: string;
	    parent_task_id?: string;
	    status: string;
	    summary: string;
	    owner?: string;
	    created_at: number;
	    last_event_at: number;
	    ended_at?: number;
	
	    static createFrom(source: any = {}) {
	        return new TaskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.session_id = source["session_id"];
	        this.parent_task_id = source["parent_task_id"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.owner = source["owner"];
	        this.created_at = source["created_at"];
	        this.last_event_at = source["last_event_at"];
	        this.ended_at = source["ended_at"];
	    }
	}
	export class TaskResult {
	    success: boolean;
	    message: string;
	    task?: TaskInfo;
	
	    static createFrom(source: any = {}) {
	        return new TaskResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.task = this.convertValues(source["task"], TaskInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ToolInfo {
	    name: string;
	    description: string;
	    safetyLevel: string;
	    isMcp: boolean;
	    serverName?: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.safetyLevel = source["safetyLevel"];
	        this.isMcp = source["isMcp"];
	        this.serverName = source["serverName"];
	    }
	}
	export class WorkspaceDTO {
	    id: string;
	    name: string;
	    type: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.path = source["path"];
	    }
	}

}

export namespace llm {
	
	export class ModelInfo {
	    id: string;
	    owned_by?: string;
	    description?: string;
	    context_window?: number;
	    max_output?: number;
	    capabilities?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.owned_by = source["owned_by"];
	        this.description = source["description"];
	        this.context_window = source["context_window"];
	        this.max_output = source["max_output"];
	        this.capabilities = source["capabilities"];
	    }
	}

}

