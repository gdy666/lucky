<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">
        <el-scrollbar height="100%">

            <div class="itemradius" :style="{
                borderRadius: 'base',
            }">

                <el-descriptions :column="4" border>

                    <div v-for="device in deviceList">
                        <el-descriptions-item label="设备" label-class-name="deviceNamelabelClass"
                            class-name="deviceNamecontentClass">
                            <el-button size="small" v-show="true">
                                {{ device.DeviceName == '' ? '未命名设备' : device.DeviceName }}
                            </el-button>

                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="deviceOnlieDeviceToBrHtml(device)"></span>
                                </template>
                                <el-button size="small" v-show="true">
                                    {{device.State}}
                                </el-button>
                            </el-tooltip>


                        </el-descriptions-item>

                        <el-descriptions-item label="操作" label-class-name="deviceOptlabelClass"
                            class-name="deviceOptcontentClass">

                            <el-button size="small" type="success" @click="wakeup(device)">
                                唤醒
                            </el-button>


                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    关机<br />
                                </template>
                                <el-button size="small" type="danger" @click="shutdown(device)"
                                    :disabled="avalidShutDownButton(device.State)">
                                    关机
                                </el-button>
                            </el-tooltip>

                            &nbsp; &nbsp;
                            <el-divider direction="vertical" />
                            &nbsp; &nbsp;

                            <el-button size="small" type="primary" @click="showAlterDeviceDialog(device)">
                                编辑
                            </el-button>

                            <el-button size="small" type="danger" @click="deleteDevice(device)">
                                删除
                            </el-button>
                        </el-descriptions-item>

                        <el-descriptions-item label="物理网卡地址" label-class-name="deviceMaclabelClass"
                            class-name="deviceMaccontentClass">
                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    ↓↓↓物理网卡地址↓↓↓<br />
                                    <span v-html="StrArrayListToBrHtml(device.MacList)"></span>
                                    <br />
                                    ↓↓↓魔法包地址↓↓↓<br />
                                    <span v-html="StrArrayListToBrHtml(device.BroadcastIPs)"></span>
                                    <br />
                                    端口: {{device.Port}} <br />

                                </template>
                                <el-button size="small" v-show="true">
                                    {{device.MacList.length==1?device.MacList[0]:device.MacList[0]+'...' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="物联网平台" label-class-name="deviceIOTlabelClass"
                            class-name="deviceIOTcontentClass">

                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="showIOT_DianDengInfo(device)"></span>
                                </template>

                                <el-button size="small" v-show="device.IOT_DianDeng_Enable"
                                    :type="showIOT_DianDengColor(device)">
                                    点灯科技
                                </el-button>
                            </el-tooltip>




                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="showIOT_BemfaInfo(device)"></span>
                                </template>

                                <el-button size="small" v-show="device.IOT_Bemfa_Enable"
                                    :type="showIOT_BemfaColor(device)">
                                    巴法云
                                </el-button>
                            </el-tooltip>


                        </el-descriptions-item>
                    </div>


                </el-descriptions>

            </div>

        </el-scrollbar>

        <el-affix position="bottom" :offset="30" class="affix-container">
            <el-button type="primary" :round=true @click="showAddDeviceDialog">添加可唤醒的设备
                <el-icon class="el-icon--right">
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix>

        <el-dialog v-if="deviceDialogShow" v-model="deviceDialogShow" :title=deviceDialogTitle draggable
            :show-close="true" :close-on-click-modal="false" width="400px">
            <el-form-item label="设备名称" label-width="120px">
                <el-input v-model="deviceForm.DeviceName" autocomplete="off" />
            </el-form-item>

            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                <template #content>
                    每行填写一个MAC<br />
                    一般情况填写一个MAC地址即可<br />
                </template>

                <el-form-item label-width="120px" label="设备MAC">
                    <el-input v-model="deviceFormMacListArea" :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                        type="textarea" wrap="off">
                    </el-input>
                </el-form-item>
            </el-tooltip>


            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                <template #content>
                    不建议使用255.255.255.255<br />
                    一般家用情况,如果路由器的管理地址是192.168.31.1,则填写192.168.31.255作为广播地址<br />
                    菜鸟没法确定的话可留空,程序会遍历全部IPV4地址发送广播<br />
                    每行填写一个广播地址<br />
                    一般情况填写一个广播地址即可<br />
                </template>

                <el-form-item label-width="120px" label="魔方包广播地址">
                    <el-input v-model="deviceFormBroadcastIPsArea" :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                        type="textarea" wrap="off">
                    </el-input>
                </el-form-item>
            </el-tooltip>


            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                <template #content>
                    默认端口为9<br />
                    没特殊情况不要修改<br />
                </template>
                <el-form-item label="端口" label-width="120px" :min="1" :max="65535">
                    <el-input-number v-model="deviceForm.Port" autocomplete="off" />
                </el-form-item>
            </el-tooltip>

            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                <template #content>
                    每次执行唤醒时重复发送魔方包的次数<br />
                    可设置范围(1-10)<br />
                </template>
                <el-form-item label="重复次数" label-width="120px" :min="1" :max="10">
                    <el-input-number v-model="deviceForm.Repeat" autocomplete="off" />
                </el-form-item>
            </el-tooltip>


            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                <template #content>
                    一般情况忽略即可<br />
                    目标唤醒设备和lucky不在同一局域网时才用得上这个开关<br />
                    发送广播的同时交给中继设备发送<br />
                </template>
                <el-form-item label="转播开关" label-width="120px" v-if="true">
                    <el-switch v-model="deviceForm.Relay" inline-prompt width="50px" active-text="启用"
                        inactive-text="禁用" />
                </el-form-item>
            </el-tooltip>

            <div class="divRadius">
                <p>第三方物联网平台对接-仅支持语音助手控制</p>

                <div class="divIOTRadius">
                    <el-form-item label="点灯科技" label-width="120px" v-if="true">
                    <el-switch v-model="deviceForm.IOT_DianDeng_Enable" inline-prompt width="50px" active-text="启用"
                        inactive-text="禁用" />
                </el-form-item>

                    <div v-show="deviceForm.IOT_DianDeng_Enable">
                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                        <template #content>
                            留空表示不启用该平台对接功能<br />
                            <br />
                            点灯科技官网:https://www.diandeng.tech/home<br />
                            <br />
                            在手机APP端注册登录-设备管理菜单-右上方+号按钮-添加点灯独立设备-选择网络接入-复制保存-secrte-key到下方<br />
                            目前仅支持点灯-独立设备类型,在设备管理修改设备名,这里的设备名就是你以后用语音助手控制操作设备时的名称.<br />
                            <br />
                            以小爱同学为例,在米家APP右下角"我的"-"其它平台设备"-右上方"添加"-选择"点灯科技"登录同步后即可.<br />
                            需要确保点灯secrte-key填写正确,在点灯APP已经显示设备在线再在米家执行同步,每次修改完设备名都要重新同步.<br />
                            小度/天猫精灵自行参考文档:https://www.diandeng.tech/doc/voice-assistant <br />
                            <br />
                            多个设备可以设置同一个secrte-key,表示多个待唤醒设备与同一个点灯设备绑定,一次语音操作同时控制多个设备的开关.不建议这样做.<br />
                            建议一个待唤醒设备对应一个点灯设备<br />
                            <br />
                            由于点灯接口偶尔发生变化,又不提供相关文档,所以哪天突然不能使用了也很正常.<br />
                        </template>
                        <el-form-item label-width="120px" label="设备密钥">
                            <el-input v-model="deviceForm.IOT_DianDeng_AUTHKEY" placeholder="设备密钥" type="text"
                                wrap="off">
                            </el-input>
                        </el-form-item>
                    </el-tooltip>
                </div>
                </div>

                <div class="divIOTRadius">
                    <el-form-item label="巴法云" label-width="120px" v-if="true">
                    <el-switch v-model="deviceForm.IOT_Bemfa_Enable" inline-prompt width="50px" active-text="启用"
                        inactive-text="禁用" />
                </el-form-item>
                    <div v-show="deviceForm.IOT_Bemfa_Enable">
                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                留空表示不启用该平台对接功能<br />
                                <br />
                                巴法云官网:https://cloud.bemfa.com/<br />
                                <br />
                                注册登录后,在控制台右上方可查看私钥.<br />
                                目前仅支持MQTT设备云,记得创建的是MQTT设备云主题,而不是TCP创客云主题.主题名要以001结尾表示插座类型.<br />
                                <br />
                                以小爱同学为例,在米家APP右下角"我的"-"其它平台设备"-右上方"添加"-选择"巴法"登录同步后即可.<br />
                                需要确保私钥和主题填写正确,在控制台已经显示订阅设备在线再在米家执行同步,主题更多设置的右上方可修改昵称(在米家等平台显示的名称),每次修改完设备名都要重新同步.<br />

                                <br />
                                多个设备可以设置同一个主题,表示多个待唤醒设备与同一个主题绑定,一次语音操作同时控制多个设备的开关.不建议这样做.<br />
                                建议一个待唤醒设备对应一个主题<br />
                                <br />
                                <br />
                            </template>


                            <el-form-item label-width="120px" label="私钥">
                                <el-input v-model="deviceForm.IOT_Bemfa_SecretKey" placeholder="设备密钥" type="text"
                                    wrap="off">
                                </el-input>
                            </el-form-item>


                        </el-tooltip>
                        <el-form-item label-width="120px" label="主题">
                            <el-input v-model="deviceForm.IOT_Bemfa_Topic" placeholder="订阅主题" type="text" wrap="off">
                            </el-input>
                        </el-form-item>

                    </div>
                </div>



            </div>




            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="deviceDialogShow = false">取消</el-button>
                    <el-button type="primary" @click="addOreAlterDevice">{{deviceDialogCommitButtonText}}</el-button>
                </span>
            </template>
        </el-dialog>


    </div>

</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { MessageShow } from '../../utils/ui'
import { StrArrayListToBrHtml, StrArrayListToArea, StringToArrayList } from '../../utils/utils'
import { GetToken, apiGetWOLDeviceList, apiAddWOLDevice, apiDeleteWOLDevice, apiAlterWOLDevice, apiWOLDeviceWakeUp, apiWOLDeviceShutDown } from '../../apis/utils'

import {
    SwitchButton,
    AlarmClock,
    Bell,
} from '@element-plus/icons-vue'

const cellclass = {
    "min-width": "50px",
    "word-break": "keep-all"
}

const CS = {
    "min-width": "300px",
    "word-break": "break-all"
}

const deviceDialogShow = ref(false)
const deviceDialogTitle = ref("")
const deviceDialogCommitButtonText = ref("")
const deviceFormMacListArea = ref("")
const deviceFormBroadcastIPsArea = ref("")
const deviceFormActionType = ref("")

const showIOT_DianDengInfo = (device) => {
    var res = ""
    if (device.DianDengClientState == "未设置") {
        res = "未设置"
        return res
    }

    res += '设备密钥:' + device.IOT_DianDeng_AUTHKEY + '<br/>'
    res += '服务器连接状态:' + device.DianDengClientState + ' <br/>'
    if (device.DianDengClientState == "已连接") {
        res += '支持小爱同学 小度 天猫精灵<br/>'
    }
    return res
}

const showIOT_BemfaInfo = (device) => {
    var res = ""
    if (device.BemfaClientState == "未设置") {
        res = "未设置"
        return res
    }

    res += '私钥:' + device.IOT_Bemfa_SecretKey + '<br/>'
    res += '订阅主题:' + device.IOT_Bemfa_Topic + '<br/>'
    res += '服务器连接状态:' + device.BemfaClientState + ' <br/>'
    if (device.BemfaClientState == "已连接") {
        res += '支持小爱同学 小度 天猫精灵 google语音 AmazonAlexa<br/>'
    }
    return res
}

const showIOT_DianDengColor = (device) => {
    if (device.DianDengClientState == "已连接") {
        return "success"
    }
    return "info"
}

const showIOT_BemfaColor = (device) => {
    if (device.BemfaClientState == "已连接") {
        return "success"
    }
    return "info"
}

const deviceList = ref([{
    Key: "",
    DeviceName: "",
    MacList: [''],
    BroadcastIPs: [''],
    Port: 9,
    Relay: true,
    Repeat: 5,
    State: "",
    OnlineMacList: [''],
    IOT_DianDeng_Enable: false,
    IOT_DianDeng_AUTHKEY: "",
    DianDengClientState: "",
    IOT_Bemfa_Enable: false,
    IOT_Bemfa_SecretKey: "",
    IOT_Bemfa_Topic: "",
    BemFaClientState: "",
},])

const deviceForm = ref({
    Key: "",
    DeviceName: "",
    MacList: [''],
    BroadcastIPs: [''],
    Port: 9,
    Relay: true,
    Repeat: 5,
    IOT_DianDeng_Enable: false,
    IOT_DianDeng_AUTHKEY: "",
    IOT_Bemfa_SecretKey: "",
    IOT_Bemfa_Topic: "",
    IOT_Bemfa_Enable: false,
})

const deviceOnlieDeviceToBrHtml = (device) => {
    if (device.OnlineMacList == undefined || device.OnlineMacList == null || device.OnlineMacList.length <= 0) {
        return "没有设备在线"
    }


    var resHtml = "在线设备列表:<br />"
    for (let i in device.OnlineMacList) {
        resHtml += device.OnlineMacList[i] + '<br />'
    }
    return resHtml
}


const deleteDevice = (device) => {

    const deviceName = device.DeviceName == "" ? device.MacList[0] : device.DeviceName;
    const deviceText = "[" + deviceName + "]"

    ElMessageBox.confirm(
        '确认要删除待唤醒设备 ' + deviceText + "?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            console.log("确认删除 " + deviceText)

            apiDeleteWOLDevice(device.Key).then((res) => {
                if (res.ret == 0) {
                    queryDeviceList();
                    MessageShow("success", "删除成功")
                } else {
                    MessageShow("error", res.msg)
                }

            }).catch((error) => {
                console.log("删除唤醒设备失败,网络请求出错:" + error)
                MessageShow("error", "删除唤醒设备失败,网络请求出错")
            })
        })
        .catch(() => {

        })

}

const wakeup = (device) => {

    const deviceName = device.DeviceName == "" ? device.MacList[0] : device.DeviceName;
    const deviceText = "[" + deviceName + "]"

    ElMessageBox.confirm(
        '确认要唤醒设备 ' + deviceText + "?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {

            apiWOLDeviceWakeUp(device.Key).then((res) => {
                if (res.ret == 0) {
                    MessageShow("success", "唤醒指令已发送")
                    queryDeviceList();
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("唤醒指令发送失败,网络请求出错:" + error)
                MessageShow("error", "唤醒指令发送失败,网络请求出错")
            })
        })



}


const shutdown = (device) => {

    const deviceName = device.DeviceName == "" ? device.MacList[0] : device.DeviceName;
    const deviceText = "[" + deviceName + "]"

    ElMessageBox.confirm(
        '确认要向设备 ' + deviceText + " 发送关机指令?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {

            apiWOLDeviceShutDown(device.Key).then((res) => {
                if (res.ret == 0) {
                    MessageShow("success", "已向在线设备发送关机指令")
                    queryDeviceList();
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("关机指令发送失败,网络请求出错:" + error)
                MessageShow("error", "关机指令发送失败,网络请求出错")
            })
        })



}

const avalidShutDownButton = (state: string) => {
    let res = state.indexOf("在线");
    console.log("show: " + res)
    if (res >= 0) {
        return false
    }

    return true
}

const addOreAlterDevice = () => {

    deviceForm.value.BroadcastIPs = StringToArrayList(deviceFormBroadcastIPsArea.value)
    deviceForm.value.MacList = StringToArrayList(deviceFormMacListArea.value)

    switch (deviceFormActionType.value) {
        case "add":
            apiAddWOLDevice(deviceForm.value).then((res) => {
                if (res.ret == 0) {
                    deviceDialogShow.value = false;
                    MessageShow("success", "设备添加成功")
                    queryDeviceList();
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("唤醒设备添加失败,网络请求出错:" + error)
                MessageShow("error", "唤醒设备添加失败,网络请求出错")
            })
            break;
        case "alter":

            apiAlterWOLDevice(deviceForm.value).then((res) => {
                if (res.ret == 0) {
                    deviceDialogShow.value = false;
                    MessageShow("success", "设备修改成功")
                    queryDeviceList();
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("唤醒设备修改失败,网络请求出错:" + error)
                MessageShow("error", "唤醒设备修改失败,网络请求出错")
            })

            break;
        default:
    }
}

const showAlterDeviceDialog = (device) => {
    deviceDialogCommitButtonText.value = "修改"
    deviceForm.value = {
        Key: device.Key,
        DeviceName: device.DeviceName,
        MacList: device.MacList,
        BroadcastIPs: device.BroadcastIPs,
        Port: device.Port,
        Relay: device.Relay,
        Repeat: device.Repeat,
        IOT_DianDeng_Enable: device.IOT_DianDeng_Enable,
        IOT_DianDeng_AUTHKEY: device.IOT_DianDeng_AUTHKEY,
        IOT_Bemfa_SecretKey: device.IOT_Bemfa_SecretKey,
        IOT_Bemfa_Topic: device.IOT_Bemfa_Topic,
        IOT_Bemfa_Enable: device.IOT_Bemfa_Enable,

    }
    deviceFormActionType.value = "alter"
    deviceFormMacListArea.value = StrArrayListToArea(device.MacList)
    deviceFormBroadcastIPsArea.value = StrArrayListToArea(device.BroadcastIPs)
    deviceDialogShow.value = true
}

const showAddDeviceDialog = () => {
    deviceDialogCommitButtonText.value = "添加"
    deviceForm.value = {
        Key: "",
        DeviceName: "",
        MacList: [''],
        BroadcastIPs: [''],
        Port: 9,
        Relay: true,
        Repeat: 5,
        IOT_DianDeng_Enable: false,
        IOT_DianDeng_AUTHKEY: "",
        IOT_Bemfa_Enable: false,
        IOT_Bemfa_SecretKey: "",
        IOT_Bemfa_Topic: "",
    }

    deviceFormActionType.value = "add"
    deviceFormMacListArea.value = ""
    deviceFormBroadcastIPsArea.value = ""
    deviceDialogShow.value = true

}


const queryDeviceList = () => {
    apiGetWOLDeviceList().then((res) => {
        //console.log(res.data)
        deviceList.value = res.list
    }).catch((error) => {
        console.log("获取设备列表出错:" + error)
        MessageShow("error", "获获取设备列表出错")
    })
}


var timerID: any

onMounted(() => {
    queryDeviceList();

    timerID = setInterval(() => {
        queryDeviceList();
    }, 2000);

})

onUnmounted(() => {
    clearInterval(timerID)
})

</script>

<style lang="scss">
.itemradius {

    border: 1px solid var(--el-border-color);
    border-radius: 0;
    margin-left: 3px;
    margin-top: 3px;
    margin-right: 3px;
    margin-bottom: 5px;
    min-width: 1200px;
}


.divRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 330px;
    padding-top: 9px;
    padding-left: 9px;
    padding-right: 9px;
}

.divIOTRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 300px;
    padding-top: 9px;
    padding-left: 9px;
    padding-right: 9px;
}

.deviceNamelabelClass {
    width: 55px,
}

.deviceNamecontentClass {
    width: 230px,
}

.deviceOptlabelClass {
    width: 55px,
}

.deviceOptcontentClass {
    width: 320px,
}

.deviceMaclabelClass {
    width: 110px,
}

.deviceMaccontentClass {
    width: 150px,
}

.deviceIOTlabelClass {
    width: 90px,
}

.deviceIOTcontentClass {
    width: 180px,
}
</style>