package client

import (
	"encoding/json"
)

type KVS struct {
	Kvs map[string]string `json:"kvs"`
}

type Job struct {
	Labels  KVS `json:"labels"`
	Configs KVS `json:"configs"`
	Secrets KVS `json:"secrets"`
}

type Node struct {
	Labels  KVS   `json:"labels"`
	Configs KVS   `json:"configs"`
	Secrets KVS   `json:"secrets"`
	Jobs    []Job `json:"jobs"`
}

// Test if specified labels are present in node
func (self *Node) TestLabels(labels KVS) bool {
	have := true
	if len(self.Labels.Kvs) != len(labels.Kvs) {
		return false
	}

	for k, _ := range labels.Kvs {
		if _, ok := self.Labels.Kvs[k]; !ok {
			have = false
			break
		}
	}

	return have
}

// Test is specified labels are present in job
func (self *Job) TestLabels(labels KVS) bool {
	have = true
	if len(self.Labels.Kvs) != len(labels.Kvs) {
		return true
	}

	for k, _ := range labels.Kvs {
		if _, ok := self.Labels.Kvs[k]; !ok {
			have = false
			break
		}
	}

	return have
}

// Marshall node informations into byte array
func (self *Node) Marshall() []byte {
	bnode, err := json.Marshal(self)
	Check(err)

	return bnode
}

// If labels are present, add new configs
func (self *Node) AddConfig(labels, data KVS, kind int) {
	switch kind {
	case SECRETS:
		for k, v := range data.Kvs {
			self.Secrets.Kvs[k] = v
		}
	default:
		for k, v := range data.Kvs {
			self.Configs.Kvs[k] = v
		}
	}
}

// If labels are present, add new configs
func (self *Job) AddConfig(labels, data KVS, kind int) {
	switch kind {
	case SECRETS:
		for k, v := range data.Kvs {
			self.Secrets.Kvs[k] = v
		}
	default:
		for k, v := range data.Kvs {
			self.Configs.Kvs[k] = v
		}
	}
}

// Select Jobs that contains labels or key-value pairs specified by user
func (self *Node) SelectJobs(selector KVS) []Job {
	jobs := []Job{}
	for _, job := range self.Jobs {
		if job.testLabels(selector) {
			jobs = append(jobs, job)
		}
	}
	return jobs
}
