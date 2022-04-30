package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"minik8s/object"
	"minik8s/pkg/klog"
	"net/http"
	"net/url"
)

type Config struct {
	Host string
}

type RESTClient struct {
	Base   *url.URL // url = base+resource+name
	Client *http.Client
}

/******************************Pod*******************************/

func (r RESTClient) CreatePods(ctx context.Context, template *object.PodTemplateSpec) error {
	pod, _ := GetPodFromTemplate(template)
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)
	attachURL := "/registry/pod/default"

	req, _ := http.NewRequest("POST", r.Base.String()+attachURL, reqBody)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("create pod fail")
	}

	result := &object.Pod{}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail")
	}
	return err
}

func (r RESTClient) UpdatePods(ctx context.Context, pod *object.Pod) error {
	podRaw, _ := json.Marshal(pod)
	reqBody := bytes.NewBuffer(podRaw)
	attachURL := "/pod"

	req, _ := http.NewRequest("POST", r.Base.String()+attachURL, reqBody)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("create pod fail")
	}
	defer resp.Body.Close()

	return nil
}

func (r RESTClient) DeletePod(ctx context.Context, podName string) error {
	attachURL := "/registry/pod/default" + podName
	req, _ := http.NewRequest("DELETE", r.Base.String()+attachURL, nil)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return errors.New("delete pod fail")
	}
	defer resp.Body.Close()

	return nil
}

// GetPodFromTemplate TODO: type conversion
func GetPodFromTemplate(template *object.PodTemplateSpec) (*object.Pod, error) {
	pod := &object.Pod{}
	pod.Spec = object.PodSpec{}
	return pod, nil
}

/********************************RS*****************************/

func (r RESTClient) GetRS(name string) (*object.ReplicaSet, error) {
	attachURL := "/rs/" + name

	req, _ := http.NewRequest("GET", r.Base.String()+attachURL, nil)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("delete pod fail")
	}

	result := &object.ReplicaSet{}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail")
	}
	return result, nil
}

func (r RESTClient) GetRSPods(name string) ([]*object.Pod, error) {
	attachURL := "/pod/" + name

	req, _ := http.NewRequest("GET", r.Base.String()+attachURL, nil)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("get rs pod fail")
	}

	var result []*object.Pod
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, &result)
	if err != nil {
		klog.Infof("[GetRSPods] Body Pods Unmarshal fail")
	}

	return result, nil
}

func (r RESTClient) UpdateRSStatus(ctx context.Context, replicaSet *object.ReplicaSet) (*object.ReplicaSet, error) {
	body, _ := json.Marshal(replicaSet)
	reqBody := bytes.NewBuffer(body)
	attachURL := "/rs/" + replicaSet.Name + "/" + "status"

	req, _ := http.NewRequest("PUT", r.Base.String()+attachURL, reqBody)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != object.SUCCESS {
		return nil, errors.New("update rs fail")
	}

	result := &object.ReplicaSet{}
	body, _ = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail")
	}
	return result, nil
}

/********************************Node*****************************/

func (r RESTClient) GetNodes() ([]*object.Node, error) {
	return nil, nil
}

/********************************watch*****************************/

// WatchRegister get ticket for message queue
func (r RESTClient) WatchRegister(resource string, name string, withPrefix bool) (*string, *int64, error) {
	attachURL := "/" + resource + "/default"
	if !withPrefix {
		attachURL += "/" + name
	}
	req, _ := http.NewRequest("PUT", r.Base.String()+attachURL, nil)
	resp, _ := r.Client.Do(req)

	if resp.StatusCode != 200 {
		return nil, nil, errors.New("api server register fail")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	result := &TicketResponse{}
	err := json.Unmarshal(body, result)
	if err != nil {
		klog.Infof("[CreatePods] Body Pods Unmarshal fail")
	}
	return &attachURL, &result.Ticket, nil

}