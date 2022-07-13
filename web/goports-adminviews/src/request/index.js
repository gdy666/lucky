import axios from 'axios'

console.log("vue run mode "+process.env.NODE_ENV)

var baseURL = "/" //
if (process.env.NODE_ENV=="development"){
	//开发环境下这个改为自己的接口地址
	baseURL = 'http://192.168.31.70:16601'
}


//var fuck = storage.getItem("cookies")
//console.log("fuck:"+fuck)

//console.log("baseURL: "+ baseURL)

// 创建一个 axios 实例
const service = axios.create({
	baseURL: baseURL, // 所有的请求地址前缀部分
	timeout: 5000, // 请求超时时间毫秒
	withCredentials: false, // 异步请求携带cookie
	headers: {
		// 设置后端需要的传参类型
		'Content-Type': 'application/json',
		//'X-Requested-With': 'XMLHttpRequest',
	},
})

// 添加请求拦截器
service.interceptors.request.use(
	function (config) {
		// 在发送请求之前做些什么
		return config
	},
	function (error) {
		// 对请求错误做些什么
		console.log(error)
		return Promise.reject(error)
	}
)

// 添加响应拦截器
service.interceptors.response.use(
	function (response) {
		//console.log(response)
		// 2xx 范围内的状态码都会触发该函数。
		// 对响应数据做点什么
		// dataAxios 是 axios 返回数据中的 data
		const dataAxios = response.data
		// 这个状态码是和后端约定的
		const code = dataAxios.reset

		//console.log("dataAxios data: "+JSON.stringify(dataAxios))
		//console.log("ret: "+dataAxios.ret)
		if (dataAxios.ret!=undefined&& dataAxios.ret==-1){
			//global.currentPage.value="set"
			console.log("登录失效")
			//window.location.href="/"
			location.hash ="#login"
			//var currentPage = ref("#login")

		}
		return dataAxios
	},
	function (error) {
		// 超出 2xx 范围的状态码都会触发该函数。
		// 对响应错误做点什么
		console.log(error)
		return Promise.reject(error)
	}
)

export default service
