// 导入axios实例
import httpRequest from '@/request/index'
import  storage  from './storage.js'

// 获取锚点
export function GetHash() {
    return location.hash
}

// 设置锚点
export function SetHash(hash) {
     location.hash = hash
}

export function GetToken(){
	//console.log("getTokenkkk: "+storage.getItem("token"))
	return storage.getItem("token")==undefined?"":storage.getItem("token");
}


// 获取状态信息
export function apiGetStatus() {
    return httpRequest({
		url: '/api/status',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()},
		
	})
}

export function apiGetLogs(pretimestamp) {
    return httpRequest({
		url: "/api/logs",
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{pre:pretimestamp,_:new Date().valueOf()}
	})
}


export function apiGetRuleList() {
    return httpRequest({
		url: '/api/rulelist',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()}
	})
}


export function apiAddRule(data) {
    return httpRequest({
		url: '/api/rule',
		method: 'post',
		headers:{'Authorization':GetToken()},
		data:data
	})
}

export function apiDeleteRule(configure) {
    return httpRequest({
		url: '/api/rule',
		method: 'delete',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),rule:configure}
	})
}

export function apiAlterRule(data) {
    return httpRequest({
		url: '/api/rule',
		method: 'put',
		headers:{'Authorization':GetToken()},
		data:data
	})
}

export function apiRuleEnable(key,enable) {
    return httpRequest({
		url: '/api/rule/enable',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),enable:enable,key:key}
	})
}

export function apiQueryBaseConfigure() {
    return httpRequest({
		url: '/api/baseconfigure',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()}
	})
}

export function apiAlterBaseConfigure(data) {
    return httpRequest({
		url: '/api/baseconfigure',
		method: 'put',
		headers:{'Authorization':GetToken()},
		data:data
	})
}

export function apiLogin(data) {
    return httpRequest({
		url: '/api/login',
		method: 'post',
		data:data
	})
}


export function apiRebootProgram() {
    return httpRequest({
		url: '/api/reboot_program',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()}
	})
}


export function apiAlterWhiteListConfigure(data) {
    return httpRequest({
		url: '/api/whitelist/configure',
		method: 'put',
		headers:{'Authorization':GetToken()},
		data:data
	})
}

export function apiGetWhiteListConfigure(data) {
    return httpRequest({
		url: '/api/whitelist/configure',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()}
	})
}

export function apiGetWhiteList(data) {
    return httpRequest({
		url: '/api/whitelist',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()},
		
	})
}


export function apiFlushWhiteList(ip,life) {
    return httpRequest({
		url: '/api/whitelist/flush',
		method: 'put',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),ip:ip,life:life}
	})
}

export function apiDeleteWhiteList(ip,life) {
    return httpRequest({
		url: '/api/whitelist',
		method: 'delete',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),ip:ip},
	})
}

export function apiGetBlackList(data) {
    return httpRequest({
		url: '/api/blacklist',
		method: 'get',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf()}
	})
}


export function apiFlushBlackList(ip,life) {
    return httpRequest({
		url: '/api/blacklist/flush',
		method: 'put',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),ip:ip,life:life}
	})
}

export function apiDeleteBlackList(ip,life) {
    return httpRequest({
		url: '/api/blacklist',
		method: 'delete',
		headers:{'Authorization':GetToken()},
		params:{_:new Date().valueOf(),ip:ip}
	})
}


export function apiGetVersion() {
    return httpRequest({
		url: '/version',
		method: 'get',
		params:{_:new Date().valueOf()}
	})
}

export function apiLogout() {
    return httpRequest({
		url: '/api/logout',
		method: 'put',
		headers:{'Authorization':GetToken()},
	})
}