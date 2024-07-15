export namespace backend {
	
	export class Crawl {
	
	
	    static createFrom(source: any = {}) {
	        return new Crawl(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

