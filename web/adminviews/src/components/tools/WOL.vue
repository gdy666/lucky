<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">
        <el-scrollbar height="100%">

            <div class="itemradius" :style="{
                borderRadius: 'base',
            }"  v-for="device in deviceList" >

                <el-descriptions :column="4" border >

                    <el-descriptions-item label="设备操作">
                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                唤醒<br />
                            </template>
                            <el-button size="small" :icon="Bell" circle type="success" @click="wakeup(device)">
                            </el-button>
                        </el-tooltip>

                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                关机<br />
                            </template>
                            <el-button size="small" :icon="SwitchButton" circle type="danger">
                            </el-button>
                        </el-tooltip>

                        &nbsp; &nbsp; 

                        <el-button size="small" type="primary" @click="showAlterDeviceDialog(device)">
                            编辑
                        </el-button>

                        <el-button size="small" type="danger" @click="deleteDevice(device)">
                            删除
                        </el-button>
                        </el-descriptions-item>

                    <el-descriptions-item label="设备名称" >
                        <el-button size="default" v-show="true">
                            {{ device.DeviceName == '' ? '未命名设备' : device.DeviceName }}
                        </el-button>
                    </el-descriptions-item>








                    

                    <el-descriptions-item label="设备MAC">
                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="StrArrayListToBrHtml(device.MacList)"></span>
                            </template>
                            <el-button size="small" v-show="true">
                                {{device.MacList.length==1?device.MacList[0]:device.MacList[0]+'...' }}
                            </el-button>
                        </el-tooltip>
                    </el-descriptions-item>

                    <el-descriptions-item label="魔法包广播地址">
                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="StrArrayListToBrHtml(device.BroadcastIPs)"></span>
                            </template>
                            <el-button size="small" v-show="true">
                                {{device.BroadcastIPs.length==1?device.BroadcastIPs[0]:device.BroadcastIPs[0]+'...' }}
                            </el-button>
                        </el-tooltip>


                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                端口<br />
                            </template>
                            <el-button size="small" v-show="true">
                                {{device.Port}}
                            </el-button>

                        </el-tooltip>

                    </el-descriptions-item>





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
            :show-close="false" :close-on-click-modal="false" width="400px">
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
import { ref, onMounted, computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { MessageShow } from '../../utils/ui'
import { StrArrayListToBrHtml, StrArrayListToArea, StringToArrayList } from '../../utils/utils'
import { GetToken, apiGetWOLDeviceList, apiAddWOLDevice, apiDeleteWOLDevice, apiAlterWOLDevice,apiWOLDeviceWakeUp } from '../../apis/utils'

import {
    SwitchButton,
    AlarmClock,
    Bell,
} from '@element-plus/icons-vue'

const deviceDialogShow = ref(false)
const deviceDialogTitle = ref("")
const deviceDialogCommitButtonText = ref("")
const deviceFormMacListArea = ref("")
const deviceFormBroadcastIPsArea = ref("")
const deviceFormActionType = ref("")

const deviceList = ref([{
    Key: "",
    DeviceName: "",
    MacList: [''],
    BroadcastIPs: [''],
    Port: 9,
    Relay: true,
    Repeat: 5,
},])

const deviceForm = ref({
    Key: "",
    DeviceName: "",
    MacList: [''],
    BroadcastIPs: [''],
    Port: 9,
    Relay: true,
    Repeat: 5,
})

const deleteDevice = (device)=>{

    const deviceName = device.DeviceName==""?device.MacList[0]:device.DeviceName;
    const deviceText = "[" + deviceName +"]"

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

const wakeup = (device)=>{
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

const showAlterDeviceDialog = (device)=>{
    deviceDialogCommitButtonText.value = "修改"
    deviceForm.value = {
        Key: device.Key,
        DeviceName: device.DeviceName,
        MacList:device.MacList,
        BroadcastIPs: device.BroadcastIPs,
        Port: device.Port,
        Relay: device.Relay,
        Repeat: device.Repeat,
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

</script>

<style scoped>

.itemradius {

border: 1px solid var(--el-border-color);
border-radius: 0;
margin-left: 3px;
margin-top: 3px;
margin-right: 3px;
margin-bottom: 5px;
min-width: 1200px;
}

</style>