import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'


//   ElMessageBox.alert(message, {
//     confirmButtonText: '好的',
//     callback: () => {
//     },
// })


export function ShowMessageBox(message: string) {
    ElMessageBox.alert(message, {
        confirmButtonText: '好的',
        callback: () => {
        },
    })
}

export function MessageShow(type:any,message: string) {
    ElMessage({
        message: message,
        type: type,
    })
}


export function Notification (type, message, duration)  {
    ElNotification({
        title: type.substring(0, 1).toUpperCase() + type.substring(1),
        message: message,
        type: type,
        duration: duration,
    })
}