export namespace bindings {
	
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
	    ssh_enabled: boolean;
	    ssh_port?: number;
	    ssh_username?: string;
	    ssh_password?: string;
	
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
	        this.ssh_enabled = source["ssh_enabled"];
	        this.ssh_port = source["ssh_port"];
	        this.ssh_username = source["ssh_username"];
	        this.ssh_password = source["ssh_password"];
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
	export class ConnectionTestResult {
	    success: boolean;
	    latency_ms: number;
	    database_version?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.latency_ms = source["latency_ms"];
	        this.database_version = source["database_version"];
	        this.error = source["error"];
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
	export class TemplateDTO {
	    id: string;
	    name: string;
	    description: string;
	    tool: string;
	    database_types: string[];
	    version: string;
	    parameters: Record<string, ParamDTO>;
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
	        this.database_types = source["database_types"];
	        this.version = source["version"];
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

