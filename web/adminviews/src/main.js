import { createApp,ref } from 'vue'
import App from './App.vue'
import './assets/common-layout.scss'
import './assets/appbase.css'
import * as ElIcon from '@element-plus/icons-vue'
import 'element-plus/theme-chalk/el-notification.css'
import 'element-plus/theme-chalk/el-menu.css'
import 'element-plus/theme-chalk/el-loading.css'
import 'element-plus/theme-chalk/el-message.css'
import 'element-plus/theme-chalk/el-message-box.css'
import 'element-plus/theme-chalk/el-button.css'
import storage from './apis/storage.js'
import {apiLogout} from './apis/utils.js'
import {PageExist,CurrentPage} from './utils/utils'


const app = createApp(App)
for (let iconName in ElIcon){
    app.component(iconName, ElIcon[iconName])
}


app.config.globalProperties.$storage = storage;



if(!PageExist(location.hash)){
    location.hash="#status"
}

//配置全局变量
//默认页面
 var currentPage = ref(location.hash)



// if (process.env.NODE_ENV=="development"){
//     currentPage.value="#relayset"
//     location.hash="#relayset"
// }

app.provide('global',{
    currentPage,
    storage,
})

window.onpopstate = function (event){
    currentPage.value=location.hash
    CurrentPage.value = location.hash

    if(location.hash == "#logout"){//注销登录
        apiLogout().then((res) => {
        }).catch((error) => {
        })
        storage.setItem("token","")
        location.hash ="#login"
        return 
    }

    if(!PageExist(location.hash)){
        console.log("location.hash["+location.hash +"]no exist")
        location.hash ="#login"
        return 
    }
}





app.mount('#app')

