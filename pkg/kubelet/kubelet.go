package kubelet

import (
	"encoding/json"
	"fmt"
	"minik8s/object"
	"minik8s/pkg/apiserver/config"
	"minik8s/pkg/client"
	"minik8s/pkg/etcdstore"
	"minik8s/pkg/klog"
	"minik8s/pkg/kubelet/monitor"
	"minik8s/pkg/kubelet/podConfig"
	"minik8s/pkg/kubelet/podManager"
	"minik8s/pkg/kubelet/types"
	"minik8s/pkg/kubeproxy"
	"minik8s/pkg/listerwatcher"
	"minik8s/pkg/netSupport"
	"minik8s/pkg/tools"
	"os"
	"path"
	"time"

	"golang.org/x/net/context"
)

type Kubelet struct {
	podManager     *podManager.PodManager
	PodConfig      *podConfig.PodConfig
	podMonitor     *monitor.DockerMonitor
	kubeNetSupport *netSupport.KubeNetSupport
	kubeProxy      *kubeproxy.KubeProxy
	ls             *listerwatcher.ListerWatcher
	stopChannel    <-chan struct{}
	Client         client.RESTClient
	Err            error
}

func NewKubelet(lsConfig *listerwatcher.Config, clientConfig client.Config, node *object.Node) *Kubelet {
	kubelet := &Kubelet{}
	kubelet.podManager = podManager.NewPodManager(clientConfig)
	restClient := client.RESTClient{
		Base: "http://" + clientConfig.Host,
	}
	kubelet.Client = restClient

	// initialize list watch
	ls, err := listerwatcher.NewListerWatcher(lsConfig)
	if err != nil {
		fmt.Printf("[NewKubelet] list watch start fail...")
	}
	kubelet.ls = ls
	kubelet.kubeNetSupport, err = netSupport.NewKubeNetSupport(lsConfig, clientConfig, node)
	if err != nil {
		fmt.Printf("[NewKubelet] new kubeNetSupport fail")
	}
	kubelet.kubeProxy = kubeproxy.NewKubeProxy(lsConfig, clientConfig)
	// initialize pod podConfig
	kubelet.PodConfig = podConfig.NewPodConfig()

	kubelet.podMonitor = monitor.NewDockerMonitor()

	return kubelet
}
func (k *Kubelet) getNodeName() string {
	netSupport := k.kubeNetSupport.GetKubeproxySnapShoot()
	return netSupport.NodeName
}

func (kl *Kubelet) Run() {
	kl.kubeNetSupport.StartKubeNetSupport()
	kl.kubeProxy.StartKubeProxy()
	updates := kl.PodConfig.GetUpdates()
	go kl.podMonitor.Listener()
	go kl.syncLoop(updates, kl)
	go kl.DoMonitor(context.Background())
	go func() {
		err := kl.ls.Watch(config.PodConfigPREFIX, kl.watchPod, kl.stopChannel)
		if err != nil {
			fmt.Printf("[kubelet] watch podConfig error" + err.Error())
		} else {
			return
		}
		time.Sleep(10 * time.Second)
	}()
	go func() {
		err := kl.ls.Watch(config.SharedDataPrefix, kl.watchSharedData, kl.stopChannel)
		if err == nil {
			return
		}
		time.Sleep(10 * time.Second)
	}()
}

func (kl *Kubelet) syncLoop(updates <-chan types.PodUpdate, handler SyncHandler) {
	for {
		kl.syncLoopIteration(updates, handler)
	}
}

func (k *Kubelet) AddPod(pod *object.Pod) error {
	return k.podManager.AddPod(pod)
}
func (k *Kubelet) DeletePod(podName string) error {
	return k.podManager.DeletePod(podName)
}

type SyncHandler interface {
	HandlePodAdditions(pods []*object.Pod)
	HandlePodUpdates(pods []*object.Pod)
	HandlePodRemoves(pods []*object.Pod)
}

// TODO: channel pod type?
func (kl *Kubelet) syncLoopIteration(ch <-chan types.PodUpdate, handler SyncHandler) bool {
	select {
	case u, open := <-ch:
		if !open {
			fmt.Printf("Update channel is closed")
			return false
		}
		switch u.Op {
		case types.UPDATE:
			handler.HandlePodUpdates(u.Pods)
		case types.ADD:
			handler.HandlePodAdditions(u.Pods)
		case types.DELETE:
			handler.HandlePodRemoves(u.Pods)
		}
	}
	return true
}

// TODO: check the message by node name. DO NOT handle pods not belong to this node
func (kl *Kubelet) watchPod(res etcdstore.WatchRes) {
	if res.ResType == etcdstore.DELETE {
		//不管delete,实际的delete通过设置pod里边的status实现
		return
	}
	pod := &object.Pod{}
	err := json.Unmarshal(res.ValueBytes, pod)
	if err != nil {
		klog.Warnf("watchNewPod bad message\n")
		return
	}
	// reject message if pod not assign pod or not belong to the node
	if pod.Spec.NodeName == "" {
		return
	}
	fmt.Printf("[watchPod] New message...\n")
	pods := []*object.Pod{pod}
	ok := kl.podManager.CheckIfPodExist(pod.Name)
	if !ok {
		//pod 不存在,
		if pod.Spec.NodeName != kl.getNodeName() {
			//pod本地不存在且和本节点无关
			return
		}
		if pod.Status.Phase != object.Delete {
			//分配给自己的pod,且Phase不为Delete
			podUp := types.PodUpdate{
				Pods: pods,
				Op:   types.ADD,
			}
			kl.PodConfig.GetUpdates() <- podUp
		} else {
			//命令删除同时已经删除，直接返回即可
			return
		}
	} else {
		//已经存在该pod
		if pod.Spec.NodeName == kl.getNodeName() {
			if pod.Status.Phase == object.Delete {
				//需要删除pod
				podUp := types.PodUpdate{
					Pods: pods,
					Op:   types.DELETE,
				}
				kl.PodConfig.GetUpdates() <- podUp
			} else {
				//节点对应上，那么是修改配置文件
				podUp := types.PodUpdate{
					Pods: pods,
					Op:   types.UPDATE,
				}
				kl.PodConfig.GetUpdates() <- podUp
			}
		} else {
			//自己存在该pod但是被分配到了其他地方, 应该删除本地的pod
			podUp := types.PodUpdate{
				Pods: pods,
				Op:   types.DELETE,
			}
			kl.PodConfig.GetUpdates() <- podUp
		}
	}
	return
}

func (kl *Kubelet) HandlePodAdditions(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] Prepare add pod:%s\npod:%+v\n", pod.Name, pod)
		err := kl.podManager.AddPod(pod)
		if err != nil {
			fmt.Println("[kubelet]AddPod error" + err.Error())
			kl.Err = err
		}
	}
}

func (kl *Kubelet) HandlePodUpdates(pods []*object.Pod) {
	//先删除原来的再增加新的
	for _, pod := range pods {
		err := kl.podManager.DeletePod(pod.Name)
		if err != nil {
			fmt.Printf("[Kubelet] Delete pod fail...")
			fmt.Printf(err.Error())
			kl.Err = err
		}
	}
	//创建新的
	for _, pod := range pods {
		err := kl.podManager.AddPod(pod)
		if err != nil {
			fmt.Printf("[Kubelet] Add pod fail...")
			fmt.Printf(err.Error())
			kl.Err = err
		}
	}
}

func (kl *Kubelet) HandlePodRemoves(pods []*object.Pod) {
	for _, pod := range pods {
		fmt.Printf("[Kubelet] Prepare delete pod:%+v\n", pod)
		err := kl.podManager.DeletePod(pod.Name)
		if err != nil {
			fmt.Printf("[Kubelet] Delete pod fail...\n")
		}
	}
}

func (kl *Kubelet) DoMonitor(ctx context.Context) {
	for {
		// fmt.Printf("[DoMonitor] New round monitoring...\n")
		podMap := kl.podManager.CopyName2pod()
		for _, pod := range podMap {
			kl.podMonitor.MetricDockerStat(ctx, pod)
		}
		time.Sleep(time.Second * 2)
	}
}

func (kl *Kubelet) watchSharedData(res etcdstore.WatchRes) {
	switch res.ResType {
	case etcdstore.PUT:
		jobZipFile := object.JobZipFile{}
		err := json.Unmarshal(res.ValueBytes, &jobZipFile)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		zipName := jobZipFile.Key + ".zip"
		unzippedDir := path.Join(config.SharedDataDirectory, jobZipFile.Key)
		err = tools.Bytes2File(jobZipFile.Zip, zipName, config.SharedDataDirectory)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		err = tools.Unzip(path.Join(config.SharedDataDirectory, zipName), unzippedDir)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		err = tools.Bytes2File(jobZipFile.Slurm, "sbatch.slurm", unzippedDir)
		if err != nil {
			klog.Errorf("%s\n", err.Error())
			return
		}
		break
	case etcdstore.DELETE:
		jobKey := path.Base(res.Key)
		err := tools.RemoveAll(path.Join(config.SharedDataDirectory, jobKey))
		if err != nil {
			klog.Errorf("%s\n", err.Error())
		}
		err = os.Remove(path.Join(config.SharedDataDirectory, jobKey+".zip"))
		if err != nil {
			klog.Errorf("%s\n", err.Error())
		}
		break
	}
}
