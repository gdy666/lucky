import useClipboard from 'vue-clipboard3'
import { ref } from 'vue'

export function CopyTotoClipboard(data: string) {
    useClipboard().toClipboard(data)
}

const ipReg = /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){6}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^::([\da-fA-F]{1,4}:){0,4}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:):([\da-fA-F]{1,4}:){0,3}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){2}:([\da-fA-F]{1,4}:){0,2}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){3}:([\da-fA-F]{1,4}:){0,1}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){4}:((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$|^:((:[\da-fA-F]{1,4}){1,6}|:)$|^[\da-fA-F]{1,4}:((:[\da-fA-F]{1,4}){1,5}|:)$|^([\da-fA-F]{1,4}:){2}((:[\da-fA-F]{1,4}){1,4}|:)$|^([\da-fA-F]{1,4}:){3}((:[\da-fA-F]{1,4}){1,3}|:)$|^([\da-fA-F]{1,4}:){4}((:[\da-fA-F]{1,4}){1,2}|:)$|^([\da-fA-F]{1,4}:){5}:([\da-fA-F]{1,4})?$|^([\da-fA-F]{1,4}:){6}:$/

export function isIP(ip :string){
    return ipReg.test(ip)
}

const MenuIndexList = ["#status",
"#log","#whitelistset",
"#whitelists","#blacklists","#set",
"#login","#ddns","#ddnstasklist","#ddnsset",
"#about","#reverseproxylist","#ssl","#portforward","#portforwardset","#wol"]

export function PageExist(page:string) {
    for(let i in MenuIndexList){
        if (MenuIndexList[i]==page){
            return true
        }
    }
    return false
}

export const CurrentPage = ref("")


export function StringToArrayList(str : string){
    let rawlist = str.split("\n")
    let resList = new Array()
    for (let i in rawlist) {
        let item = rawlist[i].replace(/^\s+|\s+$/g, '').replace(/<\/?.+?>/g, "").replace(/[\r\n]/g, "")
        if (item==""){
            continue
        }
        resList.push(item)
    }
    return resList
}

export function StrArrayListToBrHtml( strList : string[]){
    var resHtml = ""
    for ( let i in strList){
        resHtml += strList[i]  + '<br />'
    }
    return resHtml
}

export function StrArrayListToArea(strList : string[]){
    var res = ""
    for ( let i in strList){
        if(i!="0"){
            res +='\n'
        }
        res += strList[i]
       // res += strList[i]  + '\n'
    }
    return res
}

export const LogLevelList = [
    {
        value: 2,
        label: 'Error',
    },
    {
        value: 3,
        label: 'Warn',
    },
    {
        value: 4,
        label: 'Info',
    },
    {
        value: 5,
        label: 'Debug',
    },
    {
        value: 6,
        label: 'Trace',
    },
]

export const bytesToSize = (bytes) => {
    if (bytes === 0) return '0 B';
    var k = 1000, // or 1024
        sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
        i = Math.floor(Math.log(bytes) / Math.log(k));

    return (bytes / Math.pow(k, i)).toPrecision(3) + ' ' + sizes[i];
}