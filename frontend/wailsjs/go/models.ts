export namespace main {
	
	export class Species {
	    id: number;
	    common_name: string;
	    scientific_name: string;
	    kingdom: string;
	    phylum: string;
	    class: string;
	    order: string;
	    family: string;
	    last_synced: string;
	
	    static createFrom(source: any = {}) {
	        return new Species(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.common_name = source["common_name"];
	        this.scientific_name = source["scientific_name"];
	        this.kingdom = source["kingdom"];
	        this.phylum = source["phylum"];
	        this.class = source["class"];
	        this.order = source["order"];
	        this.family = source["family"];
	        this.last_synced = source["last_synced"];
	    }
	}

}

