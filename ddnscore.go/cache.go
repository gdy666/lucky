package ddnscore

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gdy666/lucky/config"
)

var taskInfoMap sync.Map
var taskInfoMapMutex sync.RWMutex
var webLastAccessDDNSTaskListLastTime int64

// 记录最后的前端请求DDNS任务列表时间
func FLushWebLastAccessDDNSTaskListLastTime() {
	atomic.StoreInt64(&webLastAccessDDNSTaskListLastTime, time.Now().Unix())
}

// webAccessAvalid 判断前端访问是否处于活跃时间内
func webAccessAvalid() bool {
	lastTime := atomic.LoadInt64(&webLastAccessDDNSTaskListLastTime)
	return time.Now().Unix()-lastTime <= 5
}

func EnableDDNSTaskByKey(key string, enable bool) error {
	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()
	taskInfo, ok := taskInfoMap.Load(key)
	if !ok {
		return fmt.Errorf("DDNSTaskInfoMap key[%s] no found", key)
	}
	if enable {
		taskInfo.(*DDNSTaskState).SetDomainUpdateStatus(UpdateWaiting, "")
	} else {
		taskInfo.(*DDNSTaskState).SetDomainUpdateStatus(UpdateStop, "")
	}
	return config.EnableDDNSTaskByKey(key, enable)
}

func DDNSTaskInfoMapUpdate(task *DDNSTaskInfo) {
	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()

	preInfo, ok := taskInfoMap.Load(task.TaskKey)
	if ok {
		var checkDomains []Domain
		//防止有域名被删除
		for i, new := range task.TaskState.Domains {
			for _, pre := range preInfo.(*DDNSTaskState).Domains {
				if strings.Compare(new.String(), pre.String()) == 0 {
					checkDomains = append(checkDomains, task.TaskState.Domains[i])
					break
				}
			}
		}
		task.TaskState.Domains = checkDomains

		if len(preInfo.(*DDNSTaskState).Domains) > 0 && preInfo.(*DDNSTaskState).Domains[0].UpdateStatus == UpdateStop {
			task.TaskState.SetDomainUpdateStatus(UpdateStop, "")
		}

		taskInfoMap.Store(task.TaskKey, &task.TaskState)
		return
	}
}

// 即时更新IP相关数据信息
func DDNSTaskInfoMapUpdateIPInfo(task *DDNSTaskInfo) {
	if !webAccessAvalid() {
		//log.Printf("前端没有访问,不即时更新")
		return
	}
	//log.Printf("前端没有访问,不即时更新")

	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()
	state, ok := taskInfoMap.Load(task.TaskKey)
	if !ok {
		return
	}
	state.(*DDNSTaskState).IpAddr = task.TaskState.IpAddr
	state.(*DDNSTaskState).IPAddrHistory = task.TaskState.IPAddrHistory
}

func DDNSTaskInfoMapUpdateDomainInfo(task *DDNSTaskInfo) {
	if !webAccessAvalid() {
		//log.Printf("前端没有访问,不即时更新")
		return
	}
	//log.Printf("前端有访问,即时更新")

	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()
	state, ok := taskInfoMap.Load(task.TaskKey)
	if !ok {
		return
	}
	state.(*DDNSTaskState).Domains = task.TaskState.Domains
}

func DDNSTaskInfoMapDelete(key string) {
	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()
	taskInfoMap.Delete(key)
}

func UpdateDomainsStateByTaskKey(key, status, message string) {
	taskInfoMapMutex.Lock()
	defer taskInfoMapMutex.Unlock()
	preInfo, ok := taskInfoMap.Load(key)
	if !ok {
		return
	}
	preInfo.(*DDNSTaskState).SetDomainUpdateStatus(status, message)
}

func GetDDNSTaskInfoList() []*DDNSTaskInfo {
	taskInfoMapMutex.RLock()
	defer taskInfoMapMutex.RUnlock()
	ddnsTaskList := config.GetDDNSTaskConfigureList()
	var res []*DDNSTaskInfo
	for i := range ddnsTaskList {
		ti := CreateDDNSTaskInfo(ddnsTaskList[i])
		res = append(res, ti)
	}
	return res
}

func GetDDNSTaskInfoByKey(key string) *DDNSTaskInfo {
	taskInfoMapMutex.RLock()
	defer taskInfoMapMutex.RUnlock()
	ddnsConf := config.GetDDNSTaskByKey(key)
	if ddnsConf == nil {
		return nil
	}
	info := CreateDDNSTaskInfo(ddnsConf)
	return info
}

func CreateDDNSTaskInfo(task *config.DDNSTask) *DDNSTaskInfo {
	var res DDNSTaskInfo
	res.DDNSTask = *task
	info, ok := taskInfoMap.Load(task.TaskKey)
	if ok {
		res.TaskState = *info.(*DDNSTaskState)
	} else {
		res.TaskState.Init(res.Domains)
		if task.Enable {
			res.TaskState.SetDomainUpdateStatus(UpdateWaiting, "")
		} else {
			res.TaskState.SetDomainUpdateStatus(UpdateStop, "")
		}
		taskInfoMap.Store(task.TaskKey, &res.TaskState)
	}
	return &res
}
