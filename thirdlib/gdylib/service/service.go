package service

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

const (
	//StateStop 未启动
	StateStop = 0
	//StateRunning 正在运行
	StateRunning = 1
	//StateStopping 正在结束
	StateStopping = 2
)

//ServiceMsg 服务消息
type ServiceMsg struct {
	Type   string
	Params []any
}

//var serviceMap map[string]*Service //用于保存已创建的服务
//var serverMapRWLock sync.RWMutex
var serviceMap = struct {
	sync.RWMutex
	services map[string]*Service
}{
	services: map[string]*Service{},
}

//Service 服务
type Service struct {
	Name         string //服务名称
	serviceMutex sync.Mutex

	State                uint8        //服务状态 //服务运行状态 , 0:未启动 1:正在运行 2:正在结束
	NextAction           string       //动作名称
	DefaultAction        string       //默认动作
	TimerFunc            func(...any) //定时回调函数
	EventFunc            func(...any) //事件回调函数
	StopFinishedCallback func(...any) //服务停止后的回调函数
	eventChan            chan any
	Timer                *time.Timer
	//Args     []interface{}

	context    context.Context
	cancelFunc context.CancelFunc
}

//NewService 创建服务对象
func NewService(name string) (*Service, error) {
	serviceMap.Lock()
	defer serviceMap.Unlock()
	if _, ok := serviceMap.services[name]; ok {
		return nil, fmt.Errorf("命名为[%s]服务已存在", name)
	}

	if _, ok := serviceMap.services[name]; !ok {
		service := Service{Name: name, State: StateStop}
		service.eventChan = make(chan any, 32)
		serviceMap.services[name] = &service
		return &service, nil
	}

	panic(fmt.Sprintf("命名为[%s]服务已存在", name))
	//return nil, fmt.Errorf("命名为[%s]服务已存在", name)
}

//SetDefaultAction 设置默认action
func (s *Service) SetDefaultAction(action string) *Service {
	s.DefaultAction = action
	return s
}

//SetTimerFunc 设置定时功能函数
func (s *Service) SetTimerFunc(timerFunc func(...any)) *Service {
	s.TimerFunc = timerFunc
	return s
}

//SetEventFunc 设置时间功能函数
func (s *Service) SetEventFunc(eventFunc func(...any)) *Service {
	s.EventFunc = eventFunc
	return s
}

//SetStopFinishedCallback 设置服务停止后的回调函数
func (s *Service) SetStopFinishedCallback(f func(...any)) *Service {
	s.StopFinishedCallback = f
	return s
}

//Start 服务启动
func (s *Service) Start(vs ...any) error {
	s.serviceMutex.Lock()
	defer s.serviceMutex.Unlock()
	if s.State == StateRunning {
		text := fmt.Sprintf("服务 [%s]已启动,无需再次启动", s.Name)
		return fmt.Errorf(text)
	}

	if s.State == StateStopping {
		text := fmt.Sprintf("服务[%s]正在结束,请结束后再次启动", s.Name)
		return fmt.Errorf(text)
	}

	s.State = StateRunning
	s.NextAction = s.DefaultAction
	log.Printf("服务[%s] 启动", s.Name)
	s.context, s.cancelFunc = context.WithCancel(context.Background())

	go s.loop(vs)
	return nil
}

//Stop 服务结束
func (s *Service) Stop() error {
	s.serviceMutex.Lock()
	defer s.serviceMutex.Unlock()
	if s.State == StateStop {
		text := fmt.Sprintf("服务[%s]未启动,无须停止", s.Name)
		return fmt.Errorf(text)
	}
	if s.State == StateStopping {
		text := fmt.Sprintf("服务[%s]正在结束,无须再次结束.", s.Name)
		return fmt.Errorf(text)
	}

	if s.cancelFunc == nil {
		return fmt.Errorf("服务[%s]context nil", s.Name)
	}
	s.cancelFunc()

	s.State = StateStopping
	return nil
}

func (s *Service) loop(params ...any) {
	defer func() {
		recoverErr := recover()
		if recoverErr == nil {
			return
		}
		log.Printf("service[%s] panic:\n%v", s.Name, recoverErr)
		log.Printf("\n%s\n", debug.Stack())
		s.State = StateStop
		log.Printf("server[%s] restart", s.Name)
		s.Start()
	}()

	s.Timer = time.NewTimer(0)

	for {

		select {
		case <-s.Timer.C:
			{
				if s.TimerFunc != nil { //如果设置了定时回调的话
					s.TimerFunc(s, params)
				}
			}
		case <-s.context.Done():
			{
				if s.State == StateStopping {
					s.State = StateStop
					log.Printf("服务[%s] 停止", s.Name)

					if s.StopFinishedCallback != nil {
						s.StopFinishedCallback()
					}

					s.context = nil
					s.cancelFunc = nil
					return
				}

				if s.State == StateStop {
					log.Printf("服务[%s] 状态有误,结束服务失败", s.Name)
					return
				}

				if s.State == StateRunning {
					log.Printf("服务[%s] 状态有误,请使用正确途径结束", s.Name)
					return
				}
			}
		case msg := <-s.eventChan:
			{
				if s.EventFunc != nil {
					s.EventFunc(s, msg)
				}
			}

		}

	}
}

//Restart 重启服务
func (s *Service) Restart(params ...any) error {
	s.Stop()

	var timeout time.Duration
	if len(params) == 0 {
		timeout = time.Second * 15
	} else {
		timeout = params[0].(time.Duration)
	}

	preTime := time.Now()
	for {
		waitTime := time.Now()
		if s.State == StateStop {
			return s.Start()
		}
		if waitTime.Sub(preTime) > timeout {
			return fmt.Errorf("重启服务[%s]超时", s.Name)
		}
		<-time.After(time.Millisecond * 100)
	}

}

//Message 发送消息给service
//t 消息类型
//params 消息内容
func (s *Service) Message(t string, params ...any) error {

	s.serviceMutex.Lock()
	defer s.serviceMutex.Unlock()

	if s.State != StateRunning {
		return fmt.Errorf("[%s]服务已关闭或者正在关闭,无法处理消息 %v,", s.Name, params)
	}
	msg := ServiceMsg{Type: t}
	for i := range params {
		msg.Params = append(msg.Params, params[i])
	}
	select {
	case s.eventChan <- msg:
	default:
		return fmt.Errorf("[%s]服务EventChan阻塞,无法处理event", s.Name)
	}
	return nil
}

//*************************************

//GetService 查询服务根据服务名称
func GetService(name string) (*Service, error) {
	serviceMap.RLock()
	defer serviceMap.RUnlock()
	if _, ok := serviceMap.services[name]; !ok {
		return nil, fmt.Errorf("service[%s]不存在", name)
	}
	service := serviceMap.services[name]
	return service, nil
}

//Message 发送消息到service,根据serviceName
func Message(serviceName string, msgType string, params ...any) error {
	service, err := GetService(serviceName)
	if err != nil {
		return fmt.Errorf("service[%s] Message faild,:%s", serviceName, err.Error())
	}
	return service.Message(msgType, params...)
}

//Stop 服务停止,根据serviceName
func Stop(serviceName string) error {
	service, err := GetService(serviceName)
	if err != nil {
		return fmt.Errorf("service[%s] stop faild:%s", serviceName, err.Error())
	}
	return service.Stop()
}

//Restart 服务重启
func Restart(serviceName string) error {
	service, err := GetService(serviceName)
	if err != nil {
		return fmt.Errorf("restart servcie[%s] faild,:%s", serviceName, err.Error())
	}
	return service.Restart()
}

//Start 服务启动
func Start(serviceName string) error {
	service, err := GetService(serviceName)
	if err != nil {
		return fmt.Errorf("servcie[%s] start faild,:%s", serviceName, err.Error())
	}
	return service.Start()
}
