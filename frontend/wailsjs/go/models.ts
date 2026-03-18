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
	        this.models = source["models"]?.map((item: any) => AIModelInfo.createFrom(item));
	        this.error = source["error"];
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

}
