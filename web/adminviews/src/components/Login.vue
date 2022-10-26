<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <div class="formradius" :style="{
            borderRadius: 'base',
        }">

            <el-form :model="form" class="SetForm" label-width="auto">

                <el-form-item label="管理账号" id="account">
                    <el-input v-model="form.Account" placeholder="管理账号" autocomplete="off" style="witdh:390px;" />
                </el-form-item>

                <el-form-item label="管理密码" id="account">
                    <el-input show-password v-model="form.Password" placeholder="管理密码" autocomplete="off"
                        style="witdh:390px;" />
                </el-form-item>


                <el-form-item>
                    <el-checkbox v-model="rememberPasswordChecked" label="记住密码" size="large" @change="rememberPasswordCheckedChange" />
                </el-form-item>


            </el-form>

            <el-button type="primary" round @click="Login">登录</el-button>

        </div>


    </div>

</template>

<script lang="ts" setup>

import { onMounted, onUnmounted, ref, inject } from 'vue'
import { apiLogin } from '../apis/utils'
import {MessageShow} from '../utils/ui'


const rememberPasswordChecked = ref(true)

const global: any = inject("global")

const form = ref({
    Account: "",
    Password: ""
})



const rememberPasswordCheckedChange= (checked )=>{
    //console.log("记住密码:"+checked)
    SaveLoginInfo(checked)
}


const SaveLoginInfo = (checked) =>{
     global.storage.setItem("rememberPassword",checked)
    if (checked){
        global.storage.setItem("loginAccount",form.value.Account)
        global.storage.setItem("loginPassword",form.value.Password)
        return 
    }
    global.storage.setItem("loginAccount","")
    global.storage.setItem("loginPassword","")
}

const ReadLoginInfo = ()=>{
    let rememberPassword = global.storage.getItem("rememberPassword")
    rememberPasswordChecked.value = rememberPassword == undefined || rememberPassword == false ? false : true;
    if(!rememberPassword){
        return 
    }
   form.value.Account= global.storage.getItem("loginAccount")==undefined?"":global.storage.getItem("loginAccount")
   form.value.Password = global.storage.getItem("loginPassword")==undefined?"":global.storage.getItem("loginPassword")
}


const Login = () => {
    if (form.value.Account == "" || form.value.Password == "") {
        MessageShow("error", "账号或密码不能为空")
        return
    }

    SaveLoginInfo(rememberPasswordChecked.value)
    apiLogin(form.value).then((res) => {
        if (res.ret == 0) {
            MessageShow("success", "登录成功")
            global.storage.setItem("token",res.token)
            //global.currentPage.value = "#set"
            location.hash="#set"
            //console.log("cookies:"+res.cookies)
            
            return
        }
        MessageShow("error", res.msg)
    }).catch((error) => {
        console.log("登录失败,网络请求出错:" + error)
        MessageShow("error", "登录失败,网络请求出错")
    })

}



const keydown = (e) => {
    if (e.keyCode == 13 && global.currentPage.value=="#login") {
        Login()
        return 
    }
    return 
}

onMounted(() => {
    window.addEventListener('keydown', keydown)


    ReadLoginInfo()
})

</script>

<style scoped>
.formradius {
    border: 0px solid var(--el-border-color);
    border-radius: 0;
    margin: 0 auto;
    width: fit-content;
    padding: 15px;
}
</style>