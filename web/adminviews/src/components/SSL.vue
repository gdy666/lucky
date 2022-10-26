<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <!-- <el-affix position="bottom" :offset="0" class="affix-container">
            <el-button type="primary" @click="showAddSSLDialog">SSL证书添加 <el-icon>
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix> -->



        <el-scrollbar height="100%">



            <div class="itemradius" :style="{
                borderRadius: 'base',
            }" v-for="ssl in SSLList">

                <el-descriptions :column="6" border>
                    <el-descriptions-item label="证书备注" :span="2">
                        <el-button size="small" v-show="true">
                            {{ ssl.Remark == '' ? '未备注' : ssl.Remark }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="添加时间" :span="2">
                        <el-button size="small" v-show="true">
                            {{ ssl.AddTime }}
                        </el-button>
                    </el-descriptions-item>

                    <el-descriptions-item label="操作" :span="2">

                        <el-tooltip :content="ssl.Enable == true ? '证书已启用' : '证书已禁用'" placement="top">
                            <el-switch v-model="ssl.Enable" inline-prompt active-text="开" inactive-text="关"
                                :before-change="sslEnableClick.bind(this, ssl.Enable, ssl)" size="large" />
                        </el-tooltip>

                        &nbsp;&nbsp;
                        <el-button size="small" type="primary" @click="showAlterRemarkDialog(ssl)">修改备注</el-button>



                        <el-button size="small" type="danger" @click="deleteSSL(ssl)">删除</el-button>
                    </el-descriptions-item>

                    <div v-for="cert in ssl.CertsInfo">
                        <!-- <el-descriptions :column="6" border> -->
                        <el-descriptions-item label="绑定域名" :span="2">
                            <el-tooltip placement="bottom" effect="dark" :trigger-keys="[]" content="">
                                <template #content>
                                    <span v-html="StrArrayListToBrHtml(cert.Domains)"></span>
                                </template>
                                <el-button size="small" v-show="true" type="primary">
                                    {{cert.Domains.length==1?cert.Domains[0]:cert.Domains[0]+'...'}}
                                </el-button>
                            </el-tooltip>
                        </el-descriptions-item>

                        <el-descriptions-item label="颁发时间" :span="2">
                            <el-button size="small" v-show="true" type="info">
                                {{ cert.NotBeforeTime }}
                            </el-button>
                        </el-descriptions-item>

                        <el-descriptions-item label="到期时间" :span="2">
                            <el-button size="small" v-show="true" type="warning">
                                {{ cert.NotAfterTime }}
                            </el-button>
                        </el-descriptions-item>


                        <!-- </el-descriptions> -->
                    </div>

                </el-descriptions>



            </div>







        </el-scrollbar>

        <el-affix position="bottom" :offset="30" class="affix-container">
            <el-button type="primary" :round=true @click="showAddSSLDialog">SSL证书添加
                <el-icon class="el-icon--right">
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix>



        <el-dialog v-if="addSSLDialogVisible" v-model="addSSLDialogVisible" title="添加SSL证书" draggable
            :show-close="false" :close-on-click-modal="false" width="400px">

            <el-form :model="addSSLForm">

                <el-form-item label="备注" label-width="auto">
                    <el-input v-model="addSSLForm.Remark" autocomplete="off" />
                </el-form-item>

                <el-form-item label="证书" label-width="auto">
                    <el-upload class="inline-block" :multiple="true" :action="getFileBase64API()"
                        :before-upload="beforeUpload" :show-file-list="false" :headers="{ 'Authorization': GetToken() }"
                        :on-success="callbackGetCreFileBase64">
                        <el-button round class='margin-change'>{{uploadCreButtontext}}</el-button>
                    </el-upload>

                </el-form-item>

                <el-form-item label="Key" label-width="auto">
                    <el-upload class="inline-block" :multiple="true" :action="getFileBase64API()"
                        :before-upload="beforeUpload" :show-file-list="false" :headers="{ 'Authorization': GetToken() }"
                        :on-success="callbackGetKeyFileBase64">
                        <el-button round class='margin-change'>{{uploadKeyButtontext}}</el-button>
                    </el-upload>
                </el-form-item>

            </el-form>

            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="addSSLDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="addSSL">添加</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog v-if="alterRemarkDialogShow" v-model="alterRemarkDialogShow" :title=alterRemarkDialogSSLText
            draggable :show-close="false" :close-on-click-modal="false" width="400px">
            <el-form-item label="备注" label-width="auto">
                <el-input v-model="alterRemarkDialogValue" autocomplete="off" />
            </el-form-item>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="alterRemarkDialogShow = false">取消</el-button>
                    <el-button type="primary" @click="alterSSLRemark">修改</el-button>
                </span>
            </template>
        </el-dialog>


    </div>

</template>


<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { MessageShow } from '../utils/ui'
import { StrArrayListToBrHtml } from '../utils/utils'
import { GetToken, apiAddSSL, apiGetSSLList, apiDeleteSSL, apiAlterSSL } from '../apis/utils'
import type { UploadProps } from 'element-plus'
var SSLList = ref([
    {
        Key: "",
        Remark: "",
        Enable: true,
        AddTime: "",
        CertsInfo: [{
            Domains: [''],
            NotBeforeTime: '',
            NotAfterTime: ''
        },]
    }
])


const addSSLDialogVisible = ref(false)
const addSSLForm = ref({ Remark: "", CertBase64: "", KeyBase64: "" })
const uploadCreButtontext = ref("")
const uploadKeyButtontext = ref("")

const getFileBase64API = () => {
    var baseURL = "/" //
    if (process.env.NODE_ENV == "development") {
        //开发环境下这个改为自己的接口地址
        baseURL = 'http://192.168.31.70:16601/'
    }
    return baseURL + "api/getfilebase64"
}


const alterRemarkDialogShow = ref(false)
const alterRemarkDialogValue = ref("")
const alterRemarkDialogSSLText = ref("")
const alterRemarkDialogSSLKey = ref("")

const showAlterRemarkDialog = (ssl) => {

    alterRemarkDialogShow.value = true
    alterRemarkDialogSSLKey.value = ssl.Key
    alterRemarkDialogValue.value = ssl.Remark
    alterRemarkDialogSSLText.value = ssl.CertsInfo[0].Domains[0];
}

const callbackGetCreFileBase64 = (res: any, uploadFile: any, uploadFiles: any) => {
    if (res.ret != 0) {
        MessageShow("error", res.msg)
        return
    }
    console.log("file:" + res.file)
    uploadCreButtontext.value = res.file
    //console.log("base64:"+res.base64)
    addSSLForm.value.CertBase64 = res.base64
}

const beforeUpload: UploadProps['beforeUpload'] = (rawFile) => {
    if (rawFile.size / 1024 / 1024 > 1) {
        MessageShow("error",'文件不能大于1M')
        return false
    }
    return true
}

const callbackGetKeyFileBase64 = (res: any, uploadFile: any, uploadFiles: any) => {
    if (res.ret != 0) {
        MessageShow("error", res.msg)
        return
    }
    console.log("file:" + res.file)
    uploadKeyButtontext.value = res.file
    //console.log("base64:"+res.base64)
    addSSLForm.value.KeyBase64 = res.base64
}


const showAddSSLDialog = () => {
    addSSLDialogVisible.value = true
    addSSLForm.value.CertBase64 = ""
    addSSLForm.value.KeyBase64 = ""
    uploadCreButtontext.value = "选择要上传的证书文件"
    uploadKeyButtontext.value = "选择要上传的Key文件"
}



const sslEnableClick = (enable, ssl) => {
    const enableText = enable == false ? "启用" : "禁用";

    const sslText = ssl.Remark != "" ? ssl.Remark : ssl.CertsInfo[0].Domains[0];

    const sslName = "[" + sslText + "]"

    return new Promise((resolve, reject) => {

        ElMessageBox.confirm(
            '确认要' + enableText + " 证书 " + sslName + "?",
            'Warning',
            {
                confirmButtonText: '确认',
                cancelButtonText: '取消',
                type: 'warning',
            }
        )
            .then(() => {
                apiAlterSSL(ssl.Key, "enable", !enable).then(res => {
                    if (res.ret == 0) {
                        resolve(true)
                        MessageShow("success", "证书  " + sslName + enableText + "成功")
                        return
                    }
                    resolve(false)
                    MessageShow("error", "证书 " + sslName + enableText + "失败: " + res.msg)

                    // if (res.syncres != undefined && res.syncres != "") {
                    //     Notification("warn", res.syncres, 0)
                    // }
                }).catch((error) => {
                    resolve(false)
                    console.log("证书 " + sslName + enableText + "失败" + ":请求出错" + error)
                    MessageShow("error", "证书 " + sslName + enableText + "失败" + ":请求出错")
                })

            })
            .catch(() => {
                resolve(false)
            })



    })
}

const alterSSLRemark = () => {


    apiAlterSSL(alterRemarkDialogSSLKey.value, "remark", alterRemarkDialogValue.value).then(res => {
        if (res.ret == 0) {
            alterRemarkDialogShow.value = false
            MessageShow("success", "证书  " + alterRemarkDialogSSLText.value + " 备注修改成功")
            querySSLList()

            return
        }
        MessageShow("error", "证书 " + alterRemarkDialogSSLText.value + " 备注修改失败: " + res.msg)
    }).catch((error) => {
        console.log("证书 " + alterRemarkDialogSSLText.value + " 备注修改失败" + ":请求出错" + error)
        MessageShow("error", "证书 " + alterRemarkDialogSSLText.value + " 备注修改失败" + ":请求出错")
    })


}





const addSSL = () => {

    if (addSSLForm.value.CertBase64 == "") {
        MessageShow("error", "请选择要保存的证书文件")
        return
    }

    if (addSSLForm.value.KeyBase64 == "") {
        MessageShow("error", "请选择要保存的Key文件")
        return
    }
    apiAddSSL(addSSLForm.value).then((res) => {
        if (res.ret == 0) {
            //let item = { IP: addWhiteListForm.value.IP, Effectivetime: res.data }
            //whitelist.value.push(item)
            addSSLDialogVisible.value = false
            MessageShow("success", "添加证书成功")
            querySSLList()
            return
        }
        MessageShow("error", res.msg)

    }).catch((error) => {
        console.log("添加SSL证书出错 " + error)
        MessageShow("error", "添加SSL证书出错 " + error)

    })
}

const deleteSSL = (ssl) => {
    const sslText = ssl.Remark != "" ? ssl.Remark : ssl.CertsInfo[0].Domains[0];
    ElMessageBox.confirm(
        '确认要删除 ' + sslText + "  的证书?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {

        apiDeleteSSL(ssl.Key).then((res) => {
            if (res.ret == 0) {
                MessageShow("success", "证书删除成功")
                querySSLList()
                return
            }
            MessageShow("error", res.msg)
        }).catch((error) => {
            console.log("证书删除失败,网络请求出错:" + error)
            MessageShow("error", "证书删除失败,网络请求出错")
        })

    })
}


const querySSLList = () => {
    apiGetSSLList().then((res) => {
        if (res.ret == 0) {
            console.log(res.list)
            if (res.list != null) {

                SSLList.value = res.list
            } else {
                // whitelist.value= []
                SSLList.value = []
            }

            return
        }
        MessageShow("error", res.msg)
        //console.log("getAdminURL "+getAdminURL())
    }).catch((error) => {
        MessageShow("error", "查询证书列表列表出错")
    })
}

const keydown = (e) => {
    if (e.keyCode != 13) {
        return
    }
    if (!addSSLDialogVisible.value) {
        return
    }
    addSSL()
}

onMounted(() => {
    querySSLList();
    window.addEventListener('keydown', keydown)

})

</script>

<style scoped>
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
    margin-bottom: 25px;
    min-width: 1350px;
}

.affix-container {
    text-align: center;
    border-radius: 4px;
    width: 3vw;
    background: var(--el-color-primary-light-9);
}
</style>