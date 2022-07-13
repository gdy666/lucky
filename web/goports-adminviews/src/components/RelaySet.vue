<template>



    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">



                <el-scrollbar height="100%">


                    <div class="itemradius" :style="{
                        borderRadius: 'base',
                    }" v-for="rule in ruleList">

                        <el-descriptions :column="4" border>

                            <el-descriptions-item label="名称&类型">

                                <el-button color="#409eff" size="default" v-show="true">
                                    {{ rule.Name == '' ? '未命名规则' : rule.Name }}
                                </el-button>


                                <el-button color="#0059b3" size="small" v-show="true"
                                    v-for="t in getRelayTypeList(rule.RelayType)">{{ t }}</el-button>
                            </el-descriptions-item>
                            <el-descriptions-item label="命令行配置">
                                <el-tooltip class="box-item" effect="dark" :content="rule.Mainconfigure"
                                    placement="top-start">
                                    <el-button type="primary" size="small" v-show="true"
                                        @click="copyRelayConfigure(rule.Mainconfigure)">
                                        复制
                                    </el-button>
                                </el-tooltip>
                            </el-descriptions-item>
                            <el-descriptions-item label="操作" :span="2">
                                <el-tooltip :content="rule.Enable == true ? '规则已启用' : '规则已禁用'" placement="top">
                                    <el-switch v-model="rule.Enable" inline-prompt active-text="开" inactive-text="关"
                                        :before-change="ruleEnableClick.bind(this, rule.Enable, rule)" size="large" />
                                </el-tooltip>

                                &nbsp;&nbsp;
                                <el-button type="primary" @click="alterRule(rule)">编辑</el-button>
                                <el-button type="danger" @click="deleteRule(rule)">删除</el-button>

                            </el-descriptions-item>



                            <el-descriptions-item label="监听IP">

                                <el-button color="#D4D7DE" size="small" v-show="true">
                                    {{ rule.ListenIP == '' ? '所有IP' : rule.ListenIP }}
                                </el-button>
                            </el-descriptions-item>
                            <el-descriptions-item label="监听端口">


                                <el-button color="#D4D7DE" size="small" v-show="true">
                                    {{ rule.ListenPorts }}
                                </el-button>

                            </el-descriptions-item>


                            <el-descriptions-item label="其它参数" :span="2">

                             <el-tooltip class="box-item" effect="dark" content="安全模式" placement="bottom">
                                    <el-button color="#6666ff" size="small" v-show="true">{{
                                            rule.Options.SafeMode=='whitelist'?'白名单':'黑名单'
                                    }}</el-button>
                                </el-tooltip>


                                <el-tooltip class="box-item" effect="dark" content="单端口最大并发数" placement="bottom">
                                    <el-button color="#6666ff" size="small" v-show="true">{{
                                            rule.Options.SingleProxyMaxConnections
                                    }}</el-button>
                                </el-tooltip>

                                <el-tooltip class="box-item" effect="dark" content="UDP包最大长度" placement="bottom">
                                    <el-button color="#626aef" size="small" v-show="ruleRelayTypeContainsUDP(rule)">{{
                                            rule.Options.UDPPackageSize
                                    }}</el-button>
                                </el-tooltip>

                                <el-tooltip class="box-item" effect="dark"
                                    :content="rule.Options.UDPProxyPerformanceMode == true ? 'UDP性能模式开启' : '性能模式关闭'"
                                    placement="bottom">
                                    <el-button color="#626aef" size="small" v-show="ruleRelayTypeContainsUDP(rule)"
                                        :disabled="rule.Options.UDPProxyPerformanceMode == true ? false : true">
                                        性能模式
                                    </el-button>
                                </el-tooltip>



                                <el-tooltip class="box-item" effect="dark"
                                    :content="rule.Options.UDPShortMode == true ? 'UDP转发 shortMode启用' : 'UDP转发 shortMode禁用'"
                                    placement="bottom">
                                    <el-button color="#626aef" size="small" v-show="ruleRelayTypeContainsUDP(rule)"
                                        :disabled="rule.Options.UDPShortMode == true ? false : true">
                                        ShortMode
                                    </el-button>
                                </el-tooltip>



                            </el-descriptions-item>



                            <el-descriptions-item label="目标IP">
                                <el-tooltip class="box-item oneLine" effect="dark" placement="bottom"
                                    :content="isBalanceRelayRule(rule) ? GetBalanceTargetList(rule) : rule.TargetIP">
                                    <el-button color="#D4D7DE" size="small" v-show="true">
                                        {{ isBalanceRelayRule(rule) ? '均衡负载' : rule.TargetIP }}
                                    </el-button>
                                </el-tooltip>
                            </el-descriptions-item>
                            <el-descriptions-item label="目标端口">
                                <el-tooltip class="box-item oneLine" effect="dark" placement="bottom"
                                    :content="isBalanceRelayRule(rule) ? GetBalanceTargetList(rule) : rule.TargetIP">
                                    <el-button color="#D4D7DE" size="small">
                                        {{ isBalanceRelayRule(rule) ? '均衡负载' : rule.TargetPorts }}
                                    </el-button>
                                </el-tooltip>
                            </el-descriptions-item>
                            <el-descriptions-item label="统计" :span="2">

                                <el-tooltip class="box-item" effect="dark" :content="'已接收 ' + GetReceiveTraffic(rule)"
                                    placement="bottom">
                                    <el-button color="#F2F6FC" size="small">
                                        <el-icon>
                                            <Download />
                                        </el-icon> {{ GetReceiveTraffic(rule) }}
                                    </el-button>
                                </el-tooltip>

                                <el-tooltip class="box-item" effect="dark" :content="'已发送 ' + GetSendTraffic(rule)"
                                    placement="bottom">
                                    <el-button color="#F2F6FC" size="small">
                                        <el-icon>
                                            <Upload />
                                        </el-icon> {{ GetSendTraffic(rule) }}
                                    </el-button>
                                </el-tooltip>

                                <el-tooltip class="box-item" effect="dark"
                                    :content="'活跃连接数 ' + GetActivityConnections(rule)" placement="bottom">
                                    <el-button color="#F2F6FC" size="small">
                                        <el-icon>
                                            <Connection />
                                        </el-icon> {{ GetActivityConnections(rule) }}
                                    </el-button>
                                </el-tooltip>

                            </el-descriptions-item>

                        </el-descriptions>




                    </div>









                </el-scrollbar>

                <el-affix position="bottom" :offset="30" class="affix-container">


                    <el-button type="primary" @click="addRule">添加转发规则 <el-icon>
                            <Plus />
                        </el-icon>
                    </el-button>


                </el-affix>

                <!--添加/修改规则对话框-->
                <el-dialog v-model="dialogFormVisible" :title="dialogTitle" draggable :before-close="handleDialogClose"
                    :show-close="false" width="650px">
                    <el-form :model="form">
                        <el-form-item label="名称" :label-width="formLabelWidth">
                            <el-input v-model="form.Name" placeholder="转发规则名称，可留空" autocomplete="off" />
                        </el-form-item>
                        <el-form-item label="转发类型" :label-width="formLabelWidth">
                            <el-checkbox v-model="checkAllRelayType" :indeterminate="isIndeterminate"
                                @change="handleCheckAllChange">全选 &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</el-checkbox>
                            <br />
                            <el-checkbox-group v-model="dialogRelayType" @change="handleCheckedProxyTypesChange">
                                <el-checkbox v-for="t in proxyTypes" :key="t" :label="t">{{
                                        t
                                }}</el-checkbox>
                            </el-checkbox-group>
                        </el-form-item>

                        <el-form-item label="安全模式" :label-width="formLabelWidth">
                            <el-radio-group v-model="form.Options.SafeMode" class="ml-4">
                            <el-radio label="blacklist">黑名单模式</el-radio>
                            <el-radio label="whitelist">白名单模式</el-radio>
    </el-radio-group>
                        </el-form-item>

                        <el-form-item label="负载均衡" :label-width="formLabelWidth">
                            <el-switch v-model="form.IsBalanceRelayType" inline-prompt active-text="开" inactive-text="关"
                                size="large" />
                        </el-form-item>

                        <el-form-item label="监听地址" :label-width="formLabelWidth">
                            <el-input v-model="form.ListenIP" autocomplete="off"
                                placeholder="留空表示监听任意地址,没特殊需求多数情况下留空即可" />
                        </el-form-item>
                        <el-form-item label="监听端口" :label-width="formLabelWidth">
                            <el-input v-model="form.ListenPorts" autocomplete="off"
                                placeholder="多个端口用,号分割,端口范围用-表示,比如 80,22,2000-20010" />
                        </el-form-item>
                        <el-form-item label="目标地址" :label-width="formLabelWidth">

                            <el-input v-model="form.TargetIP" autocomplete="off"
                                v-show="form.IsBalanceRelayType ? false : true" />

                            <el-input v-model="formBalanceTargetAddressList" autosize placeholder="一行填一个地址IP+端口"
                                type="textarea" v-show="form.IsBalanceRelayType ? true : false"></el-input>


                        </el-form-item>
                        <el-form-item label="目标端口" :label-width="formLabelWidth"
                            v-show="form.IsBalanceRelayType ? false : true">
                            <el-input v-model="form.TargetPorts" autocomplete="off" placeholder="监听端口的数量和目标端口的数量要一致" />
                        </el-form-item>
                        <el-form-item label="单端口最大并发数" :label-width="formLabelWidth">
                            <el-input-number v-model="form.Options.SingleProxyMaxConnections" :min="1" :max="65535" />
                        </el-form-item>
                        <el-form-item label="UDP最大包长度" :label-width="formLabelWidth" v-show="udpOptionsShow">
                            <el-input-number v-model="form.Options.UDPPackageSize" :min="1" :max="99999" />
                            &nbsp;&nbsp;&nbsp;
                            <el-tooltip class="box-item" effect="dark" content="通过增加协程数至与CPU核数一致提高udp小包转发性能">
                                <el-checkbox label="UDP转发性能模式" v-model="form.Options.UDPProxyPerformanceMode" />
                            </el-tooltip>
                            &nbsp;&nbsp;&nbsp;
                            <el-tooltip class="box-item" effect="dark" content="DNS转发需要打开这个开关以节约资源,其它场景自行测试">
                                <el-checkbox label="UDP short mode" v-model="form.Options.UDPShortMode" />
                            </el-tooltip>
                        </el-form-item>


                    </el-form>
                    <template #footer>
                        <span class="dialog-footer">
                            <el-button @click="dialogFormVisible = false">取消</el-button>
                            <el-button type="primary" @click="addOrAlterRuleConfirm">{{ confirmText }}</el-button>
                        </span>
                    </template>
                </el-dialog>






    </div>


</template>

<script lang="ts" setup>

import { apiGetRuleList, apiAddRule, apiDeleteRule, apiAlterRule, apiRuleEnable } from '../apis/utils'
import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'

//import { ElMessage } from 'element-plus'
import {CopyTotoClipboard} from '../utils/utils'
import { ElNotification } from 'element-plus'
import { ElMessageBox } from 'element-plus'


import {MessageShow} from '../utils/ui'
import {isIP} from '../utils/utils'


//var timerID:any
const size = ref('')
const dialogFormVisible = ref(false)
const dialogRelayType = ref([''])
const formLabelWidth = '130px'
const form = ref({
    Name: "",
    ListenIP: '',
    ListenPorts: '',
    Mainconfigure: '',
    TargetIP: '',
    TargetPorts: '',
    RelayType: "",
    IsBalanceRelayType: false,
    BalanceTargetAddressList: [''],
    Options: {
        SingleProxyMaxConnections: 128,
        UDPPackageSize: 1500,
        UDPShortMode: true,
        UDPProxyPerformanceMode: true,
        SafeMode:"blacklist",
    }

})

const formBalanceTargetAddressList = ref("")

//存储修改前的数据
const preAlterRule = ref({
    Name: "",
    ListenIP: '',
    ListenPorts: '',
    Mainconfigure: '',
    TargetIP: '',
    TargetPorts: '',
    RelayType: "",
    IsBalanceRelayType: false,
    BalanceTargetAddressList: [''],
    Options: {
        SingleProxyMaxConnections: 128,
        UDPPackageSize: 1500,
        UDPShortMode: true,
        UDPProxyPerformanceMode: true,
         SafeMode:"blacklist",
    }

})




const formOptionType = ref("")

const dialogTitle = ref("")
const confirmText = ref("")

const udpOptionsShow = ref(false)
const balanceRelayOptionsShow = ref(false)

var ruleList = ref([{
    Name: "",
    Mainconfigure: "",
    RelayType: "",
    RelayTypeList: [''],
    ListenIP: "",
    ListenPorts: "",
    TargetIP: "",
    TargetPorts: "",
    Enable: false,
    BalanceTargetAddressList: [''],
    Options: {
        UDPPackageSize: 1500,
        SingleProxyMaxConnections: 128,
        UDPProxyPerformanceMode: true,
        UDPShortMode: true,
        SafeMode:"blacklist",
    },

}]);
ruleList.value.splice(0, 1)


const handleDialogClose = (done: () => void) => {

}

const ruleEnableClick = (enable, rule) => {
    const enableText = enable == false ? "启用" : "禁用";

    const ruleName = "[" + rule.Name + "]"
    const configure = "[" + rule.Mainconfigure + "]"

    return new Promise((resolve, reject) => {

        ElMessageBox.confirm(
            '确认要' + enableText + "规则 " + ruleName + " " + configure + "?",
            'Warning',
            {
                confirmButtonText: '确认',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
            .then(() => {
                apiRuleEnable(rule.Mainconfigure, !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "规则 " + ruleName + " " + configure + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("success", "规则 " + ruleName + " " + configure + enableText + "失败")

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                }).catch((error) => {
                    resolve(false)
                    console.log("规则 " + ruleName + " " + configure + enableText + "失败" + ":请求出错" + error)
                    MessageShow("success", "规则 " + ruleName + " " + configure + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}


const addRule = () => {

    console.log("addRule")
    formOptionType.value = "add"
    form.value = {
        Name: "",
        ListenIP: '',
        Mainconfigure: '',
        ListenPorts: '',
        TargetIP: '',
        TargetPorts: '',
        RelayType: 'tcp6',
        IsBalanceRelayType: false,
        BalanceTargetAddressList: [],
        Options: {
            SingleProxyMaxConnections: 256,
            UDPPackageSize: 1500,
            UDPShortMode: false,
            UDPProxyPerformanceMode: false,
            SafeMode:"blacklist",
        }
    }
    formBalanceTargetAddressList.value = ""
    flushUdpOptionsView()

    //checkedProxyTypes.value = form.value.proxyType
    dialogTitle.value = "添加转发规则"
    dialogRelayType.value = getRelayTypeList(form.value.RelayType)
    dialogFormVisible.value = true;
    confirmText.value = "添加"
}

const converAddressListTextToList = (listStr: string) => {
    let rawlist = listStr.split("\n")
    let resList = new Array()
    for (let i in rawlist) {
        resList.push(rawlist[i].replace(/^\s+|\s+$/g, '').replace(/<\/?.+?>/g, "").replace(/[\r\n]/g, ""))
    }
    return resList
}

const addOrAlterRuleConfirm = () => {

    if (!form.value.IsBalanceRelayType) {
        form.value.BalanceTargetAddressList = []
    } else {
        form.value.BalanceTargetAddressList = converAddressListTextToList(formBalanceTargetAddressList.value)//formBalanceTargetAddressList.value.split(",")
    }

    if (!checkFormData()) {
        return
    }
    console.log("表单数据检测通过")




    switch (formOptionType.value) {
        case "add":
            apiAddRule(form.value).then((res) => {
                if (res.ret == 0) {
                    dialogFormVisible.value = false;
                    MessageShow("success", "规则添加成功")
                    console.log("刷新规则列表")
                    queryRuleList();

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("添加规则失败,网络请求出错:" + error)
                MessageShow("error", "添加规则失败,网络请求出错")
            })

            break;
        case "alter":

            if (!checkIsAlterRule()) {
                MessageShow("warning", "转发规则没有修改")
                break;
            }

            apiAlterRule(form.value).then((res) => {
                if (res.ret == 0) {
                    dialogFormVisible.value = false;
                    MessageShow("success", "规则修改成功")
                    queryRuleList();
                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("规则修改失败,网络请求出错:" + error)
                MessageShow("error", "规则修改失败,网络请求出错")
            })
            break;
        default:
            console.log("unsupport formOption")
    }

    // console.log(formOptionType.value+" "+)
    //dialogFormVisible.value = false;


}

const isNumber = (val) => {
    var regPos = /^[0-9]+.?[0-9]*/; //判断是否是数字。
    if (regPos.test(val)) {
        return true;
    }
    return false;
}

const checkFormData = () => {

    if (form.value.RelayType.length <= 0) {
        showMessageBox("转发类型至少选择一类")
        return false
    }

    if (form.value.ListenIP.length != 0 && !isIP(form.value.ListenIP)) {
        showMessageBox("监听地址IP[" + form.value.ListenIP + "]格式有误")
        return false
    }



    if (!form.value.IsBalanceRelayType && form.value.ListenPorts.length == 0) {
        showMessageBox("监听端口不能为空,端口和端口或者端口和端口范围之间用英文状态逗号,分开"
            + '\n'
            + '例如 [22,80,443,2000-2010]')
        return false
    }

    if (form.value.IsBalanceRelayType && form.value.ListenPorts.length == 0) {
        showMessageBox("均衡模式 监听端口不能为空")
        return false
    }

    if (form.value.IsBalanceRelayType && !isNumber(form.value.ListenPorts)) {
        showMessageBox("均衡模式 监听端口只能填一个")
        return false
    }

    if (form.value.IsBalanceRelayType && form.value.ListenPorts.length == 0) {
        showMessageBox("均衡模式 监听端口不能为空")
        return false
    }


    if (form.value.IsBalanceRelayType && (form.value.BalanceTargetAddressList == null || form.value.BalanceTargetAddressList.length == 0)) {
        showMessageBox("均衡模式 目标地址不能为空")
        return false
    }



    if (!form.value.IsBalanceRelayType && form.value.TargetIP.length == 0) {
        showMessageBox("目标地址不能为空")
        return false
    }

    if (!form.value.IsBalanceRelayType && !isIP(form.value.TargetIP)) {
        showMessageBox("目标地址IP[" + form.value.TargetIP + "]格式有误")
        return false
    }

    if (!form.value.IsBalanceRelayType && form.value.TargetPorts.length == 0) {
        showMessageBox("目标端口不能为空")
        return false
    }

    return true
}

const showMessageBox = (message) => {
    ElMessageBox.alert(message, {
        confirmButtonText: '好的',
        callback: () => {
        },
    })
}


const converBalanceTargetAddressListToInput = (list) => {
    var res = ""
    for (let i in list) {
        if (res.length == 0) {
            res = list[i]
        } else {
            res += "\n" + list[i]
        }
    }
    return res
}

const alterRule = (rule) => {
    console.log("alterRule " + rule)
    formOptionType.value = "alter"


    form.value = {
        Name: rule.Name,
        Mainconfigure: rule.Mainconfigure,
        ListenIP: rule.ListenIP,
        ListenPorts: rule.ListenPorts,
        TargetIP: rule.TargetIP,
        TargetPorts: rule.TargetPorts,
        RelayType: convertRelayTypeDetail(rule.RelayType),
        IsBalanceRelayType: isBalanceRelayRule(rule),
        BalanceTargetAddressList: rule.BalanceTargetAddressList,
        Options: {
            SingleProxyMaxConnections: rule.Options.SingleProxyMaxConnections,
            UDPPackageSize: rule.Options.UDPPackageSize,
            UDPShortMode: rule.Options.UDPShortMode == true ? true : false,
            UDPProxyPerformanceMode: rule.Options.UDPProxyPerformanceMode == true ? true : false,
            SafeMode:rule.Options.SafeMode,
        }
    }

    formBalanceTargetAddressList.value = converBalanceTargetAddressListToInput(rule.BalanceTargetAddressList)

    preAlterRule.value = {
        Name: rule.Name,
        Mainconfigure: rule.Mainconfigure,
        ListenIP: rule.ListenIP,
        ListenPorts: rule.ListenPorts,
        TargetIP: rule.TargetIP,
        TargetPorts: rule.TargetPorts,
        RelayType: convertRelayTypeDetail(rule.RelayType),
        IsBalanceRelayType: isBalanceRelayRule(rule),
        BalanceTargetAddressList: rule.BalanceTargetAddressList,
        Options: {
            SingleProxyMaxConnections: rule.Options.SingleProxyMaxConnections,
            UDPPackageSize: rule.Options.UDPPackageSize,
            UDPShortMode: rule.Options.UDPShortMode == true ? true : false,
            UDPProxyPerformanceMode: rule.Options.UDPProxyPerformanceMode == true ? true : false,
            SafeMode:rule.Options.SafeMode,
        }
    }

    dialogRelayType.value = getRelayTypeList(form.value.RelayType)
    flushUdpOptionsView()


    dialogTitle.value = "编辑转发规则"
    dialogFormVisible.value = true;
    confirmText.value = "确认修改"
}

const addressListIsEqual = (list1, list2: string[]) => {

    if (list1 == null && list2 == null) {
        return true
    }

    // if ((list1==null && list2!=null)|| (list1!=null && list2==null)){
    //     return false
    // }

    if ((list1 == null || list1.length == 0) && (list2 == null || list2.length == 0)) {
        return true
    }

    if (list1.length != list2.length) {
        return false
    }
    for (let i in list1) {
        if (list1[i] != list2[i]) {
            return false
        }
    }
    return true
}

const checkIsAlterRule = () => {
    if (form.value.Name == preAlterRule.value.Name
        && form.value.ListenIP == preAlterRule.value.ListenIP
        && form.value.ListenPorts == preAlterRule.value.ListenPorts
        && form.value.TargetIP == preAlterRule.value.TargetIP
        && form.value.TargetPorts == preAlterRule.value.TargetPorts
        && form.value.RelayType == preAlterRule.value.RelayType
        && form.value.Options.SingleProxyMaxConnections == preAlterRule.value.Options.SingleProxyMaxConnections
        && form.value.Options.UDPPackageSize == preAlterRule.value.Options.UDPPackageSize
        && form.value.Options.UDPShortMode == preAlterRule.value.Options.UDPShortMode
        && form.value.Options.UDPProxyPerformanceMode == preAlterRule.value.Options.UDPProxyPerformanceMode
        && form.value.Options.SafeMode == preAlterRule.value.Options.SafeMode
        && addressListIsEqual(form.value.BalanceTargetAddressList, preAlterRule.value.BalanceTargetAddressList)) {
        return false
    }
    return true
}

const deleteRule = (rule) => {

    const ruleName = "[" + rule.Name + "]"
    const configure = "[" + rule.Mainconfigure + "]"
    const ruleText = ruleName + " " + configure

    ElMessageBox.confirm(
        '确认要删除转发规则 ' + ruleText + "?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            console.log("确认删除 " + ruleText)

            apiDeleteRule(rule.Mainconfigure).then((res) => {
                if (res.ret == 0) {
                    console.log("删除成功")
                    queryRuleList();
                    MessageShow("success", res.msg)
                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }

                } else {
                    MessageShow("error", res.msg)
                }

            }).catch((error) => {
                console.log("删除规则失败,网络请求出错:" + error)
                MessageShow("error", "删除规则失败,网络请求出错")
            })
        })
        .catch(() => {

        })
}

const test = () => {
    return "aaa"
}

const getRuleType = (rule: string) => {
    const temp = rule.split('@');
    return temp[0];
}

const ruleRelayTypeContainsUDP = (rule) => {
    return rule.RelayType.indexOf("udp") != -1 ? true : false;
}



const GetSendTraffic = (rule) => {
    let ruleTrafficCount = 0
    for (let i in rule.ProxyList) {
        ruleTrafficCount += rule.ProxyList[i].TrafficOut
    }
    return bytesToSize(ruleTrafficCount)
}

const GetReceiveTraffic = (rule) => {
    let ruleTrafficCount = 0
    for (let i in rule.ProxyList) {
        ruleTrafficCount += rule.ProxyList[i].TrafficIn
    }
    return bytesToSize(ruleTrafficCount)
}

const GetActivityConnections = (rule) => {
    let count = 0
    for (let i in rule.ProxyList) {
        count += rule.ProxyList[i].CurrentConnections
    }
    return count
}

const bytesToSize = (bytes) => {
    if (bytes === 0) return '0 B';
    var k = 1000, // or 1024
        sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
        i = Math.floor(Math.log(bytes) / Math.log(k));

    return (bytes / Math.pow(k, i)).toPrecision(3) + ' ' + sizes[i];
}

const queryRuleList = () => {
    apiGetRuleList().then((res) => {
        //console.log(res.data)
        ruleList.value = res.data
    }).catch((error) => {
        console.log("获取转发规则列表出错:" + error)
        MessageShow("error", "获取转发规则列表出错")
    })
}

const copyRelayConfigure = (configure: string) => {
    CopyTotoClipboard(configure)
    MessageShow('success', '配置 ' + configure + ' 已复制到剪切板')
}

// const MessageShow = (type, message) => {
//     ElMessage({
//         message: message,
//         type: type,
//     })
// }

const Notification = (type, message, duration) => {
    ElNotification({
        title: type.substring(0, 1).toUpperCase() + type.substring(1),
        message: message,
        type: type,
        duration: duration,
    })
}

const getRelayTypeList = (releyType: string) => {
    var rawList = releyType.split(",")
    var resList = ['']

    for (let i in rawList) {
        if (rawList[i] == "udp") {
            resList.push("udp4")
            resList.push("udp6")
            continue
        }
        if (rawList[i] == "tcp") {
            resList.push("tcp4")
            resList.push("tcp6")
            continue
        }
        resList.push(rawList[i])
    }
    resList.shift()

    return resList
}

//convertRelayTypeDetail 将tcp分解成tcp4,tcp6, udp分解成udp4,udp6
const convertRelayTypeDetail = (releyType: string) => {
    let resList = getRelayTypeList(releyType)
    var res = ""
    for (let i in resList) {
        if (res.length == 0) {
            res = resList[i]
        } else {
            res += "," + resList[i]
        }
    }
    return res
}


// const queryRuleList = () =>{
//     apiGetRuleList().then((res) => {
//         console.log(res.data);

// })

//--------------------------
const checkAllRelayType = ref(false)
const isIndeterminate = ref(true)
//const checkedProxyTypes = ref(['tcp4', 'tcp6'])
const proxyTypes = ['tcp4', 'tcp6', 'udp4', 'udp6']
const handleCheckAllChange = (val: boolean) => {
    // checkedProxyTypes.value = val ? proxyTypes : []
    form.value.RelayType = val ? 'tcp4,tcp6,udp4,udp6' : ''
    dialogRelayType.value = val ? ['tcp4', 'tcp6', 'udp4', 'udp6'] : []
    isIndeterminate.value = false
    flushUdpOptionsView()
    // console.log("proxyType 全选: "+val)
    // console.log("proxyType: "+form.value.proxyType)
}
const handleCheckedProxyTypesChange = (value: string[]) => {
    const checkedCount = value.length
    checkAllRelayType.value = checkedCount === proxyTypes.length
    isIndeterminate.value = checkedCount > 0 && checkedCount < proxyTypes.length


    form.value.RelayType = getRelayTypeByList(value)
    flushUdpOptionsView()
    // console.log("proxyType: "+form.value.proxyType)
}

const sortArray = (array) => {
    array.sort((a, b) => {
        if (a.length !== b.length) {
            return a.length - b.length
        } else {
            return a.localeCompare(b);
        }
    })
    return array
}

const getRelayTypeByList = (list: string[]) => {
    var res = ""
    //list = sortArray(list)
    for (let s of list) {
        if (res.length == 0) {
            res = s
            continue
        }
        res += "," + s
    }
    return res
}


const isBalanceRelayRule = (rule) => {
    return rule.BalanceTargetAddressList != null && rule.BalanceTargetAddressList.length > 0 ? true : false;
}

const GetBalanceTargetList = (rule) => {
    if (rule.BalanceTargetAddressList == null) {
        return ""
    }
    var res = ""
    for (let i in rule.BalanceTargetAddressList) {
        if (res.length == 0) {
            res = rule.BalanceTargetAddressList[i]
        } else {
            res += ' | ' + rule.BalanceTargetAddressList[i]
        }
    }
    return res
}


//--------------------------

var timerID: any

onMounted(() => {
    queryRuleList();
    console.log("relaySet onmounted")

    queryRuleList();
    timerID = setInterval(() => {
        queryRuleList();
    }, 1000);

})

onUnmounted(() => {
    clearInterval(timerID)
})

var flushUdpOptionsView = () => {

    let list = getRelayTypeList(form.value.RelayType)

    for (let t of list) {
        if (t == "udp4" || t == "udp6") {
            udpOptionsShow.value = true
            return
        }
    }
    udpOptionsShow.value = false;
}

var flushBalanceOptionsView = () => {
    if (form.value.IsBalanceRelayType) {
        balanceRelayOptionsShow.value = true
        return
    }
    balanceRelayOptionsShow.value = false
}






</script>



<style scoped>
.itemradius {

    border: 1px solid var(--el-border-color);
    border-radius: 0;
    margin-left: 3px;
    margin-top: 3px;
    margin-right: 3px;
    margin-bottom: 25px;
    width:1200px;
}





.affix-container {
    text-align: center;
    border-radius: 4px;
    width: 3vw;

    background: var(--el-color-primary-light-9);
}
</style>

