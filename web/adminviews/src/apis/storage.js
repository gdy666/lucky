
/**
 * 使用方式
 * * * main.js挂载全局
 *  import storage from './utils/storage';
    const app = createApp(App);
    app.config.globalProperties.$storage = storage;
 * * * vue3.0 全局使用 vue实例上的数据
 *  import { defineComponent, reactive, getCurrentInstance } from "vue";
 *  let { appContext } = getCurrentInstance();
    appContext.config.globalProperties.$storage.setItem("username","aaa");
    appContext.config.globalProperties.$storage.setItem("age",20);
    appContext.config.globalProperties.$storage.clearItem("age");
    appContext.config.globalProperties.$storage.clearAll();
 */
export default {
    getStorage () { // 先获取该项目的 命名存储空间 下的storage数据 maneger
        return JSON.parse(window.localStorage.getItem("lucky") || "{}");
    },
    setItem (key, val) {
        let storage = this.getStorage()
        // console.log("setItem", storage);
        storage[key] = val; // 为当前对象添加 需要存储的值
        window.localStorage.setItem("lucky", JSON.stringify(storage)) // 保存到本地
    },
    getItem (key) {
        return this.getStorage()[key]
    },
    // 清空 当前的项目下命名存储的空间 该key项的 Storage 数据
    clearItem (key) {
        let storage = this.getStorage()
        delete storage[key]
        window.localStorage.setItem(config.namespace, JSON.stringify(storage)) // 保存到本地
    },
    // 清空所有的 当前的项目下命名存储的空间 Storage 数据
    clearAll () {
        window.localStorage.clear();
    }
}