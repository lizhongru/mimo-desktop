export namespace desktop {
	
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
	
	
	export class SessionDTO {
	    id: string;
	    modelName: string;
	    userName: string;
	    lastMessage: string;
	    workingDir: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.modelName = source["modelName"];
	        this.userName = source["userName"];
	        this.lastMessage = source["lastMessage"];
	        this.workingDir = source["workingDir"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SessionData {
	    id: string;
	    modelName: string;
	    messages: ChatMessageDTO[];
	
	    static createFrom(source: any = {}) {
	        return new SessionData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
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

