<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">
        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }">

                    <div class="whitelistConfigure">
                        <el-form :model="whiteListBaseConfigureForm" class="SetForm" label-width="auto">

                            <el-form-item label="自定义URL" id="whitelisturl">
                                <el-input v-model="whiteListBaseConfigureForm.URL" placeholder="自定义URL"
                                    autocomplete="off" style="witdh:250px;margin-bottom:4px;" />
                                <el-tooltip class="box-item oneLine" effect="dark" placement="bottom"
                                    :content="getWhiteListURL">
                                    <el-button type="info" round @click="copyRelayConfigure(getWhiteListURL)"
                                        style="margin-right: 10px;">复制</el-button>
                                </el-tooltip>
                                <a>{{ getNewWhiteListURL }}</a>
                            </el-form-item>

                            <el-form-item label="有效时长(小时)" id="whitelistActivelifeDuration">
                                <el-input-number v-model="whiteListBaseConfigureForm.ActivelifeDuration"
                                    autocomplete="off" :min="1" :max="99999" />
                            </el-form-item>

                            <el-form-item label="认证账号" id="basicAccount">
                                <el-input v-model="whiteListBaseConfigureForm.BasicAccount" placeholder="认证账号"
                                    autocomplete="off" style="witdh:250px;" />
                            </el-form-item>

                            <el-form-item label="认证密码" id="basicPassword">
                                <el-input v-model="whiteListBaseConfigureForm.BasicPassword" placeholder="认证密码"
                                    autocomplete="off" />
                            </el-form-item>
                        </el-form>
                        <el-button type="primary" round @click="SaveWhiteListConfigure">保存配置</el-button>
                    </div>


            </div>


        </el-scrollbar>
    </div>

</template>


<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import {MessageShow} from '../utils/ui'
import {CopyTotoClipboard} from '../utils/utils'
import { apiAlterWhiteListConfigure, apiGetWhiteListConfigure} from '../apis/utils'

const whiteListBaseConfigureForm = ref({
    URL: "",
    ActivelifeDuration: 36,
    BasicAccount: "",
    BasicPassword: "",
})

const preWhiteListBaseConfigureForm = ref({
    URL: "",
    ActivelifeDuration: 36,
    BasicAccount: "",
    BasicPassword: "",
})

const getWhiteListURL = computed(() => {
    if (preWhiteListBaseConfigureForm.value.URL == undefined || preWhiteListBaseConfigureForm.value.URL == "") {
        return window.location.protocol + "//" + window.location.hostname + ":" + window.location.port + "/wl"
    }
    return window.location.protocol + "//" + window.location.hostname + ":" + window.location.port + "/wl/" + preWhiteListBaseConfigureForm.value.URL
})

const getNewWhiteListURL = computed(() => {
    if (whiteListBaseConfigureForm.value.URL == undefined || whiteListBaseConfigureForm.value.URL == "") {
        return window.location.protocol + "//" + window.location.hostname + ":" + window.location.port + "/wl"
    }
    return window.location.protocol + "//" + window.location.hostname + ":" + window.location.port + "/wl/" + whiteListBaseConfigureForm.value.URL
})

const copyRelayConfigure = (url: string) => {
    CopyTotoClipboard(url)
    MessageShow('success', '白名单认证地址 ' + url + ' 已复制到剪切板')
}




const SaveWhiteListConfigure = () => {
    apiAlterWhiteListConfigure(whiteListBaseConfigureForm.value).then((res) => {
        if (res.ret == 0) {
            MessageShow("success", "保存成功")
            preWhiteListBaseConfigureForm.value = whiteListBaseConfigureForm.value
            return
        }
        MessageShow("error", res.msg)
        //console.log("getAdminURL "+getAdminURL())
    }).catch((error) => {
        MessageShow("error", "查询白名单设置出错")
    })
}

const queryWhiteListConfigure = () => {
    apiGetWhiteListConfigure().then((res) => {
        if (res.ret == 0) {
            whiteListBaseConfigureForm.value = ref(res.data).value
            preWhiteListBaseConfigureForm.value.URL = whiteListBaseConfigureForm.value.URL
            preWhiteListBaseConfigureForm.value.ActivelifeDuration = whiteListBaseConfigureForm.value.ActivelifeDuration
            preWhiteListBaseConfigureForm.value.BasicAccount = whiteListBaseConfigureForm.value.BasicAccount
            preWhiteListBaseConfigureForm.value.BasicPassword = whiteListBaseConfigureForm.value.BasicPassword
            return
        }
        MessageShow("error", res.msg)
        //console.log("getAdminURL "+getAdminURL())
    }).catch((error) => {
        MessageShow("error", "查询白名单设置出错")
    })
}

onMounted(() => {
    queryWhiteListConfigure()

})

</script>

<style scoped>
.formradius{
    border: 0px solid var(--el-border-color);
    border-radius: 0;
    margin:0 auto;
    width:fit-content;
    padding:15px;

    
}
</style>