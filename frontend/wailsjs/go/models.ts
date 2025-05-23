export namespace constants {
	
	export class PlatformInfo {
	    os: string;
	    arch: string;
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	    }
	}
	export class SharedConstants {
	    appName: string;
	    appVersion: string;
	    platformInfo: PlatformInfo;
	    UpdateServer: string;
	    needsRestart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SharedConstants(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.appVersion = source["appVersion"];
	        this.platformInfo = this.convertValues(source["platformInfo"], PlatformInfo);
	        this.UpdateServer = source["UpdateServer"];
	        this.needsRestart = source["needsRestart"];
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

export namespace types {
	
	export class Host {
	    host: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Host(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.description = source["description"];
	    }
	}
	export class HostLatencyData {
	    host: string;
	    min_latency: number;
	    avg_latency: number;
	    max_latency: number;
	    packet_loss: number;
	
	    static createFrom(source: any = {}) {
	        return new HostLatencyData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.min_latency = source["min_latency"];
	        this.avg_latency = source["avg_latency"];
	        this.max_latency = source["max_latency"];
	        this.packet_loss = source["packet_loss"];
	    }
	}
	export class HostsLatencyResponse {
	    hosts: HostLatencyData[];
	    total: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new HostsLatencyResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hosts = this.convertValues(source["hosts"], HostLatencyData);
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
	export class HostsResponse {
	    hosts: Host[];
	    total: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new HostsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hosts = this.convertValues(source["hosts"], Host);
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

}

