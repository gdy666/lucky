<template>

    <div class="ReverseProxyPageRadius" :style="{
        borderRadius: 'base',
    }">

        <el-scrollbar height="100%">

            <div class="itemradius" :style="{
                borderRadius: 'base',
            }" v-for="rule in ruleList">

                <el-descriptions :column="8" border>
                    <el-descriptions-item label="规则名称" :span="1">
                        <el-button size="small" v-show="true">
                            {{ rule.RuleName == '' ? '未命名规则' : rule.RuleName }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="监听类型" :span="2">
                        <el-button color="#0059b3" size="small" v-show="true">
                            {{ rule.Network }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="监听地址" :span="2">
                        <el-button type="success" size="small" v-show="true">
                            {{ rule.ListenIP == '' ? '所有地址' : rule.ListenIP }}
                        </el-button>
                    </el-descriptions-item>


                    <el-descriptions-item label="监听端口" :span="1">
                        <el-button color="#409eff" size="small" v-show="true">
                            {{ rule.ListenPort }}
                        </el-button>

                        
                        <el-button :type="rule.EnableTLS!=true?'info':'primary'" size="small" v-show="true">
                            {{ rule.EnableTLS==true?'TLS已启用':'TLS未启用' }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="规则操作" :span="2">
                        <el-tooltip :content="rule.Enable == true ? '规则已启用' : '规则已禁用'" placement="top">
                            <el-switch v-model="rule.Enable" inline-prompt active-text="开" inactive-text="关"
                                :before-change="ruleEnableClick.bind(this, rule.Enable, rule)" size="large" />
                        </el-tooltip>

                        &nbsp;&nbsp;
                        <el-button size="default" type="primary"
                            @click="showAddOrAlterReverseProxyRuleDialog('alter',rule)">编辑</el-button>
                        <el-button size="default" type="danger" @click="deleteReverseProxyRule(rule)">删除</el-button>
                    </el-descriptions-item>











                </el-descriptions>


                <el-descriptions :column="8" border>

                    <el-descriptions-item label="默认子规则" :span="1">
                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                默认规则<br />
                            </template>

                            <el-button type="info" size="small" v-show="true">
                                默认规则
                            </el-button>
                        </el-tooltip>
                    </el-descriptions-item>


                    <el-descriptions-item label="前端域名" :span="1">

                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                未匹配子规则的任意域名
                            </template>
                            <el-button color="#409eff" size="small" v-show="true">
                                未匹配的域名
                            </el-button>
                        </el-tooltip>
                    </el-descriptions-item>

                    <el-descriptions-item label="后端地址" :span="2">
                        <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="StrArrayListToBrHtml(rule.DefaultProxy.Locations)"></span>
                            </template>
                            <el-button color="#409eff" size="small" v-show="true">
                                {{
                                (rule.DefaultProxy.Locations==undefined||rule.DefaultProxy.Locations==null)||rule.DefaultProxy.Locations.length
                                <=0?'未设置':rule.DefaultProxy.Locations.length==1?rule.DefaultProxy.Locations[0]:rule.DefaultProxy.Locations[0]+'...'}}
                                    </el-button>
                        </el-tooltip>
                    </el-descriptions-item>



                    <el-descriptions-item label="日志" :span="2">
                        <!-- <el-button color="#409eff" size="small" v-show="true">
                            {{ rule.EnableAccessLog==true?'开启':'关闭' }}
                        </el-button>

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                访问日志最大存储条数<br />
                            </template>
                            <el-button color="#409eff" size="small" v-show="true">
                                {{ rule.AccessLogMaxNum }}
                            </el-button>
                        </el-tooltip> -->


                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                <span v-html="reverproxyLogsArrayToTooltipHtml(rule,'default')"></span>
                            </template>
                            <el-button type="info" size="small" v-show="true"
                                @click="showReverproxyLogs(rule.RuleKey,'default')">
                                查看
                            </el-button>
                        </el-tooltip>

                    </el-descriptions-item>

                    <el-descriptions-item label="安全设置" :span="2">

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content v-if="rule.DefaultProxy.EnableBasicAuth">
                                用户名:{{rule.DefaultProxy.BasicAuthUser}}<br />
                                密码:{{rule.DefaultProxy.BasicAuthPasswd}}<br />
                            </template>
                            <template #content v-if="!rule.DefaultProxy.EnableBasicAuth">
                                Basic认证未启用<br />
                            </template>
                            <el-button color="#6666ff" size="small"
                            v-show="true" :disabled="rule.DefaultProxy.EnableBasicAuth == true ? false : true">
                                {{ rule.DefaultProxy.EnableBasicAuth==false?'Basic认证未启用':'Basic认证已启用' }}
                            </el-button>
                        </el-tooltip>

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                IP过滤模式<br /></template>
                            <el-button color="#6666ff" size="small" v-show="true">
                                {{ rule.DefaultProxy.SafeIPMode=='blacklist'?'IP黑名单':'IP白名单' }}
                            </el-button>
                        </el-tooltip>

                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                            <template #content>
                                UserAgent过滤模式<br /></template>
                            <el-button color="#6666ff" size="small" v-show="true">
                                {{ rule.DefaultProxy.SafeUserAgentMode=='blacklist'?'UA黑名单':'UA白名单' }}
                            </el-button>
                        </el-tooltip>
                    </el-descriptions-item>



                    <div v-for="proxy in rule.ProxyList">

                        <el-descriptions-item label="自定义子规则" :span="1">
                            <el-button size="small" v-show="true">
                                {{ proxy.Remark==''?'未命名子规则': proxy.Remark}}
                            </el-button>
                            &nbsp;
                            <el-tooltip :content="proxy.Enable == true ? '子规则已启用' : '子规则已禁用'" placement="top">
                                <el-switch v-model="proxy.Enable" inline-prompt active-text="开" inactive-text="关"
                                    :before-change="subruleEnableClick.bind(this, proxy.Enable, rule,proxy)"
                                    size="small" />
                            </el-tooltip>

                        </el-descriptions-item>

                        <el-descriptions-item label="前端域名" :span="1">

                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="StrArrayListToBrHtml(proxy.Domains)"></span>
                                </template>
                                <el-button color="#409eff" size="small" v-show="true">
                                    {{ proxy.Domains.length==1?proxy.Domains[0]:proxy.Domains[0]+' ...' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="后端地址" :span="2">
                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="StrArrayListToBrHtml(proxy.Locations)"></span>
                                </template>
                                <el-button color="#409eff" size="small" v-show="true">
                                    {{ proxy.Locations.length==1?proxy.Locations[0]:proxy.Locations[0]+' ...' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="日志" :span="2">

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="reverproxyLogsArrayToTooltipHtml(rule,proxy.Key)"></span>
                                </template>
                                <el-button type="info" size="small" v-show="true"
                                    @click="showReverproxyLogs(rule.RuleKey,proxy.Key)">
                                    查看
                                </el-button>
                            </el-tooltip>


                        </el-descriptions-item>

                        <el-descriptions-item label="安全设置" :span="2">

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content v-if="proxy.EnableBasicAuth">
                                    用户名:{{proxy.BasicAuthUser}}<br />
                                    密码:{{proxy.BasicAuthPasswd}}<br />
                                </template>
                                <template #content v-if="!proxy.EnableBasicAuth">
                                    Basic认证未启用<br />
                                </template>
                                <el-button color="#6666ff" size="small" v-show="true" :disabled="proxy.EnableBasicAuth == true ? false : true">
                                    {{ proxy.EnableBasicAuth==false?'Basic认证未启用':'Basic认证已启用' }}
                                </el-button>
                            </el-tooltip>

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    IP过滤模式<br /></template>
                                <el-button color="#6666ff" size="small" v-show="true">
                                    {{ proxy.SafeIPMode=='blacklist'?'IP黑名单':'IP白名单' }}
                                </el-button>
                            </el-tooltip>

                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    UserAgent过滤模式<br /></template>
                                <el-button color="#6666ff" size="small" v-show="true">
                                    {{ proxy.SafeUserAgentMode=='blacklist'?'UA黑名单':'UA白名单' }}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>


                    </div>


                </el-descriptions>

            </div>


        </el-scrollbar>

        <el-affix position="bottom" :offset="30" class="affix-container">
            <el-button type="primary" :round=true @click="showAddOrAlterReverseProxyRuleDialog('add', null)">添加反向代理规则
                <el-icon class="el-icon--right">
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix>
    </div>

    <el-dialog v-if="addRuleDialogVisible" v-model="addRuleDialogVisible"
        :title="ruleFormOptionType == 'add' ? '反向代理规则添加' : '反向代理规则修改'" draggable :show-close="true"
        :close-on-click-modal="false" width="600px">

        <el-form :model="ruleForm">

            <el-form-item label="反向代理规则名称" label-width="auto">
                <el-input v-model="ruleForm.RuleName" placeholder="可留空" autocomplete="off" />
            </el-form-item>

            <el-form-item label="规则开关" label-width="auto">
                <el-switch v-model="ruleForm.Enable" inline-prompt width="50px" active-text="启用" inactive-text="停用" />
            </el-form-item>

            <div v-show="ruleForm.Enable">

                <div class="fromitemDivRadius">
                    <el-form-item label="监听类型" label-width="auto">
                        <el-checkbox v-model="checkAllListenType" :indeterminate="listenTypeIsIndeterminate"
                            @change="handleCheckAllChange">全选 &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;</el-checkbox>
                        <br />
                        <el-checkbox-group v-model="ruleFormListenType" @change="handleCheckedProxyTypesChange">
                            <el-checkbox v-for="t in listenTypes" :key="t" :label="t">{{
                            t
                            }}</el-checkbox>
                        </el-checkbox-group>
                    </el-form-item>

                    <el-form-item label="监听地址" label-width="auto">
                        <el-input v-model="ruleForm.ListenIP" placeholder="没特殊需求留空即可" autocomplete="off" />
                    </el-form-item>


                    <el-form-item label="监听端口" label-width="auto" :min="1" :max="65535">
                        <el-input-number v-model="ruleForm.ListenPort" autocomplete="off" />
                    </el-form-item>

                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                    <template #content>
                                        启用前请先添加SSL证书<br/>
                                        增加删除证书后需要重启规则新证书才生效<br/>
                                    </template>
                    <el-form-item label="TLS" label-width="auto" v-if="true">
                        <el-switch v-model="ruleForm.EnableTLS" inline-prompt width="50px" active-text="启用"
                            inactive-text="禁用" />
                    </el-form-item>
                </el-tooltip>




                </div>

                <el-collapse v-model="fromChildARuleActiveName" :accordion="true">
                    <div class="fromitemDivRadius">
                        <el-collapse-item title="默认子规则" name="default">

                            <div class="fromitemChildDivRadius">

                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                    <template #content>
                                        可留空<br />
                                        不为空时表示域名不匹配任何子规则时默认的跳转地址<br />
                                        设置多条后端地址时表示启用均衡负载,依次循环访问<br />
                                        无特殊需求留空即可<br />
                                    </template>
                                    <el-form-item label-width="auto" label="默认目标地址">
                                        <el-input v-model="ruleFormDefaultProxyLocationsArea" placeholder="没特殊需求留空即可"
                                            :autosize="{ minRows: 1, maxRows: 3 }" type="textarea">
                                        </el-input>
                                    </el-form-item>
                                </el-tooltip>



                                <el-form-item label="访问日志记录" label-width="auto">
                                    <el-switch v-model="ruleForm.DefaultProxy.EnableAccessLog" inline-prompt
                                        width="50px" active-text="开启" inactive-text="关闭" />
                                </el-form-item>




                                <div v-show="ruleForm.DefaultProxy.EnableAccessLog">

                                    <el-form-item label="日志输出级别" label-width="auto">
                                        <el-select v-model="ruleForm.DefaultProxy.LogLevel" class="m-2"
                                            placeholder="请选择">
                                            <el-option v-for="item in LogLevelList" :key="item.value"
                                                :label="item.label" :value="item.value" />
                                        </el-select>
                                    </el-form-item>

                                    <el-form-item label="日志输出到终端" label-width="auto">
                                        <el-switch v-model="ruleForm.DefaultProxy.LogOutputToConsole" inline-prompt
                                            width="50px" active-text="开启" inactive-text="关闭" />
                                    </el-form-item>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            范围(0-102400),0表示不保存日志<br />
                                        </template>
                                        <el-form-item label="访问日志记录最大条数" label-width="auto" :min="0" :max="102400">
                                            <el-input-number v-model="ruleForm.DefaultProxy.AccessLogMaxNum"
                                                autocomplete="off" />
                                        </el-form-item>
                                    </el-tooltip>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            范围(1-256)<br />
                                        </template>
                                        <el-form-item label="列表显示最新日志最大条数" label-width="auto" :min="1" :max="256">
                                            <el-input-number v-model="ruleForm.DefaultProxy.WebListShowLastLogMaxCount"
                                                autocomplete="off" />
                                        </el-form-item>
                                    </el-tooltip>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            不建议留空，留空的话日志中不显示相关信息 <br />
                                            格式中可用变量<br />
                                            #{clientIP} : 客户端IP <br />
                                            #{remoteIP} : 客户端直接连接本服务的IP(如果前端有反向代理，不一定是客户端真实IP)<br />
                                            #{tab} : 制表符<br />
                                            #{method} : 请求方法<br />
                                            #{host} : 请求host<br />
                                            #{path} : 请求path(不包含host)部分<br />
                                            #{url} : 请求url(不包含host)部分<br />

                                        </template>
                                        <el-form-item label-width="auto" label="请求信息在日志中的格式" v-if="false">
                                            <el-input v-model="ruleForm.DefaultProxy.RequestInfoLogFormat"
                                                placeholder="不建议留空，留空的话日志中不显示相关信息">
                                            </el-input>
                                        </el-form-item>
                                    </el-tooltip>
                                </div>




                                <div class="fromitemChildSafeDivRadius">
                                    <p>客户端IP获取设置</p>
                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            无特殊需求，一般情况下不需要打开这个开关<br />
                                        </template>

                                        <el-form-item label="优先从Header头部获取" label-width="auto">
                                            <el-switch v-model="ruleForm.DefaultProxy.ForwardedByClientIP" inline-prompt
                                                width="50px" active-text="启用" inactive-text="禁用" />
                                        </el-form-item>
                                    </el-tooltip>

                                    <div v-show="ruleForm.DefaultProxy.ForwardedByClientIP">

                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                信任的代理IP网段,每行填写一个网段<br />
                                                Lucky只会从信任的代理IP中的header获取客户端IP<br />
                                                0.0.0.0/0 表示信任任意Header包含IP信息的IPv4代理地址<br />
                                                ::/0 表示信任任意Header包含IP信息的IPv6代理地址<br />
                                            </template>
                                            <el-form-item label-width="auto" label="信任的代理IP网段">
                                                <el-input v-model="ruleFormTrustedCIDRsStrListArea"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                包含客户端IP的头部字段，每行填写一个字段<br />
                                                常见的字段有：<br />
                                                X-Forwarded-For <br />
                                                X-Real-IP <br />
                                            </template>
                                            <el-form-item label-width="auto" label="包含客户端IP的头部字段">
                                                <el-input v-model="ruleFormRemoteIPHeaderstArea"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                    </div>



                                </div>

                                <div class="fromitemChildSafeDivRadius">
                                    <p>追加客户端IP到指定Header</p>
                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            无特殊需求，一般情况下不需要打开这个开关<br />
                                        </template>

                                        <el-form-item label="追加客户端IP到指定Header" label-width="auto">
                                            <el-switch v-model="ruleForm.DefaultProxy.AddRemoteIPToHeader" inline-prompt
                                                width="50px" active-text="启用" inactive-text="禁用" />
                                        </el-form-item>
                                    </el-tooltip>

                                    <div v-show="ruleForm.DefaultProxy.AddRemoteIPToHeader">


                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                常用的Key有<br />
                                                X-Forwarded-For<br />
                                                X-Real-IP<br />
                                            </template>
                                            <el-form-item label-width="auto" label="自定义Header Key">
                                                <el-input v-model="ruleForm.DefaultProxy.AddRemoteIPHeaderKey"
                                                    placeholder="">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                    </div>

                                </div>

                                <div class="fromitemChildSafeDivRadius">

                                    <p>安全设置</p>

                                    <el-form-item label="BasicAuth认证" label-width="auto">
                                        <el-switch v-model="ruleForm.DefaultProxy.EnableBasicAuth" inline-prompt
                                            width="50px" active-text="启用" inactive-text="禁用" />
                                    </el-form-item>

                                    <div v-show="ruleForm.DefaultProxy.EnableBasicAuth">
                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                HTTP BasicAuth 用户名<br />
                                            </template>
                                            <el-form-item label-width="auto" label="HTTP BasicAuth 用户名">
                                                <el-input v-model="ruleForm.DefaultProxy.BasicAuthUser"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                HTTP BasicAuth 密码<br />
                                            </template>
                                            <el-form-item label-width="auto" label="HTTP BasicAuth 密码">
                                                <el-input v-model="ruleForm.DefaultProxy.BasicAuthPasswd"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>


                                    </div>



                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            没特殊使用黑名单模式即可<br />
                                        </template>
                                        <el-form-item label-width="auto" label="IP过滤模式">
                                            <el-radio-group v-model="ruleForm.DefaultProxy.SafeIPMode" class="ml-4">
                                                <el-radio label="blacklist">黑名单</el-radio>
                                                <el-radio label="whitelist">白名单</el-radio>
                                            </el-radio-group>
                                        </el-form-item>
                                    </el-tooltip>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            没特殊使用黑名单模式即可<br />
                                        </template>
                                        <el-form-item label-width="auto" label="UserAgent过滤模式">
                                            <el-radio-group v-model="ruleForm.DefaultProxy.SafeUserAgentMode"
                                                class="ml-4">
                                                <el-radio label="blacklist">黑名单</el-radio>
                                                <el-radio label="whitelist">白名单</el-radio>
                                            </el-radio-group>
                                        </el-form-item>
                                    </el-tooltip>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            自定义的UserAgent 黑/白名单内容,多组Agent分多行填入,实际的UserAgent部分匹配任意一行即是成功匹配<br />
                                            黑名单模式时,匹配成功任一条即拒绝服务<br />
                                            白名单模式时,仅匹配成功才继续服务<br />
                                        </template>
                                        <el-form-item label-width="auto" label="UserAgent过滤内容">
                                            <el-input v-model="ruleFormDefaultProxyUserAgentfilterArea"
                                                :autosize="{ minRows: 3, maxRows: 6 }" placeholder="" type="textarea"
                                                wrap="off">
                                            </el-input>
                                        </el-form-item>
                                    </el-tooltip>
                                </div>

                                <div class="fromitemChildSafeDivRadius">
                                    <p>隐私设置</p>
                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            可以通过自定义robots.txt防止爬虫对内容的抓取<br />
                                        </template>
                                        <el-form-item label="自定义robot.txt" label-width="auto">
                                            <el-switch v-model="ruleForm.DefaultProxy.CustomRobotTxt" inline-prompt
                                                width="50px" active-text="启用" inactive-text="禁用" />
                                        </el-form-item>
                                    </el-tooltip>
                                    <div v-show="ruleForm.DefaultProxy.CustomRobotTxt">
                                        <el-form-item label-width="auto" label="robot.txt">
                                            <el-input v-model="ruleForm.DefaultProxy.RobotTxt"
                                                :autosize="{ minRows: 5, maxRows: 9 }" placeholder="" type="textarea">
                                            </el-input>
                                        </el-form-item>
                                    </div>
                                </div>


                            </div>
                        </el-collapse-item>


                    </div>

                    <div class="fromitemDivRadius">
                        <p>反向代理子规则列表</p>



                        <div class="fromitemChildDivRadius" v-for="(proxy,index) in ruleForm.ProxyList">
                            <el-collapse-item :title="'第'+ (index+1) +'条 '+'[ '+ruleForm.ProxyList[index].Remark+' ]'"
                                :name=index>




                                <div>

                                    <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                        <template #content>
                                            子规则名称,可留空<br />
                                        </template>
                                        <el-form-item label="反向代理子规则名称" label-width="auto">
                                            <el-input v-model="ruleForm.ProxyList[index].Remark" placeholder="可留空"
                                                autocomplete="off" />
                                        </el-form-item>
                                    </el-tooltip>


                                    <el-form-item label="子规则开关" label-width="auto">
                                        <el-switch v-model="ruleForm.ProxyList[index].Enable" inline-prompt width="50px"
                                            active-text="启用" inactive-text="禁用" />
                                    </el-form-item>

                                    <div v-show="ruleForm.ProxyList[index].Enable">

                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                一行一条域名<br />
                                                设置多条域名时表示多条域名指向相同的后端地址<br />
                                            </template>
                                            <el-form-item label-width="auto" label="前端域名">
                                                <el-input v-model="ruleFormProxyDomainsArea[index]"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                        <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                            <template #content>
                                                一行一条后端地址<br />
                                                设置多条后端地址时表示启用均衡负载,依次循环访问<br />
                                                无特殊要求设置一条即可<br />
                                            </template>
                                            <el-form-item label-width="auto" label="后端地址">
                                                <el-input v-model="ruleFormProxyLocationsArea[index]"
                                                    :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                    type="textarea">
                                                </el-input>
                                            </el-form-item>
                                        </el-tooltip>

                                        <!--
                                [#{clientIP}][#{remoteIP}]#{tab}[#{method}][#{host}#{path}]

                            -->

                                        <el-form-item label="访问日志记录" label-width="auto">
                                            <el-switch v-model="ruleForm.ProxyList[index].EnableAccessLog" inline-prompt
                                                width="50px" active-text="开启" inactive-text="关闭" />
                                        </el-form-item>

                                        <div v-show="ruleForm.ProxyList[index].EnableAccessLog">

                                            <el-form-item label="日志输出级别" label-width="auto">
                                                <el-select v-model="ruleForm.ProxyList[index].LogLevel" class="m-2"
                                                    placeholder="请选择">
                                                    <el-option v-for="item in LogLevelList" :key="item.value"
                                                        :label="item.label" :value="item.value" />
                                                </el-select>
                                            </el-form-item>

                                            <el-form-item label="子规则日志输出到终端" label-width="auto">
                                                <el-switch v-model="ruleForm.ProxyList[index].LogOutputToConsole"
                                                    inline-prompt width="50px" active-text="启用" inactive-text="禁用" />
                                            </el-form-item>


                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    范围(0-102400),0表示不保存日志<br />
                                                </template>
                                                <el-form-item label="访问日志记录最大条数" label-width="auto" :min="0"
                                                    :max="102400">
                                                    <el-input-number v-model="ruleForm.ProxyList[index].AccessLogMaxNum"
                                                        autocomplete="off" />
                                                </el-form-item>
                                            </el-tooltip>

                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    范围(1-256)<br />
                                                </template>
                                                <el-form-item label="前端列表显示最新日志最大条数" label-width="auto" :min="1"
                                                    :max="256">
                                                    <el-input-number
                                                        v-model="ruleForm.ProxyList[index].WebListShowLastLogMaxCount"
                                                        autocomplete="off" />
                                                </el-form-item>
                                            </el-tooltip>




                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    不建议留空，留空的话日志中不显示相关信息 <br />

                                                    日志格式中可用变量<br />
                                                    #{clientIP} : 客户端IP <br />
                                                    #{remoteIP} : 客户端直接连接本服务的IP(如果前端有反向代理，不一定是客户端真实IP)<br />
                                                    #{tab} : 制表符<br />
                                                    #{method} : 请求方法<br />
                                                    #{host} : 请求host<br />
                                                    #{path} : 请求path(不包含host)部分<br />
                                                    #{url} : 请求url(不包含host)部分<br />
                                                </template>
                                                <el-form-item label="请求信息在日志中的格式" label-width="auto" v-if="false">
                                                    <el-input v-model="ruleForm.ProxyList[index].RequestInfoLogFormat"
                                                        placeholder="不建议留空，留空的话日志中不显示相关信息" autocomplete="off" />
                                                </el-form-item>
                                            </el-tooltip>

                                        </div>

                                        <div class="fromitemChildSafeDivRadius">
                                            <p>客户端IP获取设置</p>

                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    无特殊需求，一般情况下不需要打开这个开关<br />
                                                </template>
                                                <el-form-item label="优先从Header头部获取" label-width="auto">
                                                    <el-switch v-model="ruleForm.ProxyList[index].ForwardedByClientIP"
                                                        inline-prompt width="50px" active-text="启用"
                                                        inactive-text="禁用" />
                                                </el-form-item>
                                            </el-tooltip>

                                            <div v-show="ruleForm.ProxyList[index].ForwardedByClientIP">

                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        信任的代理IP网段,每行填写一个网段<br />
                                                        Lucky只会从信任的代理IP中的header获取客户端IP<br />
                                                        0.0.0.0/0 表示信任任意Header包含IP信息的IPv4代理地址<br />
                                                        ::/0 表示信任任意Header包含IP信息的IPv6代理地址<br />
                                                    </template>
                                                    <el-form-item label-width="auto" label="信任的代理IP网段">
                                                        <el-input v-model="ruleFormProxyTrustedCIDRsStrListArea[index]"
                                                            :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                            type="textarea">
                                                        </el-input>
                                                    </el-form-item>
                                                </el-tooltip>

                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        包含客户端IP的头部字段，每行填写一个字段<br />
                                                        常见的字段有：<br />
                                                        X-Forwarded-For <br />
                                                        X-Real-IP <br />
                                                    </template>
                                                    <el-form-item label-width="auto" label="包含客户端IP的头部字段">
                                                        <el-input v-model="ruleFormProxyRemoteIPHeadersArea[index]"
                                                            :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                            type="textarea">
                                                        </el-input>
                                                    </el-form-item>
                                                </el-tooltip>

                                            </div>



                                        </div>

                                        <div class="fromitemChildSafeDivRadius">
                                            <p>追加客户端IP到指定Header</p>
                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    无特殊需求，一般情况下不需要打开这个开关<br />
                                                    用于后端程序识别客户端真实IP
                                                </template>

                                                <el-form-item label="追加客户端IP到指定Header" label-width="auto">
                                                    <el-switch v-model="ruleForm.ProxyList[index].AddRemoteIPToHeader"
                                                        inline-prompt width="50px" active-text="启用"
                                                        inactive-text="禁用" />
                                                </el-form-item>
                                            </el-tooltip>

                                            <div v-show="ruleForm.ProxyList[index].AddRemoteIPToHeader">


                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        常用的Key有<br />
                                                        X-Forwarded-For<br />
                                                        X-Real-IP<br />
                                                    </template>
                                                    <el-form-item label-width="auto" label="自定义Header Key">
                                                        <el-input
                                                            v-model="ruleForm.ProxyList[index].AddRemoteIPHeaderKey"
                                                            placeholder="">
                                                        </el-input>
                                                    </el-form-item>
                                                </el-tooltip>

                                            </div>

                                        </div>

                                        <div class="fromitemChildSafeDivRadius">

                                            <p>安全设置</p>

                                            <el-form-item label="BasicAuth认证" label-width="auto">
                                                <el-switch v-model="ruleForm.ProxyList[index].EnableBasicAuth"
                                                    inline-prompt width="50px" active-text="启用" inactive-text="禁用" />
                                            </el-form-item>

                                            <div v-show="ruleForm.ProxyList[index].EnableBasicAuth">
                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        HTTP BasicAuth 用户名<br />
                                                    </template>
                                                    <el-form-item label-width="auto" label="HTTP BasicAuth 用户名">
                                                        <el-input v-model="ruleForm.ProxyList[index].BasicAuthUser"
                                                            :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                            type="textarea">
                                                        </el-input>
                                                    </el-form-item>
                                                </el-tooltip>

                                                <el-tooltip class="box-item" effect="dark" :trigger-keys="[]"
                                                    content="">
                                                    <template #content>
                                                        HTTP BasicAuth 密码<br />
                                                    </template>
                                                    <el-form-item label-width="auto" label="HTTP BasicAuth 密码">
                                                        <el-input v-model="ruleForm.ProxyList[index].BasicAuthPasswd"
                                                            :autosize="{ minRows: 1, maxRows: 3 }" placeholder=""
                                                            type="textarea">
                                                        </el-input>
                                                    </el-form-item>
                                                </el-tooltip>


                                            </div>



                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    没特殊使用黑名单模式即可<br />
                                                </template>
                                                <el-form-item label-width="auto" label="IP过滤模式">
                                                    <el-radio-group v-model="ruleForm.ProxyList[index].SafeIPMode"
                                                        class="ml-4">
                                                        <el-radio label="blacklist">黑名单</el-radio>
                                                        <el-radio label="whitelist">白名单</el-radio>
                                                    </el-radio-group>
                                                </el-form-item>
                                            </el-tooltip>

                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    没特殊使用黑名单模式即可<br />
                                                </template>
                                                <el-form-item label-width="auto" label="UserAgent过滤模式">
                                                    <el-radio-group
                                                        v-model="ruleForm.ProxyList[index].SafeUserAgentMode"
                                                        class="ml-4">
                                                        <el-radio label="blacklist">黑名单</el-radio>
                                                        <el-radio label="whitelist">白名单</el-radio>
                                                    </el-radio-group>
                                                </el-form-item>
                                            </el-tooltip>

                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    自定义的UserAgent 黑/白名单内容,多组Agent分多行填入,实际的UserAgent部分匹配任意一行即是成功匹配<br />
                                                    黑名单模式时,匹配成功任一条即拒绝服务<br />
                                                    白名单模式时,仅匹配成功才继续服务<br />
                                                </template>
                                                <el-form-item label-width="auto" label="UserAgent过滤内容">
                                                    <el-input v-model="ruleFormProxyUserAgentfilterArea[index]"
                                                        :autosize="{ minRows: 3, maxRows: 6 }" placeholder=""
                                                        type="textarea" wrap="off">
                                                    </el-input>
                                                </el-form-item>
                                            </el-tooltip>
                                        </div>

                                        <div class="fromitemChildSafeDivRadius">
                                            <p>隐私设置</p>
                                            <el-tooltip class="box-item" effect="dark" :trigger-keys="[]" content="">
                                                <template #content>
                                                    可以通过自定义robots.txt防止爬虫对内容的抓取<br />
                                                </template>
                                                <el-form-item label="自定义robot.txt" label-width="auto">
                                                    <el-switch v-model="ruleForm.ProxyList[index].CustomRobotTxt"
                                                        inline-prompt width="50px" active-text="启用"
                                                        inactive-text="禁用" />
                                                </el-form-item>
                                            </el-tooltip>
                                            <div v-show="ruleForm.ProxyList[index].CustomRobotTxt">
                                                <el-form-item label-width="auto" label="robot.txt">
                                                    <el-input v-model="ruleForm.ProxyList[index].RobotTxt"
                                                        :autosize="{ minRows: 5, maxRows: 9 }" placeholder=""
                                                        type="textarea">
                                                    </el-input>
                                                </el-form-item>
                                            </div>
                                        </div>

                                    </div>





                                    <el-form-item>
                                        <el-button type="danger" :round=true
                                            @click="deleteProxyToRuleFormProxyList(index)">
                                            删除</el-button>
                                    </el-form-item>

                                </div>

                            </el-collapse-item>
                        </div>





                        <el-button type="primary" :round=true @click="addProxyToRuleFormProxyList">添加反向代理转发子规则
                        </el-button>
                    </div>
                </el-collapse>
            </div>



        </el-form>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="addRuleDialogVisible = false">取消</el-button>
                <el-button type="primary" @click="exeAddOrAlterRuleOption">{{ ruleFormOptionType == "add" ? '添加' :
                '修改'
                }}
                </el-button>
            </span>
        </template>
    </el-dialog>


    <el-dialog v-if="reverproxyLogsVisible" v-model="reverproxyLogsVisible" :close-on-click-modal="false" width="900px">

        <div>

            <el-scrollbar max-height="95vh" class="reverseProxyLogs" element-loading-background="transparent">
                {{reverseProxyLogsDialogLogsContentView}}

            </el-scrollbar>


            <el-pagination :page-size=reverseProxyLogsPageSize :page-sizes="[10,20,50, 100, 200, 300,400,500]" :small="false"
                :disabled="false" :background="false" layout="total, sizes, prev, pager, next, jumper"
                :current-page="reverseProxyLogsDialogCurrentPage" :total=reverseProxyLogsTotal
                @size-change="handleReverseRroxyLogsSizeChange" @current-change="handleReverproxyLogsCurrentChange"
                @prev-click="handleReverproxyLogsPreClick" @next-click="handleReverproxyLogsNextClick" />
        </div>

    </el-dialog>

</template>


<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessageBox, CheckboxValueType } from 'element-plus'
import { MessageShow, Notification, ShowMessageBox } from '../utils/ui'
import {
    apiAddReverseProxyRule,
    apiGeReverseProxyRuleList,
    apiAlterReverseProxyRule,
    apiDeleteReverseProxyRule,
    apiReverseProxyRuleEnable,
    apiReverseProxyRuleLogs
} from '../apis/utils'

import { StringToArrayList, CopyTotoClipboard, StrArrayListToBrHtml, StrArrayListToArea, LogLevelList } from '../utils/utils'


const addRuleDialogVisible = ref(false)
const reverproxyLogsVisible = ref(false)
const ruleFormOptionType = ref("")
const fromChildARuleActiveName = ref(0)
const reverseProxyLogsPageSize = ref(50)
const reverseProxyLogsTotal = ref(0)
const reverseProxyLogsDialogData = ref([{
    ProxyKey: "",
    ClientIP: "",
    LogContent: "",
    LogTime: ""
}])
const reverseProxyLogsDialogRuleKey = ref("")
const reverseProxyLogsDialogProxyKey = ref("")
const reverseProxyLogsDialogCurrentPage = ref(1)

const reverseProxyLogsDialogLogsContentView = ref("")


const ruleList = ref([{
    RuleName: "",
    RuleKey: "",
    Enable: false,
    Network: "",
    ListenIP: "",
    ListenPort: 666,
    EnableTLS: false,
    DefaultProxy: {
        Locations: [""],
        EnableAccessLog: true,
        LogLevel: 4,
        LogOutputToConsole: false,
        AccessLogMaxNum: 100,
        WebListShowLastLogMaxCount: 30,
        RequestInfoLogFormat: "",
        ForwardedByClientIP: false,
        TrustedCIDRsStrList: [""],
        RemoteIPHeaders: [""],
        AddRemoteIPToHeader: false,
        AddRemoteIPHeaderKey: "",
        EnableBasicAuth: true,
        BasicAuthUser: "",
        BasicAuthPasswd: "",
        SafeIPMode: "",
        SafeUserAgentMode: "",
        UserAgentfilter: [""],

        CustomRobotTxt: false,
        RobotTxt: ""
    },

    ProxyList: [
        {
            Enable: true,
            Key: "",
            Remark: "",
            Domains: [""],
            Locations: [""],
            LogLevel: 4,
            LogOutputToConsole: false,
            AccessLogMaxNum: 1000,
            WebListShowLastLogMaxCount: 30,
            RequestInfoLogFormat: "",
            ForwardedByClientIP: false,
            TrustedCIDRsStrList: [""],
            RemoteIPHeaders: [""],
            EnableBasicAuth: false,
            BasicAuthUser: "",
            BasicAuthPasswd: "",
            SafeIPMode: "",
            SafeUserAgentMode: "",
            UserAgentfilter: [""],

            CustomRobotTxt: false,
            RobotTxt: ""
        }
    ],
}

])

const ruleForm = ref({
    RuleName: "",
    RuleKey: "",
    Enable: false,
    Network: "",
    ListenIP: "",
    ListenPort: 666,
    EnableTLS: false,
    DefaultProxy: {
        Key: "",
        Locations: [""],
        EnableAccessLog: true,
        LogLevel: 4,
        LogOutputToConsole: false,
        AccessLogMaxNum: 100,
        WebListShowLastLogMaxCount: 30,
        RequestInfoLogFormat: "",
        ForwardedByClientIP: false,
        TrustedCIDRsStrList: [""],
        RemoteIPHeaders: [""],
        AddRemoteIPToHeader: false,
        AddRemoteIPHeaderKey: "",
        EnableBasicAuth: true,
        BasicAuthUser: "",
        BasicAuthPasswd: "",
        SafeIPMode: "",
        SafeUserAgentMode: "",
        UserAgentfilter: [""],
        CustomRobotTxt: false,
        RobotTxt: ""
    },

    ProxyList: [
        {
            Enable: true,
            Key: "",
            Remark: "",
            Domains: [""],
            Locations: [""],
            EnableAccessLog: true,
            LogLevel: 4,
            LogOutputToConsole: false,
            AccessLogMaxNum: 1000,
            WebListShowLastLogMaxCount: 30,
            RequestInfoLogFormat: "",
            ForwardedByClientIP: false,
            TrustedCIDRsStrList: [""],
            RemoteIPHeaders: [""],
            AddRemoteIPToHeader: false,
            AddRemoteIPHeaderKey: "",

            EnableBasicAuth: false,
            BasicAuthUser: "",
            BasicAuthPasswd: "",
            SafeIPMode: "",
            SafeUserAgentMode: "",
            UserAgentfilter: [""],

            CustomRobotTxt: false,
            RobotTxt: ""

        }
    ],
})
const ruleFormTrustedCIDRsStrListArea = ref("")
const ruleFormRemoteIPHeaderstArea = ref("")

const ruleFormDefaultProxyUserAgentfilterArea = ref("")
const ruleFormDefaultProxyLocationsArea = ref("")
const ruleFormListenType = ref([""])
const ruleFormProxyDomainsArea = ref([""])
const ruleFormProxyLocationsArea = ref([""])
const ruleFormProxyUserAgentfilterArea = ref([""])
const ruleFormProxyTrustedCIDRsStrListArea = ref([""])
const ruleFormProxyRemoteIPHeadersArea = ref([""])

const showReverproxyLogs = (ruleKey, proxyKey) => {
    reverproxyLogsVisible.value = true

    reverseProxyLogsDialogCurrentPage.value = 1
    reverseProxyLogsTotal.value = 0
    reverseProxyLogsPageSize.value = 10
    reverseProxyLogsDialogRuleKey.value = ruleKey
    reverseProxyLogsDialogProxyKey.value = proxyKey


    apiReverseProxyRuleLogs(ruleKey, proxyKey, reverseProxyLogsPageSize.value, reverseProxyLogsDialogCurrentPage.value).then((res) => {
        //console.log(res.data)
        if (res.ret == 0) {
            reverseProxyLogsPageSize.value = res.pageSize
            reverseProxyLogsTotal.value = res.total
            reverseProxyLogsDialogData.value = res.logs
            flushReverseProxyLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })
}

const handleReverproxyLogsPreClick = (page: number) => {
    // console.log(page)
    reverseProxyLogsDialogCurrentPage.value = page - 1

}


const handleReverproxyLogsNextClick = (page: number) => {
    //console.log(page)
    reverseProxyLogsDialogCurrentPage.value = page + 1

}



const handleReverseRroxyLogsSizeChange = (pageSize: number) => {
    reverseProxyLogsPageSize.value = pageSize
    reverseProxyLogsDialogCurrentPage.value = 1

    apiReverseProxyRuleLogs(reverseProxyLogsDialogRuleKey.value, reverseProxyLogsDialogProxyKey.value, reverseProxyLogsPageSize.value, reverseProxyLogsDialogCurrentPage.value).then((res) => {
        if (res.ret == 0) {
            reverseProxyLogsPageSize.value = res.pageSize
            reverseProxyLogsTotal.value = res.total
            reverseProxyLogsDialogData.value = res.logs
            flushReverseProxyLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })

}


const handleReverproxyLogsCurrentChange = (page: number) => {
    reverseProxyLogsDialogCurrentPage.value = page

    apiReverseProxyRuleLogs(reverseProxyLogsDialogRuleKey.value, reverseProxyLogsDialogProxyKey.value, reverseProxyLogsPageSize.value, reverseProxyLogsDialogCurrentPage.value).then((res) => {
        if (res.ret == 0) {
            reverseProxyLogsPageSize.value = res.pageSize
            reverseProxyLogsTotal.value = res.total
            reverseProxyLogsDialogData.value = res.logs
            flushReverseProxyLogsDialogLogsContentView()
            return
        }
        MessageShow("error", "获取日志出错")
    }).catch((error) => {
        console.log("获取日志出错:" + error)
        MessageShow("error", "获取日志出错")
    })
}

const flushReverseProxyLogsDialogLogsContentView = () => {
    //reverseProxyLogsDialogLogsContentView
    // {{log.LogTime}} &nbsp; &nbsp; &nbsp; {{log.LogContent}}
    reverseProxyLogsDialogLogsContentView.value = ""
    for (let index in reverseProxyLogsDialogData.value) {
        let log = reverseProxyLogsDialogData.value[index]
        if (index!="0"){
            reverseProxyLogsDialogLogsContentView.value += "\n"
        }
        reverseProxyLogsDialogLogsContentView.value += log.LogTime + "\t" + log.LogContent + "\n"
    }
   
}

const showAddOrAlterReverseProxyRuleDialog = (optionType: string, rule: any) => {
    addRuleDialogVisible.value = true
    ruleFormOptionType.value = optionType

    fromChildARuleActiveName.value = -1

    switch (optionType) {
        case "add":
            {
                ruleForm.value = {
                    RuleName: "",
                    RuleKey: "",
                    Enable: true,
                    Network: "tcp6",
                    ListenIP: "",
                    ListenPort: 16666,
                    EnableTLS: false,
                    DefaultProxy: {
                        Key: "default",
                        Locations: [],
                        EnableAccessLog: true,
                        LogLevel: 4,
                        LogOutputToConsole: false,
                        AccessLogMaxNum: 1000,
                        WebListShowLastLogMaxCount: 10,
                        RequestInfoLogFormat: "[#{clientIP}][#{remoteIP}]#{tab}[#{method}][#{host}#{url}]",
                        ForwardedByClientIP: false,
                        TrustedCIDRsStrList: [""],
                        RemoteIPHeaders: [""],
                        AddRemoteIPToHeader: false,
                        AddRemoteIPHeaderKey: "",
                        EnableBasicAuth: false,
                        BasicAuthUser: "",
                        BasicAuthPasswd: "",
                        SafeIPMode: "blacklist",
                        SafeUserAgentMode: "blacklist",
                        UserAgentfilter: [""],
                        CustomRobotTxt: false,
                        RobotTxt: "User-agent:  *\nDisallow:  /"
                    },
                    ProxyList: [],
                }
                ruleFormProxyDomainsArea.value = []
                ruleFormProxyLocationsArea.value = []
                ruleFormProxyUserAgentfilterArea.value = []
                ruleFormProxyTrustedCIDRsStrListArea.value = []
                ruleFormProxyRemoteIPHeadersArea.value = []
                // addProxyToRuleFormProxyList()
                ruleFormTrustedCIDRsStrListArea.value = `0.0.0.0/0
::/0`
                ruleFormRemoteIPHeaderstArea.value = `X-Forwarded-For
X-Real-IP`
                ruleFormDefaultProxyUserAgentfilterArea.value = ""
                ruleFormDefaultProxyLocationsArea.value = ""
            }
            break;
        case "alter":
            {
                ruleForm.value = rule
                ruleFormProxyDomainsArea.value = []
                ruleFormProxyLocationsArea.value = []
                ruleFormProxyUserAgentfilterArea.value = []
                ruleFormProxyTrustedCIDRsStrListArea.value = []
                ruleFormProxyRemoteIPHeadersArea.value = []
                ruleFormTrustedCIDRsStrListArea.value = StrArrayListToArea(ruleForm.value.DefaultProxy.TrustedCIDRsStrList)
                ruleFormRemoteIPHeaderstArea.value = StrArrayListToArea(ruleForm.value.DefaultProxy.RemoteIPHeaders)
                ruleFormDefaultProxyLocationsArea.value = StrArrayListToArea(ruleForm.value.DefaultProxy.Locations)
                ruleFormDefaultProxyUserAgentfilterArea.value = StrArrayListToArea(ruleForm.value.DefaultProxy.UserAgentfilter)
                for (let i in ruleForm.value.ProxyList) {
                    ruleFormProxyDomainsArea.value.push(StrArrayListToArea(ruleForm.value.ProxyList[i].Domains))
                    ruleFormProxyLocationsArea.value.push(StrArrayListToArea(ruleForm.value.ProxyList[i].Locations))
                    ruleFormProxyUserAgentfilterArea.value.push(StrArrayListToArea(ruleForm.value.ProxyList[i].UserAgentfilter))
                    ruleFormProxyTrustedCIDRsStrListArea.value.push(StrArrayListToArea(ruleForm.value.ProxyList[i].TrustedCIDRsStrList))
                    ruleFormProxyRemoteIPHeadersArea.value.push(StrArrayListToArea(ruleForm.value.ProxyList[i].RemoteIPHeaders))
                }
            }
            break;
        default:
    }



    if (ruleForm.value.Network == "tcp4") {
        ruleFormListenType.value = ["tcp4"]
    } else if (ruleForm.value.Network == "tcp6") {
        ruleFormListenType.value = ["tcp6"]
    } else if (ruleForm.value.Network == "tcp") {
        ruleFormListenType.value = ["tcp4", "tcp6"]
    }
}



const addProxyToRuleFormProxyList = () => {
    ruleFormProxyDomainsArea.value.push("")
    ruleFormProxyLocationsArea.value.push("")
    ruleFormProxyUserAgentfilterArea.value.push("")
    ruleFormProxyTrustedCIDRsStrListArea.value.push(`0.0.0.0/0
::/0`)
    ruleFormProxyRemoteIPHeadersArea.value.push(`X-Forwarded-For
X-Real-IP`)

    ruleForm.value.ProxyList.push({
        Enable: true,
        Key: "",
        Remark: "",
        Domains: [""],
        Locations: [""],
        EnableAccessLog: true,
        LogLevel: 4,
        LogOutputToConsole: false,
        AccessLogMaxNum: 1000,
        WebListShowLastLogMaxCount: 10,
        RequestInfoLogFormat: "[#{clientIP}][#{remoteIP}]#{tab}[#{method}][#{host}#{url}]",
        ForwardedByClientIP: false,
        TrustedCIDRsStrList: [""],
        RemoteIPHeaders: [""],
        AddRemoteIPToHeader: false,
        AddRemoteIPHeaderKey: "",
        EnableBasicAuth: false,
        BasicAuthUser: "",
        BasicAuthPasswd: "",
        SafeIPMode: "blacklist",
        SafeUserAgentMode: "blacklist",
        UserAgentfilter: [""],
        CustomRobotTxt: false,
        RobotTxt: "User-agent:  *\nDisallow:  /"
    })

    var len = ruleForm.value.ProxyList.length
    fromChildARuleActiveName.value = len - 1
}

const deleteProxyToRuleFormProxyList = (index: number) => {


    ElMessageBox.confirm(
        '确认要删除第 ' + (index + 1) + " 条反向代理设置?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {

        ruleForm.value.ProxyList.splice(index, 1)
        ruleFormProxyDomainsArea.value.splice(index, 1)
        ruleFormProxyLocationsArea.value.splice(index, 1)
        ruleFormProxyUserAgentfilterArea.value.splice(index, 1)
        ruleFormProxyTrustedCIDRsStrListArea.value.splice(index, 1)
        ruleFormProxyRemoteIPHeadersArea.value.splice(index, 1)
    })
}

const deleteReverseProxyRule = (rule) => {
    var ruleName = rule.RuleName == '' ? '未命名' : rule.RuleName

    ElMessageBox.confirm(
        '确认要删除 ' + ruleName + " 反向代理规则?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {

        apiDeleteReverseProxyRule(rule.RuleKey).then((res) => {
            if (res.ret == 0) {
                MessageShow("success", "反向代理规则删除成功")
                queryRuleList()
                return
            }
            MessageShow("error", res.msg)
        }).catch((error) => {
            console.log("反向代理规则删除失败,网络请求出错:" + error)
            MessageShow("error", "反向代理规则删除失败,网络请求出错")
        })

    })
}

const ruleEnableClick = (enable, rule) => {
    const enableText = enable == false ? "启用" : "禁用";

    const ruleName = "[" + (rule.RuleName == '' ? '未命名' : rule.RuleName) + "]"

    return new Promise((resolve, reject) => {

        ElMessageBox.confirm(
            '确认要' + enableText + "反向代理规则 " + ruleName + "?",
            'Warning',
            {
                confirmButtonText: '确认',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
            .then(() => {
                apiReverseProxyRuleEnable(rule.RuleKey, "", !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "反向代理规则  " + ruleName + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("error", "反向代理规则 " + ruleName + enableText + "失败: " + res.msg)

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                }).catch((error) => {
                    resolve(false)
                    console.log("反向代理规则 " + ruleName + enableText + "失败" + ":请求出错" + error)
                    MessageShow("error", "反向代理规则 " + ruleName + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}

const subruleEnableClick = (enable, rule, proxy) => {
    const enableText = enable == false ? "启用" : "禁用";

    const ruleName = "[" + (rule.RuleName == '' ? '未命名' : rule.RuleName) + "]"
    const proxyName = "[" + (proxy.Remark == '' ? '未命名' : proxy.Remark) + "]"

    return new Promise((resolve, reject) => {

        ElMessageBox.confirm(
            '确认要' + enableText + "反向代理规则 " + ruleName + "的子规则 " + proxyName + "?",
            'Warning',
            {
                confirmButtonText: '确认',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
            .then(() => {
                apiReverseProxyRuleEnable(rule.RuleKey, proxy.Key, !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "反向代理规则  " + ruleName + "的子规则 " + proxyName + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("error", "反向代理规则 " + ruleName + "的子规则 " + proxyName + enableText + "失败: " + res.msg)

                    if (res.syncres != undefined && res.syncres != "") {
                        Notification("warn", res.syncres, 0)
                    }
                }).catch((error) => {
                    resolve(false)
                    console.log("反向代理规则 " + ruleName + "的子规则 " + proxyName + enableText + "失败" + ":请求出错" + error)
                    MessageShow("error", "反向代理规则 " + ruleName + "的子规则 " + proxyName + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}



const exeAddOrAlterRuleOption = () => {

    // if (ruleForm.value.ProxyList.length <= 0) {
    //     MessageShow("error", "至少添加一条反向代理转发规则")
    //     return
    // }

    ruleForm.value.DefaultProxy.TrustedCIDRsStrList = StringToArrayList(ruleFormTrustedCIDRsStrListArea.value)
    ruleForm.value.DefaultProxy.RemoteIPHeaders = StringToArrayList(ruleFormRemoteIPHeaderstArea.value)
    ruleForm.value.DefaultProxy.UserAgentfilter = StringToArrayList(ruleFormDefaultProxyUserAgentfilterArea.value)
    ruleForm.value.DefaultProxy.Locations = StringToArrayList(ruleFormDefaultProxyLocationsArea.value)
    // area 转换
    for (let i in ruleForm.value.ProxyList) {
        ruleForm.value.ProxyList[i].Domains = StringToArrayList(ruleFormProxyDomainsArea.value[i])
        ruleForm.value.ProxyList[i].Locations = StringToArrayList(ruleFormProxyLocationsArea.value[i])
        ruleForm.value.ProxyList[i].UserAgentfilter = StringToArrayList(ruleFormProxyUserAgentfilterArea.value[i])
        ruleForm.value.ProxyList[i].TrustedCIDRsStrList = StringToArrayList(ruleFormProxyTrustedCIDRsStrListArea.value[i])
        ruleForm.value.ProxyList[i].RemoteIPHeaders = StringToArrayList(ruleFormProxyRemoteIPHeadersArea.value[i])
    }

    for (let i in ruleForm.value.ProxyList) {
        let indexNum = parseInt(i) + 1
        let proxy = ruleForm.value.ProxyList[i]
        console.log("proxy domains length: " + proxy.Domains.length)
        if (proxy.Domains.length <= 0) {
            MessageShow("error", "第 " + indexNum + " 条反向代理转发规则中域名不能为空")
            return
        }
        if (proxy.Locations.length <= 0) {
            MessageShow("error", "第 " + indexNum + " 条反向代理转发规则中后端地址不能为空")
            return
        }
    }




    switch (ruleFormOptionType.value) {
        case "add":
            apiAddReverseProxyRule(ruleForm.value).then((res) => {
                if (res.ret == 0) {
                    addRuleDialogVisible.value = false;
                    MessageShow("success", "反向代理规则添加成功")

                    queryRuleList()
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("反向代理规则添加失败,网络请求出错:" + error)
                MessageShow("error", "反向代理规则添加失败,网络请求出错")
            })

            break;
        case "alter":
            apiAlterReverseProxyRule(ruleForm.value).then((res) => {
                if (res.ret == 0) {
                    addRuleDialogVisible.value = false;
                    MessageShow("success", "反向代理规则修改成功")
                    queryRuleList()
                    return
                }
                MessageShow("error", res.msg)
            }).catch((error) => {
                console.log("反向代理规则修改失败,网络请求出错:" + error)
                MessageShow("error", "反向代理规则修改失败,网络请求出错")
            })

            break;
        default:
            console.log("错误的ruleFormOptionType " + ruleFormOptionType.value)
    }
}

const queryRuleList = () => {
    apiGeReverseProxyRuleList().then((res) => {
        //console.log(res.data)
        if (res.ret == 0) {
            if (res.list == undefined || res.list == null) {
                ruleList.value = []
            } else {
                ruleList.value = res.list
            }
            return
        }



    }).catch((error) => {
        console.log("获取反向代理规则列表出错:" + error)
        MessageShow("error", "获取反向代理规则列表出错")
    })
}

const reverproxyLogsArrayToTooltipHtml = (rule, proxyKey) => {
    var res = ""
    if (rule.AccessLogs == undefined || rule.AccessLogs == null) {
        res = "暂无日志"
        return res
    }


    //console.log(JSON.stringify(rule.AccessLogs))

    console.log()

    for (var key of Object.keys(rule.AccessLogs)) {
        console.log("key:" + key + "   proxyKey:" + proxyKey)
        if (key != proxyKey) {
            continue
        }

        if (rule.AccessLogs[key] == undefined || rule.AccessLogs[key] == null || rule.AccessLogs[key].lengh == 0) {
            break
        }



        for (let i in rule.AccessLogs[key]) {

            let log = rule.AccessLogs[key][i]
            if(i!="0"){
                res += '<br />'
            }
            res += log.LogTime + "&nbsp;&nbsp;&nbsp;" + log.LogContent + '<br />'
        }

    }





    // rule.AccessLogs.forEach(function (proxyLogs) {

    //     console.log("fuck  "+JSON.stringify(proxyLogs))
    //    // res += log.LogTime + "&nbsp;&nbsp;&nbsp;" + log.LogContent + '<br />'
    // });

    if (res == "") {
        res = "暂无日志"
    }

    return res
}

const checkAllListenType = ref(false)
const listenTypeIsIndeterminate = ref(true)

const listenTypes = ['tcp4', 'tcp6']

const handleCheckAllChange = (val: CheckboxValueType) => {
    ruleForm.value.Network = val ? "tcp" : ''
    ruleFormListenType.value = val ? ['tcp4', 'tcp6'] : []
    listenTypeIsIndeterminate.value = false

}
const handleCheckedProxyTypesChange = (value: CheckboxValueType[]) => {
    const checkedCount = value.length
    checkAllListenType.value = checkedCount === listenTypes.length
    listenTypeIsIndeterminate.value = checkedCount > 0 && checkedCount < listenTypes.length


    ruleForm.value.Network = getListenTypeByList(value)
}

const getListenTypeByList = (list: CheckboxValueType[]) => {
    let listLength = list.length
    switch (listLength) {
        case 0:
            return ""
        case 1:
            return list[0] + ''
        case 2:
            return "tcp"
        default:
            return ""
    }
}

var timerID: any
onMounted(() => {
    queryRuleList();

    timerID = setInterval(() => {
        queryRuleList();
    }, 1500);
})

onUnmounted(() => {
    clearInterval(timerID)
})


</script>

<style scoped>
.ReverseProxyPageRadius {
    height: 90vh;
    width: 100%;
    max-width: 1600px;
    border: 1px solid var(--el-border-color);
    border-radius: 0;
    margin: 20px
}

.affix-container {
    text-align: center;
    border-radius: 4px;
    width: 3vw;
    background: var(--el-color-primary-light-9);
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
    border: 4px solid var(--el-border-color);
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

.fromitemChildSafeDivRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin-left: 3px;
    margin-top: 15px;
    margin-right: 3px;
    margin-bottom: 15px;
    width: 465px;
    padding-top: 9px;
    padding-left: 9px;
    padding-right: 9px;
}

.formradius {
    border: 0px solid var(--el-border-color);
    border-radius: 0;
    margin: 0 auto;
    width: fit-content;
    padding: 15px;
}

.itemradius {

    border: 1px solid var(--el-border-color);
    border-radius: 0;
    margin-left: 3px;
    margin-top: 3px;
    margin-right: 3px;
    margin-bottom: 10px;
    min-width: 1350px;
}

.reverseProxyLogs {
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
</style>