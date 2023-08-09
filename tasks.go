package main

type TaskCommon struct {
	Deployment_id int64
	Id            int64
}

type WasmTask struct {
	TaskCommon
	Source string
	Func   string
}

type DockerTask struct {
	TaskCommon
	Url string
}
