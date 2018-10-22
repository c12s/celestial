package helper

import (
	"strings"
)

const (
	namespaces = "namespaces"
	labels     = "labels"
	status     = "status"
)

/*
  /regions/regionid/clusters/clusterid/nodes/nodeid

 /regions
  -/regionid
     -/clusters
       -/clusterid
         -/nodes
           -/nodeid
             -/namespace:labels
             -/stats
             -/configs
             -/secrets
             -/actions
             -/jobs
               -/jobid
                 -/namespace:labels
                 -/configs
                 -/secrets
                 -/containers...

/namespaces
   /namespace
     -data
     /labels
       -kv pairs
     /status
       -status value

/namespaces/labels/namespace -> [k:v, k:v]
/namespaces/namespace -> {data}
/namespaces/namespace/status -> "status"

*/

func Merge(m1, m2 map[string]string) {
	for k, v := range m2 {
		m1[k] = v
	}
}

func NSKey(ns string) string {
	// mid := fmt.Sprintf("%s:%s", namespace, name)
	s := []string{namespaces, ns}
	return strings.Join(s, "/")
}

func NSLabelsKey(name string) string {
	prefix := NSKey(labels)
	s := []string{prefix, name}
	return strings.Join(s, "/")
}

func NSStatusKey(name string) string {
	prefix := NSKey(name)
	s := []string{prefix, status}
	return strings.Join(s, "/")
}

func NS() string {
	return namespaces
}

func Labels() string {
	return strings.Join([]string{namespaces, labels}, "/")
}
