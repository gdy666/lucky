<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }">



                <el-form :model="form" class="SetForm" label-width="auto">

                    <div class="AdminListenDivRadius">
                        <p>服务端设置</p>

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                被控制端不需要打开这个<br />
                            </template>
                            <el-form-item label="服务端开关">
                                <el-switch v-model="form.Server.Enable" class="mb-1" inline-prompt
                                    style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949" width="50px"
                                    active-text="开启" inactive-text="关闭" />
                            </el-form-item>
                        </el-tooltip>

                        <div v-show="form.Server.Enable">

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    客户端登录需要填写Token一致<br />
                                </template>

                                <el-form-item label="认证Token">
                                    <el-input v-model="form.Server.Token" placeholder="Token" autocomplete="off" />
                                </el-form-item>
                            </el-tooltip>

                        </div>

                    </div>

                    <div class="AdminListenDivRadius">

                        <p>客户端设置</p>


                        <el-form-item label="客户端开关">
                            <el-switch v-model="form.Client.Enable" class="mb-1" inline-prompt
                                style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949" width="50px"
                                active-text="开启" inactive-text="关闭" />
                            <el-divider direction="vertical" />
                            <el-button size="small" v-show="true">
                                {{ clientState }}
                            </el-button>

                            <el-divider direction="vertical" v-show="clientstateMsg == '' ? false : true" />


                            <el-button size="small" v-show="clientstateMsg == '' ? false : true" type="danger">
                                {{ clientstateMsg }}
                            </el-button>

                        </el-form-item>


                        <div v-show="form.Client.Enable">

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    这里填你的lucky后台管理地址<br />
                                    比如:<br />
                                    http://192.168.31.1:16601<br />
                                    或者<br />
                                    https://192.168.31.1:16601<br />
                                    保存配置后会自动转为ws://或者wss://开头的websocket协议地址.<br />
                                </template>

                                <el-form-item label="服务端地址">
                                    <el-input v-model="form.Client.ServerURL" placeholder="服务器地址" autocomplete="off" />
                                </el-form-item>
                            </el-tooltip>

                            <el-form-item label="Token">
                                <el-input v-model="form.Client.Token" placeholder="Token" autocomplete="off" />
                            </el-form-item>


                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    当lucky服务端和待唤醒设备不在同一局域网时,可以通过具备转发的客户端转发唤醒魔法包<br />
                                </template>
                                <el-form-item label="转发唤醒包">
                                    <el-switch v-model="form.Client.Relay" class="mb-1" inline-prompt
                                        style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949"
                                        width="50px" active-text="开启" inactive-text="关闭" />
                                </el-form-item>
                            </el-tooltip>


                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    在lucky服务器网络唤醒设备列表显示的设备名称,不能为空<br />
                                </template>
                                <el-form-item label="设备名称">
                                    <el-input v-model="form.Client.DeviceName" placeholder="设备名称" autocomplete="off" />
                                </el-form-item>
                            </el-tooltip>


                            <el-form-item label="网卡物理地址">
                                <el-input v-model="form.Client.Mac" placeholder="网卡物理地址" autocomplete="off">


                                    <template #append>

                                        <el-select v-model="form.Client.Mac" placeholder="网卡选择" style="width: auto"
                                            @change="interfaceMacChange">

                                            <div v-for="info in ipv4InterfaceList">
                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        <span
                                                            v-html="StrArrayListToBrHtml(getIPList(info.AddressList))"></span>
                                                    </template>
                                                    <el-option :label="info.NetInterfaceName"
                                                        :value=info.HardwareAddr />
                                                </el-tooltip>
                                            </div>


                                        </el-select>
                                    </template>
                                </el-input>
                            </el-form-item>

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    255.255.255.255在多数环境失效,不要使用<br />
                                </template>
                                <el-form-item label="广播地址">
                                    <el-input v-model="form.Client.BroadcastIP" placeholder="广播地址" autocomplete="off">
                                        <template #append>
                                            <el-select v-model="form.Client.BroadcastIP" placeholder="局域网选择"
                                                style="width: 180px">

                                                <div v-for="info in BroadcastIPInputAddressList">

                                                    <el-option :label="info.IP" :value=info.BroadcastIP />
                                                </div>
                                            </el-select>
                                        </template>
                                    </el-input>
                                </el-form-item>
                            </el-tooltip>

                            <el-form-item label="端口">
                                <el-input-number v-model="form.Client.Port" autocomplete="off" />
                            </el-form-item>

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    唤醒魔方包重复发送次数<br />
                                </template>
                                <el-form-item label="重复次数">
                                    <el-input-number v-model="form.Client.Repeat" autocomplete="off" />
                                </el-form-item>
                            </el-tooltip>

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    windows关机指令: Shutdown /s /t 0 <br />
                                    linux关机指令: poweroff <br />
                                    如果你发现你的windows关机后怎么也唤不醒,可以试试使用休眠/睡眠 指令: rundll32.exe powrprof.dll,SetSuspendState
                                    0,1,0 <br />
                                    注意新手注意别使用其它未知指令,比如"rm -rf"...<br />
                                    数据丢失概不负责<br />
                                </template>
                                <el-form-item label="关机指令">
                                    <el-input v-model="form.Client.PowerOffCMD" placeholder="关机指令" autocomplete="off" >
                                        <template #append>
                                            <el-select v-model="form.Client.PowerOffCMD" placeholder="关机指令"
                                                style="width: 160px">

                                                <div v-for="cmd in defualtCMDList">

                                                    <el-option :label="cmd.label" :value=cmd.value />
                                                </div>
                                            </el-select>
                                        </template>
                                    </el-input>
                                </el-form-item>
                            </el-tooltip>

                        </div>

                    </div>

                    <div class="AdminListenDivRadius" v-show="serviceStatus >= 0 ? true : false">

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                当前Lucky windows服务状态<br />
                               
                            </template>

                            <el-button type="info" size="small">
                                {{ serviceStatus == 0 ? 'lucky服务未安装' : serviceStatus == 1 ? 'lucky服务已启动' : serviceStatus == 2 ? 'luck服务已停止' : '未知服务状态' }}
                            </el-button>
                        </el-tooltip>


                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                安装windows服务,lukcy可能需要以管理员身份运行才能设置成功<br />
                                实现lucky开机自启<br/>
                            </template>
                            <el-button v-show='serviceStatus == 0 ? true : false' type="warning" round
                                @click="optionLuckyService('install')">安装Lucky windows服务</el-button>

                        </el-tooltip>
                        <el-button type="danger" v-show="serviceStatus>0?true:false" round @click="optionLuckyService('unstall')">删除Lucky windows服务
                        </el-button>


                        <div v-show="serviceStatus>0?true:false" >

                        <el-button type="info" size="small" @click="optionLuckyService('start')" v-show="serviceStatus==2?true:false">
                               启动lucky服务
                        </el-button>

                        <el-button type="info" size="small" @click="optionLuckyService('stop')" v-show="serviceStatus==1?true:false">
                               停止lucky服务
                        </el-button>

                        <el-button type="info" size="small" @click="optionLuckyService('restart')">
                               重启lucky服务
                        </el-button>
                        </div>



                    </div>

                </el-form>

                <el-button type="primary" round @click="RequestAlterConfigure">保存修改</el-button>


            </div>

        </el-scrollbar>
    </div>

</template>


<script lang="ts" setup>

import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'
import { apiGetWOLServiceConfigure, apiAlterWOLServiceConfigure, apiGetIPv4Interface, apiOptionsLuckyService, apiGetRestoreConfigureConfirm } from '../../apis/utils'
import { ElMessageBox } from 'element-plus'

import { StrArrayListToBrHtml } from '../../utils/utils'
import { MessageShow } from '../../utils/ui'
import FileSaver from 'file-saver'
import { anyTypeAnnotation } from '@babel/types'

const formLabelWidth = '10vw'
console.log("window.location.href " + window.location.href)
console.log("window.location.port " + window.location.port)
console.log("window.location.host " + window.location.host)
console.log("window.location " + JSON.stringify(window.location))
const disableRebootButton = ref(false)

const BroadcastIPInputAddressList = ref([{ IP: "", BroadcastIP: "" }])


const defualtCMDList = ref([
    {label:"windows关机指令",value:"Shutdown /s /t 0"},
    {label:"Linux关机指令",value:"poweroff"},
    {label:"windows休眠指令",value:"rundll32.exe powrprof.dll,SetSuspendState 0,1,0"},
    {label:"自定义",value:""}])


const ipv4InterfaceList = ref([{
    NetInterfaceName: "",
    HardwareAddr: "",
    AddressList: [{ IP: "", BroadcastIP: "" },]
},])

const getIPList = (AddressList) => {
    let iplist = new Array()
    for (let i in AddressList) {
        iplist.push(AddressList[i].IP)
    }
    return iplist
}

const interfaceMacChange = (val) => {
    if (val == "") {
        return
    }
    for (let i in ipv4InterfaceList.value) {
        if (ipv4InterfaceList.value[i].HardwareAddr == val) {
            BroadcastIPInputAddressList.value = ipv4InterfaceList.value[i].AddressList
            break
        }
    }

    if (BroadcastIPInputAddressList.value.length <= 0) {
        return
    }
    form.value.Client.BroadcastIP = BroadcastIPInputAddressList.value[0].BroadcastIP
}





const rawData = {
    Server: {
        Enable: false,
        Token: ""
    },
    Client: {
        Enable: false,
        ServerURL: "",
        Token: "",
        Relay: false,
        Key: "",
        DeviceName: "",
        Mac: "",
        BroadcastIP: "",
        Port: 9,
        Repeat: 5,
        PowerOffCMD: "",
    }
}
const clientState = ref("")
const clientstateMsg = ref("")
const serviceStatus = ref(-1)

const form = ref(rawData)
const preFormData = ref(rawData)









const queryIPv4InterfaceList = () => {
    apiGetIPv4Interface().then((res) => {
        if (res.ret == 0) {
            ipv4InterfaceList.value = res.list
            interfaceMacChange(form.value.Client.Mac)
            return
        }
        MessageShow("error", "获取Ipv4网卡信息列表出错" + res.msg)
    }).catch((error) => {
        console.log("获取Ipv4网卡信息列表出错:" + error)
        MessageShow("error", "获取Ipv4网卡信息列表出错")
    })
}

const queryConfigure = () => {
    apiGetWOLServiceConfigure().then((res) => {
        if (res.ret == 0) {
            form.value = res.configure
            clientState.value = res.ClientState
            clientstateMsg.value = res.ClientStateMsg
            serviceStatus.value = res.serviceStatus
            return
        }
        MessageShow("error", "获取唤醒服务配置出错")
    }).catch((error) => {
        console.log("获取唤醒服务配置出错:" + error)
        MessageShow("error", "获取唤醒服务配置出错")
    })
}

const optionLuckyService = (op) => {



    var optionServiceText = "安装Lucky服务"
    if (op == 'unstall') {
        optionServiceText = "卸载Lucky服务"
    }else if (op == 'start'){
        optionServiceText = "启动Lucky服务"
    }else if (op == 'stop'){
        optionServiceText = '停止lucky服务'
    }else if (op=="restart"){
        optionServiceText = '重启lucky服务'
    }

    var warnText = ""

    if (op=="install"){
        warnText = "\n安装成功后lucky会重启并以windows后台服务方式启动,到时需重新登录后台"
    }else if(op=="unstall"){
        warnText="\n卸载后lucky也会随之退出,如有需要请手动启动"
    }


    ElMessageBox.confirm(
        '确认要 ' + optionServiceText + " ?"+warnText,
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            apiOptionsLuckyService(op).then((res) => {
                if (res.ret == 0) {
                    MessageShow("success", optionServiceText + "成功")
                    if (op == 1) {
                        MessageShow("success", "请重新启动系统后执行输入后台网站,确认lucky服务已正常自启.")
                    }

                    if (res.msg!=""){
                        MessageShow("success", res.msg) 
                    }
                    serviceStatus.value = res.status

                    return
                }
                console.log(optionServiceText + " 出错:" + res.msg)

                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("服务状态修改出错,网络请求出错:" + error)
                MessageShow("error", "服务状态修改出错,网络请求出错")
            })

        })


}

const RequestAlterConfigure = () => {
    apiAlterWOLServiceConfigure(form.value).then((res) => {
        if (res.ret == 0) {
            MessageShow("success", "修改成功")
            form.value = res.configure

            setTimeout(() => {
                queryConfigure()
            }, 2000)


            return
        }
        console.log("保存配置出错:" + res.msg)
        MessageShow("error", res.msg)
    }).catch((error) => {
        console.log("配置修改失败,网络请求出错:" + error)
        MessageShow("error", "配置修改失败,网络请求出错")
    })
}



//var timerID: any

onMounted(() => {
    queryConfigure()
    queryIPv4InterfaceList()

    // timerID = setInterval(() => {
    //     queryConfigure()
    //     queryIPv4InterfaceList()
    // }, 2000);

})

onUnmounted(() => {
    //clearInterval(timerID)
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
    width: 600px;
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

.inline-block {
    display: inline-block;
    margin-right: 10px;
}

.margin-change {
    display: inline-block;
    margin-left: 10px;
}
</style>