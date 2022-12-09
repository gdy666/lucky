package ddnsgo

import (
	"fmt"
	"log"
	"time"

	"github.com/gdy666/lucky/config"
	ddnsconf "github.com/gdy666/lucky/module/ddns/conf"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

func init() {
	ddnsconf.SetGetDDNSConfigureFunc(GetDDNSConfigure)
}

func DDNSTaskListConfigureCheck() {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].DNS.ForceInterval < 60 {
			config.Configure.DDNSTaskList[i].DNS.ForceInterval = 60
		} else if config.Configure.DDNSTaskList[i].DNS.ForceInterval > 360000 {
			config.Configure.DDNSTaskList[i].DNS.ForceInterval = 360000
		}

		if config.Configure.DDNSTaskList[i].HttpClientTimeout < 3 {
			config.Configure.DDNSTaskList[i].HttpClientTimeout = 3
		} else if config.Configure.DDNSTaskList[i].HttpClientTimeout > 60 {
			config.Configure.DDNSTaskList[i].HttpClientTimeout = 60
		}
	}
}

func DDNSTaskSetWebhookCallResult(taskKey string, result bool, message string) {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	taskIndex := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return
	}

	log.Printf("DDNSTaskSetWebhookCallResult %s", taskKey)

}

func GetDDNSTaskConfigureList() []*ddnsconf.DDNSTask {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	var resList []*ddnsconf.DDNSTask

	for i := range config.Configure.DDNSTaskList {
		task := config.Configure.DDNSTaskList[i]
		resList = append(resList, &task)
	}
	return resList
}

func GetDDNSTaskByKey(taskKey string) *ddnsconf.DDNSTask {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	taskIndex := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return nil
	}
	res := config.Configure.DDNSTaskList[taskIndex]

	return &res
}

func DDNSTaskListAdd(task *ddnsconf.DDNSTask) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	task.TaskKey = stringsp.GetRandomString(16)
	task.ModifyTime = time.Now().Unix()
	config.Configure.DDNSTaskList = append(config.Configure.DDNSTaskList, *task)
	return config.Save()
}

func DDNSTaskListDelete(taskKey string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	taskIndex := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("找不到需要删除的DDNS任务")
	}

	config.Configure.DDNSTaskList = DeleteDDNSTaskListlice(config.Configure.DDNSTaskList, taskIndex)
	return config.Save()
}

func EnableDDNSTaskByKey(taskKey string, enable bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	taskIndex := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}
	if taskIndex == -1 {
		return fmt.Errorf("开关DDNS任务失败,TaskKey不存在")
	}
	config.Configure.DDNSTaskList[taskIndex].Enable = enable

	return config.Save()
}

func UpdateTaskToDDNSTaskList(taskKey string, task ddnsconf.DDNSTask) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	taskIndex := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.DDNSTaskList[i].TaskKey == taskKey {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("找不到需要更新的DDNS任务")
	}

	config.Configure.DDNSTaskList[taskIndex].TaskName = task.TaskName
	config.Configure.DDNSTaskList[taskIndex].TaskType = task.TaskType
	config.Configure.DDNSTaskList[taskIndex].Enable = task.Enable
	config.Configure.DDNSTaskList[taskIndex].GetType = task.GetType
	config.Configure.DDNSTaskList[taskIndex].URL = task.URL
	config.Configure.DDNSTaskList[taskIndex].NetInterface = task.NetInterface
	config.Configure.DDNSTaskList[taskIndex].IPReg = task.IPReg
	config.Configure.DDNSTaskList[taskIndex].Domains = task.Domains
	config.Configure.DDNSTaskList[taskIndex].DNS = task.DNS
	config.Configure.DDNSTaskList[taskIndex].Webhook = task.Webhook
	config.Configure.DDNSTaskList[taskIndex].TTL = task.TTL
	config.Configure.DDNSTaskList[taskIndex].HttpClientTimeout = task.HttpClientTimeout

	config.Configure.DDNSTaskList[taskIndex].ModifyTime = time.Now().Unix()

	return config.Save()
}

func DeleteDDNSTaskListlice(a []ddnsconf.DDNSTask, deleteIndex int) []ddnsconf.DDNSTask {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

func GetDDNSConfigure() ddnsconf.DDNSConfigure {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	conf := config.Configure.DDNSConfigure
	return conf
}

func SetDDNSConfigure(conf *ddnsconf.DDNSConfigure) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	if conf.Intervals < 30 {
		conf.Intervals = 30
	}

	if conf.Intervals > 3600 {
		conf.Intervals = 3600
	}

	if conf.FirstCheckDelay < 0 {
		conf.FirstCheckDelay = 0
	}

	if conf.FirstCheckDelay > 3600 {
		conf.FirstCheckDelay = 3600
	}

	config.Configure.DDNSConfigure = *conf
	return config.Save()
}
