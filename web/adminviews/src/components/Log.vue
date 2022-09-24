<template>
    <!-- <div class="logterm">
        {{weblogsContent}}
    </div> -->
     <el-scrollbar max-height="95vh" class="logtermv2" v-loading="logLoading" element-loading-background="transparent">
 {{weblogsContent}}
 <br>
 <br>
 <br>
 <br>
 <br>
 <br>
    </el-scrollbar>
</template>


<script setup lang="ts">

import {apiGetLogs } from '../apis/utils'
import { onMounted,onUnmounted,ref,inject   } from 'vue'
import {GetHash,SetHash} from'../apis/utils.js'

const global:any = inject("global")
var preLogTimestamp = ""
var preStartTime = ""
const weblogsContent = ref("")
var logLoading = ref(true)




function queryLastlogs() {
    // if(GetHash()!="#log"){
    //     return ;
    // }

    
    apiGetLogs(preLogTimestamp).then((res) => {
        logLoading.value =false

        if (preStartTime!=res.starttime){
            weblogsContent.value = ""
            preStartTime = res.starttime
        }

        if (res.logs!=null && res.logs.length > 0) {
            // if (res.logs[0].timestamp!=preLogTimestamp){
            //     console.log("fuckkkk")
            //     return 
            // }
            

            preLogTimestamp = res.logs[res.logs.length - 1].timestamp
            console.log("fff "+res.logs[res.logs.length - 1].log)
            console.log("追加日志 "+preLogTimestamp)
            for(var i=0;i<res.logs.length;i++){
                weblogsContent.value += res.logs[i].log +"\n"
            }
        }
       
    })

}

var timerID:any
 


onMounted(() => {
    queryLastlogs();
    timerID = setInterval(() => {
      queryLastlogs();
    }, 1000);
})

onUnmounted(()=>{
    clearInterval(timerID) 
})







</script>

<style>

.logtermv2 {
    background-color: black;
    height: 95vh;
    width: 100%;
    color: white;
    text-align: left;
    padding-left: 3px;

    border: 10px;
    overflow-y: auto;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap:break-word;
}



.logterm::-webkit-scrollbar {
    width: 5px;
}

.logterm::-webkit-scrollbar-thumb {
    background-color: #c0c0c0;
    border-radius: 10%;

}
</style>

