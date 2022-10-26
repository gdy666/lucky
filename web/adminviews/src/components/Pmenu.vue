  <template>
    <el-menu :default-active="activeIndex" class="el-menu-demo menu" mode="horizontal" :ellipsis="false"
        @select="handleSelect">

        <el-sub-menu index="#menu" v-if="global.currentPage.value != '#login' ? true : false">
            <template #title>
                <el-icon>
                    <Menu />
                </el-icon>
                <span>菜单</span>
            </template>


            <el-menu-item index="#status">
                <el-icon>
                    <DataAnalysis />
                </el-icon>
                <template #title>总览</template>
            </el-menu-item>
            <el-menu-item index="#log">
                <el-icon>
                    <document />
                </el-icon>
                <template #title>程序日志</template>
            </el-menu-item>

            <el-divider style="margin-top: 0px;margin-bottom: 0px;" />

            <el-sub-menu index="#relay">
                <template #title>
                    <el-icon>
                        <Position />
                    </el-icon>
                    <span>端口转发</span>
                </template>

                <!-- <el-menu-item index="#relayset">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>转发规则</template>
                </el-menu-item> -->

                <el-menu-item index="#portforward">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>转发规则列表</template>
                </el-menu-item>

                <el-menu-item index="#portforwardset">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>设置</template>
                </el-menu-item>

            </el-sub-menu>

            <el-sub-menu index="#reverseproxy">
                <template #title>
                    <el-icon>
                        <Connection />
                    </el-icon>
                    <span>反向代理</span>
                </template>

                <el-menu-item index="#reverseproxylist">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>反向代理规则列表</template>
                </el-menu-item>

            </el-sub-menu>

            <el-sub-menu index="#ddns">
                <template #title>
                    <el-icon>
                        <Promotion />
                    </el-icon>
                    <span>动态域名</span>
                </template>

                <el-menu-item index="#ddnstasklist">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>动态域名任务列表</template>
                </el-menu-item>

                <el-menu-item index="#ddnsset">
                    <el-icon>
                        <Setting />
                    </el-icon>
                    <template #title>动态域名设置</template>
                </el-menu-item>
            </el-sub-menu>

            <el-sub-menu index="#wol">
                <template #title>
                    <el-icon>
                        <Bell />
                    </el-icon>
                    <span>网络唤醒</span>
                </template>


                <el-menu-item index="#wol">
                    <el-icon>
                        <Bell />
                    </el-icon>
                    <template #title>网络唤醒设备列表</template>
                </el-menu-item>

                <el-menu-item index="#wolset">
                    <el-icon>
                        <setting />
                    </el-icon>
                    <template #title>网络唤醒设置</template>
                </el-menu-item>

                </el-sub-menu>




            <el-divider style="margin-top: 0px;margin-bottom: 0px;" />



            <el-sub-menu index="#safe">
                <template #title>
                    <el-icon>
                        <Guide />
                    </el-icon>
                    <span>安全相关</span>
                </template>

                <el-sub-menu index="#safe">
                <template #title>
                    <el-icon>
                        <Guide />
                    </el-icon>
                    <span>IP过滤设置</span>
                </template>

                <el-menu-item index="#whitelistset">
                    <el-icon>
                        <Setting />
                    </el-icon>
                    <template #title>白名单设置</template>
                </el-menu-item>
                <el-menu-item index="#whitelists">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>白名单列表</template>
                </el-menu-item>


                <el-menu-item index="#blacklists">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>黑名单列表</template>
                </el-menu-item>

            </el-sub-menu>

            <el-menu-item index="#ssl">
                <el-icon>
                    <Lock />
                </el-icon>
                <template #title>SSL证书</template>
            </el-menu-item>


            </el-sub-menu>
            





            

            <el-menu-item index="#set">
                <el-icon>
                    <setting />
                </el-icon>
                <template #title>设置</template>
            </el-menu-item>



            <el-menu-item index="#about">
                <el-icon>
                    <Pointer />
                </el-icon>
                <template #title>关于</template>
            </el-menu-item>


            <el-divider style="margin-top: 0px;margin-bottom: 0px;" />

            <el-menu-item index="#logout">
                <el-icon>
                    <Close />
                </el-icon>
                <template #title>退出登录</template>
            </el-menu-item>
        </el-sub-menu>




        <div class="flex-grow" />

        <el-menu-item index="#logo">Lucky {{ version }}</el-menu-item>

    </el-menu>
</template>





<script setup lang="ts">
import { inject, ref, onMounted } from 'vue';
import { SetHash, apiGetVersion } from '../apis/utils.js'
import { ElMessageBox } from 'element-plus'
const global: any = inject("global")


const activeIndex = ref('#set')
const version = ref("")

console.log("currentPage[menu]:" + global.currentPage.value)


const queryVersion = () => {
    apiGetVersion().then((res) => {
        if (res.ret == 0) {
            version.value = res.version
            return
        }

    }).catch((error) => {

    })
}


function handleOpen(key, keyPath) {
    //console.log(key, keyPath);
}
function handleClose(key, keyPath) {
    //console.log(key, keyPath);
}
function handleSelect(key, keyPath, item, routeResult) {
    //console.log("选择菜单")
    //console.log(key, keyPath, item, routeResult);
    // switchView(key);
    switch (key) {
        case "#logout":
            ElMessageBox.confirm(
                '确定要注销登录?',
                'Warning',
                {
                    confirmButtonText: '确认',
                    cancelButtonText: '点错了',
                    type: 'warning',
                }
            )
                .then(() => {
                    SetHash(key)
                })
                .catch(() => {

                })
            break;
        case "#logo":
            //window.open("https://github.com/gdy666/lucky", "_blank");
            location.hash = "#about"
            break;
        default:
            SetHash(key)
            break;
    }

}

onMounted(() => {

    queryVersion()

})

</script>

<style>
.menu {
    background-color: #d9ecff;
}



.flex-grow {
    flex-grow: 1;
}
</style>