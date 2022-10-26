<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

        <el-affix position="top" :offset="0" class="affix-container">
            <el-button type="primary" @click="showAddBlackListDialog">黑名单添加 <el-icon>
                    <Plus />
                </el-icon>
            </el-button>
        </el-affix>

        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }">



                <el-table :data="Blacklist" style="width: 700px" height="85vh">
                    <el-table-column prop="IP" label="IP" width="200" />
                    <el-table-column prop="Effectivetime" label="有效时间" width="200" />
                    <el-table-column fixed="right" label="操作" width="300">
                        <template #default="list">
                            <el-button link type="primary" size="small"
                                @click="flushBlackListEffectivetime(list.$index, Blacklist[list.$index], 0, '确认要刷新IP[' + Blacklist[list.$index].IP + ']的有效时间?')">
                                刷新有效时间</el-button>
                            <el-button link type="primary" size="small"
                                @click="flushBlackListEffectivetime(list.$index, Blacklist[list.$index], 666666, '确认要设置IP[' + Blacklist[list.$index].IP + ']为长期有效?')">
                                设置永久有效</el-button>
                            <el-button link type="primary" size="small"
                                @click="deleteBlackList(list.$index, Blacklist[list.$index])">删除</el-button>
                        </template>
                    </el-table-column>

                </el-table>





            </div>


        </el-scrollbar>


        <el-dialog v-model="addBlackListDialogVisible" title="添加黑名单IP" draggable :show-close="false" :close-on-click-modal="false" width="400px">

            <el-form :model="addBlackListForm">
                <el-form-item label="IP" label-width="auto">
                    <el-input v-model="addBlackListForm.IP" autocomplete="off" />
                </el-form-item>
                <el-form-item label="有效时间(小时)" label-width="auto">
                    <el-input-number v-model="addBlackListForm.Life" :min="1" :max="999999" />
                </el-form-item>

            </el-form>

            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="addBlackListDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="addBlackList">添加</el-button>
                </span>
            </template>
        </el-dialog>


    </div>

</template>


<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { MessageShow } from '../utils/ui'
import { isIP } from '../utils/utils'

import { apiGetBlackList, apiFlushBlackList, apiDeleteBlackList } from '../apis/utils'
var Blacklist = ref([{ IP: "", Effectivetime: "" }])
Blacklist.value.splice(0, 1)

const addBlackListDialogVisible = ref(false)
const addBlackListForm = ref({ IP: "", Life: 0 })

const showAddBlackListDialog = () => {
    addBlackListDialogVisible.value = true
    addBlackListForm.value.IP = ""
    addBlackListForm.value.Life = 666666
}

const flushBlackListEffectivetime = (index, item, life, text) => {
    ElMessageBox.confirm(
        text,
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            flushBlackListlife(index, item.IP, life)
        })
        .catch(() => {

        })
}

const flushBlackListlife = (index, ip, life) => {
    apiFlushBlackList(ip, life).then((res) => {
        if (res.ret == 0) {
            Blacklist.value[index].Effectivetime = res.data
            return
        }
        MessageShow("error", res.msg)

    }).catch((error) => {
        MessageShow("error", "刷新IP[" + ip + "]有效时间出错")
    })
}



const addBlackList = () => {

    // if (!isIP(addBlackListForm.value.IP)) {
    //     MessageShow("error", "IP格式有误,请检查修正后再添加")
    //     return
    // }

    apiFlushBlackList(addBlackListForm.value.IP, addBlackListForm.value.Life).then((res) => {
        if (res.ret == 0) {
            let item = { IP: addBlackListForm.value.IP, Effectivetime: res.data }
            Blacklist.value.push(item)
            addBlackListDialogVisible.value = false
            // MessageShow("success", "黑名单添加成功")
            return
        }
        MessageShow("error", res.msg)

    }).catch((error) => {
        MessageShow("error", "刷新IP[" + addBlackListForm.value.IP + "]有效时间出错")
    })
}

const deleteBlackList = (index, item) => {
    ElMessageBox.confirm(
        '确认要删除IP [' + item.IP + "]的黑名单记录?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            apiDeleteBlackList(item.IP).then((res) => {
                if (res.ret == 0) {
                    Blacklist.value.splice(index, 1)
                    return
                }
                MessageShow("error", res.msg)

            }).catch((error) => {
                MessageShow("error", "删除[" + item.IP + "]的黑名单记录出错")
            })
        })
        .catch(() => {

        })
}


const queryBlackList = () => {
    apiGetBlackList().then((res) => {
        if (res.ret == 0) {
            if (res.data != null) {
                Blacklist.value = res.data
            } else {
                Blacklist.value = []
            }
            return
        }
        MessageShow("error", res.msg)
        //console.log("getAdminURL "+getAdminURL())
    }).catch((error) => {
        MessageShow("error", "查询黑名单列表出错")
    })
}

const keydown = (e) => {
    if (e.keyCode != 13) {
        return
    }
    if (!addBlackListDialogVisible.value) {
        return
    }
    addBlackList()
}



onMounted(() => {
    queryBlackList();
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
</style>