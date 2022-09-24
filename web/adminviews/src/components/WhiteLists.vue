<template>

    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

                         <el-affix position="bottom" :offset="0" class="affix-container">
                    <el-button type="primary" @click="showAddWhiteListDialog">白名单添加 <el-icon>
                            <Plus />
                        </el-icon>
                    </el-button>
        </el-affix>

        <el-scrollbar height="100%">


            <div class="formradius" :style="{
                borderRadius: 'base',
            }" >




                <el-table :data="whitelist" style="width: 700px"   height="85vh">
                    <el-table-column prop="IP" label="IP" width="200" />
                    <el-table-column prop="Effectivetime" label="有效时间" width="200" />
                    <el-table-column fixed="right" label="操作" width="300">
                        <template #default="list">
                            <el-button link type="primary" size="small"
                                @click="flushWhiteListEffectivetime(list.$index, whitelist[list.$index], 0, '确认要刷新IP[' + whitelist[list.$index].IP + ']的有效时间?')">
                                刷新有效时间</el-button>
                            <el-button link type="primary" size="small"
                                @click="flushWhiteListEffectivetime(list.$index, whitelist[list.$index], 666666, '确认要设置IP[' + whitelist[list.$index].IP + ']为长期有效?')">
                                设置永久有效</el-button>
                            <el-button link type="primary" size="small"
                                @click="deleteWhiteList(list.$index, whitelist[list.$index])">删除</el-button>
                        </template>
                    </el-table-column>

                </el-table>


            </div>


        </el-scrollbar>



        <el-dialog v-model="addWhiteListDialogVisible" title="添加白名单IP" draggable :show-close="false" :close-on-click-modal="false" width="400px">

            <el-form :model="addWhiteListForm">
                <el-form-item label="IP" label-width="auto">
                    <el-input v-model="addWhiteListForm.IP" autocomplete="off" />
                </el-form-item>
                <el-form-item label="有效时间(小时)" label-width="auto">
                    <el-input-number v-model="addWhiteListForm.Life" :min="1" :max="999999" />
                </el-form-item>

            </el-form>

            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="addWhiteListDialogVisible = false">取消</el-button>
                    <el-button type="primary" @click="addWhiteList">添加</el-button>
                </span>
            </template>
        </el-dialog>


    </div>

</template>


<script lang="ts" setup>
import { ref, onMounted, computed } from 'vue'
import {  ElMessageBox } from 'element-plus'
import {MessageShow} from '../utils/ui'
import {isIP} from '../utils/utils'
import { apiGetWhiteList, apiFlushWhiteList, apiDeleteWhiteList, apiGetBlackList, apiFlushBlackList, apiDeleteBlackList } from '../apis/utils'
var whitelist = ref([{ IP: "", Effectivetime: "" }])
whitelist.value.splice(0, 1)

const addWhiteListDialogVisible = ref(false)
const addWhiteListForm = ref({ IP: "", Life: 0 })

const showAddWhiteListDialog = () => {
    addWhiteListDialogVisible.value = true
    addWhiteListForm.value.IP = ""
    addWhiteListForm.value.Life = 24
}

const flushWhiteListEffectivetime = (index, item, life, text) => {
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
            flushWhiteListlife(index, item.IP, life)
        })
        .catch(() => {

        })
}

const flushWhiteListlife = (index, ip, life) => {
    apiFlushWhiteList(ip, life).then((res) => {
        if (res.ret == 0) {
            whitelist.value[index].Effectivetime = res.data
            return
        }
        MessageShow("error", res.msg)

    }).catch((error) => {
        console.log( "刷新IP[" + addWhiteListForm.value.IP + "]有效时间出错 "+error)
        MessageShow("error", "刷新IP[" + ip + "]有效时间出错")
    })
}



const addWhiteList = () => {

    // if (!isIP(addWhiteListForm.value.IP)) {
    //     MessageShow("error", "IP格式有误,请检查修正后再添加")
    //     return
    // }

    apiFlushWhiteList(addWhiteListForm.value.IP, addWhiteListForm.value.Life).then((res) => {
        if (res.ret == 0) {
            let item = { IP: addWhiteListForm.value.IP, Effectivetime: res.data }
            whitelist.value.push(item)
            addWhiteListDialogVisible.value = false
            return
        }
        MessageShow("error", res.msg)

    }).catch((error) => {
        console.log( "刷新IP[" + addWhiteListForm.value.IP + "]有效时间出错 "+error)
        MessageShow("error", "刷新IP[" + addWhiteListForm.value.IP + "]有效时间出错")
    })
}

const deleteWhiteList = (index, item) => {
    ElMessageBox.confirm(
        '确认要删除IP [' + item.IP + "]的白名单记录?",
        'Warning',
        {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning',
        }
    )
        .then(() => {
            apiDeleteWhiteList(item.IP).then((res) => {
                if (res.ret == 0) {
                    whitelist.value.splice(index, 1)
                    return
                }
                MessageShow("error", res.msg)

            }).catch((error) => {
                MessageShow("error", "删除[" + item.IP + "]的白名单记录出错")
            })
        })
        .catch(() => {

        })
}



const queryWhiteList = () => {
    apiGetWhiteList().then((res) => {
        if (res.ret == 0) {
            if (res.data!=null){
                whitelist.value = res.data
            }else{
                whitelist.value= []
            }
           
            return
        }
        MessageShow("error", res.msg)
        //console.log("getAdminURL "+getAdminURL())
    }).catch((error) => {
        MessageShow("error", "查询白名单列表出错")
    })
}

const keydown = (e)=>{
    if (e.keyCode != 13) {
    return 
  }
  if(!addWhiteListDialogVisible.value ){
    return 
  }
  addWhiteList()
}

onMounted(() => {
    queryWhiteList();
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