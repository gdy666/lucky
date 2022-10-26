<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }" v-loading="logLoading" element-loading-background="transparent">

        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }">

                <el-form :model="form" class="SetForm" label-width="auto">

                    <!-- <el-tooltip content="如果不需要DDNS动态域名服务请不要打开这个开关" placement="top">

                        <el-form-item label="动态域名服务开关" id="adminListen">
                            <el-switch v-model="form.Enable" class="mb-1" inline-prompt
                                style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949" width="50px"
                                active-text="开启" inactive-text="停用" />
                        </el-form-item>
                    </el-tooltip>

                    <el-tooltip content="多数嵌入式设备启用这个开关会导致https访问失败" placement="top">

                        <el-form-item label="Http(s) 客户端 安全证书验证" id="adminListen">
                            <el-switch v-model="form.HttpClientSecureVerify" class="mb-1" inline-prompt
                                style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949" width="50px"
                                active-text="启用" inactive-text="禁用" />
                        </el-form-item> -->
                    <!-- </el-tooltip> -->

                    <div class="AdminListenDivRadius">


                    

                    <el-tooltip content="同一端口tcp和udp类型各算一个,(0-1024)" placement="top">
                        <el-form-item label="全局端口转发最大数量" label-width="auto" min="0" max="1024">
                            <el-input-number v-model="form.PortForwardsLimit" autocomplete="off" />
                        </el-form-item>
                    </el-tooltip>

                    <el-tooltip content="端口转发全局TCP最大并发连接数,(0-4096)" placement="top">
                        <el-form-item label="端口转发全局TCP最大并发连接数" label-width="auto" min="0" max="4096">
                            <el-input-number v-model="form.TCPPortforwardMaxConnections" autocomplete="off" />
                        </el-form-item>
                    </el-tooltip>

                    <el-tooltip content="端口转发全局UDP读取目标地址数据协程数限制,(0-4096)" placement="top">
                        <el-form-item label="端口转发全局UDP读取目标地址数据协程数限制" label-width="auto" min="0" max="4096">
                            <el-input-number v-model="form.UDPReadTargetDataMaxgoroutineCount" autocomplete="off" />
                        </el-form-item>
                    </el-tooltip>


                </div>

                    <!-- <el-tooltip content="DDNS任务每次执行的时间间隔,最小30秒,最长3600秒" placement="top">
                        <el-form-item label="时间间隔(秒)" label-width="auto" :min="30" :max="3600">
                            <el-input-number v-model="form.Intervals" autocomplete="off" />
                        </el-form-item>
                    </el-tooltip> -->


                </el-form>

                <el-button type="primary" round @click="RequestAlterPortForwardConfigure">保存修改</el-button>


            </div>


        </el-scrollbar>
    </div>

</template>


<script lang="ts" setup>

import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'
import { apiQueryPortForwardConfigure, apiAlterPortForwardConfigure } from '../apis/utils'


import { MessageShow } from '../utils/ui'


const logLoading = ref(true)


const rawData = {
    // Enable: false,
    // HttpClientSecureVerify: false,
    // Intervals: 0,
    PortForwardsLimit: 0,
    TCPPortforwardMaxConnections:0,
    UDPReadTargetDataMaxgoroutineCount:0,
}

const form = ref(rawData)
const preFormData = ref(rawData)

const resetFormData = () => {
    form.value.PortForwardsLimit = preFormData.value.PortForwardsLimit
    form.value.TCPPortforwardMaxConnections = preFormData.value.TCPPortforwardMaxConnections
}

const syncToPreFormData = (data: any) => {
    preFormData.value.PortForwardsLimit = data.value.PortForwardsLimit
    preFormData.value.TCPPortforwardMaxConnections = data.TCPPortforwardMaxConnections
}







const queryPortForwardsConfigure = () => {
    apiQueryPortForwardConfigure().then((res) => {

        if (res.ret == 0) {
            logLoading.value = false
            form.value = res.configure
            syncToPreFormData(form)
            return
        }
        MessageShow("error", "获取端口转发配置出错")
    }).catch((error) => {
        MessageShow("error", "获取端口转发配置出错")
    })
}

const RequestAlterPortForwardConfigure = () => {
    apiAlterPortForwardConfigure(form.value).then((res) => {
        if (res.ret == 0) {
            MessageShow("success", "配置修改成功")
            queryPortForwardsConfigure()
            return
        }
        resetFormData()
        MessageShow("error", res.msg)
    }).catch((error) => {
        console.log("配置修改失败,网络请求出错:" + error)
        MessageShow("error", "配置修改失败,网络请求出错")
        resetFormData()
    })
}




onMounted(() => {
    queryPortForwardsConfigure()

})

</script>


<style scoped>

.AdminListenDivRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 456;
    padding-top: 9px;
    padding-left: 9px;
    padding-right: 9px;
}


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
</style>