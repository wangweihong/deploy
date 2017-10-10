package controllers

import (
	"encoding/json"
	"fmt"
	jk "ufleet-deploy/pkg/resource/job"
	pk "ufleet-deploy/pkg/resource/pod"
)

type JobState struct {
	Name        string   `json:"name"`
	User        string   `json:"user"`
	Workspace   string   `json:"workspace"`
	Group       string   `json:"group"`
	Images      []string `json:"images"`
	Containers  []string `json:"containers"`
	PodNum      int      `json:"podnum"`
	ClusterIP   string   `json:"clusterip"`
	CompleteNum int      `json:"completenum"`
	ParamNum    int      `json:"paramnum"`
	Succeeded   int      `json:"succeeded"`
	Failed      int      `json:"failed"`
	CreateTime  int64    `json:"createtime"`
	StartTime   int64    `json:"starttime"`
	Reason      string   `json:"reason"`
	//	Pods       []string `json:"pods"`
	PodStatus []pk.Status `json:"podstatus"`
}

type JobController struct {
	baseController
}

func GetJobState(ji jk.JobInterface) JobState {
	var js JobState
	job := ji.Info()
	js.Name = job.Name
	js.User = job.User
	js.Workspace = job.Workspace
	js.Group = job.Group

	runtime, err := ji.GetRuntime()
	if err != nil {
		js.Reason = err.Error()
		return js
	}

	js.CreateTime = runtime.Job.CreationTimestamp.Unix()
	if runtime.Job.Spec.Parallelism != nil {
		js.ParamNum = int(*runtime.Job.Spec.Parallelism)
	}
	if runtime.Job.Spec.Completions != nil {
		js.CompleteNum = int(*runtime.Job.Spec.Completions)
	}

	if runtime.Job.Status.StartTime != nil {
		js.StartTime = runtime.Job.Status.StartTime.Unix()
	}

	js.Succeeded = int(runtime.Job.Status.Succeeded)
	js.Failed = int(runtime.Job.Status.Failed)

	for _, v := range runtime.Job.Spec.Template.Spec.Containers {
		js.Containers = append(js.Containers, v.Name)
	}

	js.PodNum = len(runtime.Pods)
	if js.PodNum != 0 {
		pod := runtime.Pods[0]
		js.ClusterIP = pod.Status.HostIP
		for _, v := range pod.Spec.Containers {
			js.Images = append(js.Images, v.Image)
		}
	}

	for _, v := range runtime.Pods {
		ps := pk.V1PodToPodStatus(*v)
		js.PodStatus = append(js.PodStatus, *ps)

	}

	return js

}

// ListJobs
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *JobController) ListGroupJobs() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := jk.Controller.List(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//jobs := make([]jk.Job, 0)
	jss := make([]JobState, 0)

	for _, v := range pis {
		js := GetJobState(v)
		jss = append(jss, js)
	}

	this.normalReturn(jss)
}

// ListGroupsJobs
// @Title Job
// @Description   Job
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *JobController) ListGroupsJobs() {

	groups := make([]string, 0)
	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.errReturn(err, 500)
		return
	}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &groups)
	if err != nil {
		err = fmt.Errorf("try to unmarshal data \"%v\" fail for %v", string(this.Ctx.Input.RequestBody), err)
		this.errReturn(err, 500)
		return
	}

	pis := make([]jk.JobInterface, 0)
	for _, v := range groups {
		tmp, err := jk.Controller.ListGroup(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	//jobs := make([]jk.Job, 0)
	jss := make([]JobState, 0)

	for _, v := range pis {
		js := GetJobState(v)
		jss = append(jss, js)
	}

	this.normalReturn(jss)
}

// DeleteJob
// @Title Job
// @Description  DeleteeJob
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param job path string true "任务名"
// @Success 201 {string} create success!
// @Failure 500
// @router /:job/group/:group/workspace/:workspace [Delete]
func (this *JobController) DeleteJob() {

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	job := this.Ctx.Input.Param(":job")

	err := jk.Controller.Delete(group, workspace, job, jk.DeleteOption{})
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}
