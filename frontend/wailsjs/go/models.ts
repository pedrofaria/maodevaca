export namespace main {
	
	export class Account {
	    id: number;
	    groupId?: number;
	    groupName: string;
	    name: string;
	    amount: number;
	    dueDay: number;
	    active: boolean;
	    notes: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.groupId = source["groupId"];
	        this.groupName = source["groupName"];
	        this.name = source["name"];
	        this.amount = source["amount"];
	        this.dueDay = source["dueDay"];
	        this.active = source["active"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Group {
	    id: number;
	    name: string;
	    icon: string;
	    color: string;
	    sortOrder: number;
	    accountCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.sortOrder = source["sortOrder"];
	        this.accountCount = source["accountCount"];
	    }
	}
	export class Income {
	    id: number;
	    sourceId: number;
	    sourceName: string;
	    sourceIcon: string;
	    sourceColor: string;
	    amount: number;
	    date: string;
	    year: number;
	    month: number;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Income(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceId = source["sourceId"];
	        this.sourceName = source["sourceName"];
	        this.sourceIcon = source["sourceIcon"];
	        this.sourceColor = source["sourceColor"];
	        this.amount = source["amount"];
	        this.date = source["date"];
	        this.year = source["year"];
	        this.month = source["month"];
	        this.description = source["description"];
	    }
	}
	export class IncomeSource {
	    id: number;
	    name: string;
	    icon: string;
	    color: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new IncomeSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Payment {
	    id: number;
	    accountId: number;
	    accountName: string;
	    groupName: string;
	    amount: number;
	    paidOn: string;
	    year: number;
	    month: number;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new Payment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.accountId = source["accountId"];
	        this.accountName = source["accountName"];
	        this.groupName = source["groupName"];
	        this.amount = source["amount"];
	        this.paidOn = source["paidOn"];
	        this.year = source["year"];
	        this.month = source["month"];
	        this.notes = source["notes"];
	    }
	}
	export class MonthSummary {
	    year: number;
	    month: number;
	    incomesTotal: number;
	    expensesTotal: number;
	    balance: number;
	    incomes: Income[];
	    payments: Payment[];
	
	    static createFrom(source: any = {}) {
	        return new MonthSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.incomesTotal = source["incomesTotal"];
	        this.expensesTotal = source["expensesTotal"];
	        this.balance = source["balance"];
	        this.incomes = this.convertValues(source["incomes"], Income);
	        this.payments = this.convertValues(source["payments"], Payment);
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
	export class PayAccountInput {
	    accountId: number;
	    year: number;
	    month: number;
	    amount: number;
	    paidOn: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new PayAccountInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountId = source["accountId"];
	        this.year = source["year"];
	        this.month = source["month"];
	        this.amount = source["amount"];
	        this.paidOn = source["paidOn"];
	        this.notes = source["notes"];
	    }
	}
	
	export class YearRow {
	    year: number;
	    month: number;
	    label: string;
	    incomesTotal: number;
	    expensesTotal: number;
	    balance: number;
	
	    static createFrom(source: any = {}) {
	        return new YearRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.month = source["month"];
	        this.label = source["label"];
	        this.incomesTotal = source["incomesTotal"];
	        this.expensesTotal = source["expensesTotal"];
	        this.balance = source["balance"];
	    }
	}
	export class YearSummary {
	    year: number;
	    rows: YearRow[];
	    incomesTotal: number;
	    expensesTotal: number;
	    balance: number;
	
	    static createFrom(source: any = {}) {
	        return new YearSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.year = source["year"];
	        this.rows = this.convertValues(source["rows"], YearRow);
	        this.incomesTotal = source["incomesTotal"];
	        this.expensesTotal = source["expensesTotal"];
	        this.balance = source["balance"];
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

