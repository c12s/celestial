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
func (self *Node) testLabels(labels KVS) bool {
	retChan := make(chan bool)
	if len(self.Labels.Kvs) != len(labels.Kvs) {
		return false
	}

	go func() {
		for k, _ := range labels.Kvs {
			if _, ok := self.Labels.Kvs[k]; !ok {
				retChan <- false
				break
			}
		}
		retChan <- true
	}()
	return <-retChan
}

// Test is specified labels are present in job
func (self *Job) testLabels(labels KVS) bool {
	retChan := make(chan bool)
	if len(self.Labels.Kvs) != len(labels.Kvs) {
		return true
	}

	go func() {
		for k, _ := range labels.Kvs {
			if _, ok := self.Labels.Kvs[k]; !ok {
				retChan <- false
				break
			}
		}
		retChan <- true
	}()
	return <-retChan
}

// Marshall node informations into byte array
func (self *Node) marshall() []byte {
	bnode, err := json.Marshal(self)
	Check(err)

	return bnode
}

// If labels are present, add new configs
func (self *Node) addConfig(labels, data KVS, kind int) {
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
func (self *Job) addConfig(labels, data KVS, kind int) {
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
// Return Job chanel from witch jobs will arrive
func (self *Node) selectJobs(selector KVS) <-chan Job {
	jobChan := make(chan Job)
	go func() {
		for _, job := range self.Jobs {
			if job.testLabels(selector) {
				jobChan <- job
			}
		}
		close(jobChan)
	}()

	return jobChan
}
