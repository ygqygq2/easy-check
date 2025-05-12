export namespace constants {
	
	export class SharedConstants {
	    appName: string;
	    appVersion: string;
	    UpdateServer: string;
	
	    static createFrom(source: any = {}) {
	        return new SharedConstants(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.appVersion = source["appVersion"];
	        this.UpdateServer = source["UpdateServer"];
	    }
	}

}

