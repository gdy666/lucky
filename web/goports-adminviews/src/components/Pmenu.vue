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

            <el-sub-menu index="#relay">
                <template #title>
                    <el-icon>
                        <Position />
                    </el-icon>
                    <span>端口转发</span>
                </template>

                <el-menu-item index="#relayset">
                    <el-icon>
                        <List />
                    </el-icon>
                    <template #title>转发规则</template>
                </el-menu-item>


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






            <el-menu-item index="#set">
                <el-icon>
                    <setting />
                </el-icon>
                <template #title>设置</template>
            </el-menu-item>



            <el-menu-item index="#logout">
                <el-icon>
                    <Close />
                </el-icon>
                <template #title>退出登录</template>
            </el-menu-item>
        </el-sub-menu>




        <div class="flex-grow" />

        <el-menu-item index="#logo">goports {{ version }}</el-menu-item>

    </el-menu>
</template>





<script setup lang="ts">
import { inject, ref, onMounted } from 'vue';
import {  SetHash, apiGetVersion } from '../apis/utils.js'
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
            window.open("https://github.com/ljymc/goports", "_blank");
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