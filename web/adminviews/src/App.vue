<template>

  <div class="common-layout">
    <el-container>
      <el-header class="header" id="header" >
        <!-- <p class="title">大吉</p>
        <p class="version">version:{{version}}</p> -->
        <Pmenu class="menu"></Pmenu>
      </el-header>
      <el-container>
        <!-- <el-aside width="initial" id="aside" direction="vertical" >
      
        <Pmenu v-if="global.currentPage.value=='#login'?false:true"></Pmenu>
          
        </el-aside> -->
        <el-container class="body">
          <el-main id="pageContent">
            <Log v-if="global.currentPage.value=='#log'?true:false"></Log>
            <Status v-if="global.currentPage.value=='#status'?true:false"></Status>
            <Pset v-if="global.currentPage.value=='#set'?true:false"></Pset>
            <Login v-if="global.currentPage.value=='#login'?true:false"></Login>
            <WhiteListSet v-if="global.currentPage.value=='#whitelistset'?true:false"></WhiteListSet>
            <WhiteLists v-if="global.currentPage.value=='#whitelists'?true:false"></WhiteLists>
            <BlackLists v-if="global.currentPage.value=='#blacklists'?true:false"></BlackLists>
            <DDNS v-if="global.currentPage.value=='#ddnstasklist'?true:false"></DDNS>
            <DDNSSet v-if="global.currentPage.value=='#ddnsset'?true:false"></DDNSSet>
            <ReverseProxy v-if="global.currentPage.value=='#reverseproxylist'?true:false"></ReverseProxy>
            <SSL v-if="global.currentPage.value=='#ssl'?true:false"></SSL>
            <About v-if="global.currentPage.value=='#about'?true:false"></About>
            <PortForward v-if="global.currentPage.value=='#portforward'"></PortForward>
            <PortForwardSet v-if="global.currentPage.value=='#portforwardset'"></PortForwardSet>
            <WOL v-if="global.currentPage.value=='#wol'"></WOL>
          </el-main>

        </el-container>
      </el-container>
    </el-container>
  </div>
</template>




<script setup lang="ts">
import { onMounted,ref,inject ,computed } from 'vue'
import Status from './components/status.vue'
import Log from './components/log.vue';
import Pmenu from './components/pmenu.vue';
import Pset from './components/pset.vue';
import Login from './components/login.vue';
import WhiteListSet from './components/WhiteListSet.vue';
import WhiteLists from './components/WhiteLists.vue';
import BlackLists from './components/BlackLists.vue';
import DDNS from './components/DDNS.vue';
import ReverseProxy from './components/reverseproxy.vue';


import {apiGetVersion} from "./apis/utils.js"
import DDNSSet from './components/DDNSSet.vue';
import SSL from './components/SSL.vue'

//console.log("111") 

const global:any = inject("global")

const version = ref("0.0.0")

const queryVersion = ()=>{
  apiGetVersion().then((res) => {
        if (res.ret == 0) {
            version.value = res.version
            return
        }

    }).catch((error) => {

    })
}


onMounted(() => {
  //console.log("222")
//queryVersion()

})






</script>




<style scoped>

#pageContent{
  margin:0;
  height: 95vh;
  overflow: hidden;
  padding-left: 1px;
  padding-right: 0px;
  width: 100%;
}

body{
  margin: 0;
  width:100%;
}





  #header {
    background-color: #0d8bbb;
    height: fit-content;
    width:100%;
    padding:0
  }

  .title {
    float:left;
    text-align: left;
    margin-left: 10%;
    font-size:25px;
  }

    .menu {
    float:left;

    height: 30px;
    width:100vw;
  }

  
  .version{
    float:right;
    
  }

  .title,.version{
    width:100%;
    color: aliceblue;
    font-family: "Helvetica Neue",Helvetica,"PingFang SC","Hiragino Sans GB","Microsoft YaHei","微软雅黑",Arial,sans-serif;
  }
</style>


