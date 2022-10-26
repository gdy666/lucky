<template>
    <div class="PageRadius" :style="{
        borderRadius: 'base',
    }">

    
    <div class="InfoDivRadius">
        <div class="line">
            {{Info.AppName}}&nbsp;&nbsp;&nbsp;version:{{Info.Version}}&nbsp; {{Info.OS}}({{Info.ARCH}})
        </div>
        <div class="line">
          
        </div>
        <div class="line">
            作者:古大羊 &nbsp;{{Info.GoVersion}}&nbsp;编译时间:{{Info.Date}}
        </div>

        <div class="line">
            
<el-link type="primary" href="tencent://message/?uin=272288813&Site=&Menu=yes" target="_blank">QQ联系作者</el-link>
&nbsp;&nbsp;&nbsp;邮箱: 272288813@qq.com

        </div>

        <div class="line">
             Lukcy交流 QQ群:&nbsp;&nbsp;602427029
        </div>

        <div class="line">
             Github&nbsp;&nbsp;<el-link type="primary" href="https://github.com/gdy666/lucky" target="_blank">https://github.com/gdy666/lucky</el-link>
        </div>

        <div class="line">
             Gitee&nbsp;&nbsp;<el-link type="primary" href="https://gitee.com/gdy666/lucky" target="_blank">https://gitee.com/gdy666/lucky</el-link>
        </div>

        <div>
            本项目借鉴引用或参考的第三方开源项目: <el-link type="primary" href="https://github.com/fatedier/frp" target="_blank">frp</el-link> <el-link type="primary" href="https://github.com/jeessy2/ddns-go" target="_blank">ddns-go</el-link>
        </div>

        <div class="line">
             
        </div>
    </div>

    </div>
</template>


<script lang="ts" setup>

import { onMounted, onUnmounted, ref, computed, reactive } from 'vue'
import { apiGetAPPInfo } from '../apis/utils'
import {MessageShow} from '../utils/ui'

var Info = ref({
    AppName:"Lucky",
    Version:"1.0.0",
    OS:"unknow",
    ARCH:"unknow",
    Date:"2022-07-25",
    GoVersion:""
})


const queryAPPInfo = ()=>{
     apiGetAPPInfo().then((res) => {
        if (res.ret==0){
            Info.value = res.info
            return 
        }
        MessageShow("error", "获取App信息出错")
    }).catch((error) => {
        console.log("获获取App信息出错:" + error)
        MessageShow("error", "获取App信息出错")
    })
}

onMounted(() => {
    queryAPPInfo()

})

</script>


<style scoped>


.InfoDivRadius {
    border: 2px solid var(--el-border-color);
    border-radius: 10px;
    margin:auto;
    margin-top:50px;
    width: 495px;
    height: fit-content;
    padding:auto;
    padding-top: 20px;
    padding-bottom: 30px;
   
}

.line {
    margin-bottom: 5px;
}

</style>