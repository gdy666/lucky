<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">



        <el-scrollbar height="100%">

            <div class="itemradius" :style="{
                borderRadius: 'base',
            }" v-for="task, taskIndex in taskList">

                <el-descriptions :column="4" border>

                    <el-descriptions-item label="DDNS任务名称">
                        <el-button type="" size="default" v-show="true">
                            {{ task.TaskName == '' ? '未命名任务' : task.TaskName }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="DDNS类型">
                        <el-button color="#409eff" size="default" v-show="true">
                            {{ task.TaskType }}
                        </el-button>
                    </el-descriptions-item>


                    <el-descriptions-item label="任务操作" :span="2">
                        <el-tooltip :content="task.Enable == true ? '任务已启用' : '任务已禁用'" placement="top">
                            <el-switch v-model="task.Enable" inline-prompt active-text="开" inactive-text="关"
                                :before-change="ruleEnableClick.bind(this, task.Enable, task)" size="large" />
                        </el-tooltip>

                        &nbsp;&nbsp;
                        <el-button type="primary" @click="showAddOrAlterDDNSTaskDialog('alter', task)">编辑</el-button>
                        <el-button type="danger" @click="deleteTask(task)">删除</el-button>

                    </el-descriptions-item>

                    <el-descriptions-item label="DNS服务商">
                        <el-button color="#409eff" size="small" v-show="true">
                            {{ GetDNSServerName(task.DNS) }}
                        </el-button>


                        <el-tooltip :v-if="task.DNS.HttpClientProxyType == '' ? false : true"
                            :content="task.DNS.HttpClientProxyType == '' ? '未设置代理' : '类型:[' + task.DNS.HttpClientProxyType + '] 代理服务器地址:[' + task.DNS.HttpClientProxyAddr + ']'"
                            placement="top">
                            <el-button :type="task.DNS.HttpClientProxyType == '' ? '' : 'success'" size="small">
                                {{ task.DNS.HttpClientProxyType == '' ? '未设置代理' : '已设置代理' }}
                            </el-button>
                        </el-tooltip>

                    </el-descriptions-item>

                    <el-descriptions-item label="获取公网IP方式">
                        <el-button color="#409eff" size="default" v-show="true">
                            {{ task.GetType == "url" ? "URL" : "网卡" }}
                        </el-button>
                    </el-descriptions-item>


                    <el-descriptions-item label="公网IP">

                        <el-tooltip :placement="taskIndex == 0 ? 'bottom' : 'top'" effect="dark" :trigger-keys="[]"
                            content="">
                            <template #content>
                                <span v-html="getIPHistroyListHtml(task.TaskState.IPAddrHistory)"></span>
                            </template>
                            <el-button color="#409eff" size="default" v-show="true"
                                @click="copyWanIP(task.TaskState.IpAddr)">
                                {{ task.TaskState.IpAddr == "" ? '尚未获取到公网IP' : task.TaskState.IpAddr }}
                            </el-button>
                        </el-tooltip>
                    </el-descriptions-item>






                    <el-descriptions-item label="TTL">
                        <el-button size="small" type="" v-show="true">
                            {{ GetTTLText(task.TTL) }}
                        </el-button>
                    </el-descriptions-item>


                    <div v-if="task.WebhookEnable">

                        <el-descriptions-item label="WebHook" :span="1">

                            <el-button type="success" size="small">
                                已启用
                            </el-button>

                            <el-button :type="task.WebhookProxy == '' ? '' : 'success'" size="small">
                                {{ task.WebhookProxy == '' ? '未设置代理' : '已设置代理' }}
                            </el-button>
                        </el-descriptions-item>

                        <el-descriptions-item label="WebHook 触发时间" :span="task.TaskState.WebhookCallTime == '' ? 3 : 1">

                            <el-tooltip :placement="taskIndex == 0 ? 'bottom' : 'top'" effect="dark" :trigger-keys="[]"
                                content="">
                                <template #content>
                                    <span
                                        v-html="getWebhookCallHistroyListHtml(task.TaskState.WebhookCallHistroy)"></span>
                                </template>
                                <el-button color="#409eff" size="default">
                                    {{ task.TaskState.WebhookCallTime == "" ? '从未触发' :
                                    task.TaskState.WebhookCallTime
                                    }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>


                        <el-descriptions-item label="WebHook 触发结果"
                            v-if="task.TaskState.WebhookCallTime == '' ? false : true"
                            :span="task.TaskState.WebhookCallErrorMsg == '' ? 2 : 1">
                            <el-tooltip :placement="taskIndex == 0 ? 'bottom' : 'top'" effect="dark" :trigger-keys="[]"
                                content="">
                                <template #content>
                                    <span
                                        v-html="getWebhookCallHistroyListHtml(task.TaskState.WebhookCallHistroy)"></span>
                                </template>
                                <el-button color="#409eff" size="default">
                                    {{ task.TaskState.WebhookCallResult == true ? "成功" : '失败' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>









                        <el-descriptions-item label="WebHook 触发错误原因"
                            v-if="task.TaskState.WebhookCallErrorMsg == '' ? false : true" :span="1">
                            <el-button color="#409eff" size="default"
                                @click="showErrorMessage(task.TaskState.WebhookCallErrorMsg)">
                                点击查看
                            </el-button>
                        </el-descriptions-item>



                    </div>

                    <!-- <el-descriptions-item label="WebHook URL" :span="2" v-if="task.WebhookURL == '' ? false : true">
                        <el-button color="#409eff" size="default">
                            {{ task.WebhookURL }}
                        </el-button>
                    </el-descriptions-item>


                    <el-descriptions-item label="WebHook RequestBody" :span="2"
                        v-if="task.WebhookURL == '' ? false : true">
                        <el-button color="#409eff" size="default">
                            {{ task.WebhookRequestBody }}
                        </el-button>
                    </el-descriptions-item> -->


                    <div v-for="domain in task.TaskState.Domains">
                        <el-descriptions-item label="域名">
                            <el-button color="#409eff" size="default"
                                @click="copyDomain(domain.SubDomain, domain.DomainName)">
                                {{ domain.SubDomain == '' ? domain.DomainName : domain.SubDomain + "." +
                                domain.DomainName }}
                            </el-button>
                        </el-descriptions-item>

                        <el-descriptions-item label="同步结果">

                            <el-tooltip :placement="taskIndex == 0 ? 'bottom' : 'top'" effect="dark" :trigger-keys="[]"
                                content="">
                                <template #content>
                                    <span v-html="GetSyncUpdateHistroyListHtml(domain.UpdateHistroy)"></span>

                                </template>

                                <el-button
                                    :type="domain.UpdateStatus == '失败' ? 'danger' : task.Enable ? 'success' : 'info'"
                                    size="small">
                                    {{ task.Enable ? domain.UpdateStatus : '停止同步' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="最后检测时间" :span="domain.Message == '' ? 2 : 1">
                            <el-tooltip :placement="taskIndex == 0 ? 'bottom' : 'top'" effect="dark" :trigger-keys="[]"
                                content="">
                                <template #content>
                                    <span v-html="GetSyncUpdateHistroyListHtml(domain.UpdateHistroy)"></span>

                                </template>
                                <el-button color="#409eff" size="default">
                                    {{ domain.LastUpdateStatusTime }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="错误信息" v-if="domain.Message == '' ? false : true">
                            <el-button color="#409eff" size="default" @click="showErrorMessage(domain.Message)">
                                点击查看
                            </el-button>
                        </el-descriptions-item>

                    </div>


                </el-descriptions>
            </div>


        </el-scrollbar>

        <el-affix position="bottom" :offset="30" class="affix-container">
            <el-button type="primary" @click="showAddOrAlterDDNSTaskDialog('add', null)">添加DDNS任务 <el-icon>
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix>




        <el-dialog v-model="errorMessageVisible" title="详细信息" draggable :show-close="true" :close-on-click-modal="false"
            width="600px">

            <el-input v-model="errorMessage" :autosize="{ minRows: 3, maxRows: 10 }" type="textarea">
            </el-input>
        </el-dialog>


        <!-- 添加/修改DDNS任务 对话框-->
        <el-dialog v-model="addDDNSDialogVisible" :title="DDNSFormOptionType == 'add' ? 'DDNS任务添加' : 'DDNS任务修改'"
            draggable :show-close="true" :close-on-click-modal="false" width="600px">

            <el-form :model="DDNSForm">
                <el-form-item label="DDNS任务名称" label-width="auto">
                    <el-input v-model="DDNSForm.TaskName" placeholder="可留空" autocomplete="off" />
                </el-form-item>


                <el-form-item label="任务开关" label-width="auto">
                    <el-switch v-model="DDNSForm.Enable" inline-prompt width="50px" active-text="启用"
                        inactive-text="停用" />
                </el-form-item>



                <div v-show="DDNSForm.Enable">



                    <div class="fromitemDivRadius">
                        <p>DNS服务商设置</p>

                        <div class="fromitemChildDivRadius">

                            <el-form-item label="DNS服务商" label-width="auto">
                                <el-select v-model="DDNSForm.DNS.Name" class="m-2" placeholder="请选择">
                                    <el-option v-for="item in DNSServerList" :key="item.value" :label="item.label"
                                        :value="item.value" />
                                </el-select>
                            </el-form-item>

                            <!-- <el-form-item label-width="auto">
                                <el-link type="primary" :href="getDNSHelpLink()" target="_blank">{{ getDNSHelpLinkText()
                                }}
                                </el-link>

                                &nbsp; <div v-html="getDNSHelpTipsHtml()"></div>
                            </el-form-item> -->

                            <div v-if="DDNSForm.DNS.Name == 'alidns'">
                                <el-form-item label-width="auto">
                                    <el-link type="primary" style="font-size: small;"
                                        href="https://ram.console.aliyun.com/manage/ak?spm=5176.12818093.nav-right.dak.488716d0mHaMgg"
                                        target="_blank">
                                        创建 AccessKey
                                    </el-link>
                                </el-form-item>
                            </div>

                            <div v-if="DDNSForm.DNS.Name == 'baiducloud'">

                                <el-tooltip class="box-item" effect="dark"
                                    content="需调用 API,而百度云相关API仅对申请用户开放，使用前请先提交工单申请。">

                                    <el-form-item label-width="auto">




                                        <el-link type="primary" style="font-size: small;"
                                            href="https://console.bce.baidu.com/iam/?_=1651763238057#/iam/accesslist"
                                            target="_blank">
                                            创建 AccessKey
                                        </el-link>
                                        &nbsp; &nbsp; &nbsp;
                                        <el-link type="primary" style="font-size: small;"
                                            href="https://ticket.bce.baidu.com/#/ticket/create~productId=60&questionId=393&channel=2"
                                            target="_blank">
                                            申请工单
                                        </el-link>

                                    </el-form-item>
                                </el-tooltip>
                            </div>


                            <div v-if="DDNSForm.DNS.Name == 'cloudflare'">
                                <el-form-item label-width="auto">
                                    <el-link type="primary" style="font-size: small;"
                                        href="https://dash.cloudflare.com/profile/api-tokens" target="_blank">
                                        创建令牌->编辑区域 DNS (使用模板)
                                    </el-link>
                                </el-form-item>
                            </div>

                            <div v-if="DDNSForm.DNS.Name == 'dnspod'">
                                <el-form-item label-width="auto">
                                    <el-link type="primary" style="font-size: small;"
                                        href="https://console.dnspod.cn/account/token" target="_blank">
                                        创建密钥
                                    </el-link>
                                </el-form-item>
                            </div>


                            <div v-if="DDNSForm.DNS.Name == 'huaweicloud'">
                                <el-form-item label-width="auto">
                                    <el-link type="primary" style="font-size: small;"
                                        href="https://console.huaweicloud.com/iam/?locale=zh-cn#/mine/accessKey"
                                        target="_blank">
                                        新增访问密钥
                                    </el-link>
                                </el-form-item>
                            </div>

                            <div v-if="DDNSForm.DNS.Name == 'porkbun'">
                                <el-form-item label-width="auto">
                                    <el-link type="primary" style="font-size: small;"
                                        href="https://porkbun.com/account/api" target="_blank">
                                        创建 Access
                                    </el-link>
                                </el-form-item>
                            </div>

                            <div v-if="DDNSForm.DNS.Name == 'callback'">
                                <el-form-item label-width="auto">
                                    <p style="font-size:1px">支持的变量 #{ip}, #{domain}, #{recordType}, #{ttl}</p>
                                </el-form-item>

                            </div>

                            <el-form-item :label="getIDLabel()" v-show="showDNSIDFormItem()" label-width="auto">
                                <el-input v-model="DDNSForm.DNS.ID" autocomplete="off" />
                            </el-form-item>

                            <el-form-item :label="getSecretLabel()" v-show="showDNSSecretFormItem()" label-width="auto">
                                <el-input v-model="DDNSForm.DNS.Secret" autocomplete="off" />
                            </el-form-item>


                            <el-tooltip placement="top">
                                <template #content>
                                    强制同步,当DNS解析域名开关打开时会先通过DNS解析进行IP比对,比对一致依然不会强制同步,只要不手动修改域名IP这个值设置大一些完全没问题,可设范围(60-360000)<br />
                                    强制同步检查会在每一轮定时批量执行DDNS任务中进行,所以实际强制同步时间不会很精确
                                </template>
                                <el-form-item label="强制同步(秒)" label-width="auto" :min="60" :max="360000">
                                    <el-input-number v-model="DDNSForm.DNS.ForceInterval" autocomplete="off" />
                                </el-form-item>
                            </el-tooltip>

                        </div>


                        <p v-show="DDNSForm.DNS.Name == 'callback' ? true : false">自定义(Callback)服务商设置</p>

                        <!--Callback 相关-->
                        <div v-show="DDNSForm.DNS.Name == 'callback' ? true : false" class="fromitemChildDivRadius">



                            <el-form-item label="Callback DNS服务商" label-width="auto">
                                <el-select v-model="DDNSForm.DNS.Callback.Server" class="m-2" placeholder="请选择">
                                    <el-option v-for="item in DNSCallbackServerList" :key="item.value"
                                        :label="item.label" :value="item.value" />
                                </el-select>
                            </el-form-item>

                            <div v-show="DDNSForm.DNS.Callback.Server == 'other' ? false : true">

                                <div v-if="DDNSForm.DNS.Callback.Server == 'meibu'" style="font-size: small;">
                                    <el-tooltip content="注意:每步 IPv4和IPv6的接口不相同,免费二级域名不能同时支持IPv4和IPv6" placement="top">
                                        <el-form-item label-width="auto">
                                            <el-link type="primary" style="font-size: small;"
                                                href="http://www.meibu.com/regedit.shtml" target="_blank">
                                                每步-免费二级域名注册
                                            </el-link>
                                        </el-form-item>
                                    </el-tooltip>
                                </div>


                                <div v-if="DDNSForm.DNS.Callback.Server == 'noip'" style="font-size: small;">
                                    <el-form-item label-width="auto">
                                        <el-link type="primary" style="font-size: small;" href="https://www.noip.com"
                                            target="_blank">
                                            No-IP官网
                                        </el-link>
                                    </el-form-item>
                                </div>

                                <div v-if="DDNSForm.DNS.Callback.Server == 'dynv6'" style="font-size: small;">
                                    <el-form-item label-width="auto">
                                        <el-link type="primary" style="font-size: small;" href="https://dynv6.com/"
                                            target="_blank">
                                            Dynv6官网
                                        </el-link>
                                        &nbsp; &nbsp; &nbsp;
                                        <el-link type="primary" style="font-size: small;" href="https://dynv6.com/keys"
                                            target="_blank">
                                            Token创建
                                        </el-link>
                                    </el-form-item>
                                </div>

                                <div v-if="DDNSForm.DNS.Callback.Server == 'dynu'" style="font-size: small;">
                                    <el-form-item label-width="auto">
                                        <el-link type="primary" style="font-size: small;" href="https://www.dynu.com/"
                                            target="_blank">
                                            Dynu官网
                                        </el-link>
                                        &nbsp; &nbsp; &nbsp;
                                        <el-link type="primary" style="font-size: small;"
                                            href="https://www.dynu.com/zh-CN/ControlPanel/ManageCredentials"
                                            target="_blank">
                                            IP更新密码设置
                                        </el-link>
                                    </el-form-item>
                                </div>

                                <el-form-item label-width="auto">
                                    <el-button color="#409eff" size="default" @click="autocompleteCallbackForm">
                                        根据DNS服务商自动填充参数模版
                                    </el-button>
                                </el-form-item>
                            </div>


                            <el-tooltip class="box-item" effect="dark" content="">
                                <template #content>接口地址<br />
                                    支持的变量<br />
                                    #{ip} : 外网IP<br />
                                    #{domain} : 域名<br />
                                    #{recordType} : A 或者 AAAA <br />
                                    #{ttl} : TTL值</template>
                                <el-form-item label="接口地址" label-width="auto">
                                    <el-input v-model="DDNSForm.DNS.Callback.URL" autocomplete="off" />
                                </el-form-item>

                            </el-tooltip>

                            <el-form-item label="请求方法" label-width="auto">
                                <el-select v-model="DDNSForm.DNS.Callback.Method" class="m-2" placeholder="请选择">
                                    <el-option v-for="item in CallbackMethodList" :key="item.value" :label="item.label"
                                        :value="item.value" />
                                </el-select>
                            </el-form-item>




                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    一行一条Header(key:value)<br />
                                    支持的变量<br />
                                    #{ip} : 外网IP<br />
                                    #{domain} : 域名<br />
                                    #{recordType} : A 或者 AAAA <br />
                                    #{ttl} : TTL值<br />
                                    如果需要使用BasicAuth,请使用下面两行Header设置BasicAuth的账号和密码<br />
                                    BasicAuthUserName:你的账号<br />
                                    BasicAuthPassword:你的密码</template>
                                <el-form-item label-width="auto" label="请求Headers">
                                    <el-input v-model="DDNSFormCallbackHeaderArea"
                                        :autosize="{ minRows: 3, maxRows: 5 }" placeholder="" type="textarea">
                                    </el-input>
                                </el-form-item>
                            </el-tooltip>


                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>请求主体 requestBody<br />
                                    支持的变量<br />
                                    #{ip} : 外网IP<br />
                                    #{domain} : 域名<br />
                                    #{recordType} : A 或者 AAAA <br />
                                    #{ttl} : TTL值</template>
                                <el-form-item label-width="auto" label="请求主体"
                                    v-show="DDNSForm.DNS.Callback.Method == 'get' ? false : true">
                                    <el-input v-model="DDNSForm.DNS.Callback.RequestBody"
                                        :autosize="{ minRows: 3, maxRows: 5 }" placeholder="" type="textarea">
                                    </el-input>
                                </el-form-item>
                            </el-tooltip>


                            <el-tooltip content="禁用接口调用成功字符串检测,开启后仅以http StatusCode==200判断接口是否成功调用." placement="top">

                                <el-form-item label="禁用接口调用成功字符串检测" label-width="auto">
                                    <el-switch v-model="DDNSForm.DNS.Callback.DisableCallbackSuccessContentCheck"
                                        inline-prompt width="50px" active-text="是" inactive-text="否" />
                                </el-form-item>
                            </el-tooltip>

                            <div v-show="!DDNSForm.DNS.Callback.DisableCallbackSuccessContentCheck">
                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                    <template #content>用于判断记录接口是否成功调用,多种表示成功的不同字符串请分多行写<br />
                                        支持的变量<br />
                                        #{ip} : 外网IP<br />
                                        #{domain} : 域名<br />
                                        #{recordType} : A 或者 AAAA <br />
                                        #{ttl} : TTL值</template>
                                    <el-form-item label="接口调用成功包含的字符串" label-width="auto">
                                        <el-input v-model="DDNSFormCallbackSuccessContentArea"
                                            :autosize="{ minRows: 3, maxRows: 5 }" type="textarea" autocomplete="off"
                                            placeholder="" />
                                    </el-form-item>
                                </el-tooltip>
                            </div>

                        </div>


                        <div class="fromitemChildDivRadius">


                            <el-tooltip content="调用DNS服务商接口更新或增加域名之前先通过DNS查询域名对应IP,降低对DNS服务商接口的访问频率,建议启用"
                                placement="top">

                                <el-form-item label="DNS解析检测域名" label-width="auto">
                                    <el-switch v-model="DDNSForm.DNS.ResolverDoaminCheck" inline-prompt width="50px"
                                        active-text="开启" inactive-text="禁用" @change="ResolverDoaminCheckChange" />
                                </el-form-item>
                            </el-tooltip>

                            <div v-show="DDNSForm.DNS.ResolverDoaminCheck">

                                <el-tooltip class="box-item" :trigger-keys="[]" effect="dark"
                                    content="一行一个DNS服务器地址(带端口)">
                                    <el-form-item label-width="auto" label="DNS服务器列表">
                                        <el-input v-model="DDNSFormDNSServerListArea"
                                            :autosize="{ minRows: 3, maxRows: 5 }" placeholder="一行一个DNS服务器地址(带端口)"
                                            type="textarea">
                                        </el-input>
                                    </el-form-item>
                                </el-tooltip>


                            </div>

                        </div>

                        <p>DNS接口调用额外设置</p>
                        <div class="fromitemChildDivRadius">
                            <el-form-item label="DNS接口调用使用的网络类型" label-width="auto">
                                <el-select v-model="DDNSForm.DNS.CallAPINetwork" class="m-2" placeholder="请选择">
                                    <el-option v-for="item in TCPNetworkTypeList" :key="item.value" :label="item.label"
                                        :value="item.value" />
                                </el-select>
                            </el-form-item>

                        </div>


                        <p>DNS接口调用代理设置</p>
                        <div class="fromitemChildDivRadius">

                            <el-form-item label="DNS接口调用 代理设置" label-width="auto">
                                <el-select v-model="DDNSForm.DNS.HttpClientProxyType" class="m-2" placeholder="请选择">
                                    <el-option v-for="item in HttpProxyTypeList" :key="item.value" :label="item.label"
                                        :value="item.value" />
                                </el-select>
                            </el-form-item>

                            <div v-show="DDNSForm.DNS.HttpClientProxyType == '' ? false : true">

                                <el-form-item label="代理服务器IP" label-width="auto">
                                    <el-input v-model="DDNSForm.DNS.HttpClientProxyAddr" autocomplete="off" />
                                </el-form-item>

                                <el-form-item label="代理服务器认证用户" label-width="auto">
                                    <el-input v-model="DDNSForm.DNS.HttpClientProxyUser" autocomplete="off"
                                        placeholder="没有可留空" />
                                </el-form-item>

                                <el-form-item label="代理服务器认证密码" label-width="auto">
                                    <el-input v-model="DDNSForm.DNS.HttpClientProxyPassword" autocomplete="off"
                                        placeholder="没有可留空" />
                                </el-form-item>

                            </div>

                        </div>



                    </div>








                    <div class="fromitemDivRadius">

                        <el-form-item label="公网IP类型" label-width="auto">
                            <el-radio-group v-model="DDNSForm.TaskType" class="ml-4" @change="ddnsTaskTypeChange">
                                <el-radio label="IPv4">IPv4</el-radio>
                                <el-radio label="IPv6">IPv6</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item label="获取公网IP方式" label-width="auto">
                            <el-radio-group v-model="DDNSForm.GetType" class="ml-4">
                                <el-radio label="url">通过接口获取</el-radio>
                                <el-radio label="netInterface">通过网卡获取</el-radio>
                            </el-radio-group>
                        </el-form-item>


                        <div v-if="DDNSForm.GetType == 'url' ? true : false">
                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="一行一个接口">
                                <el-form-item label-width="auto" label="接口列表">
                                    <el-input v-model="DDNSFormURLArea" :autosize="{ minRows: 5, maxRows: 20 }"
                                        placeholder="一行一个接口" type="textarea">
                                    </el-input>
                                </el-form-item>
                            </el-tooltip>
                        </div>


                        <div v-if="DDNSForm.GetType == 'netInterface' ? true : false">
                            <el-form-item label-width="auto" label="网卡列表">
                                <el-select v-model="DDNSForm.NetInterface" class="m-2">

                                    <el-tooltip class="box-item" effect="dark" v-for="item in currentNetInterface"
                                        :key="item.NetInterfaceName" :content="JSON.stringify(item.AddressList)">

                                        <el-option :label="item.NetInterfaceName" :value="item.NetInterfaceName" />
                                    </el-tooltip>

                                    <!-- <el-option  v-for="item in currentNetInterface" :key="item.NetInterfaceName" :label="item.NetInterfaceName"
                                        :value="item.NetInterfaceName" /> -->

                                </el-select>
                            </el-form-item>

                            <el-tooltip class="box-item" effect="dark"
                                content="留空表示匹配选中网卡第1个IP, 纯数字n表示匹配第n个IP, 24*表示匹配以24开头的第一个IP, *24表示匹配以24结尾的第一个IP, 还可以填写正则表达式">
                                <el-form-item label-width="auto" label="IP选择匹配规则">
                                    <el-input v-model="DDNSForm.IPReg" :autosize="{ minRows: 5, maxRows: 20 }"
                                        placeholder="留空表示选择当前网卡第一个IP">
                                    </el-input>

                                    <el-button color="#409eff" size="small" @click="IPRegTest">
                                        IP选择匹配测试
                                    </el-button>
                                </el-form-item>
                            </el-tooltip>
                        </div>




                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="一行一条域名">
                            <el-form-item label-width="auto" label="域名列表">
                                <el-input v-model="DDNSFormDomiansArea" :autosize="{ minRows: 3, maxRows: 9 }"
                                    placeholder="一行一条域名" type="textarea">
                                </el-input>
                            </el-form-item>
                        </el-tooltip>


                        <el-tooltip class="box-item" effect="dark" content="如账号支持更小的 TTL , 可修改。 IP 有变化时才会更新 TTL">
                            <el-form-item label-width="auto" label="TTL">
                                <el-select v-model="DDNSForm.TTL" class="m-2">
                                    <el-option v-for="item in TTLList" :key="item.value" :label="item.label"
                                        :value="item.value" />
                                </el-select>
                            </el-form-item>
                        </el-tooltip>

                    </div>








                    <div class="fromitemDivRadius">



                        <el-form-item label="Webhook" label-width="auto">
                            <el-tooltip class="box-item" effect="dark"
                                content="Webhook 仅当IP改变,域名更新/添加成功或失败时才会触发Webhook">
                                <el-switch v-model="DDNSForm.WebhookEnable" inline-prompt width="50px" active-text="开启"
                                    inactive-text="禁用" />
                            </el-tooltip>
                            <el-tooltip class="box-item" effect="dark"
                                content="如果选择了DNS代理设置先请保存任务再手动触发测试,否则在测试中可能代理设置不生效">

                                <el-button color="#409eff" v-show="DDNSForm.WebhookEnable" size="small"
                                    @click="WebHookTest" style="margin-left:30px;">
                                    Webhook手动触发测试
                                </el-button>
                            </el-tooltip>
                        </el-form-item>


                        <div v-show="DDNSForm.WebhookEnable">

                            <el-tooltip class="box-item" effect="dark" content="获取IP失败时同样触发Webhook,默认不开启">
                                <el-form-item label="获取IP失败时触发Webhook" label-width="auto">
                                    <el-switch v-model="DDNSForm.WebhookCallOnGetIPfail" inline-prompt width="50px"
                                        active-text="启用" inactive-text="禁用" />
                                </el-form-item>
                            </el-tooltip>

                            <!-- <el-form-item label-width="auto">
                                <el-link type="primary" href="https://github.com/jeessy2/ddns-go#webhook"
                                    target="_blank">
                                    点击参考官方 Webhook 说明
                                </el-link>

                                &nbsp; <div v-html="getDNSHelpTipsHtml()"></div>
                            </el-form-item> -->

                            <div class="fromitemChildDivRadius">

                                <el-form-item label="常见Webhook(消息推送)服务商" label-width="auto">
                                    <el-select v-model="webhookSelect" class="m-2" placeholder="请选择"
                                        @change="WebhookServerSelectChange">
                                        <el-option v-for="item in WebhookServerList" :key="item.value"
                                            :label="item.label" :value="item.value" />
                                    </el-select>说明


                                </el-form-item>

                                <div v-show="webhookSelect == 'custom'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">
                                            支持自定义webhook
                                        </p>

                                    </el-form-item>
                                </div>

                                <div v-show="webhookSelect == 'serverjiang'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">
                                            https://sctapi.ftqq.com/[SendKey].send?title=主人IP变了#{ipAddr},你的公网IP变了#{ipAddr},域名更新成功列表：#{successDomains},域名更新失败列表：#{failedDomains}
                                        </p>

                                    </el-form-item>
                                </div>

                                <div v-show="webhookSelect == 'bark'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">
                                            https://api.day.app/[YOUR_KEY]/主人IP变了#{ipAddr},你的公网IP变了#{ipAddr},域名更新成功列表：#{successDomains},域名更新失败列表：#{failedDomains}
                                        </p>

                                    </el-form-item>
                                </div>

                                <div v-show="webhookSelect == 'dingding'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">钉钉电脑端 -> 群设置 -> 智能群助手 -> 添加机器人 -> 自定义</p>
                                        <p style="font-size:1px">只勾选自定义关键词,输入的关键字必须包含在RequestBody的content中,如:你的公网IP变了
                                        </p>
                                        <p style="font-size:1px">接口调用成功包含的字符串填入 {"errcode":0,"errmsg":"ok"}</p>
                                        <p style="font-size:1px">方法请求选择POST,RequestBody 示例如下</p>
                                    </el-form-item>

                                    <el-form-item label-width="auto">
                                        <el-input v-model="WebhookServerListArea" type="textarea" rows="5">
                                        </el-input>
                                    </el-form-item>

                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">调用成功包含的字符串填入 {"errcode":0,"errmsg":"ok"}</p>
                                    </el-form-item>
                                </div>

                                <div v-show="webhookSelect == 'feishu'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">飞书电脑端 -> 群设置 -> 添加机器人 -> 自定义机器人</p>
                                        <p style="font-size:1px">
                                            安全设置只勾选自定义关键词,输入的关键字必须包含在RequestBody的content中,如：你的公网IP变了</p>
                                        <p style="font-size:1px">接口调用成功包含的字符串填入
                                            {"StatusCode":0,"StatusMessage":"success"}</p>
                                        <p style="font-size:1px">方法请求选择POST,RequestBody 示例如下</p>
                                    </el-form-item>

                                    <el-form-item label-width="auto">
                                        <el-input v-model="WebhookServerListArea" type="textarea" rows="5">
                                        </el-input>
                                    </el-form-item>


                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">调用成功包含的字符串填入 {"StatusCode":0,"StatusMessage":"success"}
                                        </p>
                                    </el-form-item>
                                </div>

                                <div v-show="webhookSelect == 'weixinpro'" style="color:blue;">
                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">下载企业微信→左上角三横杠→全新创建企业→个人组件团队(创建个人的企业群聊)</p>
                                        <p style="font-size:1px">进入群聊添加 [群机器人] 复制机器人Webhook地址</p>
                                        <p style="font-size:1px">方法请求选择POST,RequestBody 示例如下</p>
                                    </el-form-item>

                                    <el-form-item label-width="auto">
                                        <el-input v-model="WebhookServerListArea" type="textarea" rows="5">
                                        </el-input>
                                    </el-form-item>


                                    <el-form-item label-width="auto">
                                        <p style="font-size:1px">调用成功包含的字符串填入 {"errcode":0,"errmsg":"ok"}</p>
                                    </el-form-item>


                                </div>

                            </div>

                            <div class="fromitemChildDivRadius">

                                <el-tooltip class="box-item" effect="dark" content="">
                                    <template #content>支持的变量 <br />
                                        #{ipAddr} : 当前公网IP<br />
                                        #{time} : 触发Webhook的时间 <br />
                                        #{successDomains} : 更新/添加成功的域名列表,域名之间用,号分隔<br />
                                        #{successDomainsLine} : 更新/添加成功的域名列表,域名之间用'\n'分隔<br />
                                        #{failedDomains} : 更新/添加失败的域名列表,域名之间用,号分隔<br />
                                        #{failedDomainsLine} : 更新/添加失败的域名列表,域名之间用'\n'分隔</template>
                                    <el-form-item label="接口地址" label-width="auto">
                                        <el-input v-model="DDNSForm.WebhookURL" autocomplete="off" />
                                    </el-form-item>

                                </el-tooltip>


                                <el-form-item label="请求方法" label-width="auto">
                                    <el-select v-model="DDNSForm.WebhookMethod" class="m-2" placeholder="请选择">
                                        <el-option v-for="item in CallbackMethodList" :key="item.value"
                                            :label="item.label" :value="item.value" />
                                    </el-select>
                                </el-form-item>



                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                    <template #content>
                                        一行一条Header(key:value) <br />
                                        支持的变量 :<br />
                                        #{time} : 触发Webhook的时间 <br />
                                        #{ipAddr} : 当前公网IP <br />
                                        #{successDomains} : 更新/添加成功的域名列表,域名之间用,号分隔<br />
                                        #{successDomainsLine} : 更新/添加成功的域名列表,域名之间用'\n'分隔<br />
                                        #{failedDomains} : 更新/添加失败的域名列表,域名之间用,号分隔<br />
                                        #{failedDomainsLine} : 更新/添加失败的域名列表,域名之间用'\n'分隔<br />
                                        如果需要使用BasicAuth,请使用下面两行Header设置BasicAuth的账号和密码<br />
                                        BasicAuthUserName:你的账号<br />
                                        BasicAuthPassword:你的密码</template>
                                    <el-form-item label-width="auto" label="Headers">
                                        <el-input v-model="DDNSFormWebhookHeadersArea"
                                            :autosize="{ minRows: 3, maxRows: 5 }" placeholder="" type="textarea">
                                        </el-input>
                                    </el-form-item>
                                </el-tooltip>




                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                    <template #content>支持的变量<br />
                                        #{time} : 触发Webhook的时间 <br />
                                        #{ipAddr} : 当前公网IP<br />
                                        #{successDomains} : 更新/添加成功的域名列表,域名之间用,号分隔<br />
                                        #{successDomainsLine} : 更新/添加成功的域名列表,域名之间用'\n'分隔<br />
                                        #{failedDomains} : 更新/添加失败的域名列表,域名之间用,号分隔<br />
                                        #{failedDomainsLine} : 更新/添加失败的域名列表,域名之间用'\n'分隔</template>
                                    <el-form-item label="RequestBody" label-width="auto"
                                        v-show="DDNSForm.WebhookMethod == 'get' ? false : true">
                                        <el-input v-model="DDNSForm.WebhookRequestBody" type="textarea" rows="5"
                                            placeholder="">
                                        </el-input>
                                    </el-form-item>
                                </el-tooltip>



                                <el-tooltip content="禁用Webhook接口调用成功字符串检测,开启后仅以http StatusCode==200判断接口是否成功调用."
                                    placement="top">

                                    <el-form-item label="禁用Webhook接口调用成功字符串检测" label-width="auto">
                                        <el-switch v-model="DDNSForm.WebhookDisableCallbackSuccessContentCheck"
                                            inline-prompt width="50px" active-text="是" inactive-text="否" />
                                    </el-form-item>
                                </el-tooltip>


                                <div v-show="!DDNSForm.WebhookDisableCallbackSuccessContentCheck">
                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>用于判断记录Webhook 接口是否成功调用<br />
                                            多种表示成功的不同字符串请分多行写<br />
                                            支持的变量 <br />
                                            #{ipAddr} : 当前公网IP<br />
                                            #{successDomains} : 更新/添加成功的域名列表,域名之间用,号分隔<br />
                                            #{successDomainsLine} : 更新/添加成功的域名列表,域名之间用'\n'分隔<br />
                                            #{failedDomains} : 更新/添加失败的域名列表,域名之间用,号分隔<br />
                                            #{failedDomainsLine} : 更新/添加失败的域名列表,域名之间用'\n'分隔</template>
                                        <el-form-item label="接口调用成功包含的字符串" label-width="auto">
                                            <el-input v-model="DDNSFormWebhookSuccessContentArea"
                                                :autosize="{ minRows: 3, maxRows: 5 }" type="textarea"
                                                autocomplete="off" placeholder="" />
                                        </el-form-item>
                                    </el-tooltip>
                                </div>


                            </div>


                            <div class="fromitemChildDivRadius">

                                <el-form-item label="代理设置" label-width="auto">
                                    <el-select v-model="DDNSForm.WebhookProxy" class="m-2" placeholder="请选择">
                                        <el-option v-for="item in WebhookProxyTypeList" :key="item.value"
                                            :label="item.label" :value="item.value" />
                                    </el-select>
                                </el-form-item>

                                <div
                                    v-show="DDNSForm.WebhookProxy == '' || DDNSForm.WebhookProxy == 'dns' ? false : true">

                                    <el-form-item label="代理服务器IP" label-width="auto">
                                        <el-input v-model="DDNSForm.WebhookProxyAddr" autocomplete="off" />
                                    </el-form-item>

                                    <el-form-item label="代理服务器认证用户" label-width="auto">
                                        <el-input v-model="DDNSForm.WebhookProxyUser" autocomplete="off"
                                            placeholder="没有可留空" />
                                    </el-form-item>

                                    <el-form-item label="代理服务器认证密码" label-width="auto">
                                        <el-input v-model="DDNSForm.WebhookProxyPassword" autocomplete="off"
                                            placeholder="没有可留空" />
                                    </el-form-item>

                                </div>

                            </div>

                        </div>


                    </div>

                    <div class="fromitemDivRadius" style="padding-top:10px;">
                        <el-tooltip content="Http Client 超时时间,没必要不要改,可设置范围 (3-60)" placement="top">
                            <el-form-item label="HttpClient timeout(秒)" label-width="auto" :min="3" :max="60">
                                <el-input-number v-model="DDNSForm.HttpClientTimeout" autocomplete="off" />
                            </el-form-item>
                        </el-tooltip>
                    </div>
                </div>

            </el-form>

            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="addDDNSDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="exeAddOrAlterDDNSOption">{{ DDNSFormOptionType == "add" ? '添加' :
                    '修改'
                    }}
                    </el-button>
                </span>
            </template>
        </el-dialog>


    </div>

</template>


<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessageBox } from 'element-plus'
import { MessageShow, Notification, ShowMessageBox } from '../utils/ui'

import { StringToArrayList, CopyTotoClipboard } from '../utils/utils'
import { apiWebhookTest, apiAddDDNSTask, apiAlterDDNSTask, apiDeleteDDNSTask, apiGetDDNSTaskList, apiDDNSTaskEnable, apiGetNetinterfaces, apiGetIPRegTest } from '../apis/utils'
import { networkInterfaces } from 'os'

const errorMessageVisible = ref(false)
const errorMessage = ref("")

const showErrorMessage = (message: string) => {
    errorMessageVisible.value = true
    errorMessage.value = message
}

var netInterfaceList = ref({
    IPv6NewInterfaces: [{ NetInterfaceName: "", AddressList: [""] }],
    IPv4NewInterfaces: [{ NetInterfaceName: "", AddressList: [""] }]
})
var currentNetInterface = ref([{ NetInterfaceName: "", AddressList: [""] }])

var taskList = ref([{
    TaskName: "",
    TaskKey: "",
    TaskType: "IPv6",
    Enable: true,
    GetType: "url",
    URL: [""],
    NetInterface: "",
    IPReg: "",
    Domains: [""],
    HttpClientTimeout: 20,
    DNS: {
        Name: "alidns",
        ID: "",
        Secret: "",
        ForceInterval: 600,
        ResolverDoaminCheck: false,
        DNSServerList: [''],
        HttpClientProxyType: "",
        HttpClientProxyAddr: "",
        HttpClientProxyUser: "",
        HttpClientProxyPassword: "",
        Callback: {
            URL: "",
            Method: "",
            Headers: [""],
            RequestBody: "",
            Server: "",
            DisableCallbackSuccessContentCheck: false,
            CallbackSuccessContent: [""]
        },

    },
    WebhookEnable: false,
    WebhookURL: "",
    WebhookMethod: "",
    WebhookHeaders: [""],
    WebhookRequestBody: "",
    WebhookDisableCallbackSuccessContentCheck: false,
    WebhookSuccessContent: [""],
    WebhookProxy: "",
    WebhookProxyAddr: "",
    WebhookProxyUser: "",
    WebhookProxyPassword: "",

    TTL: "",
    TaskState: {
        WebhookCallTime: "",
        WebhookCallResult: false,
        WebhookCallErrorMsg: "",
        IPAddrHistory: [{ IPaddr: "", RecordTime: "" }],
        WebhookCallHistroy: [{ CallTime: "", CallResult: "" }],
        IpAddr: "",
        Domains: [{
            DomainName: "",
            SubDomain: "",
            UpdateStatus: "",
            LastUpdateStatusTime: "",
            Message: "",
            UpdateHistroy: [{ UpdateStatus: "", UpdateTime: "" }]
        }]
    },
}]);
taskList.value.splice(0, 1)


const rhtml = ref("")
const webhookSelect = ref("")
const WebhookServerListArea = ref("")

rhtml.value = ` <el-link type="info">info</el-link>`


const HttpProxyTypeList = [
    { value: "", label: "禁用" },
    { value: "http", label: "HTTP" },
    { value: "https", label: "HTTPS" },
    { value: "socks5", label: "SOCKS5" },
]

const WebhookProxyTypeList = [
    { value: "", label: "禁用" },
    { value: "dns", label: "使用DNS服务商同样设置" },
    { value: "http", label: "HTTP" },
    { value: "https", label: "HTTPS" },
    { value: "socks5", label: "SOCKS5" },
]

const TTLList = [
    { value: "", label: "自动" },
    { value: "1", label: "1秒" },
    { value: "5", label: "5秒" },
    { value: "10", label: "10秒" },
    { value: "60", label: "1分钟" },
    { value: "120", label: "2分钟" },
    { value: "600", label: "10分钟" },
    { value: "1800", label: "30分钟" },
    { value: "3600", label: "1小时" },
]

const GetTTLText = (ttl: string) => {
    for (let i in TTLList) {
        if (TTLList[i].value == ttl) {
            return TTLList[i].label
        }
    }
    return ttl + "秒"
}

const DNSCallbackServerList = [
    {
        value: 'meibu',
        label: '每步',
    },
    {
        value: 'noip',
        label: 'No-IP',
    },
    {
        value: 'dynv6',
        label: 'Dynv6',
    },
    {
        value: 'dynu',
        label: 'Dynu',
    },
    {
        value: 'other',
        label: '其它',
    },
]

const TCPNetworkTypeList = [
    {
        value: '',
        label: 'tcp',
    },
    {
        value: 'tcp4',
        label: 'tcp4',
    },
    {
        value: 'tcp6',
        label: 'tcp6',
    },
]

const WebHookTest = () => {
    console.log("WebHookTest")


    // WebhookHeaders        []string `json:"WebhookHeaders"`
    // WebhookSuccessContent []string `json:"WebhookSuccessContent"` //接口调用成功包含的内容


    let requestData = {
        WebhookURL: DDNSForm.value.WebhookURL,
        WebhookMethod: DDNSForm.value.WebhookMethod,
        WebhookRequestBody: DDNSForm.value.WebhookRequestBody,
        WebhookProxy: DDNSForm.value.WebhookProxy,
        WebhookProxyAddr: DDNSForm.value.WebhookProxyAddr,
        WebhookProxyUser: DDNSForm.value.WebhookProxyUser,
        WebhookProxyPassword: DDNSForm.value.WebhookProxyPassword,
        WebhookHeaders: StringToArrayList(DDNSFormWebhookHeadersArea.value),
        WebhookSuccessContent: StringToArrayList(DDNSFormWebhookSuccessContentArea.value)
    }


    apiWebhookTest(alterTaskKey.value, requestData).then((res) => {
        if (res.ret == 0) {
            console.log("apiWebhookTest: " + JSON.stringify(res))
            let msg = "Webhook接口调用结果:\n" + res.msg
            msg += "\n\n\n"
            msg += "Web接口反馈的完整内容:\n" + res.Response

            showErrorMessage(msg)
            return
        }
        MessageShow("error", res.msg)
    }).catch((error) => {
        console.log("webhook接口调用出错,error:" + error)
        MessageShow("error", "webhook接口调用出错")
    })
}

const getDNSCallbackServerLaber = (value: string) => {
    for (let i in DNSCallbackServerList) {
        if (DNSCallbackServerList[i].value == value) {
            return DNSCallbackServerList[i].label
        }
    }
    return "未支持的DNSCallbackServerLaber"
}

const defaultIPv6DNSServerList = [
    "[2001:4860:4860::8888]:53", //谷歌
    "[2001:4860:4860::8844]:53", //谷歌
    "[2606:4700:4700::64]:53",   //cloudflare
    "[2606:4700:4700::6400]:53", //cloudflare
    "[240C::6666]:53",           //下一代互联网北京研究中心
    "[240C::6644]:53",           //下一代互联网北京研究中心
    "[2402:4e00::]:53",          //dnspod
    //"[2400:3200::1]:53",         //阿里
    //		"[2400:3200:baba::1]:53",    //阿里
    "[240e:4c:4008::1]:53",  //中国电信
    "[240e:4c:4808::1]:53",  //中国电信
    "[2408:8899::8]:53",     //中国联通
    "[2408:8888::8]:53",     //中国联通
    "[2409:8088::a]:53",     //中国移动
    "[2409:8088::b]:53",     //中国移动
    "[2001:dc7:1000::1]:53", //CNNIC
    "[2400:da00::6666]:53",  //百度
]

const defaultIPv4DNSServerList = [
    "1.1.1.1:53",
    "1.2.4.8:53",
    "8.8.8.8:53",
    "9.9.9.9:53",
    "8.8.4.4:53",
    "114.114.114.114:53",
    "223.5.5.5:53",
    "223.6.6.6:53",
    "101.226.4.6:53",
    "218.30.118.6:53",
    "119.28.28.28:53",
]

const WebhookServerList = [
    {
        value: 'serverjiang',
        label: 'Server酱',
    },
    {
        value: 'bark',
        label: 'Bark',
    },
    {
        value: 'dingding',
        label: '钉钉',
    },
    {
        value: 'feishu',
        label: '飞书',
    },
    {
        value: 'weixinpro',
        label: '企业微信',
    },
    {
        value: 'custom',
        label: '自定义',
    },
]


const DNSServerList = [
    {
        value: 'alidns',
        label: 'Alidns(阿里云)',
    },
    {
        value: 'baiducloud',
        label: '百度云',
    },
    {
        value: 'cloudflare',
        label: 'Cloudflare',
    },
    {
        value: 'dnspod',
        label: 'Dnspod(腾讯云)',
    },

    {
        value: 'huaweicloud',
        label: '华为云',
    },
    {
        value: 'porkbun',
        label: 'Porkbun',
    },
    {
        value: 'callback',
        label: '自定义(Callback)',
    },
]

const CallbackMethodList = [
    {
        value: 'get',
        label: 'GET',
    },
    {
        value: 'post',
        label: 'POST',
    },
    {
        value: 'put',
        label: 'PUT',
    }
]


const getIPHistroyListHtml = (ipHistroy) => {
    let res = ""

    for (let i in ipHistroy) {
        let ipText = ipHistroy[i].IPaddr;
        if (ipText == "") {
            ipText = "获取IP失败"
        }

        res += ipHistroy[i].RecordTime + "&nbsp; &nbsp; &nbsp;" + ipText + '<br />'
    }

    return res
}

const GetSyncUpdateHistroyListHtml = (updateHistroy) => {
    let res = ""

    for (let i in updateHistroy) {
        let state = updateHistroy[i].UpdateStatus;

        res += updateHistroy[i].UpdateTime + "&nbsp; &nbsp; &nbsp;" + state + '<br />'
    }

    return res
}

const getWebhookCallHistroyListHtml = (histroy) => {
    let res = "仅记录程序本次启动以来的Webhook调用记录<br />"

    for (let i in histroy) {
        let result = histroy[i].CallResult;
        res += histroy[i].CallTime + "&nbsp; &nbsp; &nbsp;" + result + '<br />'
    }

    return res
}

const WebhookServerSelectChange = (server: string) => {
    switch (server) {
        case "dingding":
            let dingding_msg = {
                msgtype: "markdown",
                markdown: {
                    title: "DDNS域名同步反馈",
                    text: '#### DDNS域名同步反馈 \n - IP地址：#{ipAddr} \n - 域名更新成功列表：#{successDomainsLine}\n - 域名更新失败列表：#{failedDomainsLine}\n - Webhook触发时间:  \n  #{time}'
                },
            }
            WebhookServerListArea.value = JSON.stringify(dingding_msg, null, 2);
            break;
        case 'feishu':
            let feishu_msg = {
                msg_type: "post",
                content: {
                    post: {
                        zh_cn: {
                            title: "DDNS域名同步反馈",
                            content: [
                                [{ tag: "text", text: "IP地址：#{ipAddr}" }],
                                [{ tag: "text", text: "域名更新成功列表：#{su.ccessDomainsLine}" }],
                                [{ tag: "text", text: "域名更新失败列表：#{failedDomainsLine}" }],
                                [{ tag: "text", text: "Webhook触发时间: \n#{time}" }],
                            ]
                        }
                    }
                }
            }
            WebhookServerListArea.value = JSON.stringify(feishu_msg, null, 2)
            break
        case 'weixinpro':
            let weixin_msg = {
                msgtype: "markdown",
                markdown: {
                    content: '#### DDNS域名同步反馈 \n##### IP地址：\n#{ipAddr} \n##### 域名更新成功列表：\n#{successDomainsLine}\n##### 域名更新失败列表：\n#{failedDomainsLine}\n##### Webhook触发时间: \n#{time}'
                }
            }
            WebhookServerListArea.value = JSON.stringify(weixin_msg, null, 2)
            break
        default:
    }
}

const copyDomain = (SubDomain: string, domain: string) => {

    let content = SubDomain == '' ? domain : SubDomain + "." + domain;

    CopyTotoClipboard(content)
    MessageShow('success', '域名 ' + content + ' 已复制到剪切板')
}

const copyWanIP = (ip: string) => {
    if (ip == "") {
        return
    }
    CopyTotoClipboard(ip)
    MessageShow('success', 'IP ' + ip + ' 已复制到剪切板')
}


const IPRegTest = () => {
    // apiGetIPRegTest.

    apiGetIPRegTest(DDNSForm.value.TaskType, DDNSForm.value.NetInterface, DDNSForm.value.IPReg).then((res) => {
        //console.log(res.data)
        if (res.ret == 0) {
            //taskList.value = res.data
            console.log("IP选择匹配测试结果:" + res.ip)
            // MessageShow("success", "IP选择匹配测试结果:"+res.ip)
            let message = ""
            if (res.ip == "") {
                message = "IP选择匹配不到任何IP"
            } else {
                message = "IP选择匹配测试结果: " + res.ip
            }
            ShowMessageBox(message)
            return
        }

        MessageShow("error", "IP选择匹配测试出错")

    }).catch((error) => {
        console.log("IP选择匹配测试出错:" + error)
        MessageShow("error", "IP选择匹配测试出错")
    })
}


const GetDNSServerName = (dns: any) => {
    for (let i in DNSServerList) {
        if (DNSServerList[i].value == dns.Name) {
            if (dns.Name != "callback") {
                return DNSServerList[i].label
            }
            return getDNSCallbackServerLaber(dns.Callback.Server) + ' (自定义)'
        }
    }
    return "未知DNS服务商"
}

const getIDLabel = () => {
    switch (DDNSForm.value.DNS.Name) {
        case "alidns":
            return "AccessKey ID"
        case "dnspod":
            return "ID"
        case "cloudflare":
            return ""
        case "huaweicloud":
            return "Access Key Id"
        case "baiducloud":
            return "AccessKey ID"
        case "porkbun":
            return "API Key"
        case "callback":
            return "URL"
        default:
            return "未支持服务商类型"
    }
}



const getSecretLabel = () => {
    switch (DDNSForm.value.DNS.Name) {
        case "alidns":
            return "AccessKey Secret"
        case "dnspod":
            return "Token"
        case "cloudflare":
            return "Token"
        case "huaweicloud":
            return "Secret Access Key"
        case "baiducloud":
            return "AccessKey Secret"
        case "porkbun":
            return "Secret Key"
        case "callback":
            return "RequestBody"
        default:
            return "未支持服务商类型"
    }
}
const showDNSIDFormItem = () => {
    switch (DDNSForm.value.DNS.Name) {
        case "alidns":
            return true
        case "dnspod":
            return true
        case "cloudflare":
            return false
        case "huaweicloud":
            return true
        case "baiducloud":
            return true
        case "porkbun":
            return true
        case "callback":
            return false
        default:
            return false
    }
}


const showDNSSecretFormItem = () => {
    switch (DDNSForm.value.DNS.Name) {
        case "alidns":
            return true
        case "dnspod":
            return true
        case "cloudflare":
            return true
        case "huaweicloud":
            return true
        case "baiducloud":
            return true
        case "porkbun":
            return true
        case "callback":
            return false
        default:
            return false
    }
}







const ddnsTaskTypeChange = (label) => {


    console.log("ddnsTaskTypeChange label:" + label)

    if (DDNSForm.value.TaskType != preDDNSFrom.value.TaskType) {
        DDNSForm.value.URL = []
        DDNSForm.value.DNS.DNSServerList = []
    } else {
        DDNSForm.value.URL = preDDNSFrom.value.URL
        DDNSForm.value.DNS.DNSServerList = preDDNSFrom.value.DNS.DNSServerList
    }
    DDNSFormURLArea.value = getFormURLAreaValueFromFromURLList()
    DDNSFormDNSServerListArea.value = getFormDNSServerListAreaValueFromFromDNSServerList()

    queryNetinterfaces()
}

const addDDNSDialogVisible = ref(false)
const alterTaskKey = ref("")
const DDNSForm = ref(
    {
        TaskName: "",
        TaskType: "IPv6",
        Enable: true,
        GetType: "url",
        URL: [""],
        NetInterface: "",
        IPReg: "",
        Domains: [""],
        HttpClientTimeout: 60,
        DNS: {
            Name: "alidns",
            ID: "",
            Secret: "",
            ForceInterval: 3600,
            ResolverDoaminCheck: false,
            CallAPINetwork: "",
            DNSServerList: [""],
            HttpClientProxyType: "",
            HttpClientProxyAddr: "",
            HttpClientProxyUser: "",
            HttpClientProxyPassword: "",
            Callback: {
                URL: "",
                Method: "",
                Headers: [""],
                RequestBody: "",
                Server: "",
                DisableCallbackSuccessContentCheck: false,
                CallbackSuccessContent: [""],
            },
        },
        WebhookEnable: false,
        WebhookCallOnGetIPfail: false,
        WebhookURL: "",
        WebhookMethod: "",
        WebhookHeaders: [""],
        WebhookRequestBody: "",
        WebhookDisableCallbackSuccessContentCheck: false,
        WebhookSuccessContent: [""],
        WebhookProxy: "",
        WebhookProxyAddr: "",
        WebhookProxyUser: "",
        WebhookProxyPassword: "",
        TTL: ""
    }
)

const preDDNSFrom = ref(
    {
        TaskName: "",
        TaskType: "IPv6",
        Enable: true,
        GetType: "url",
        URL: [""],
        NetInterface: "",
        IPReg: "",
        Domains: [""],
        HttpClientTimeout: 20,
        DNS: {
            Name: "alidns",
            ID: "",
            Secret: "",
            ForceInterval: 3600,
            ResolverDoaminCheck: false,
            DNSServerList: [''],
            HttpClientProxyType: "",
            CallAPINetwork: "",
            HttpClientProxyAddr: "",
            HttpClientProxyUser: "",
            HttpClientProxyPassword: "",
            Callback: {
                URL: "",
                Method: "",
                Headers: [""],
                RequestBody: "",
                Server: "",
                DisableCallbackSuccessContentCheck: false,
                CallbackSuccessContent: [""],
            },
        },
        WebhookEnable: false,
        WebhookCallOnGetIPfail: false,
        WebhookURL: "",
        WebhookMethod: "",
        WebhookHeaders: [""],
        WebhookRequestBody: "",
        WebhookDisableCallbackSuccessContentCheck: false,
        WebhookSuccessContent: [""],
        WebhookProxy: "",
        WebhookProxyAddr: "",
        WebhookProxyUser: "",
        WebhookProxyPassword: "",
        TTL: ""
    }
)

const DDNSFormCallbackHeaderArea = ref("")
const DDNSFormURLArea = ref("")
const DDNSFormWebhookHeadersArea = ref("")
const DDNSFormWebhookSuccessContentArea = ref("")
const DDNSFormDNSServerListArea = ref("")
const DDNSFormCallbackSuccessContentArea = ref("")
const DDNSFormDomiansArea = ref("")
const DDNSFormOptionType = ref("")
const checkIPv4URLList = ["https://4.ipw.cn", "http://v4.ip.zxinc.org/getip", "https://myip4.ipip.net", "https://www.taobao.com/help/getip.php", "https://ddns.oray.com/checkip", "https://ip.3322.net", "https://v4.myip.la"]
const checkIPv6URLList = ["https://6.ipw.cn", "https://ipv6.ddnspod.com", "http://v6.ip.zxinc.org/getip", "https://speed.neu6.edu.cn/getIP.php", "https://v6.ident.me", "https://v6.myip.la"]

const showAddOrAlterDDNSTaskDialog = (optionType: string, task: any) => {
    //console.log("optionType fuck:" + optionType)

    webhookSelect.value = ""

    queryNetinterfaces()
    DDNSFormOptionType.value = optionType
    if (optionType == "add") {


        DDNSForm.value.TaskName = ""
        DDNSForm.value.TaskType = "IPv6"
        DDNSForm.value.Enable = true
        DDNSForm.value.GetType = "url"
        DDNSForm.value.URL = checkIPv6URLList
        DDNSForm.value.NetInterface = ""
        DDNSForm.value.IPReg = ""
        DDNSForm.value.Domains = [""]
        DDNSForm.value.HttpClientTimeout = 20,
            DDNSForm.value.DNS = {
                Name: "alidns",
                ID: "",
                Secret: "",
                ForceInterval: 3600,
                ResolverDoaminCheck: true,
                DNSServerList: [],
                HttpClientProxyType: "",
                CallAPINetwork: "",
                HttpClientProxyAddr: "",
                HttpClientProxyUser: "",
                HttpClientProxyPassword: "",
                Callback: {
                    URL: "",
                    Method: "get",
                    Headers: [""],
                    RequestBody: "",
                    Server: "other",
                    DisableCallbackSuccessContentCheck: false,
                    CallbackSuccessContent: [],
                }
            }
        DDNSForm.value.WebhookEnable = false
        DDNSForm.value.WebhookCallOnGetIPfail = false
        DDNSForm.value.WebhookURL = ""
        DDNSForm.value.WebhookMethod = "get"
        DDNSForm.value.WebhookHeaders = []
        DDNSForm.value.WebhookRequestBody = ""
        DDNSForm.value.WebhookDisableCallbackSuccessContentCheck = false,
            DDNSForm.value.WebhookSuccessContent = []
        DDNSForm.value.WebhookProxy = ""
        DDNSForm.value.WebhookProxyAddr = ""
        DDNSForm.value.WebhookProxyUser = ""
        DDNSForm.value.WebhookProxyPassword = ""
        DDNSForm.value.TTL = ""
        DDNSFormURLArea.value = getFormURLAreaValueFromFromURLList()
        DDNSFormDNSServerListArea.value = getFormDNSServerListAreaValueFromFromDNSServerList()
        DDNSFormWebhookHeadersArea.value = getFormWebhookHeadersAreaValueFromFormWebhookHeadersList()

        DDNSFormWebhookHeadersArea.value = ""
        DDNSFormWebhookSuccessContentArea.value = ""

        DDNSFormDomiansArea.value = ""
        DDNSFormCallbackSuccessContentArea.value = ""


        preDDNSFrom.value.TaskName = ""
        preDDNSFrom.value.TaskType = "IPv6"
        preDDNSFrom.value.Enable = true
        preDDNSFrom.value.GetType = "url"
        preDDNSFrom.value.URL = StringToArrayList(DDNSFormURLArea.value)
        preDDNSFrom.value.NetInterface = ""
        preDDNSFrom.value.IPReg = ""
        preDDNSFrom.value.Domains = [""]
        preDDNSFrom.value.HttpClientTimeout = 20,
            preDDNSFrom.value.DNS = {
                Name: "alidns",
                ID: "",
                Secret: "",
                ForceInterval: 3600,
                ResolverDoaminCheck: true,
                DNSServerList: [],
                HttpClientProxyType: "",
                CallAPINetwork: "",
                HttpClientProxyAddr: "",
                HttpClientProxyUser: "",
                HttpClientProxyPassword: "",
                Callback: {
                    URL: "",
                    Method: "get",
                    Headers: [""],
                    RequestBody: "",
                    DisableCallbackSuccessContentCheck: false,
                    CallbackSuccessContent: [],
                    Server: "other",
                }
            }
        preDDNSFrom.value.WebhookEnable = false
        preDDNSFrom.value.WebhookCallOnGetIPfail = false
        preDDNSFrom.value.WebhookURL = ""
        preDDNSFrom.value.WebhookMethod = "get"
        preDDNSFrom.value.WebhookHeaders = []
        preDDNSFrom.value.WebhookRequestBody = ""
        preDDNSFrom.value.WebhookDisableCallbackSuccessContentCheck = false,
            preDDNSFrom.value.WebhookSuccessContent = []
        preDDNSFrom.value.WebhookProxy = ""
        preDDNSFrom.value.WebhookProxyAddr = ""
        preDDNSFrom.value.WebhookProxyUser = ""
        preDDNSFrom.value.WebhookProxyPassword = ""
        preDDNSFrom.value.TTL = ""
    } else {

        DDNSForm.value.TaskName = task.TaskName
        DDNSForm.value.TaskType = task.TaskType
        DDNSForm.value.Enable = task.Enable
        DDNSForm.value.GetType = task.GetType
        DDNSForm.value.URL = task.URL
        DDNSForm.value.NetInterface = task.NetInterface
        DDNSForm.value.IPReg = task.IPReg
        DDNSForm.value.Domains = task.Domains
        DDNSForm.value.HttpClientTimeout = task.HttpClientTimeout
        DDNSForm.value.DNS = {
            Name: task.DNS.Name,
            ID: task.DNS.ID,
            Secret: task.DNS.Secret,
            ForceInterval: task.DNS.ForceInterval,
            ResolverDoaminCheck: task.DNS.ResolverDoaminCheck,
            DNSServerList: task.DNS.DNSServerList,
            HttpClientProxyType: task.DNS.HttpClientProxyType,
            CallAPINetwork: task.DNS.CallAPINetwork,
            HttpClientProxyAddr: task.DNS.HttpClientProxyAddr,
            HttpClientProxyUser: task.DNS.HttpClientProxyUser,
            HttpClientProxyPassword: task.DNS.HttpClientProxyPassword,
            Callback: {
                URL: task.DNS.Callback.URL,
                Method: task.DNS.Callback.Method,
                Headers: task.DNS.Callback.Headers,
                RequestBody: task.DNS.Callback.RequestBody,
                Server: task.DNS.Callback.Server,
                DisableCallbackSuccessContentCheck: task.DNS.Callback.DisableCallbackSuccessContentCheck,
                CallbackSuccessContent: task.DNS.Callback.CallbackSuccessContent
            }
        }
        DDNSForm.value.WebhookEnable = task.WebhookEnable
        DDNSForm.value.WebhookCallOnGetIPfail = task.WebhookCallOnGetIPfail
        DDNSForm.value.WebhookURL = task.WebhookURL
        DDNSForm.value.WebhookRequestBody = task.WebhookRequestBody
        DDNSForm.value.WebhookMethod = task.WebhookMethod
        DDNSForm.value.WebhookHeaders = task.WebhookHeaders
        DDNSForm.value.WebhookDisableCallbackSuccessContentCheck = task.WebhookDisableCallbackSuccessContentCheck
        DDNSForm.value.WebhookSuccessContent = task.WebhookSuccessContent
        DDNSForm.value.WebhookProxy = task.WebhookProxy
        DDNSForm.value.WebhookProxyAddr = task.WebhookProxyAddr
        DDNSForm.value.WebhookProxyUser = task.WebhookProxyUser
        DDNSForm.value.WebhookProxyPassword = task.WebhookProxyPassword
        DDNSForm.value.TTL = task.TTL
        DDNSFormURLArea.value = getFormURLAreaValueFromFromURLList()
        DDNSFormDomiansArea.value = getFormDomainsAreaValueFromFromDomainsList()
        DDNSFormCallbackHeaderArea.value = getFormCallbackHeaderAreaValueFromFormCallbackHeaderList()
        DDNSFormCallbackSuccessContentArea.value = getFormCallbackSuccessContentAreaValueFromFormCallbackSuccessContentList()
        DDNSFormDNSServerListArea.value = getFormDNSServerListAreaValueFromFromDNSServerList()
        DDNSFormWebhookHeadersArea.value = getFormWebhookHeadersAreaValueFromFormWebhookHeadersList()
        DDNSFormWebhookSuccessContentArea.value = getFormWebhookSuccessContentAreaValueFromFormWebhookSuccessContenttList()


        preDDNSFrom.value.TaskName = task.TaskName
        preDDNSFrom.value.TaskType = task.TaskType
        preDDNSFrom.value.Enable = task.Enable
        preDDNSFrom.value.GetType = task.GetType
        preDDNSFrom.value.URL = StringToArrayList(DDNSFormURLArea.value)
        preDDNSFrom.value.NetInterface = task.NetInterface
        preDDNSFrom.value.IPReg = task.IPReg
        preDDNSFrom.value.Domains = task.Domains
        preDDNSFrom.value.HttpClientTimeout = task.HttpClientTimeout
        preDDNSFrom.value.DNS = {
            Name: task.DNS.Name,
            ID: task.DNS.ID,
            Secret: task.DNS.Secret,
            ForceInterval: task.DNS.ForceInterval,
            ResolverDoaminCheck: task.DNS.ResolverDoaminCheck,
            DNSServerList: task.DNS.DNSServerList,
            HttpClientProxyType: task.DNS.HttpClientProxyType,
            CallAPINetwork: task.DNS.CallAPINetwork,
            HttpClientProxyAddr: task.DNS.HttpClientProxyAddr,
            HttpClientProxyUser: task.DNS.HttpClientProxyUser,
            HttpClientProxyPassword: task.DNS.HttpClientProxyPassword,
            Callback: {
                URL: task.DNS.Callback.URL,
                Method: task.DNS.Callback.Method,
                Headers: task.DNS.Callback.Headers,
                RequestBody: task.DNS.Callback.RequestBody,
                CallbackSuccessContent: task.DNS.Callback.CallbackSuccessContent,
                DisableCallbackSuccessContentCheck: task.DNS.Callback.DisableCallbackSuccessContentCheck,
                Server: task.DNS.Callback.Server,
            }
        }
        preDDNSFrom.value.WebhookEnable = task.WebhookEnable
        preDDNSFrom.value.WebhookMethod = task.WebhookMethod
        preDDNSFrom.value.WebhookHeaders = task.WebhookHeaders
        preDDNSFrom.value.WebhookCallOnGetIPfail = task.WebhookCallOnGetIPfail
        preDDNSFrom.value.WebhookURL = task.WebhookURL
        preDDNSFrom.value.WebhookDisableCallbackSuccessContentCheck = task.WebhookDisableCallbackSuccessContentCheck
        preDDNSFrom.value.WebhookSuccessContent = task.WebhookSuccessContent
        preDDNSFrom.value.WebhookRequestBody = task.WebhookRequestBody
        preDDNSFrom.value.WebhookProxy = task.WebhookProxy
        preDDNSFrom.value.WebhookProxyAddr = task.WebhookProxyAddr
        preDDNSFrom.value.WebhookProxyUser = task.WebhookProxyUser
        preDDNSFrom.value.WebhookProxyPassword = task.WebhookProxyPassword
        preDDNSFrom.value.TTL = task.TTL
        alterTaskKey.value = task.TaskKey
    }




    addDDNSDialogVisible.value = true

}

const ResolverDoaminCheckChange = (change) => {
    console.log("ResolverDoaminCheckChange: " + change);
    // DDNSForm.value.DNS.DNSServerList = 
}

const ruleEnableClick = (enable, task) => {
    const enableText = enable == false ? "启用" : "禁用";

    const taskName = "[" + task.TaskName + "]"

    return new Promise((resolve, reject) => {

        ElMessageBox.confirm(
            '确认要' + enableText + "DDNS任务 " + taskName + "?",
            'Warning',
            {
                confirmButtonText: '确认',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
            .then(() => {
                apiDDNSTaskEnable(task.TaskKey, !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "DDNS任务  " + taskName + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("error", "DDNS任务 " + taskName + enableText + "失败")

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                }).catch((error) => {
                    resolve(false)
                    console.log("DDNS任务 " + taskName + enableText + "失败" + ":请求出错" + error)
                    MessageShow("error", "DDNS任务 " + taskName + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}


const deleteTask = (task) => {

    const ruleName = "[" + task.TaskName + "]"

    const ruleText = ruleName

    ElMessageBox.confirm(
        '确认要删除DDNS任务 ' + ruleText + "?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            console.log("确认删除DDNS任务 " + ruleText)

            apiDeleteDDNSTask(task.TaskKey).then((res) => {
                if (res.ret == 0) {
                    //console.log("删除成功")
                    queryDDNSTaskList();
                    MessageShow("success", "删除成功")
                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }

                } else {
                    MessageShow("error", res.msg)
                }

            }).catch((error) => {
                console.log("删除DDNS任务失败,网络请求出错:" + error)
                MessageShow("error", "删除DDNS任务失败,网络请求出错")
            })
        })
        .catch(() => {

        })
}

const exeAddOrAlterDDNSOption = () => {


    // console.log("点击了添加: "+ DDNSFormOptionType.value)
    DDNSForm.value.URL = StringToArrayList(DDNSFormURLArea.value)
    DDNSForm.value.Domains = StringToArrayList(DDNSFormDomiansArea.value)
    DDNSForm.value.DNS.Callback.Headers = StringToArrayList(DDNSFormCallbackHeaderArea.value)
    DDNSForm.value.DNS.Callback.CallbackSuccessContent = StringToArrayList(DDNSFormCallbackSuccessContentArea.value)
    DDNSForm.value.DNS.DNSServerList = StringToArrayList(DDNSFormDNSServerListArea.value)
    DDNSForm.value.WebhookHeaders = StringToArrayList(DDNSFormWebhookHeadersArea.value)
    DDNSForm.value.WebhookSuccessContent = StringToArrayList(DDNSFormWebhookSuccessContentArea.value)

    if (DDNSForm.value.URL.length == 0) {
        if (DDNSForm.value.TaskType == "IPv6") {
            DDNSForm.value.URL = checkIPv6URLList
        } else {
            DDNSForm.value.URL = checkIPv4URLList
        }
        DDNSFormURLArea.value = getFormURLAreaValueFromFromURLList()
    }


    let checkDDNSFormAvalidRes = checkDDNSFormAvalid()

    // console.log("fuck[" + checkDDNSFormAvalidRes + "]")

    if (checkDDNSFormAvalidRes.length > 0) {
        ShowMessageBox(checkDDNSFormAvalidRes)
        return
    }


    if (DDNSFormOptionType.value == "add") {
        console.log("add DDNS任务")

        apiAddDDNSTask(DDNSForm.value).then((res) => {
            if (res.ret == 0) {
                addDDNSDialogVisible.value = false;
                MessageShow("success", "DDNS任务添加成功")

                console.log("刷新DDNS任务列表")
                queryDDNSTaskList()
                //queryTaskList();

                if (res.syncres != undefined && res.syncres != "") {
                    Notification("warn", res.syncres, 0)
                }
                return
            }
            MessageShow("error", res.msg)
        }).catch((error) => {
            console.log("DDNS任务添加失败,网络请求出错:" + error)
            MessageShow("error", "DDNS任务添加失败,网络请求出错")
        })
    } else {
        apiAlterDDNSTask(alterTaskKey.value, DDNSForm.value).then((res) => {
            if (res.ret == 0) {
                addDDNSDialogVisible.value = false;
                MessageShow("success", "DDNS任务修改成功")
                console.log("刷新DDNS任务列表")
                //queryTaskList();
                queryDDNSTaskList()
                if (res.syncres != undefined && res.syncres != "") {
                    Notification("warn", res.syncres, 0)
                }
                return
            }
            MessageShow("error", res.msg)
        }).catch((error) => {
            console.log("DDNS任务添加失败,网络请求出错:" + error)
            MessageShow("error", "DDNS任务添加失败,网络请求出错")
        })
    }

}




const queryDDNSTaskList = () => {
    apiGetDDNSTaskList().then((res) => {
        //console.log(res.data)
        if (res.ret == 0) {
            taskList.value = res.data
            return
        }

        if (res.ret == 6) {
            MessageShow("warning", "请先在动态域名设置中启用DDNS动态域名服务")
            location.hash = "#ddnsset"
            return
        }

    }).catch((error) => {
        console.log("获取DDNS任务列表出错:" + error)
        MessageShow("error", "获取DDNS任务列表出错")
    })
}

const queryNetinterfaces = () => {
    apiGetNetinterfaces().then((res) => {
        if (res.ret == 0) {
            netInterfaceList.value = res.data
            //console.log("data:"+res.data)
            // console.log("网卡列表：" + netInterfaceList.value)
            if (DDNSForm.value.TaskType == "IPv6") {
                currentNetInterface.value = netInterfaceList.value.IPv6NewInterfaces
            } else {
                currentNetInterface.value = netInterfaceList.value.IPv4NewInterfaces
            }

            if (DDNSForm.value.NetInterface == "" && currentNetInterface.value.length > 0) {
                DDNSForm.value.NetInterface = currentNetInterface.value[0].NetInterfaceName
            }
            // console.log("currentNetInterface "+ currentNetInterface.value )
        }



    }).catch((error) => {
        console.log("获取网卡列表出错:" + error)
        MessageShow("error", "获取网卡列表出错")
    })
}

const checkDDNSFormAvalid = () => {
    let data = DDNSForm.value
    //if (data.DNS)

    let checkDNSDataResMsg = checkDNSData(data.DNS)

    if (checkDNSDataResMsg.length > 0) {
        return checkDNSDataResMsg
    }



    if (data.Domains.length == 0) {
        return "域名列表不能为空"
    }

    if (data.GetType == "url") {
        if (data.URL.length <= 0) {
            return "接口列表不能为空"
        }
    }

    if (data.GetType == "netInterface") {
        if (data.NetInterface == "") {
            return "请选择获取IP的网卡"
        }
    }

    if (data.DNS.HttpClientProxyType != "") {
        if (data.DNS.HttpClientProxyAddr == "") {
            return "DNS 代理设置服务器地址不能为空"
        }
    }

    if (data.WebhookEnable) {
        if (data.WebhookURL == "") {
            return "请填写Webhook 接口地址"
        }

        if (data.WebhookMethod == "") {
            return "请选择Webhook 请求方法"
        }

        if (data.WebhookProxy != "" && data.WebhookProxy != "dns") {
            if (data.WebhookProxyAddr == "") {
                return "Webhook 代理设置服务器地址不能为空"
            }
        }

        if (!data.WebhookDisableCallbackSuccessContentCheck && data.WebhookSuccessContent.length == 0) {
            return "Webhook接口调用成功包含的字符串不能为空,如果要指定为空请禁用检测"
        }
    }





    //清空无用数据
    if (data.DNS.Name == "callback") {
        data.DNS.ID = ""
        data.DNS.Secret = ""
        if (data.DNS.Callback.Method == "get") {
            data.DNS.Callback.RequestBody = ""
        }



    } else {
        data.DNS.Callback = {
            URL: "",
            Method: "",
            Headers: [],
            RequestBody: "", Server: "", CallbackSuccessContent: [], DisableCallbackSuccessContentCheck: false
        }
    }

    if (data.WebhookEnable) {
        if (data.WebhookMethod == 'get') {
            data.WebhookRequestBody = ""
        }
    }



    return ""
}

const checkDNSData = (dns: any) => {
    //switch
    switch (dns.Name) {
        case "cloudflare":
            if (dns.Secret == "") {
                return "Cloudflare Token不能为空"
            }
            break;
        case "callback":
            if (dns.Callback.Method == "") {
                return "请选择Callback的请求方法"
            }

            if (dns.Callback.URL == "") {
                return "Callback 接口地址不能为空"
            }

            if (!dns.Callback.DisableCallbackSuccessContentCheck && dns.Callback.CallbackSuccessContent.length == 0) {
                return "接口调用成功包含的字符串不能为空,如果要指定为空请禁用检测"
            }

            break;
        default:
            if (dns.Secret == "" || dns.ID == "") {
                return "DNS服务商相关参数不能为空"
            }
    }
    return ""
}

const autocompleteCallbackForm = () => {
    ElMessageBox.confirm(
        '确认要根据Callback DNS服务商[' + getDNSCallbackServerLaber(DDNSForm.value.DNS.Callback.Server) + ']和公网IP类型:[' + DDNSForm.value.TaskType + ']自动填充参数模版?',
        'Warning',
        {
            confirmButtonText: '自动填充',
            cancelButtonText: '不需要',
            type: 'warning',
        }
    )
        .then(() => {
            switch (DDNSForm.value.DNS.Callback.Server) {
                case "meibu":
                    if (DDNSForm.value.TaskType == "IPv6") {
                        DDNSForm.value.DNS.Callback.URL = 'http://v6.meibu.com/v6.asp?name=#{domain}&pwd=这里替换为你的密码'
                    } else {
                        DDNSForm.value.DNS.Callback.URL = 'http://ipv4.meibu.com/ipv4.asp?ID=lucky&name=#{domain}&pwd=这里替换为你的密码'

                    }
                    DDNSFormCallbackHeaderArea.value = ""
                    DDNSForm.value.DNS.Callback.Headers = []
                    DDNSForm.value.DNS.Callback.Method = 'get'
                    DDNSForm.value.DNS.Callback.RequestBody = ""
                    DDNSForm.value.DNS.Callback.DisableCallbackSuccessContentCheck = false
                    DDNSForm.value.DNS.Callback.CallbackSuccessContent = ["chenggong", "chongfu", "ok"]
                    DDNSFormCallbackSuccessContentArea.value = 'chenggong\nchongfu\nok'
                    break;
                case "noip":
                    DDNSForm.value.DNS.Callback.URL = 'http://你的账号:你的密码@dynupdate.no-ip.com/nic/update?hostname=#{domain}&myip=#{ip}'
                    DDNSFormCallbackHeaderArea.value = ""
                    DDNSForm.value.DNS.Callback.Headers = []
                    DDNSForm.value.DNS.Callback.Method = 'get'
                    DDNSForm.value.DNS.Callback.RequestBody = ""
                    DDNSForm.value.DNS.Callback.DisableCallbackSuccessContentCheck = false
                    DDNSForm.value.DNS.Callback.CallbackSuccessContent = ["nochg #{ip}", "good #{ip}"]
                    DDNSFormCallbackSuccessContentArea.value = 'nochg #{ip}\ngood #{ip}'
                    break;
                case "dynv6":
                    if (DDNSForm.value.TaskType == "IPv6") {
                        DDNSForm.value.DNS.Callback.URL = 'https://dynv6.com/api/update?hostname=#{domain}&token=这里替换为你的Token&ipv6=#{ip}'
                    } else {
                        DDNSForm.value.DNS.Callback.URL = 'https://dynv6.com/api/update?hostname=#{domain}&token=这里替换为你的Token&ipv4=#{ip}'
                    }
                    DDNSFormCallbackHeaderArea.value = ""
                    DDNSForm.value.DNS.Callback.Headers = []
                    DDNSForm.value.DNS.Callback.Method = 'get'
                    DDNSForm.value.DNS.Callback.RequestBody = ""
                    DDNSForm.value.DNS.Callback.DisableCallbackSuccessContentCheck = false
                    DDNSForm.value.DNS.Callback.CallbackSuccessContent = ["addresses updated"]
                    DDNSFormCallbackSuccessContentArea.value = 'addresses updated\n'
                    break;
                case "dynu":
                    if (DDNSForm.value.TaskType == "IPv6") {
                        DDNSForm.value.DNS.Callback.URL = 'https://api.dynu.com/nic/update?hostname=#{domain}&myip=no&myipv6=#{ip}&username=用户名&password=登录密码或IP更新密码'
                    } else {
                        DDNSForm.value.DNS.Callback.URL = 'https://api.dynu.com/nic/update?hostname=#{domain}&myip=#{ip}&myipv6=no}&username=用户名&password=登录密码或IP更新密码'
                    }
                    DDNSFormCallbackHeaderArea.value = ""
                    DDNSForm.value.DNS.Callback.Headers = []
                    DDNSForm.value.DNS.Callback.Method = 'get'
                    DDNSForm.value.DNS.Callback.RequestBody = ""
                    DDNSForm.value.DNS.Callback.DisableCallbackSuccessContentCheck = false
                    DDNSForm.value.DNS.Callback.CallbackSuccessContent = ["nochg", "good #{ip}"]
                    DDNSFormCallbackSuccessContentArea.value = 'nochg\ngood #{ip}'
                    break;
                    break
                default:
            }


        })
        .catch(() => {

        })
}


const getFormURLAreaValueFromFromURLList = () => {
    if (DDNSForm.value.URL == null || DDNSForm.value.URL.length <= 0 || (DDNSForm.value.URL.length == 1 && DDNSForm.value.URL[0] == "")) {
        if (DDNSForm.value.TaskType == "IPv6") {
            DDNSForm.value.URL = checkIPv6URLList
        } else {
            DDNSForm.value.URL = checkIPv4URLList
        }
    }
    var res = ""
    for (let i in DDNSForm.value.URL) {
        if (i == "0") {
            res = DDNSForm.value.URL[i]
        } else {
            res += "\n" + DDNSForm.value.URL[i]
        }
    }

    return res
}

const getFormDNSServerListAreaValueFromFromDNSServerList = () => {

    if (DDNSForm.value.DNS.DNSServerList == null || DDNSForm.value.DNS.DNSServerList.length <= 0 || (DDNSForm.value.DNS.DNSServerList.length == 1 && DDNSForm.value.DNS.DNSServerList[0] == "")) {
        if (DDNSForm.value.TaskType == "IPv6") {
            DDNSForm.value.DNS.DNSServerList = defaultIPv6DNSServerList
        } else {
            DDNSForm.value.DNS.DNSServerList = defaultIPv4DNSServerList
        }
    }


    var res = ""
    for (let i in DDNSForm.value.DNS.DNSServerList) {
        if (i == "0") {
            res = DDNSForm.value.DNS.DNSServerList[i]
        } else {
            res += "\n" + DDNSForm.value.DNS.DNSServerList[i]
        }
    }

    return res
}


const getFormDomainsAreaValueFromFromDomainsList = () => {
    if (DDNSForm.value.Domains == null || DDNSForm.value.Domains.length == 0) {
        return ""
    }


    var res = ""
    for (let i in DDNSForm.value.Domains) {
        if (i == "0") {
            res = DDNSForm.value.Domains[i]
        } else {
            res += "\n" + DDNSForm.value.Domains[i]
        }
    }

    return res
}


const getFormCallbackHeaderAreaValueFromFormCallbackHeaderList = () => {
    if (DDNSForm.value.DNS.Callback.Headers == null || DDNSForm.value.DNS.Callback.Headers.length == 0) {
        return ""
    }


    var res = ""
    for (let i in DDNSForm.value.DNS.Callback.Headers) {
        if (i == "0") {
            res = DDNSForm.value.DNS.Callback.Headers[i]
        } else {
            res += "\n" + DDNSForm.value.DNS.Callback.Headers[i]
        }
    }
    return res
}

const getFormWebhookHeadersAreaValueFromFormWebhookHeadersList = () => {
    if (DDNSForm.value.WebhookHeaders == null || DDNSForm.value.WebhookHeaders.length == 0) {
        return ""
    }

    var res = ""
    for (let i in DDNSForm.value.WebhookHeaders) {
        if (i == "0") {
            res = DDNSForm.value.WebhookHeaders[i]
        } else {
            res += "\n" + DDNSForm.value.WebhookHeaders[i]
        }
    }
    return res
}


const getFormCallbackSuccessContentAreaValueFromFormCallbackSuccessContentList = () => {
    if (DDNSForm.value.DNS.Callback.CallbackSuccessContent == null || DDNSForm.value.DNS.Callback.CallbackSuccessContent.length == 0) {
        return ""
    }


    var res = ""
    for (let i in DDNSForm.value.DNS.Callback.CallbackSuccessContent) {
        if (i == "0") {
            res = DDNSForm.value.DNS.Callback.CallbackSuccessContent[i]
        } else {
            res += "\n" + DDNSForm.value.DNS.Callback.CallbackSuccessContent[i]
        }
    }
    return res
}

const getFormWebhookSuccessContentAreaValueFromFormWebhookSuccessContenttList = () => {
    if (DDNSForm.value.WebhookSuccessContent == null || DDNSForm.value.WebhookSuccessContent.length == 0) {
        return ""
    }


    var res = ""
    for (let i in DDNSForm.value.WebhookSuccessContent) {
        if (i == "0") {
            res = DDNSForm.value.WebhookSuccessContent[i]
        } else {
            res += "\n" + DDNSForm.value.WebhookSuccessContent[i]
        }
    }
    return res
}

const deleteWhiteList = (index, item) => {
    // ElMessageBox.confirm(
    //     '确认要删除IP [' + item.IP + "]的白名单记录?",
    //     'Warning',
    //     {
    //         confirmButtonText: '确认',
    //         cancelButtonText: '取消',
    //         type: 'warning',
    //     }
    // )
    //     .then(() => {
    //         apiDeleteWhiteList(item.IP).then((res) => {
    //             if (res.ret == 0) {
    //                 whitelist.value.splice(index, 1)
    //                 return
    //             }
    //             MessageShow("error", res.msg)

    //         }).catch((error) => {
    //             MessageShow("error", "删除[" + item.IP + "]的白名单记录出错")
    //         })
    //     })
    //     .catch(() => {

    //     })
}





// const keydown = (e) => {
//     if (e.keyCode != 13) {
//         return
//     }
//     if (!addDDNSDialogVisible.value) {
//         return
//     }
//     exeAddOrAlterDDNSOption()
// }

var timerID: any
onMounted(() => {
    queryDDNSTaskList();

    timerID = setInterval(() => {
        queryDDNSTaskList();
    }, 500);
})

onUnmounted(() => {
    clearInterval(timerID)
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

.fromitemDivRadius {
    border: 5px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 530px;
    padding-left: 9px;
    padding-right: 9px;
}

.fromitemChildDivRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 495px;
    padding-top: 9px;
    padding-left: 9px;
    padding-right: 9px;
}

.affix-container {
    text-align: center;
    border-radius: 4px;
    width: 3vw;

    background: var(--el-color-primary-light-9);
}
</style>