<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <el-scrollbar height="100%">
            <div class="itemradius" :style="{
                borderRadius: 'base',
            }" v-for="rule in ruleList">


                <el-descriptions :column="4" border>

                    <el-descriptions-item label="规则名称">

                        <el-button  size="default" v-show="true">
                            {{ rule.Name == '' ? '未命名规则' : rule.Name }}
                        </el-button>

                    </el-descriptions-item>
                    <el-descriptions-item label="转发类型">

                        <el-button color="#0059b3" size="small" v-show="true" v-for="t in rule.ForwardTypes">{{ t }}
                        </el-button>

                    </el-descriptions-item>
                    <el-descriptions-item label="操作" :span="2">
                        <el-tooltip :content="rule.Enable == true ? '规则已启用' : '规则已禁用'" placement="top">
                            <el-switch v-model="rule.Enable" inline-prompt active-text="开" inactive-text="关"
                                :before-change="ruleEnableClick.bind(this, rule.Enable, rule)" size="large" />
                        </el-tooltip>
                        &nbsp;&nbsp;
                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="proxyLogsArrayToTooltipHtml(rule)"></span>
                            </template>
                            <el-button type="info" plain size="small" v-show="true" @click="showProxyLogs(rule.Key)">
                                日志
                            </el-button>
                        </el-tooltip>

                        <el-button type="primary" @click="alterRule(rule)">编辑</el-button>
                        <el-button type="danger" @click="deleteRule(rule)">删除</el-button>

                    </el-descriptions-item>



                    <el-descriptions-item label="监听IP">

                        <el-button color="#D4D7DE" size="small" v-show="true">
                            {{ rule.ListenAddress == '' ? '所有IP' : rule.ListenAddress }}
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


                        <el-tooltip class="box-item" effect="dark" content="单端口TCP最大连接数" placement="bottom">
                            <el-button color="#6666ff" size="small" v-show="true">{{
                            rule.Options.SingleProxyMaxTCPConnections
                            }}</el-button>
                        </el-tooltip>

                        <el-tooltip class="box-item" effect="dark" content="UDP包最大长度" placement="bottom">
                            <el-button color="#626aef" size="small" v-show="ruleForwardTypeContainsUDP(rule)">{{
                            rule.Options.UDPPackageSize
                            }}</el-button>
                        </el-tooltip>

                        <el-tooltip class="box-item" effect="dark"
                            :content="rule.Options.UDPProxyPerformanceMode == true ? 'UDP性能模式开启' : '性能模式关闭'"
                            placement="bottom">
                            <el-button color="#626aef" size="small" v-show="ruleForwardTypeContainsUDP(rule)"
                                :disabled="rule.Options.UDPProxyPerformanceMode == true ? false : true">
                                性能模式
                            </el-button>
                        </el-tooltip>



                        <el-tooltip class="box-item" effect="dark"
                            :content="rule.Options.UDPShortMode == true ? 'UDP转发 shortMode启用' : 'UDP转发 shortMode禁用'"
                            placement="bottom">
                            <el-button color="#626aef" size="small" v-show="ruleForwardTypeContainsUDP(rule)"
                                :disabled="rule.Options.UDPShortMode == true ? false : true">
                                ShortMode
                            </el-button>
                        </el-tooltip>



                    </el-descriptions-item>



                    <el-descriptions-item label="目标地址">
                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="StrArrayListToBrHtml(rule.TargetAddressList)"></span>
                            </template>
                            <el-button color="#409eff" size="small" v-show="true">
                                {{
                                (rule.TargetAddressList==undefined||rule.TargetAddressList==null)||rule.TargetAddressList.length
                                <=0?'未设置':rule.TargetAddressList.length==1?rule.TargetAddressList[0]:rule.TargetAddressList[0]+'...'}}
                                    </el-button>
                        </el-tooltip>
                    </el-descriptions-item>
                    <el-descriptions-item label="目标端口">
                        <el-tooltip class="box-item oneLine" effect="dark" placement="bottom"
                            :content="rule.TargetPorts">
                            <el-button color="#D4D7DE" size="small">
                                {{rule.TargetPorts }}
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
                            :content="'当前TCP连接数 ' + GetTCPActivityConnections(rule)" placement="bottom">
                            <el-button v-if="isTCPRule(rule)" color="#F2F6FC" size="small">
                                TCP&nbsp; &nbsp;
                                <el-icon>
                                    <Connection />
                                </el-icon> {{ GetTCPActivityConnections(rule) }}
                            </el-button>
                        </el-tooltip>

                        <el-tooltip class="box-item" effect="dark"
                            :content="'UDP转发目标地址数据协程数 ' + GetUDPActivityConnections(rule)" placement="bottom">
                            <el-button v-if="isUDPRule(rule)" color="#F2F6FC" size="small">
                                UDP&nbsp; &nbsp;
                                <el-icon>
                                    <Connection />
                                </el-icon> {{ GetUDPActivityConnections(rule) }}
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
        <el-dialog v-model="dialogFormVisible" :title="dialogTitle" draggable :show-close="false"
            :close-on-click-modal="false" width="650px">
            <el-form :model="formData">
                <el-form-item label="名称" :label-width="formLabelWidth">
                    <el-input v-model="formData.Name" placeholder="转发规则名称，可留空" autocomplete="off" />
                </el-form-item>
                <el-form-item label="转发类型" :label-width="formLabelWidth">
                    <el-checkbox v-model="checkAllForwardType" :indeterminate="isIndeterminate"
                        @change="handleCheckAllChange">全选 &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</el-checkbox>
                    <br />
                    <el-checkbox-group v-model="dialogForwardTypes" @change="handleCheckedForwardTypesChange">
                        <el-checkbox v-for="t in forwardTypes" :key="t" :label="t">{{
                        t
                        }}</el-checkbox>
                    </el-checkbox-group>
                </el-form-item>

                <el-form-item label="安全模式" :label-width="formLabelWidth">
                    <el-radio-group v-model="formData.Options.SafeMode" class="ml-4">
                        <el-radio label="blacklist">黑名单模式</el-radio>
                        <el-radio label="whitelist">白名单模式</el-radio>
                    </el-radio-group>
                </el-form-item>


                <el-form-item label="监听地址" :label-width="formLabelWidth">
                    <el-input v-model="formData.ListenAddress" autocomplete="off"
                        placeholder="留空表示监听任意地址,没特殊需求多数情况下留空即可" />
                </el-form-item>
                <el-form-item label="监听端口" :label-width="formLabelWidth">
                    <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                        <template #content>
                            多个端口用,号分割,端口范围用-表示,比如 80,22,2000-20010<br />
                            表示监听80端口,22端口,20000-20010范围内的11个端口<br />
                        </template>

                        <el-input v-model="formData.ListenPorts" autocomplete="off"
                            placeholder="多个端口用,号分割,端口范围用-表示,比如 80,22,20000-20010" />
                    </el-tooltip>
                </el-form-item>
                <el-form-item label="目标地址" :label-width="formLabelWidth">

                    <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                        <template #content>
                            没特殊需要填一行即可<br />
                            多行表示启用均衡负载,每次选择不同的目标地址<br />
                            可填写IP或者域名<br />
                        </template>

                        <el-input v-model="formTargetAddressListArea" autosize placeholder="一行填一个地址IP" type="textarea">
                        </el-input>
                    </el-tooltip>


                </el-form-item>
                <el-form-item label="目标端口" :label-width="formLabelWidth">
                    <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                        <template #content>
                            多个端口用,号分割,端口范围用-表示,比如 80,22,2000-20010<br />
                            表示监听80端口,22端口,20000-20010范围内的11个端口<br />
                        </template>

                        <el-input v-model="formData.TargetPorts" autocomplete="off" placeholder="监听端口的数量和目标端口的数量要一致" />
                    </el-tooltip>
                </el-form-item>

                <div v-show="tcpOptionsShow" class="dialogRadius">

                    <el-form-item label="单端口TCP最大连接数" :label-width="formLabelWidth">
                        <el-input-number v-model="formData.Options.SingleProxyMaxTCPConnections" :min="1"
                            :max="65535" />
                    </el-form-item>

                </div>

                <div v-show="udpOptionsShow" class="dialogRadius">

                    
                    <el-form-item label="单端口UDP读取目标地址数据协程数限制" label-width="auto" v-show="udpOptionsShow">
                        <el-input-number v-model="formData.Options.SingleProxyMaxUDPReadTargetDatagoroutineCount" :min="0" :max="4096" />
                        &nbsp;&nbsp;&nbsp;
                    </el-form-item>


                    <el-form-item label="UDP最大包长度" :label-width="formLabelWidth" v-show="udpOptionsShow">
                        <el-input-number v-model="formData.Options.UDPPackageSize" :min="1" :max="99999" />
                        &nbsp;&nbsp;&nbsp;
                    </el-form-item>

                    <el-form-item label="&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;">
                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                通过增加协程数至与CPU核数一致提高udp小包转发性能<br />
                                一般情况不要开启<br />
                            </template>
                            <el-checkbox label="UDP转发性能模式" v-model="formData.Options.UDPProxyPerformanceMode" />
                        </el-tooltip>
                    </el-form-item>


                    <el-form-item label="&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;">
                        <el-tooltip class="box-item" effect="dark" content="DNS转发需要打开这个开关以节约资源,其它场景自行测试">
                            <el-checkbox label="UDP short mode" v-model="formData.Options.UDPShortMode" />
                        </el-tooltip>
                    </el-form-item>
                </div>







                <el-form-item label="日志输出级别" :label-width="formLabelWidth">
                    <el-select v-model="formData.LogLevel" class="m-2" placeholder="请选择">
                        <el-option v-for="item in LogLevelList" :key="item.value" :label="item.label"
                            :value="item.value" />
                    </el-select>
                </el-form-item>

                <el-form-item label="日志输出到终端" :label-width="formLabelWidth">
                    <el-switch v-model="formData.LogOutputToConsole" inline-prompt width="50px" active-text="开启"
                        inactive-text="关闭" />
                </el-form-item>

                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                    <template #content>
                        范围(0-102400),0表示不保存日志<br />
                    </template>
                    <el-form-item label="访问日志记录最大条数" :label-width="formLabelWidth" :min="0" :max="102400">
                        <el-input-number v-model="formData.AccessLogMaxNum" autocomplete="off" />
                    </el-form-item>
                </el-tooltip>

                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                    <template #content>
                        范围(1-64)<br />
                    </template>
                    <el-form-item label="前端列表显示最新日志最大条数" :label-width="formLabelWidth" :min="1" :max="64">
                        <el-input-number v-model="formData.WebListShowLastLogMaxCount" autocomplete="off" />
                    </el-form-item>
                </el-tooltip>




            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogFormVisible = false">取消</el-button>
                    <el-button type="primary" @click="addOrAlterRuleConfirm">{{ confirmText }}</el-button>
                </span>
            </template>
        </el-dialog>


        <el-dialog v-if="portforwardLogsVisible" v-model="portforwardLogsVisible" :close-on-click-modal="false"
            width="1100px">

            <div>

                <el-scrollbar max-height="95vh" class="portforwardLogs" element-loading-background="transparent">
                    {{portforwardLogsDialogLogsContentView}}

                </el-scrollbar>


                <el-pagination :page-size=portforwardLogsPageSize :page-sizes="[10,25,50, 100, 200, 300,400,500]"
                    :small="false" :disabled="false" :background="false"
                    layout="total, sizes, prev, pager, next, jumper" :current-page="portforwardLogsDialogCurrentPage"
                    :total=portforwardLogsTotal @size-change="handlePortforwardLogsSizeChange"
                    @current-change="handlePortforwardLogsCurrentChange" @prev-click="handlePortforwardLogsPreClick"
                    @next-click="handlePortforwardLogsNextClick" />
            </div>

        </el-dialog>

    </div>
</template>

<script lang="ts" setup>

import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'
import { CheckboxValueType, ElNotification } from 'element-plus'
import { isIP, StringToArrayList, bytesToSize, StrArrayListToBrHtml, StrArrayListToArea, LogLevelList } from '../utils/utils'
import { ShowMessageBox, MessageShow, Notification } from '../utils/ui'
import {
    apiGetPortForwardRuleList,
    apiAddPortForwardRule,
    apiDeletePortForwardRule,
    apiAlterPortForwardRule,
    apiPortForwardRuleEnable,
    apiPortforwardRuleLogs
} from '../apis/utils'
import { ElMessageBox } from 'element-plus'


const portforwardLogsPageSize = ref(50)
const portforwardLogsTotal = ref(0)
const portforwardLogsDialogData = ref([{
    ProxyKey: "",
    ClientIP: "",
    LogContent: "",
    LogTime: ""
}])
const portforwardLogsDialogKey = ref("")
const portforwardLogsDialogCurrentPage = ref(1)
const portforwardLogsVisible = ref(false)
const portforwardLogsDialogLogsContentView = ref("")



const checkAllForwardType = ref(false)
const isIndeterminate = ref(true)
const dialogForwardTypes = ref([''])
const formTargetAddressListArea = ref('')
const udpOptionsShow = ref(false)
const tcpOptionsShow = ref(false)



const forwardTypes = ['tcp4', 'tcp6', 'udp4', 'udp6']
const handleCheckAllChange = (val: CheckboxValueType) => {
    // checkedProxyTypes.value = val ? proxyTypes : []
    formData.value.ForwardTypes = val ? ['tcp', 'udp'] : []
    dialogForwardTypes.value = val ? ['tcp4', 'tcp6', 'udp4', 'udp6'] : []
    isIndeterminate.value = false
    flushUdpOptionsView()
    flushTcpOptionsView()
    // console.log("proxyType 全选: "+val)
    // console.log("proxyType: "+form.value.proxyType)
}

const handleCheckedForwardTypesChange = (value: CheckboxValueType[]) => {
    const checkedCount = value.length
    checkAllForwardType.value = checkedCount === forwardTypes.length
    isIndeterminate.value = checkedCount > 0 && checkedCount < forwardTypes.length
    formData.value.ForwardTypes = convertCheckedDataToForwardType(value)

    // console.log(formData.value.ForwardTypes)
    flushUdpOptionsView()
    flushTcpOptionsView()
}

const GetReceiveTraffic = (rule) => {
    let ruleTrafficCount = 0
    for (let i in rule.ProxyList) {
        ruleTrafficCount += rule.ProxyList[i].TrafficIn
    }
    return bytesToSize(ruleTrafficCount)
}

const GetSendTraffic = (rule) => {
    let ruleTrafficCount = 0
    for (let i in rule.ProxyList) {
        ruleTrafficCount += rule.ProxyList[i].TrafficOut
    }
    return bytesToSize(ruleTrafficCount)
}

const isTCPRule = (rule) => {
    for (let i in rule.ForwardTypes) {

        if (rule.ForwardTypes[i].indexOf("tcp") == 0) {
            return true
        }
    }
    return false
}

const isUDPRule = (rule) => {
    for (let i in rule.ForwardTypes) {
        if (rule.ForwardTypes[i].indexOf("udp") == 0) {
            return true
        }
    }
    return false
}

const GetTCPActivityConnections = (rule) => {
    let count = 0
    for (let i in rule.ProxyList) {
        if (rule.ProxyList[i].Proxy.indexOf("tcp") != 0) {
            continue
        }
        count += rule.ProxyList[i].CurrentConnections
    }
    return count
}

const GetUDPActivityConnections = (rule) => {
    let count = 0
    for (let i in rule.ProxyList) {
        if (rule.ProxyList[i].Proxy.indexOf("udp") != 0) {
            continue
        }
        count += rule.ProxyList[i].CurrentConnections
    }
    return count
}

const convertForwardTypeToCheckedData = (list: string[]) => {
    var res = ['']
    res.splice(0, 1)
    for (let t of list) {
        if (t == "tcp") {
            res.push('tcp4')
            res.push('tcp6')
            continue
        }
        if (t == "udp") {
            res.push('udp4')
            res.push('udp6')
            continue
        }

        res.push(t + '')
    }
    return res
}

const ruleForwardTypeContainsUDP = (rule) => {
    for (let i in rule.ForwardTypes) {
        if (rule.ForwardTypes[i].indexOf("udp") != -1) {
            return true
        }
    }
    return false
}

const alterRule = (rule) => {
    console.log("alterRule " + rule)
    formOptionType.value = "alter"


    formData.value = {
        Name: rule.Name,
        Key: rule.Key,
        ListenAddress: rule.ListenAddress,
        ListenPorts: rule.ListenPorts,
        TargetAddressList: rule.TargetAddressList,
        TargetPorts: rule.TargetPorts,
        ForwardTypes: rule.ForwardTypes,
        Enable: rule.Enable,
        LogLevel: rule.LogLevel,
        LogOutputToConsole: rule.LogOutputToConsole,
        AccessLogMaxNum: rule.AccessLogMaxNum,
        WebListShowLastLogMaxCount: rule.WebListShowLastLogMaxCount,
        Options: {
            SingleProxyMaxUDPReadTargetDatagoroutineCount: rule.Options.SingleProxyMaxUDPReadTargetDatagoroutineCount,
            SingleProxyMaxTCPConnections: rule.Options.SingleProxyMaxTCPConnections,
            UDPPackageSize: rule.Options.UDPPackageSize,
            UDPShortMode: rule.Options.UDPShortMode == true ? true : false,
            UDPProxyPerformanceMode: rule.Options.UDPProxyPerformanceMode == true ? true : false,
            SafeMode: rule.Options.SafeMode,
        }
    }

    dialogForwardTypes.value = convertForwardTypeToCheckedData(rule.ForwardTypes)
    formTargetAddressListArea.value = StrArrayListToArea(rule.TargetAddressList)




    // dialogRelayType.value = getRelayTypeList(form.value.RelayType)
    flushUdpOptionsView()
    flushTcpOptionsView()

    dialogTitle.value = "编辑转发规则"
    dialogFormVisible.value = true;
    confirmText.value = "确认修改"
}


const proxyLogsArrayToTooltipHtml = (rule) => {
    var res = ""
    if (rule.LastLogs == undefined || rule.LastLogs.length == 0) {
        res = "暂无日志"
        return res
    }

    for (var i in rule.LastLogs) {
        let time = rule.LastLogs[i].LogTime
        let content = rule.LastLogs[i].LogContent
        res += time + "&nbsp;&nbsp;&nbsp;" + content + '<br />'
    }


    if (res == "") {
        res = "暂无日志"
    }

    return res
}


const showProxyLogs = (key: string) => {
    portforwardLogsVisible.value = true

    portforwardLogsDialogCurrentPage.value = 1
    portforwardLogsTotal.value = 0
    portforwardLogsPageSize.value = 25
    portforwardLogsDialogKey.value = key


    apiPortforwardRuleLogs(key, portforwardLogsPageSize.value, portforwardLogsDialogCurrentPage.value).then((res) => {
        //console.log(res.data)
        if (res.ret == 0) {
            portforwardLogsPageSize.value = res.pageSize
            portforwardLogsTotal.value = res.total
            portforwardLogsDialogData.value = res.logs
            flushPortforwardLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })
}

const handlePortforwardLogsPreClick = (page: number) => {
    // console.log(page)
    portforwardLogsDialogCurrentPage.value = page - 1

}


const handlePortforwardLogsNextClick = (page: number) => {
    //console.log(page)
    portforwardLogsDialogCurrentPage.value = page + 1

}



const handlePortforwardLogsSizeChange = (pageSize: number) => {
    portforwardLogsPageSize.value = pageSize
    portforwardLogsDialogCurrentPage.value = 1

    apiPortforwardRuleLogs(portforwardLogsDialogKey.value, portforwardLogsPageSize.value, portforwardLogsDialogCurrentPage.value).then((res) => {
        if (res.ret == 0) {
            portforwardLogsPageSize.value = res.pageSize
            portforwardLogsTotal.value = res.total
            portforwardLogsDialogData.value = res.logs
            flushPortforwardLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })

}


const handlePortforwardLogsCurrentChange = (page: number) => {
    portforwardLogsDialogCurrentPage.value = page

    apiPortforwardRuleLogs(portforwardLogsDialogKey.value, portforwardLogsPageSize.value, portforwardLogsDialogCurrentPage.value).then((res) => {
        if (res.ret == 0) {
            portforwardLogsPageSize.value = res.pageSize
            portforwardLogsTotal.value = res.total
            portforwardLogsDialogData.value = res.logs
            flushPortforwardLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })
}


const flushPortforwardLogsDialogLogsContentView = () => {
    portforwardLogsDialogLogsContentView.value = ""
    for (let index in portforwardLogsDialogData.value) {
        let log = portforwardLogsDialogData.value[index]
        // if (index!="0"){
        //     portforwardLogsDialogLogsContentView.value += "\n"
        // }
        portforwardLogsDialogLogsContentView.value += log.LogTime + "\t" + log.LogContent + "\n"
    }

}

const convertCheckedDataToForwardType = (list: CheckboxValueType[]) => {
    var res = ['']
    res.splice(0, 1)

    let hasTcp4 = false
    let hasTcp6 = false
    let hasUdp4 = false
    let hasUdp6 = false
    let hasUdp = false
    let hasTcp = false
    for (let t of list) {
        switch (t) {
            case 'tcp4':
                hasTcp4 = true;
                break;
            case 'tcp6':
                hasTcp6 = true;
                break;
            case "udp4":
                hasUdp4 = true;
                break;
            case "udp6":
                hasUdp6 = true;
                break;
            default:
        }
    }

    if (hasTcp4 && hasTcp6) {
        hasTcp = true
    }

    if (hasUdp4 && hasUdp6) {
        hasUdp = true
    }

    if (hasTcp) {
        res.push("tcp")
    } else {
        if (hasTcp4) {
            res.push("tcp4")
        } else if (hasTcp6) {
            res.push("tcp6")
        }
    }

    if (hasUdp) {
        res.push("udp")
    } else {
        if (hasUdp4) {
            res.push("udp4")
        } else if (hasUdp6) {
            res.push("udp6")
        }
    }

    return res
}

var flushUdpOptionsView = () => {

    for (let t of formData.value.ForwardTypes) {
        if (t == "udp4" || t == "udp6" || t == 'udp') {
            udpOptionsShow.value = true
            return
        }
    }
    udpOptionsShow.value = false;
}

var flushTcpOptionsView = () => {
    for (let t of formData.value.ForwardTypes) {
        if (t == "tcp4" || t == "tcp6" || t == 'tcp') {
            tcpOptionsShow.value = true
            return
        }
    }
    tcpOptionsShow.value = false
}


var ruleList = ref([{
    Name: "",
    Key: "",
    ForwardTypes: [''],
    ListenAddress: "",
    ListenPorts: "",
    TargetAddressList: [''],
    TargetPorts: "",
    Enable: false,
    LogLevel: 1,
    LogOutputToConsole: false,
    AccessLogMaxNum: 1024,
    WebListShowLastLogMaxCount: 20,
    Options: {
        UDPPackageSize: 1500,
        SingleProxyMaxUDPReadTargetDatagoroutineCount: 64,
        SingleProxyMaxTCPConnections: 128,
        UDPProxyPerformanceMode: true,
        UDPShortMode: true,
        SafeMode: "blacklist",
    },

}]);
ruleList.value.splice(0, 1)
const dialogFormVisible = ref(false)
const formLabelWidth = '160px'

const formData = ref({
    Name: "",
    Key: "",
    ListenAddress: '',
    ListenPorts: '',
    TargetAddressList: [''],
    TargetPorts: '',
    ForwardTypes: [''],
    Enable: true,
    LogLevel: 4,
    LogOutputToConsole: false,
    AccessLogMaxNum: 1000,
    WebListShowLastLogMaxCount: 20,
    Options: {
        SingleProxyMaxTCPConnections: 128,
        SingleProxyMaxUDPReadTargetDatagoroutineCount: 32,
        UDPPackageSize: 1500,
        UDPShortMode: true,
        UDPProxyPerformanceMode: true,
        SafeMode: "blacklist",
    }
})
const formOptionType = ref("")
const dialogTitle = ref("")
const confirmText = ref("")

const addRule = () => {
    formOptionType.value = "add"
    formData.value = {
        Name: "",
        Key: "",
        ListenAddress: '',
        ListenPorts: '',
        TargetAddressList: [''],
        TargetPorts: '',
        ForwardTypes: ['tcp6'],
        Enable: true,
        LogLevel: 4,
        LogOutputToConsole: false,
        AccessLogMaxNum: 1024,
        WebListShowLastLogMaxCount: 20,
        Options: {
            SingleProxyMaxTCPConnections: 256,
            SingleProxyMaxUDPReadTargetDatagoroutineCount: 32,
            UDPPackageSize: 1500,
            UDPShortMode: false,
            UDPProxyPerformanceMode: false,
            SafeMode: "blacklist",
        }
    }
    dialogTitle.value = "添加转发规则"
    dialogFormVisible.value = true;
    confirmText.value = "添加"
    dialogForwardTypes.value = ['tcp6']
    flushUdpOptionsView()
    flushTcpOptionsView()
}

const addOrAlterRuleConfirm = () => {

    formData.value.TargetAddressList = StringToArrayList(formTargetAddressListArea.value)

    if (!checkFormData()) {
        return
    }

    switch (formOptionType.value) {
        case "add":
            console.log(JSON.stringify(formData.value))
            apiAddPortForwardRule(formData.value).then((res) => {
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
            apiAlterPortForwardRule(formData.value).then((res) => {
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
    }
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
                apiPortForwardRuleEnable(rule.Key, !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "规则 " + ruleName + " " + configure + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("error", "规则 " + ruleName + " " + configure + enableText + "失败")

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                }).catch((error) => {
                    resolve(false)
                    console.log("规则 " + ruleName + " " + configure + enableText + "失败" + ":请求出错" + error)
                    MessageShow("error", "规则 " + ruleName + " " + configure + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}

const deleteRule = (rule) => {

    const ruleName = "[" + rule.Name + "]"
    const ruleText = ruleName

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

            apiDeletePortForwardRule(rule.Key).then((res) => {
                if (res.ret == 0) {
                    //console.log("删除成功")
                    queryRuleList();
                    MessageShow("success", "删除成功")
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

const checkFormData = () => {
    if (formData.value.ForwardTypes.length <= 0) {
        ShowMessageBox("转发类型至少选择一类")
        return false
    }

    if (formData.value.ListenAddress.length != 0 && !isIP(formData.value.ListenAddress)) {
        ShowMessageBox("监听地址IP[" + formData.value.ListenAddress + "]格式有误")
        return false
    }

    if (formData.value.ListenPorts.length == 0) {
        ShowMessageBox("监听端口不能为空,端口和端口或者端口和端口范围之间用英文状态逗号,分开"
            + '\n'
            + '例如 [22,80,443,2000-2010]')
        return false
    }

    if (formData.value.TargetAddressList.length == 0) {
        ShowMessageBox("目标地址不能为空")
        return false
    }

    if (formData.value.TargetPorts.length == 0) {
        ShowMessageBox("目标端口不能为空")
        return false
    }

    if (formData.value.WebListShowLastLogMaxCount > 64 || formData.value.WebListShowLastLogMaxCount <= 0) {
        ShowMessageBox("前端列表显示最新日志最大条数 范围1-64")
        return false
    }

    return true
}

const queryRuleList = () => {
    apiGetPortForwardRuleList().then((res) => {
        //console.log(res.data)
        ruleList.value = res.list
    }).catch((error) => {
        console.log("获取转发规则列表出错:" + error)
        MessageShow("error", "获取转发规则列表出错")
    })
}

var timerID: any

onMounted(() => {
    queryRuleList();

    timerID = setInterval(() => {
        queryRuleList();
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

.portforwardLogs {
    background-color: black;
    height: fit-content;
    width: 100%;
    color: white;
    text-align: left;
    padding-left: 3px;

    border: 10px;
    overflow-y: auto;
    overflow-x: auto;
    white-space: pre-wrap;


}


.dialogRadius {
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
</style>