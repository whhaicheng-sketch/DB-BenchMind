export namespace bindings {
	
	export class AIAssistantConfig {
	    id: string;
	    name: string;
	    provider: string;
	    api_host: string;
	    api_endpoint: string;
	    api_key?: string;
	    model: string;
	    temperature: number;
	    description: string;
	    enter_action: string;
	    compare_with_others: boolean;
	    language: string;
	
	    static createFrom(source: any = {}) {
	        return new AIAssistantConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.provider = source["provider"];
	        this.api_host = source["api_host"];
	        this.api_endpoint = source["api_endpoint"];
	        this.api_key = source["api_key"];
	        this.model = source["model"];
	        this.temperature = source["temperature"];
	        this.description = source["description"];
	        this.enter_action = source["enter_action"];
	        this.compare_with_others = source["compare_with_others"];
	        this.language = source["language"];
	    }
	}
	export class AIChatTestRequest {
	    provider: string;
	    api_host: string;
	    api_endpoint: string;
	    api_key: string;
	    model: string;
	    prompt: string;
	    temperature: number;
	
	    static createFrom(source: any = {}) {
	        return new AIChatTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.api_host = source["api_host"];
	        this.api_endpoint = source["api_endpoint"];
	        this.api_key = source["api_key"];
	        this.model = source["model"];
	        this.prompt = source["prompt"];
	        this.temperature = source["temperature"];
	    }
	}
	export class AIChatTestResult {
	    success: boolean;
	    latency_ms: number;
	    content?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIChatTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.content = source["content"];
	        this.error = source["error"];
	    }
	}
	export class AIModelInfo {
	    id: string;
	    name?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class AIModelQueryRequest {
	    provider: string;
	    api_host: string;
	    api_key: string;
	
	    static createFrom(source: any = {}) {
	        return new AIModelQueryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.api_host = source["api_host"];
	        this.api_key = source["api_key"];
	    }
	}
	export class AIModelQueryResult {
	    success: boolean;
	    models?: AIModelInfo[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIModelQueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.models = this.convertValues(source["models"], AIModelInfo);
	        this.error = source["error"];
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
	export class AITestRequest {
	    provider: string;
	    api_host: string;
	    api_endpoint: string;
	    api_key: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AITestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.api_host = source["api_host"];
	        this.api_endpoint = source["api_endpoint"];
	        this.api_key = source["api_key"];
	        this.model = source["model"];
	    }
	}
	export class AITestResult {
	    success: boolean;
	    latency_ms: number;
	    message?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AITestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class BenchmarkBinding {
	
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class BenchmarkResultDTO {
	    run_id: string;
	    tps: number;
	    tpm: number;
	    max_tps: number;
	    avg_tps: number;
	    max_tpm: number;
	    avg_tpm: number;
	    latency_avg_ms: number;
	    latency_min_ms: number;
	    latency_max_ms: number;
	    latency_p95_ms: number;
	    latency_p99_ms: number;
	    error_count: number;
	    error_rate_percent: number;
	    total_transactions: number;
	    duration_seconds: number;
	    connection_name?: string;
	    template_name?: string;
	    database_type?: string;
	    threads?: number;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkResultDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.run_id = source["run_id"];
	        this.tps = source["tps"];
	        this.tpm = source["tpm"];
	        this.max_tps = source["max_tps"];
	        this.avg_tps = source["avg_tps"];
	        this.max_tpm = source["max_tpm"];
	        this.avg_tpm = source["avg_tpm"];
	        this.latency_avg_ms = source["latency_avg_ms"];
	        this.latency_min_ms = source["latency_min_ms"];
	        this.latency_max_ms = source["latency_max_ms"];
	        this.latency_p95_ms = source["latency_p95_ms"];
	        this.latency_p99_ms = source["latency_p99_ms"];
	        this.error_count = source["error_count"];
	        this.error_rate_percent = source["error_rate_percent"];
	        this.total_transactions = source["total_transactions"];
	        this.duration_seconds = source["duration_seconds"];
	        this.connection_name = source["connection_name"];
	        this.template_name = source["template_name"];
	        this.database_type = source["database_type"];
	        this.threads = source["threads"];
	    }
	}
	export class BenchmarkRunDTO {
	    id: string;
	    task_id: string;
	    state: string;
	    created_at: string;
	    started_at?: string;
	    completed_at?: string;
	    duration?: string;
	    duration_ms?: number;
	    result?: BenchmarkResultDTO;
	    error_message?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkRunDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.task_id = source["task_id"];
	        this.state = source["state"];
	        this.created_at = source["created_at"];
	        this.started_at = source["started_at"];
	        this.completed_at = source["completed_at"];
	        this.duration = source["duration"];
	        this.duration_ms = source["duration_ms"];
	        this.result = this.convertValues(source["result"], BenchmarkResultDTO);
	        this.error_message = source["error_message"];
	        this.message = source["message"];
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
	export class BenchmarkListResult {
	    runs: BenchmarkRunDTO[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runs = this.convertValues(source["runs"], BenchmarkRunDTO);
	        this.error = source["error"];
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
	export class BenchmarkOptionsDTO {
	    skip_prepare: boolean;
	    skip_cleanup: boolean;
	    warmup_time: number;
	    dry_run: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkOptionsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skip_prepare = source["skip_prepare"];
	        this.skip_cleanup = source["skip_cleanup"];
	        this.warmup_time = source["warmup_time"];
	        this.dry_run = source["dry_run"];
	    }
	}
	
	
	export class BenchmarkStartRequest {
	    connection_id: string;
	    template_id: string;
	    parameters: Record<string, any>;
	    options: BenchmarkOptionsDTO;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkStartRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection_id = source["connection_id"];
	        this.template_id = source["template_id"];
	        this.parameters = source["parameters"];
	        this.options = this.convertValues(source["options"], BenchmarkOptionsDTO);
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
	export class BenchmarkStartResult {
	    run_id: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkStartResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.run_id = source["run_id"];
	        this.error = source["error"];
	    }
	}
	export class BenchmarkStatusResult {
	    run?: BenchmarkRunDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.run = this.convertValues(source["run"], BenchmarkRunDTO);
	        this.error = source["error"];
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
	export class ConnectionCreateRequest {
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    database: string;
	    username: string;
	    password: string;
	    ssl_mode: string;
	    sid?: string;
	    service_name?: string;
	    connect_type?: string;
	    identifier_type?: string;
	    tns_name?: string;
	    connect_as?: string;
	    ssh_enabled: boolean;
	    ssh_port?: number;
	    ssh_username?: string;
	    ssh_password?: string;
	    winrm_enabled: boolean;
	    winrm_port?: number;
	    winrm_use_https: boolean;
	    winrm_username?: string;
	    winrm_password?: string;
	    ai_assistants?: AIAssistantConfig[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionCreateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database = source["database"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.ssl_mode = source["ssl_mode"];
	        this.sid = source["sid"];
	        this.service_name = source["service_name"];
	        this.connect_type = source["connect_type"];
	        this.identifier_type = source["identifier_type"];
	        this.tns_name = source["tns_name"];
	        this.connect_as = source["connect_as"];
	        this.ssh_enabled = source["ssh_enabled"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_username = source["ssh_username"];
	        this.ssh_password = source["ssh_password"];
	        this.winrm_enabled = source["winrm_enabled"];
	        this.winrm_port = source["winrm_port"];
	        this.winrm_use_https = source["winrm_use_https"];
	        this.winrm_username = source["winrm_username"];
	        this.winrm_password = source["winrm_password"];
	        this.ai_assistants = this.convertValues(source["ai_assistants"], AIAssistantConfig);
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
	export class ConnectionDTO {
	    id: string;
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    database?: string;
	    username: string;
	    password?: string;
	    ssl_mode?: string;
	    sid?: string;
	    service_name?: string;
	    connect_type?: string;
	    identifier_type?: string;
	    tns_name?: string;
	    connect_as?: string;
	    ssh_enabled: boolean;
	    ssh_port?: number;
	    ssh_username?: string;
	    ssh_password?: string;
	    winrm_enabled: boolean;
	    winrm_port?: number;
	    winrm_use_https: boolean;
	    winrm_username?: string;
	    winrm_password?: string;
	    trust_server_certificate: boolean;
	    ai_assistants?: AIAssistantConfig[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database = source["database"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.ssl_mode = source["ssl_mode"];
	        this.sid = source["sid"];
	        this.service_name = source["service_name"];
	        this.connect_type = source["connect_type"];
	        this.identifier_type = source["identifier_type"];
	        this.tns_name = source["tns_name"];
	        this.connect_as = source["connect_as"];
	        this.ssh_enabled = source["ssh_enabled"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_username = source["ssh_username"];
	        this.ssh_password = source["ssh_password"];
	        this.winrm_enabled = source["winrm_enabled"];
	        this.winrm_port = source["winrm_port"];
	        this.winrm_use_https = source["winrm_use_https"];
	        this.winrm_username = source["winrm_username"];
	        this.winrm_password = source["winrm_password"];
	        this.trust_server_certificate = source["trust_server_certificate"];
	        this.ai_assistants = this.convertValues(source["ai_assistants"], AIAssistantConfig);
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
	export class ConnectionCreateResult {
	    connection?: ConnectionDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionCreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection = this.convertValues(source["connection"], ConnectionDTO);
	        this.error = source["error"];
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
	
	export class ConnectionDeleteResult {
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class ConnectionListResult {
	    connections: ConnectionDTO[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connections = this.convertValues(source["connections"], ConnectionDTO);
	        this.error = source["error"];
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
	export class SSHWinRMTestResult {
	    success: boolean;
	    latency_ms: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHWinRMTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.error = source["error"];
	    }
	}
	export class ConnectionTestResult {
	    success: boolean;
	    latency_ms: number;
	    database_version?: string;
	    error?: string;
	    ssh_result?: SSHWinRMTestResult;
	    winrm_result?: SSHWinRMTestResult;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.database_version = source["database_version"];
	        this.error = source["error"];
	        this.ssh_result = this.convertValues(source["ssh_result"], SSHWinRMTestResult);
	        this.winrm_result = this.convertValues(source["winrm_result"], SSHWinRMTestResult);
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
	export class ConnectionUpdateRequest {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    database: string;
	    username: string;
	    password?: string;
	    ssl_mode: string;
	    sid?: string;
	    service_name?: string;
	    connect_type?: string;
	    identifier_type?: string;
	    tns_name?: string;
	    connect_as?: string;
	    ssh_enabled: boolean;
	    ssh_port?: number;
	    ssh_username?: string;
	    ssh_password?: string;
	    winrm_enabled: boolean;
	    winrm_port?: number;
	    winrm_use_https: boolean;
	    winrm_username?: string;
	    winrm_password?: string;
	    ai_assistants?: AIAssistantConfig[];
	
	    static createFrom(source: any = {}) {
	        return new ConnectionUpdateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database = source["database"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.ssl_mode = source["ssl_mode"];
	        this.sid = source["sid"];
	        this.service_name = source["service_name"];
	        this.connect_type = source["connect_type"];
	        this.identifier_type = source["identifier_type"];
	        this.tns_name = source["tns_name"];
	        this.connect_as = source["connect_as"];
	        this.ssh_enabled = source["ssh_enabled"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_username = source["ssh_username"];
	        this.ssh_password = source["ssh_password"];
	        this.winrm_enabled = source["winrm_enabled"];
	        this.winrm_port = source["winrm_port"];
	        this.winrm_use_https = source["winrm_use_https"];
	        this.winrm_username = source["winrm_username"];
	        this.winrm_password = source["winrm_password"];
	        this.ai_assistants = this.convertValues(source["ai_assistants"], AIAssistantConfig);
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
	export class ConnectionUpdateResult {
	    connection?: ConnectionDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionUpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection = this.convertValues(source["connection"], ConnectionDTO);
	        this.error = source["error"];
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
	export class AnomalyPreview {
	    type?: string;
	    severity?: string;
	    summary?: string;
	    value?: number;

	    static createFrom(source: any = {}) {
	        return new AnomalyPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.severity = source["severity"];
	        this.summary = source["summary"];
	        this.value = source["value"];
	    }
	}
	export class WindowPreview {
	    name?: string;
	    sample_count?: number;

	    static createFrom(source: any = {}) {
	        return new WindowPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sample_count = source["sample_count"];
	    }
	}
	export class DetailedDataPreview {
	    bundle_filename?: string;
	    compressed_size_bytes?: number;
	    sampling_policy?: string;
	    retained_windows?: WindowPreview[];
	    anomaly_windows?: AnomalyPreview[];
	    raw_sample_count?: number;
	    schema_preview?: {[key: string]: any};

	    static createFrom(source: any = {}) {
	        return new DetailedDataPreview(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bundle_filename = source["bundle_filename"];
	        this.compressed_size_bytes = source["compressed_size_bytes"];
	        this.sampling_policy = source["sampling_policy"];
	        this.retained_windows = source["retained_windows"]?.map(function(element: any) {
	            return WindowPreview.createFrom(element);
	        });
	        this.anomaly_windows = source["anomaly_windows"]?.map(function(element: any) {
	            return AnomalyPreview.createFrom(element);
	        });
	        this.raw_sample_count = source["raw_sample_count"];
	        this.schema_preview = source["schema_preview"];
	    }
	}
	export class DetailedDataResult {
	    preview?: DetailedDataPreview;
	    markdown?: string;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new DetailedDataResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preview = source["preview"] && DetailedDataPreview.createFrom(source["preview"]);
	        this.markdown = source["markdown"];
	        this.error = source["error"];
	    }
	}
	export class DeleteAllResult {
	    count?: number;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new DeleteAllResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.error = source["error"];
	    }
	}
	export class DeleteReportResult {
	    success?: boolean;
	    error?: string;

	    static createFrom(source: any = {}) {
	        return new DeleteReportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class ExportFilePathsResult {
	    metrics?: string;
	    monitoring?: string;
	    raw?: string;
	    html?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportFilePathsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metrics = source["metrics"];
	        this.monitoring = source["monitoring"];
	        this.raw = source["raw"];
	        this.html = source["html"];
	        this.error = source["error"];
	    }
	}
	export class ExportHTMLResult {
	    html?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportHTMLResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.html = source["html"];
	        this.error = source["error"];
	    }
	}
	export class ExportJSONResult {
	    data?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportJSONResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.error = source["error"];
	    }
	}
	export class ListReportsOptionsDTO {
	    page: number;
	    page_size: number;
	    suite_id?: string;
	    status?: string;
	    connection_id?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListReportsOptionsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.suite_id = source["suite_id"];
	        this.status = source["status"];
	        this.connection_id = source["connection_id"];
	    }
	}
	export class ListSuitesOptionsDTO {
	    page: number;
	    page_size: number;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListSuitesOptionsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.page_size = source["page_size"];
	        this.status = source["status"];
	    }
	}
	export class MonitorBinding {
	
	
	    static createFrom(source: any = {}) {
	        return new MonitorBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class MonitorConfigDTO {
	    buffer_size: number;
	    sample_rate_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new MonitorConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.buffer_size = source["buffer_size"];
	        this.sample_rate_ms = source["sample_rate_ms"];
	    }
	}
	export class MonitorDataDTO {
	    current_tpm: number;
	    current_tps: number;
	    tpm_points: collector.MetricPoint[];
	    tps_points: collector.MetricPoint[];
	    stats?: collector.MetricStats;
	
	    static createFrom(source: any = {}) {
	        return new MonitorDataDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current_tpm = source["current_tpm"];
	        this.current_tps = source["current_tps"];
	        this.tpm_points = this.convertValues(source["tpm_points"], collector.MetricPoint);
	        this.tps_points = this.convertValues(source["tps_points"], collector.MetricPoint);
	        this.stats = this.convertValues(source["stats"], collector.MetricStats);
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
	export class MonitorStateDTO {
	    is_running: boolean;
	    run_id: string;
	    tpm_count: number;
	    tps_count: number;
	    system_running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MonitorStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_running = source["is_running"];
	        this.run_id = source["run_id"];
	        this.tpm_count = source["tpm_count"];
	        this.tps_count = source["tps_count"];
	        this.system_running = source["system_running"];
	    }
	}
	export class ParamDTO {
	    type: string;
	    label: string;
	    default: any;
	    min?: number;
	    max?: number;
	    options?: string[];
	    extra?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ParamDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.default = source["default"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.options = source["options"];
	        this.extra = source["extra"];
	    }
	}
	export class ParamDefinition {
	    name: string;
	    type: string;
	    label: string;
	    default: any;
	    min?: number;
	    max?: number;
	    options?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ParamDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.label = source["label"];
	        this.default = source["default"];
	        this.min = source["min"];
	        this.max = source["max"];
	        this.options = source["options"];
	    }
	}
	export class ReportDTO {
	    id: string;
	    suite_id: string;
	    suite_item_id?: string;
	    source_type: string;
	    connection_id: string;
	    connection_name?: string;
	    database_type: string;
	    template_id?: string;
	    template_name?: string;
	    started_at: string;
	    ended_at?: string;
	    duration_ms?: number;
	    status: string;
	    error_message?: string;
	    tpm?: number;
	    tps?: number;
	    qps?: number;
	    throughput?: number;
	    latency_avg_ms?: number;
	    latency_p95_ms?: number;
	    latency_p99_ms?: number;
	    error_count?: number;
	    tags?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.suite_id = source["suite_id"];
	        this.suite_item_id = source["suite_item_id"];
	        this.source_type = source["source_type"];
	        this.connection_id = source["connection_id"];
	        this.connection_name = source["connection_name"];
	        this.database_type = source["database_type"];
	        this.template_id = source["template_id"];
	        this.template_name = source["template_name"];
	        this.started_at = source["started_at"];
	        this.ended_at = source["ended_at"];
	        this.duration_ms = source["duration_ms"];
	        this.status = source["status"];
	        this.error_message = source["error_message"];
	        this.tpm = source["tpm"];
	        this.tps = source["tps"];
	        this.qps = source["qps"];
	        this.throughput = source["throughput"];
	        this.latency_avg_ms = source["latency_avg_ms"];
	        this.latency_p95_ms = source["latency_p95_ms"];
	        this.latency_p99_ms = source["latency_p99_ms"];
	        this.error_count = source["error_count"];
	        this.tags = source["tags"];
	    }
	}
	export class ReportListResult {
	    reports: ReportDTO[];
	    total: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reports = this.convertValues(source["reports"], ReportDTO);
	        this.total = source["total"];
	        this.error = source["error"];
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
	export class ReportMetricsResult {
	    metrics?: report.MetricsData;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportMetricsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metrics = this.convertValues(source["metrics"], report.MetricsData);
	        this.error = source["error"];
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
	export class ReportResult {
	    report?: ReportDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.report = this.convertValues(source["report"], ReportDTO);
	        this.error = source["error"];
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
	export class SSHTestRequest {
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	
	export class SuiteDTO {
	    id: string;
	    name?: string;
	    execution_mode?: string;
	    failure_policy?: string;
	    cleanup_enabled: boolean;
	    suite_manifest_json_path?: string;
	    status: string;
	    started_at?: string;
	    ended_at?: string;
	    total_items: number;
	    completed_items: number;
	    success_items: number;
	    failed_items: number;
	    skipped_items: number;
	    suite_report_json_path?: string;
	    suite_report_html_path?: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SuiteDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.execution_mode = source["execution_mode"];
	        this.failure_policy = source["failure_policy"];
	        this.cleanup_enabled = source["cleanup_enabled"];
	        this.suite_manifest_json_path = source["suite_manifest_json_path"];
	        this.status = source["status"];
	        this.started_at = source["started_at"];
	        this.ended_at = source["ended_at"];
	        this.total_items = source["total_items"];
	        this.completed_items = source["completed_items"];
	        this.success_items = source["success_items"];
	        this.failed_items = source["failed_items"];
	        this.skipped_items = source["skipped_items"];
	        this.suite_report_json_path = source["suite_report_json_path"];
	        this.suite_report_html_path = source["suite_report_html_path"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class SuiteListResult {
	    suites: SuiteDTO[];
	    total: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SuiteListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.suites = this.convertValues(source["suites"], SuiteDTO);
	        this.total = source["total"];
	        this.error = source["error"];
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
	export class SuiteResult {
	    suite?: SuiteDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SuiteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.suite = this.convertValues(source["suite"], SuiteDTO);
	        this.error = source["error"];
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
	export class SystemHistoryDTO {
	    cpu: collector.SystemMetricPoint[];
	    disk_io: collector.SystemMetricPoint[];
	    disk_space: collector.SystemMetricPoint[];
	
	    static createFrom(source: any = {}) {
	        return new SystemHistoryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], collector.SystemMetricPoint);
	        this.disk_io = this.convertValues(source["disk_io"], collector.SystemMetricPoint);
	        this.disk_space = this.convertValues(source["disk_space"], collector.SystemMetricPoint);
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
	export class SystemMetricsDTO {
	    cpu_percent: number;
	    disk_read_bps: number;
	    disk_write_bps: number;
	    disk_used_percent: number;
	    disk_used_gb: number;
	    disk_total_gb: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemMetricsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu_percent = source["cpu_percent"];
	        this.disk_read_bps = source["disk_read_bps"];
	        this.disk_write_bps = source["disk_write_bps"];
	        this.disk_used_percent = source["disk_used_percent"];
	        this.disk_used_gb = source["disk_used_gb"];
	        this.disk_total_gb = source["disk_total_gb"];
	    }
	}
	export class TaskActionResult {
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class TaskBinding {
	
	
	    static createFrom(source: any = {}) {
	        return new TaskBinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class TaskDraftRequest {
	    task_name: string;
	    database_type?: string;
	    template_id: string;
	    connection_id: string;
	    action: string;
	    preview_token?: string;
	    overrides: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new TaskDraftRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task_name = source["task_name"];
	        this.database_type = source["database_type"];
	        this.template_id = source["template_id"];
	        this.connection_id = source["connection_id"];
	        this.action = source["action"];
	        this.preview_token = source["preview_token"];
	        this.overrides = source["overrides"];
	    }
	}
	export class TaskListResult {
	    tasks: task.ExecutionTask[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tasks = this.convertValues(source["tasks"], task.ExecutionTask);
	        this.error = source["error"];
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
	export class TaskLogsRequest {
	    task_id: string;
	    limit: number;
	    query: string;
	    phase: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskLogsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task_id = source["task_id"];
	        this.limit = source["limit"];
	        this.query = source["query"];
	        this.phase = source["phase"];
	    }
	}
	export class TaskLogsResult {
	    lines: task.LogLine[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskLogsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lines = this.convertValues(source["lines"], task.LogLine);
	        this.error = source["error"];
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
	export class TaskResult {
	    task?: task.ExecutionTask;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task = this.convertValues(source["task"], task.ExecutionTask);
	        this.error = source["error"];
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
	export class TemplateDTO {
	    id: string;
	    name: string;
	    description: string;
	    tool: string;
	    profile_type: string;
	    goal: string;
	    readonly: boolean;
	    source_alignment: string;
	    prepare_config: Record<string, any>;
	    run_config: Record<string, any>;
	    cleanup_config: Record<string, any>;
	    metrics: string[];
	    tags: string[];
	    test_position: string;
	    dbFamily: string;
	    workloadFamily: string;
	    is_builtin: boolean;
	    version: string;
	    createdAt: string;
	    compatibility: template.Compatibility;
	    phases: template.PhaseSet;
	    runtime: template.Runtime;
	    toolConfig: template.ToolConfig;
	    database_types: string[];
	    parameters?: Record<string, ParamDTO>;
	    custom_data?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new TemplateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tool = source["tool"];
	        this.profile_type = source["profile_type"];
	        this.goal = source["goal"];
	        this.readonly = source["readonly"];
	        this.source_alignment = source["source_alignment"];
	        this.prepare_config = source["prepare_config"];
	        this.run_config = source["run_config"];
	        this.cleanup_config = source["cleanup_config"];
	        this.metrics = source["metrics"];
	        this.tags = source["tags"];
	        this.test_position = source["test_position"];
	        this.dbFamily = source["dbFamily"];
	        this.workloadFamily = source["workloadFamily"];
	        this.is_builtin = source["is_builtin"];
	        this.version = source["version"];
	        this.createdAt = source["createdAt"];
	        this.compatibility = this.convertValues(source["compatibility"], template.Compatibility);
	        this.phases = this.convertValues(source["phases"], template.PhaseSet);
	        this.runtime = this.convertValues(source["runtime"], template.Runtime);
	        this.toolConfig = this.convertValues(source["toolConfig"], template.ToolConfig);
	        this.database_types = source["database_types"];
	        this.parameters = this.convertValues(source["parameters"], ParamDTO, true);
	        this.custom_data = source["custom_data"];
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
	export class TemplateDeleteResult {
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class TemplateListResult {
	    templates: TemplateDTO[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.templates = this.convertValues(source["templates"], TemplateDTO);
	        this.error = source["error"];
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
	export class TemplateParamsResult {
	    params: ParamDefinition[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateParamsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.params = this.convertValues(source["params"], ParamDefinition);
	        this.error = source["error"];
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
	export class TemplateResult {
	    template?: TemplateDTO;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TemplateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.template = this.convertValues(source["template"], TemplateDTO);
	        this.error = source["error"];
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
	export class WinRMTestRequest {
	    host: string;
	    port: number;
	    username: string;
	    password: string;
	    use_https: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WinRMTestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.use_https = source["use_https"];
	    }
	}

}

export namespace collector {
	
	export class MetricPoint {
	    // Go type: time
	    timestamp: any;
	    tps: number;
	    tpm: number;
	    latency_ms: number;
	    errors: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.tps = source["tps"];
	        this.tpm = source["tpm"];
	        this.latency_ms = source["latency_ms"];
	        this.errors = source["errors"];
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
	export class MetricStats {
	    tps_avg: number;
	    tps_max: number;
	    tps_min: number;
	    tpm_avg: number;
	    tpm_max: number;
	    tpm_min: number;
	    latency_avg: number;
	    latency_max: number;
	    latency_min: number;
	    tpm_stddev: number;
	    tpm_cv: number;
	    tps_stddev: number;
	    tps_cv: number;
	    tpm_direction_changes: number;
	    tps_direction_changes: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tps_avg = source["tps_avg"];
	        this.tps_max = source["tps_max"];
	        this.tps_min = source["tps_min"];
	        this.tpm_avg = source["tpm_avg"];
	        this.tpm_max = source["tpm_max"];
	        this.tpm_min = source["tpm_min"];
	        this.latency_avg = source["latency_avg"];
	        this.latency_max = source["latency_max"];
	        this.latency_min = source["latency_min"];
	        this.tpm_stddev = source["tpm_stddev"];
	        this.tpm_cv = source["tpm_cv"];
	        this.tps_stddev = source["tps_stddev"];
	        this.tps_cv = source["tps_cv"];
	        this.tpm_direction_changes = source["tpm_direction_changes"];
	        this.tps_direction_changes = source["tps_direction_changes"];
	    }
	}
	export class SystemMetricPoint {
	    timestamp: number;
	    cpu_percent: number;
	    disk_read_bps: number;
	    disk_write_bps: number;
	    disk_used_percent: number;
	
	    static createFrom(source: any = {}) {
	        return new SystemMetricPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.cpu_percent = source["cpu_percent"];
	        this.disk_read_bps = source["disk_read_bps"];
	        this.disk_write_bps = source["disk_write_bps"];
	        this.disk_used_percent = source["disk_used_percent"];
	    }
	}

}

export namespace report {
	
	export class MetricsBenchmark {
	    connection_id?: string;
	    connection_name?: string;
	    database_type?: string;
	    template_id?: string;
	    template_name?: string;
	
	    static createFrom(source: any = {}) {
	        return new MetricsBenchmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection_id = source["connection_id"];
	        this.connection_name = source["connection_name"];
	        this.database_type = source["database_type"];
	        this.template_id = source["template_id"];
	        this.template_name = source["template_name"];
	    }
	}
	export class MetricsTimeSeriesItem {
	    // Go type: time
	    timestamp: any;
	    tps?: number;
	    latency_avg?: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricsTimeSeriesItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.tps = source["tps"];
	        this.latency_avg = source["latency_avg"];
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
	export class MetricsSummaryData {
	    tpm?: number;
	    tps?: number;
	    qps?: number;
	    throughput?: number;
	    latency_avg_ms?: number;
	    latency_p95_ms?: number;
	    latency_p99_ms?: number;
	    error_count?: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricsSummaryData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tpm = source["tpm"];
	        this.tps = source["tps"];
	        this.qps = source["qps"];
	        this.throughput = source["throughput"];
	        this.latency_avg_ms = source["latency_avg_ms"];
	        this.latency_p95_ms = source["latency_p95_ms"];
	        this.latency_p99_ms = source["latency_p99_ms"];
	        this.error_count = source["error_count"];
	    }
	}
	export class MetricsExecution {
	    status?: string;
	    started_at?: string;
	    duration_ms?: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricsExecution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.started_at = source["started_at"];
	        this.duration_ms = source["duration_ms"];
	    }
	}
	export class MetricsData {
	    schema_version: string;
	    report_id: string;
	    suite_id?: string;
	    suite_item_id?: string;
	    generated_at?: string;
	    benchmark?: MetricsBenchmark;
	    execution?: MetricsExecution;
	    summary?: MetricsSummaryData;
	    percentiles?: Record<string, number>;
	    time_series?: MetricsTimeSeriesItem[];
	
	    static createFrom(source: any = {}) {
	        return new MetricsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schema_version = source["schema_version"];
	        this.report_id = source["report_id"];
	        this.suite_id = source["suite_id"];
	        this.suite_item_id = source["suite_item_id"];
	        this.generated_at = source["generated_at"];
	        this.benchmark = this.convertValues(source["benchmark"], MetricsBenchmark);
	        this.execution = this.convertValues(source["execution"], MetricsExecution);
	        this.summary = this.convertValues(source["summary"], MetricsSummaryData);
	        this.percentiles = source["percentiles"];
	        this.time_series = this.convertValues(source["time_series"], MetricsTimeSeriesItem);
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
	
	

}

export namespace task {
	
	export class ConnectionSnapshot {
	    id: string;
	    name: string;
	    type: string;
	    host: string;
	    port: number;
	    database?: string;
	    username: string;
	    ssh_enabled: boolean;
	    ssh_host?: string;
	    ssh_port?: number;
	    ssh_username?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database = source["database"];
	        this.username = source["username"];
	        this.ssh_enabled = source["ssh_enabled"];
	        this.ssh_host = source["ssh_host"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_username = source["ssh_username"];
	    }
	}
	export class LogLine {
	    timestamp: string;
	    phase: string;
	    stream: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new LogLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.phase = source["phase"];
	        this.stream = source["stream"];
	        this.content = source["content"];
	    }
	}
	export class SystemMetricSummary {
	    current: number;
	    series: MetricSeriesPoint[];
	
	    static createFrom(source: any = {}) {
	        return new SystemMetricSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.series = this.convertValues(source["series"], MetricSeriesPoint);
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
	export class MetricSeriesPoint {
	    timestamp: number;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricSeriesPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.value = source["value"];
	    }
	}
	export class MetricSummary {
	    current: number;
	    avg: number;
	    max: number;
	    series: MetricSeriesPoint[];
	
	    static createFrom(source: any = {}) {
	        return new MetricSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.avg = source["avg"];
	        this.max = source["max"];
	        this.series = this.convertValues(source["series"], MetricSeriesPoint);
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
	export class UnifiedMetrics {
	    tps: MetricSummary;
	    tpm: MetricSummary;
	    cpu_user: SystemMetricSummary;
	    cpu_sys: SystemMetricSummary;
	    cpu_iowait: SystemMetricSummary;
	    cpu_steal: SystemMetricSummary;
	    disk_read_bps: SystemMetricSummary;
	    disk_write_bps: SystemMetricSummary;
	    disk_read_latency_ms: SystemMetricSummary;
	    disk_write_latency_ms: SystemMetricSummary;
	    system_enabled: boolean;
	    system_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new UnifiedMetrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tps = this.convertValues(source["tps"], MetricSummary);
	        this.tpm = this.convertValues(source["tpm"], MetricSummary);
	        this.cpu_user = this.convertValues(source["cpu_user"], SystemMetricSummary);
	        this.cpu_sys = this.convertValues(source["cpu_sys"], SystemMetricSummary);
	        this.cpu_iowait = this.convertValues(source["cpu_iowait"], SystemMetricSummary);
	        this.cpu_steal = this.convertValues(source["cpu_steal"], SystemMetricSummary);
	        this.disk_read_bps = this.convertValues(source["disk_read_bps"], SystemMetricSummary);
	        this.disk_write_bps = this.convertValues(source["disk_write_bps"], SystemMetricSummary);
	        this.disk_read_latency_ms = this.convertValues(source["disk_read_latency_ms"], SystemMetricSummary);
	        this.disk_write_latency_ms = this.convertValues(source["disk_write_latency_ms"], SystemMetricSummary);
	        this.system_enabled = source["system_enabled"];
	        this.system_message = source["system_message"];
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
	export class TaskTiming {
	    prepare_ms: number;
	    run_ms: number;
	    cleanup_ms: number;
	    total_ms: number;
	    run_duration_input_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new TaskTiming(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prepare_ms = source["prepare_ms"];
	        this.run_ms = source["run_ms"];
	        this.cleanup_ms = source["cleanup_ms"];
	        this.total_ms = source["total_ms"];
	        this.run_duration_input_ms = source["run_duration_input_ms"];
	    }
	}
	export class PhaseRecord {
	    phase: string;
	    status: string;
	    run_id?: string;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    ended_at?: any;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new PhaseRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.phase = source["phase"];
	        this.status = source["status"];
	        this.run_id = source["run_id"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.ended_at = this.convertValues(source["ended_at"], null);
	        this.message = source["message"];
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
	export class Readiness {
	    template_selected: boolean;
	    connection_selected: boolean;
	    action_supported: boolean;
	    runtime_valid: boolean;
	    db_valid: boolean;
	    db_message?: string;
	    ssh_available: boolean;
	    ssh_checked: boolean;
	    ssh_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new Readiness(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.template_selected = source["template_selected"];
	        this.connection_selected = source["connection_selected"];
	        this.action_supported = source["action_supported"];
	        this.runtime_valid = source["runtime_valid"];
	        this.db_valid = source["db_valid"];
	        this.db_message = source["db_message"];
	        this.ssh_available = source["ssh_available"];
	        this.ssh_checked = source["ssh_checked"];
	        this.ssh_message = source["ssh_message"];
	    }
	}
	export class TemplateSnapshot {
	    id: string;
	    name: string;
	    tool: string;
	    db_family: string;
	    workload_family: string;
	    phases: Record<string, boolean>;
	    parameters: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new TemplateSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.tool = source["tool"];
	        this.db_family = source["db_family"];
	        this.workload_family = source["workload_family"];
	        this.phases = source["phases"];
	        this.parameters = source["parameters"];
	    }
	}
	export class ExecutionTask {
	    id: string;
	    preview_token?: string;
	    name: string;
	    action: string;
	    status: string;
	    current_phase: string;
	    template_snapshot: TemplateSnapshot;
	    connection_snapshot: ConnectionSnapshot;
	    resolved_params: Record<string, any>;
	    readiness: Readiness;
	    phase_history: PhaseRecord[];
	    timing: TaskTiming;
	    metrics: UnifiedMetrics;
	    log_tail: LogLine[];
	    benchmark_tool: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    started_at?: any;
	    // Go type: time
	    completed_at?: any;
	    error_message?: string;
	    run_log_paths?: Record<string, string>;
	    system_log_paths?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionTask(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.preview_token = source["preview_token"];
	        this.name = source["name"];
	        this.action = source["action"];
	        this.status = source["status"];
	        this.current_phase = source["current_phase"];
	        this.template_snapshot = this.convertValues(source["template_snapshot"], TemplateSnapshot);
	        this.connection_snapshot = this.convertValues(source["connection_snapshot"], ConnectionSnapshot);
	        this.resolved_params = source["resolved_params"];
	        this.readiness = this.convertValues(source["readiness"], Readiness);
	        this.phase_history = this.convertValues(source["phase_history"], PhaseRecord);
	        this.timing = this.convertValues(source["timing"], TaskTiming);
	        this.metrics = this.convertValues(source["metrics"], UnifiedMetrics);
	        this.log_tail = this.convertValues(source["log_tail"], LogLine);
	        this.benchmark_tool = source["benchmark_tool"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
	        this.error_message = source["error_message"];
	        this.run_log_paths = source["run_log_paths"];
	        this.system_log_paths = source["system_log_paths"];
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
	
	
	
	
	
	
	
	

}

export namespace template {
	
	export class Compatibility {
	    supportedDatabases?: string[];
	    supportedVersions?: string[];
	    compatibilityNotes?: string;
	    requiresPrivileges?: string[];
	    constraints?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Compatibility(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supportedDatabases = source["supportedDatabases"];
	        this.supportedVersions = source["supportedVersions"];
	        this.compatibilityNotes = source["compatibilityNotes"];
	        this.requiresPrivileges = source["requiresPrivileges"];
	        this.constraints = source["constraints"];
	    }
	}
	export class Concurrency {
	    mode: string;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new Concurrency(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.value = source["value"];
	    }
	}
	export class HammerDBConfig {
	    benchmark?: string;
	    virtualUsers?: number;
	    warehouses?: number;
	    scaleFactor?: number;
	    timeProfile: boolean;
	    stepTesting: boolean;
	    xmlConnectPool: boolean;
	    advancedNotes?: string;
	
	    static createFrom(source: any = {}) {
	        return new HammerDBConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.benchmark = source["benchmark"];
	        this.virtualUsers = source["virtualUsers"];
	        this.warehouses = source["warehouses"];
	        this.scaleFactor = source["scaleFactor"];
	        this.timeProfile = source["timeProfile"];
	        this.stepTesting = source["stepTesting"];
	        this.xmlConnectPool = source["xmlConnectPool"];
	        this.advancedNotes = source["advancedNotes"];
	    }
	}
	export class PhaseConfig {
	    enabled: boolean;
	    required: boolean;
	    params?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new PhaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.required = source["required"];
	        this.params = source["params"];
	    }
	}
	export class PhaseSet {
	    prepare: PhaseConfig;
	    warmup: PhaseConfig;
	    run: PhaseConfig;
	    cleanup: PhaseConfig;
	
	    static createFrom(source: any = {}) {
	        return new PhaseSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prepare = this.convertValues(source["prepare"], PhaseConfig);
	        this.warmup = this.convertValues(source["warmup"], PhaseConfig);
	        this.run = this.convertValues(source["run"], PhaseConfig);
	        this.cleanup = this.convertValues(source["cleanup"], PhaseConfig);
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
	export class Runtime {
	    concurrency: Concurrency;
	    durationSeconds: number;
	    warmupSeconds: number;
	    rampUpSeconds: number;
	    reportIntervalSeconds: number;
	    percentile: number;
	    iterations: number;
	    rateLimit: number;
	    validationEnabled: boolean;
	    notes?: string;
	
	    static createFrom(source: any = {}) {
	        return new Runtime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.concurrency = this.convertValues(source["concurrency"], Concurrency);
	        this.durationSeconds = source["durationSeconds"];
	        this.warmupSeconds = source["warmupSeconds"];
	        this.rampUpSeconds = source["rampUpSeconds"];
	        this.reportIntervalSeconds = source["reportIntervalSeconds"];
	        this.percentile = source["percentile"];
	        this.iterations = source["iterations"];
	        this.rateLimit = source["rateLimit"];
	        this.validationEnabled = source["validationEnabled"];
	        this.notes = source["notes"];
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
	export class SwingbenchConfig {
	    benchmark?: string;
	    frontend?: string;
	    configMode?: string;
	    wizardOperation?: string;
	    userCount?: number;
	    runTimeSeconds?: number;
	    minThinkTime?: number;
	    maxThinkTime?: number;
	    xmlOverrides?: string;
	
	    static createFrom(source: any = {}) {
	        return new SwingbenchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.benchmark = source["benchmark"];
	        this.frontend = source["frontend"];
	        this.configMode = source["configMode"];
	        this.wizardOperation = source["wizardOperation"];
	        this.userCount = source["userCount"];
	        this.runTimeSeconds = source["runTimeSeconds"];
	        this.minThinkTime = source["minThinkTime"];
	        this.maxThinkTime = source["maxThinkTime"];
	        this.xmlOverrides = source["xmlOverrides"];
	    }
	}
	export class SysbenchConfig {
	    dbDriver?: string;
	    scriptType?: string;
	    tables?: number;
	    tableSize?: number;
	    reportChecks: boolean;
	    extraCliArgs?: string;
	
	    static createFrom(source: any = {}) {
	        return new SysbenchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dbDriver = source["dbDriver"];
	        this.scriptType = source["scriptType"];
	        this.tables = source["tables"];
	        this.tableSize = source["tableSize"];
	        this.reportChecks = source["reportChecks"];
	        this.extraCliArgs = source["extraCliArgs"];
	    }
	}
	export class ToolConfig {
	    sysbench: SysbenchConfig;
	    swingbench: SwingbenchConfig;
	    hammerdb: HammerDBConfig;
	
	    static createFrom(source: any = {}) {
	        return new ToolConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sysbench = this.convertValues(source["sysbench"], SysbenchConfig);
	        this.swingbench = this.convertValues(source["swingbench"], SwingbenchConfig);
	        this.hammerdb = this.convertValues(source["hammerdb"], HammerDBConfig);
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

}

export namespace usecase {
	
	export class TemplateMetadata {
	    id: string;
	    name: string;
	    description: string;
	    tool: string;
	    is_builtin: boolean;
	    param_count: number;
	    database_types: string[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tool = source["tool"];
	        this.is_builtin = source["is_builtin"];
	        this.param_count = source["param_count"];
	        this.database_types = source["database_types"];
	    }
	}

}

