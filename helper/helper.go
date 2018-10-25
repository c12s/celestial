package helper

import (
	"sort"
	"strings"
	"time"
)

const (
	namespaces = "namespaces"
	labels     = "labels"
	status     = "status"

	topology = "topology"
	nodes    = "nodes"
	actions  = "actions"
	configs  = "configs"
	secrets  = "secrets"
	undone   = "undone"
)

/*
namespaces
   /namespace
     -data
     /labels
       -kv pairs
     /status
       -status value

namespaces/labels/namespace -> [k:v, k:v]
namespaces/namespace -> {data}
namespaces/namespace/status -> "status"
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

func NSLabels() string {
	return strings.Join([]string{namespaces, labels}, "/")
}

func Compare(a, b []string, strict bool) bool {
	for _, akv := range a {
		for _, bkv := range b {
			if akv == bkv && !strict {
				return true
			}
		}
	}
	return true
}

func Labels(lbs map[string]string) []string {
	lbls := []string{}
	for k, v := range lbs {
		kv := strings.Join([]string{k, v}, ":")
		lbls = append(lbls, kv)
	}
	sort.Strings(lbls)
	return lbls
}

func SplitLabels(value string) []string {
	ls := strings.Split(value, ",")
	sort.Strings(ls)
	return ls
}

/*
topology/regionid/clusters/clusterid/nodes/nodeid

topology/
  -/regionid
     -/clusters
       -/clusterid
         -/nodes
           -/nodeid
             -/namespace:labels
             -/configs
             -/secrets
             -/actions

topology/labels/regionid/clusterid/nodeid -> [k:v, k:v]
topology/regionid/clusterid/nodes/nodeid -> {stats}
topology/regionid/clusterid/nodeid/undone -> [k:v, k:v] unpushed artifacts
topology/regionid/clusterid/nodeid/configs -> {config list with status}
topology/regionid/clusterid/nodeid/secrets -> {secrets list with status} [ecnrypted]
topology/regionid/clusterid/nodeid/actions -> {actions list history with status}
*/

// topology/regionid/clusterid/nodeid -> [k:v, k:v]
func ACSNodeKey(rid, cid, nid string) string {
	s := []string{topology, rid, cid, nid}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodes
func ACSNodesKey(rid, cid string) string {
	s := []string{topology, rid, cid, nodes}
	return strings.Join(s, "/")
}

// topology/labels/regionid/clusterid/nodeid
func ACSLabelsKey(rid, cid, nid string) string {
	s := []string{topology, labels, rid, cid, nid}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/labels
func SearchACSLabelsKey(rid, cid string) string {
	s := []string{topology, rid, cid, labels}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodeid/undone
func ACSUndoneKey(key string) string {
	s := []string{key, undone}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodeid/actions
func ANodeKey(rid, cid, nid string) string {
	prefix := ACSNodeKey(rid, cid, nid)
	s := []string{prefix, actions}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodeid/configs
func CNodeKey(rid, cid, nid string) string {
	prefix := ACSNodeKey(rid, cid, nid)
	s := []string{prefix, configs}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodeid/secrets
func SNodeKey(rid, cid, nid string) string {
	prefix := ACSNodeKey(rid, cid, nid)
	s := []string{prefix, secrets}
	return strings.Join(s, "/")
}

// topology/regionid/clusterid/nodeid/{artifact} [configs | secrets | actions]
func Join(keyPart, artifact string) string {
	s := []string{keyPart, artifact}
	return strings.Join(s, "/")
}

// construct variable path
func JoinParts(artifact string, parts ...string) string {
	s := []string{}
	for _, part := range parts {
		s = append(s, part)
	}

	if artifact != "" {
		s = append(s, artifact)
	}
	return strings.Join(s, "/")
}

func Timestamp() int64 {
	return time.Now().Unix()
}
