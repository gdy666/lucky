import { ElMessage, ElMessageBox } from 'element-plus'


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


