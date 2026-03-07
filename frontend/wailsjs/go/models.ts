export namespace data {
	
	export class AIConfig {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    name: string;
	    baseUrl: string;
	    apiKey: string;
	    modelName: string;
	    maxTokens: number;
	    temperature: number;
	    timeOut: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.name = source["name"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.modelName = source["modelName"];
	        this.maxTokens = source["maxTokens"];
	        this.temperature = source["temperature"];
	        this.timeOut = source["timeOut"];
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
	export class FundBasic {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    code: string;
	    name: string;
	    fullName: string;
	    type: string;
	    establishment: string;
	    scale: string;
	    company: string;
	    manager: string;
	    rating: string;
	    trackingTarget: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedUnitTime: string;
	    netAccumulated?: number;
	    netGrowth1?: number;
	    netGrowth3?: number;
	    netGrowth6?: number;
	    netGrowth12?: number;
	    netGrowth36?: number;
	    netGrowth60?: number;
	    netGrowthYTD?: number;
	    netGrowthAll?: number;
	
	    static createFrom(source: any = {}) {
	        return new FundBasic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.fullName = source["fullName"];
	        this.type = source["type"];
	        this.establishment = source["establishment"];
	        this.scale = source["scale"];
	        this.company = source["company"];
	        this.manager = source["manager"];
	        this.rating = source["rating"];
	        this.trackingTarget = source["trackingTarget"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedUnitTime = source["netEstimatedUnitTime"];
	        this.netAccumulated = source["netAccumulated"];
	        this.netGrowth1 = source["netGrowth1"];
	        this.netGrowth3 = source["netGrowth3"];
	        this.netGrowth6 = source["netGrowth6"];
	        this.netGrowth12 = source["netGrowth12"];
	        this.netGrowth36 = source["netGrowth36"];
	        this.netGrowth60 = source["netGrowth60"];
	        this.netGrowthYTD = source["netGrowthYTD"];
	        this.netGrowthAll = source["netGrowthAll"];
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
	export class FollowedFund {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    code: string;
	    name: string;
	    netUnitValue?: number;
	    netUnitValueDate: string;
	    netEstimatedUnit?: number;
	    netEstimatedUnitTime: string;
	    netAccumulated?: number;
	    netEstimatedRate?: number;
	    fundBasic: FundBasic;
	
	    static createFrom(source: any = {}) {
	        return new FollowedFund(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.netUnitValue = source["netUnitValue"];
	        this.netUnitValueDate = source["netUnitValueDate"];
	        this.netEstimatedUnit = source["netEstimatedUnit"];
	        this.netEstimatedUnitTime = source["netEstimatedUnitTime"];
	        this.netAccumulated = source["netAccumulated"];
	        this.netEstimatedRate = source["netEstimatedRate"];
	        this.fundBasic = this.convertValues(source["fundBasic"], FundBasic);
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
	export class Group {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    name: string;
	    sort: number;
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.name = source["name"];
	        this.sort = source["sort"];
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
	export class GroupStock {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    stockCode: string;
	    groupId: number;
	    groupInfo: Group;
	
	    static createFrom(source: any = {}) {
	        return new GroupStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.stockCode = source["stockCode"];
	        this.groupId = source["groupId"];
	        this.groupInfo = this.convertValues(source["groupInfo"], Group);
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
	export class FollowedStock {
	    StockCode: string;
	    Name: string;
	    Volume: number;
	    CostPrice: number;
	    Price: number;
	    PriceChange: number;
	    ChangePercent: number;
	    AlarmChangePercent: number;
	    AlarmPrice: number;
	    // Go type: time
	    Time: any;
	    Sort: number;
	    Cron?: string;
	    IsDel: number;
	    Groups: GroupStock[];
	    AiConfigId: number;
	
	    static createFrom(source: any = {}) {
	        return new FollowedStock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StockCode = source["StockCode"];
	        this.Name = source["Name"];
	        this.Volume = source["Volume"];
	        this.CostPrice = source["CostPrice"];
	        this.Price = source["Price"];
	        this.PriceChange = source["PriceChange"];
	        this.ChangePercent = source["ChangePercent"];
	        this.AlarmChangePercent = source["AlarmChangePercent"];
	        this.AlarmPrice = source["AlarmPrice"];
	        this.Time = this.convertValues(source["Time"], null);
	        this.Sort = source["Sort"];
	        this.Cron = source["Cron"];
	        this.IsDel = source["IsDel"];
	        this.Groups = this.convertValues(source["Groups"], GroupStock);
	        this.AiConfigId = source["AiConfigId"];
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
	
	
	
	export class KLineData {
	    day: string;
	    open: string;
	    high: string;
	    low: string;
	    close: string;
	    volume: string;
	
	    static createFrom(source: any = {}) {
	        return new KLineData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.open = source["open"];
	        this.high = source["high"];
	        this.low = source["low"];
	        this.close = source["close"];
	        this.volume = source["volume"];
	    }
	}
	export class SentimentResult {
	    Score: number;
	    Category: number;
	    PositiveCount: number;
	    NegativeCount: number;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new SentimentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Score = source["Score"];
	        this.Category = source["Category"];
	        this.PositiveCount = source["PositiveCount"];
	        this.NegativeCount = source["NegativeCount"];
	        this.Description = source["Description"];
	    }
	}
	export class SettingConfig {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    tushareToken: string;
	    localPushEnable: boolean;
	    dingPushEnable: boolean;
	    dingRobot: string;
	    updateBasicInfoOnStart: boolean;
	    refreshInterval: number;
	    openAiEnable: boolean;
	    prompt: string;
	    checkUpdate: boolean;
	    questionTemplate: string;
	    crawlTimeOut: number;
	    kDays: number;
	    enableDanmu: boolean;
	    browserPath: string;
	    enableNews: boolean;
	    darkTheme: boolean;
	    browserPoolSize: number;
	    enableFund: boolean;
	    enablePushNews: boolean;
	    enableOnlyPushRedNews: boolean;
	    sponsorCode: string;
	    httpProxy: string;
	    httpProxyEnabled: boolean;
	    aiConfigs: AIConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SettingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.tushareToken = source["tushareToken"];
	        this.localPushEnable = source["localPushEnable"];
	        this.dingPushEnable = source["dingPushEnable"];
	        this.dingRobot = source["dingRobot"];
	        this.updateBasicInfoOnStart = source["updateBasicInfoOnStart"];
	        this.refreshInterval = source["refreshInterval"];
	        this.openAiEnable = source["openAiEnable"];
	        this.prompt = source["prompt"];
	        this.checkUpdate = source["checkUpdate"];
	        this.questionTemplate = source["questionTemplate"];
	        this.crawlTimeOut = source["crawlTimeOut"];
	        this.kDays = source["kDays"];
	        this.enableDanmu = source["enableDanmu"];
	        this.browserPath = source["browserPath"];
	        this.enableNews = source["enableNews"];
	        this.darkTheme = source["darkTheme"];
	        this.browserPoolSize = source["browserPoolSize"];
	        this.enableFund = source["enableFund"];
	        this.enablePushNews = source["enablePushNews"];
	        this.enableOnlyPushRedNews = source["enableOnlyPushRedNews"];
	        this.sponsorCode = source["sponsorCode"];
	        this.httpProxy = source["httpProxy"];
	        this.httpProxyEnabled = source["httpProxyEnabled"];
	        this.aiConfigs = this.convertValues(source["aiConfigs"], AIConfig);
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
	export class StockBasic {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    ts_code: string;
	    symbol: string;
	    name: string;
	    area: string;
	    industry: string;
	    fullname: string;
	    enname: string;
	    cnspell: string;
	    market: string;
	    exchange: string;
	    curr_type: string;
	    list_status: string;
	    list_date: string;
	    delist_date: string;
	    is_hs: string;
	    act_name: string;
	    act_ent_type: string;
	    bk_name: string;
	    bk_code: string;
	
	    static createFrom(source: any = {}) {
	        return new StockBasic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.ts_code = source["ts_code"];
	        this.symbol = source["symbol"];
	        this.name = source["name"];
	        this.area = source["area"];
	        this.industry = source["industry"];
	        this.fullname = source["fullname"];
	        this.enname = source["enname"];
	        this.cnspell = source["cnspell"];
	        this.market = source["market"];
	        this.exchange = source["exchange"];
	        this.curr_type = source["curr_type"];
	        this.list_status = source["list_status"];
	        this.list_date = source["list_date"];
	        this.delist_date = source["delist_date"];
	        this.is_hs = source["is_hs"];
	        this.act_name = source["act_name"];
	        this.act_ent_type = source["act_ent_type"];
	        this.bk_name = source["bk_name"];
	        this.bk_code = source["bk_code"];
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
	export class StockInfo {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    "日期": string;
	    "时间": string;
	    "股票代码": string;
	    "股票名称": string;
	    "上次当前价格": number;
	    "当前价格": string;
	    "成交的股票数": string;
	    "成交金额": string;
	    "今日开盘价": string;
	    "昨日收盘价": string;
	    "今日最高价": string;
	    "今日最低价": string;
	    "竞买价": string;
	    "竞卖价": string;
	    "买一报价": string;
	    "买一申报": string;
	    "买二报价": string;
	    "买二申报": string;
	    "买三报价": string;
	    "买三申报": string;
	    "买四报价": string;
	    "买四申报": string;
	    "买五报价": string;
	    "买五申报": string;
	    "卖一报价": string;
	    "卖一申报": string;
	    "卖二报价": string;
	    "卖二申报": string;
	    "卖三报价": string;
	    "卖三申报": string;
	    "卖四报价": string;
	    "卖四申报": string;
	    "卖五报价": string;
	    "卖五申报": string;
	    "市场": string;
	    "盘前盘后": string;
	    "盘前盘后涨跌幅": string;
	    changePercent: number;
	    changePrice: number;
	    highRate: number;
	    lowRate: number;
	    costPrice: number;
	    costVolume: number;
	    profit: number;
	    profitAmount: number;
	    profitAmountToday: number;
	    sort: number;
	    alarmChangePercent: number;
	    alarmPrice: number;
	    Groups: GroupStock[];
	
	    static createFrom(source: any = {}) {
	        return new StockInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this["日期"] = source["日期"];
	        this["时间"] = source["时间"];
	        this["股票代码"] = source["股票代码"];
	        this["股票名称"] = source["股票名称"];
	        this["上次当前价格"] = source["上次当前价格"];
	        this["当前价格"] = source["当前价格"];
	        this["成交的股票数"] = source["成交的股票数"];
	        this["成交金额"] = source["成交金额"];
	        this["今日开盘价"] = source["今日开盘价"];
	        this["昨日收盘价"] = source["昨日收盘价"];
	        this["今日最高价"] = source["今日最高价"];
	        this["今日最低价"] = source["今日最低价"];
	        this["竞买价"] = source["竞买价"];
	        this["竞卖价"] = source["竞卖价"];
	        this["买一报价"] = source["买一报价"];
	        this["买一申报"] = source["买一申报"];
	        this["买二报价"] = source["买二报价"];
	        this["买二申报"] = source["买二申报"];
	        this["买三报价"] = source["买三报价"];
	        this["买三申报"] = source["买三申报"];
	        this["买四报价"] = source["买四报价"];
	        this["买四申报"] = source["买四申报"];
	        this["买五报价"] = source["买五报价"];
	        this["买五申报"] = source["买五申报"];
	        this["卖一报价"] = source["卖一报价"];
	        this["卖一申报"] = source["卖一申报"];
	        this["卖二报价"] = source["卖二报价"];
	        this["卖二申报"] = source["卖二申报"];
	        this["卖三报价"] = source["卖三报价"];
	        this["卖三申报"] = source["卖三申报"];
	        this["卖四报价"] = source["卖四报价"];
	        this["卖四申报"] = source["卖四申报"];
	        this["卖五报价"] = source["卖五报价"];
	        this["卖五申报"] = source["卖五申报"];
	        this["市场"] = source["市场"];
	        this["盘前盘后"] = source["盘前盘后"];
	        this["盘前盘后涨跌幅"] = source["盘前盘后涨跌幅"];
	        this.changePercent = source["changePercent"];
	        this.changePrice = source["changePrice"];
	        this.highRate = source["highRate"];
	        this.lowRate = source["lowRate"];
	        this.costPrice = source["costPrice"];
	        this.costVolume = source["costVolume"];
	        this.profit = source["profit"];
	        this.profitAmount = source["profitAmount"];
	        this.profitAmountToday = source["profitAmountToday"];
	        this.sort = source["sort"];
	        this.alarmChangePercent = source["alarmChangePercent"];
	        this.alarmPrice = source["alarmPrice"];
	        this.Groups = this.convertValues(source["Groups"], GroupStock);
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
	export class ToolItem {
	    name: string;
	    type: string;
	    enabled: boolean;
	    config: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new ToolItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.enabled = source["enabled"];
	        this.config = source["config"];
	    }
	}
	export class ToolConfig {
	    version: string;
	    tools: ToolItem[];
	
	    static createFrom(source: any = {}) {
	        return new ToolConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.tools = this.convertValues(source["tools"], ToolItem);
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

export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

export namespace models {
	
	export class AIResponseResult {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    chatId: string;
	    modelName: string;
	    stockCode: string;
	    stockName: string;
	    question: string;
	    content: string;
	    IsDel: number;
	
	    static createFrom(source: any = {}) {
	        return new AIResponseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.chatId = source["chatId"];
	        this.modelName = source["modelName"];
	        this.stockCode = source["stockCode"];
	        this.stockName = source["stockName"];
	        this.question = source["question"];
	        this.content = source["content"];
	        this.IsDel = source["IsDel"];
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
	export class Position {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    deleted_at: gorm.DeletedAt;
	    stock_code: string;
	    stock_name: string;
	    quantity: number;
	    buy_price: number;
	    // Go type: time
	    buy_date: any;
	    current_price: number;
	    profit_loss: number;
	    profit_loss_pct: number;
	    market_value: number;
	    notes: string;
	    is_active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted_at = this.convertValues(source["deleted_at"], gorm.DeletedAt);
	        this.stock_code = source["stock_code"];
	        this.stock_name = source["stock_name"];
	        this.quantity = source["quantity"];
	        this.buy_price = source["buy_price"];
	        this.buy_date = this.convertValues(source["buy_date"], null);
	        this.current_price = source["current_price"];
	        this.profit_loss = source["profit_loss"];
	        this.profit_loss_pct = source["profit_loss_pct"];
	        this.market_value = source["market_value"];
	        this.notes = source["notes"];
	        this.is_active = source["is_active"];
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
	export class PositionAnalysis {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    deleted_at: gorm.DeletedAt;
	    position_id: number;
	    overall_advice: string;
	    confidence: number;
	    suggested_buy_price?: number;
	    suggested_sell_price?: number;
	    stop_loss_price?: number;
	    technical_analysis: string;
	    fundamental_analysis: string;
	    risk_analysis: string;
	    raw_response: string;
	
	    static createFrom(source: any = {}) {
	        return new PositionAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted_at = this.convertValues(source["deleted_at"], gorm.DeletedAt);
	        this.position_id = source["position_id"];
	        this.overall_advice = source["overall_advice"];
	        this.confidence = source["confidence"];
	        this.suggested_buy_price = source["suggested_buy_price"];
	        this.suggested_sell_price = source["suggested_sell_price"];
	        this.stop_loss_price = source["stop_loss_price"];
	        this.technical_analysis = source["technical_analysis"];
	        this.fundamental_analysis = source["fundamental_analysis"];
	        this.risk_analysis = source["risk_analysis"];
	        this.raw_response = source["raw_response"];
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
	export class Prompt {
	    ID: number;
	    name: string;
	    content: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Prompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.name = source["name"];
	        this.content = source["content"];
	        this.type = source["type"];
	    }
	}
	export class RecommendationItem {
	    rank: number;
	    stock_code: string;
	    stock_name: string;
	    current_price: number;
	    price_change: number;
	    volume: number;
	    market_value: number;
	    technical_analysis: string;
	    fundamental_analysis: string;
	    reason: string;
	    target_price: number;
	    target_change_percent: number;
	    risk_level: string;
	    risk_tips: string;
	    score: number;
	    trade_suggestion: string;
	    recommended_price: number;
	    previous_close: number;
	    buy_price_range: string;
	    stop_loss_price: number;
	    sector_concept: string;
	    remarks: string;
	    // Go type: time
	    recommended_at: any;
	    is_followed: boolean;
	    macd: string;
	    kdj: string;
	    rsi: number;
	    trend: string;
	    pe: number;
	    pb: number;
	    roe: number;
	    revenue_growth: number;
	    profit_growth: number;
	
	    static createFrom(source: any = {}) {
	        return new RecommendationItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.stock_code = source["stock_code"];
	        this.stock_name = source["stock_name"];
	        this.current_price = source["current_price"];
	        this.price_change = source["price_change"];
	        this.volume = source["volume"];
	        this.market_value = source["market_value"];
	        this.technical_analysis = source["technical_analysis"];
	        this.fundamental_analysis = source["fundamental_analysis"];
	        this.reason = source["reason"];
	        this.target_price = source["target_price"];
	        this.target_change_percent = source["target_change_percent"];
	        this.risk_level = source["risk_level"];
	        this.risk_tips = source["risk_tips"];
	        this.score = source["score"];
	        this.trade_suggestion = source["trade_suggestion"];
	        this.recommended_price = source["recommended_price"];
	        this.previous_close = source["previous_close"];
	        this.buy_price_range = source["buy_price_range"];
	        this.stop_loss_price = source["stop_loss_price"];
	        this.sector_concept = source["sector_concept"];
	        this.remarks = source["remarks"];
	        this.recommended_at = this.convertValues(source["recommended_at"], null);
	        this.is_followed = source["is_followed"];
	        this.macd = source["macd"];
	        this.kdj = source["kdj"];
	        this.rsi = source["rsi"];
	        this.trend = source["trend"];
	        this.pe = source["pe"];
	        this.pb = source["pb"];
	        this.roe = source["roe"];
	        this.revenue_growth = source["revenue_growth"];
	        this.profit_growth = source["profit_growth"];
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
	export class StockPickReport {
	    id: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    deleted_at: gorm.DeletedAt;
	    user_query: string;
	    query_summary: string;
	    result: string;
	    market_analysis: string;
	    filter_logic: string;
	    total_scanned: number;
	    candidates_count: number;
	    recommendations: RecommendationItem[];
	    tools_used: string;
	    ai_config_id: number;
	    ai_model: string;
	    status: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new StockPickReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted_at = this.convertValues(source["deleted_at"], gorm.DeletedAt);
	        this.user_query = source["user_query"];
	        this.query_summary = source["query_summary"];
	        this.result = source["result"];
	        this.market_analysis = source["market_analysis"];
	        this.filter_logic = source["filter_logic"];
	        this.total_scanned = source["total_scanned"];
	        this.candidates_count = source["candidates_count"];
	        this.recommendations = this.convertValues(source["recommendations"], RecommendationItem);
	        this.tools_used = source["tools_used"];
	        this.ai_config_id = source["ai_config_id"];
	        this.ai_model = source["ai_model"];
	        this.status = source["status"];
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
	export class StockPickReportItem {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    QuerySummary: string;
	    RecommendCount: number;
	    Status: string;
	
	    static createFrom(source: any = {}) {
	        return new StockPickReportItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.QuerySummary = source["QuerySummary"];
	        this.RecommendCount = source["RecommendCount"];
	        this.Status = source["Status"];
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
	export class StockPickReportsResponse {
	    items: StockPickReportItem[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new StockPickReportsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], StockPickReportItem);
	        this.total = source["total"];
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
	export class VersionInfo {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    version: string;
	    content: string;
	    icon: string;
	    alipay: string;
	    wxpay: string;
	    wxgzh: string;
	    buildTimeStamp: number;
	    officialStatement: string;
	    IsDel: number;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.version = source["version"];
	        this.content = source["content"];
	        this.icon = source["icon"];
	        this.alipay = source["alipay"];
	        this.wxpay = source["wxpay"];
	        this.wxgzh = source["wxgzh"];
	        this.buildTimeStamp = source["buildTimeStamp"];
	        this.officialStatement = source["officialStatement"];
	        this.IsDel = source["IsDel"];
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

