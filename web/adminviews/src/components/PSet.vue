<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }">

                <el-form :model="form" class="SetForm" label-width="auto">
                    <el-form-item label="后台管理端口" id="adminListen">
                        <el-input-number v-model="form.AdminWebListenPort" autocomplete="off" />
                    </el-form-item>

                    <el-form-item label="外网访问" id="adminListen">
                        <el-switch v-model="form.AllowInternetaccess" class="mb-1" inline-prompt
                            style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949" width="50px"
                            active-text="允许" inactive-text="禁止" />
                    </el-form-item>

                    <el-form-item label="管理登录账号" id="adminAccount">
                        <el-input v-model="form.AdminAccount" placeholder="管理登录账号" autocomplete="off"
                            style="witdh:390px;" />
                    </el-form-item>

                    <el-form-item label="管理登录密码" id="adminPassword">
                        <el-input v-model="form.AdminPassword" placeholder="管理登录密码" autocomplete="off" />
                    </el-form-item>

                    <el-form-item label="日志记录最大条数" id="logMaxSize">
                        <el-input-number v-model="form.LogMaxSize" autocomplete="off" :min="1024" :max="40960" />
                    </el-form-item>


                    <el-form-item label="全局最大端口代理数量" id="proxyCountLimit">
                        <el-input-number v-model="form.ProxyCountLimit" autocomplete="off" :min="1" :max="1024" />
                    </el-form-item>

                    <el-form-item label="全局最大连接数" id="globalMaxConnections">
                        <el-input-number v-model="form.GlobalMaxConnections" autocomplete="off" :min="1" :max="65535" />
                    </el-form-item>



                </el-form>

                <el-button type="primary" round @click="RequestAlterConfigure">保存修改</el-button>
                <el-button type="info" round @click="resetFormData">撤销改动</el-button>
                <el-button type="danger" round @click="rebootProgram" :disabled="disableRebootButton">重启程序</el-button>
                <el-button type="success" round @click="backupConfigure">备份配置</el-button>

                <el-upload class="inline-block" :action="getRestoreConfigureAPI()" :show-file-list="false"
                    :headers="{ 'Authorization': GetToken() }" :limit="1" :on-success="callRestoreConfigureAPI">
                    <el-button round class='margin-change'>恢复配置</el-button>
                </el-upload>







            </div>


        </el-scrollbar>
    </div>

</template>


<script lang="ts" setup>

import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'
import { apiQueryBaseConfigure, apiAlterBaseConfigure, apiRebootProgram, apiGetConfigure, GetToken, apiGetRestoreConfigureConfirm } from '../apis/utils'
import { ElMessageBox } from 'element-plus'

import { MessageShow } from '../utils/ui'
import FileSaver from 'file-saver'
import { anyTypeAnnotation } from '@babel/types'

const formLabelWidth = '10vw'
console.log("window.location.href " + window.location.href)
console.log("window.location.port " + window.location.port)
console.log("window.location.host " + window.location.host)
console.log("window.location " + JSON.stringify(window.location))
const disableRebootButton = ref(false)

const getAdminURL = () => {
    return window.location.protocol + "//" + window.location.hostname + ":" + preFormData.value.AdminWebListenPort
}

const getRestoreConfigureAPI = () => {
    var baseURL = "/" //
    if (process.env.NODE_ENV == "development") {
        //开发环境下这个改为自己的接口地址
        baseURL = 'http://192.168.31.70:16601/'
    }
    return baseURL + "api/configure"

}

const callRestoreConfigureAPI = (res: any, uploadFile: any, uploadFiles: any) => {
    //  console.log("response ret"+res.ret +" msg:"+res.msg)
    if (res.ret != 0) {
        MessageShow("error", res.msg)
        return
    }
    console.log("restoreKey: " + res.restoreConfigureKey)

    let fileName  = res.file

    ElMessageBox.confirm(
        "确认要将[" + fileName + "]替换为大吉现有配置?替换完成后大吉会自动重启",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            apiGetRestoreConfigureConfirm(res.restoreConfigureKey).then(res => {
                if (res.ret != 0) {
                    MessageShow("error", "将[" + fileName + "]替换为大吉现有配置出错:" + res.msg)
                    return
                }

                MessageShow("success", "将[" + fileName + "]替换为大吉现有配置成功")

                setTimeout(() => {
                    window.location.href = window.location.protocol + "//" + window.location.hostname + ":" + res.port;
                }, 2000)

            }).catch((error) => {
                console.log("网络出错:" + error)
                MessageShow("error", "将[" + res.file + "]替换为大吉现有配置出错:" + error)
            })

        })
        .catch(() => {

        })
}

const rawData = {
    AdminWebListenPort: 1,
    AdminAccount: "",
    AdminPassword: "",
    ProxyCountLimit: 1,
    GlobalMaxConnections: 1,
    AllowInternetaccess: false,
    LogMaxSize:1024,
}

const form = ref(rawData)
const preFormData = ref(rawData)



const backupConfigure = () => {
    //var data = {res:1,dd:2,msg:"ggggg"}
    //const blob = new Blob([JSON.stringify(data, null, 2)], {type: 'application/json'})

    apiGetConfigure().then((res) => {
        //const blob = new Blob([res], {type: 'application/json'})
        //let fileName = new Date().format("yyyy-MM-dd hh:mm:ss")
        if (res.ret != 0) {
            MessageShow("error", "获取配置出错")
            return
        }
        let blob = new Blob([res.configure], { type: 'application/json' })
        FileSaver.saveAs(blob, "lucky_" + res.time + ".conf")

    }).catch((error) => {
        console.log("获取配置出错:" + error)
        MessageShow("error", "获取配置出错")
    })

}



const resetFormData = () => {
    form.value.AdminWebListenPort = preFormData.value.AdminWebListenPort
    form.value.AdminAccount = preFormData.value.AdminAccount
    form.value.AdminPassword = preFormData.value.AdminPassword
    form.value.ProxyCountLimit = preFormData.value.ProxyCountLimit
    form.value.GlobalMaxConnections = preFormData.value.GlobalMaxConnections
    form.value.AllowInternetaccess = preFormData.value.AllowInternetaccess
}

const syncToPreFormData = (data: any) => {
    preFormData.value.AdminWebListenPort = data.value.AdminWebListenPort
    preFormData.value.AdminAccount = data.value.AdminAccount
    preFormData.value.AdminPassword = data.value.AdminPassword
    preFormData.value.ProxyCountLimit = data.value.ProxyCountLimit
    preFormData.value.GlobalMaxConnections = data.value.GlobalMaxConnections
    preFormData.value.AllowInternetaccess = data.value.AllowInternetaccess
}



const rebootProgram = () => {
    disableRebootButton.value = true;

    ElMessageBox.confirm(
        '确定要重启lucky?',
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '点错了',
            type: 'warning',
        }
    )
        .then(() => {
            apiRebootProgram().then((res) => {
                MessageShow("success", "重启成功,3秒后自动跳转到新登录连接")

                setTimeout(() => {
                    window.location.href = getAdminURL()
                }, 3000)

                //console.log("getAdminURL "+getAdminURL())
            }).catch((error) => {
                disableRebootButton.value = false;
                console.log("重启操作出错:" + error)
                MessageShow("error", "重启操作出错")
            })
        })
        .catch(() => {
            disableRebootButton.value = false;
        })



}

const queryConfigure = () => {
    apiQueryBaseConfigure().then((res) => {
        if (res.ret == 0) {
            form.value = res.baseconfigure
            syncToPreFormData(form)
            return
        }
        MessageShow("error", "获取基本配置出错")
    }).catch((error) => {
        console.log("获取转发规则列表出错:" + error)
        MessageShow("error", "获取基本配置出错")
    })
}

const RequestAlterConfigure = () => {
    apiAlterBaseConfigure(form.value).then((res) => {
        if (res.ret == 0) {
            MessageShow("success", "配置修改成功")
            syncToPreFormData(form)
            return
        }
        MessageShow("error", res.msg)
    }).catch((error) => {
        console.log("配置修改失败,网络请求出错:" + error)
        MessageShow("error", "配置修改失败,网络请求出错")
    })
}




onMounted(() => {
    queryConfigure()

})

</script>


<style scoped>
.SetForm {
    margin-top: 15px;
    margin-left: 20px;
}

.formradius {
    border: 0px solid var(--el-border-color);
    border-radius: 0;
    margin: 0 auto;
    width: fit-content;
    padding: 15px;
}

#adminListen {
    width: 360px;
}

#adminAccount {
    width: 30vw;
    max-width: 360px;
    min-width: 300px;
}


#adminPassword {
    width: 30vw;
    max-width: 360px;
    min-width: 300px;
}


#proxyCountLimit {
    width: 360px;
}


#globalMaxConnections {
    width: 360px;
}

.inline-block {
    display: inline-block;
    margin-right: 10px;
}

.margin-change {
    display: inline-block;
    margin-left: 10px;
}
</style>