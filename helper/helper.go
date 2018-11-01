package helper

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const (
	namespaces = "namespaces"
	labels     = "labels"
	status     = "status"

	topology = "topology"
	regions  = "regions"
	nodes    = "nodes"
	actions  = "actions"
	configs  = "configs"
	secrets  = "secrets"
	undone   = "undone"
)

/*
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
TODO: **MAYBE** NEW KEY SPACE
topology/regions/labels/regionid/clusterid/nodeid -> [k:v, k:v]
topology/regions/regionid/clusterid/nodes/nodeid -> {stats}
topology/regions/configs/regionid/clusterid/nodeid -> {config list with status}
topology/regions/secrets/regionid/clusterid/nodeid -> {secrets list with status} [ecnrypted]
topology/regions/actions/regionid/clusterid/nodeid/timestamp -> {actions list history with status}
*/

// topology/regions/regionid/clusterid/nodeid -> [k:v, k:v]
func ACSNodeKey(rid, cid, nid string) string {
	s := []string{topology, regions, rid, cid, nid}
	return strings.Join(s, "/")
}

// topology/regions/regionid/clusterid/nodes
func ACSNodesKey(rid, cid string) string {
	s := []string{topology, regions, rid, cid, nodes}
	return strings.Join(s, "/")
}

// topology/regions/labels/regionid/clusterid/nodeid
func ACSLabelsKey(rid, cid, nid string) string {
	s := []string{topology, regions, labels, rid, cid, nid}
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

func JoinFull(parts ...string) string {
	s := []string{}
	for _, part := range parts {
		s = append(s, part)
	}
	return strings.Join(s, "/")
}

func Timestamp() int64 {
	return time.Now().Unix()
}

// from: topology/regions/labels/regionid/clusterid/nodeid -> topology/regions/{replacement}/regionid/clusterid/nodeid
// {configs | secrets | actions}
func Key(path, replacement string) string {
	return strings.Replace(path, labels, replacement, -1)
}

func SearchKey(regionid, clusterid string) (string, error) {
	if regionid == "*" && clusterid == "*" {
		return JoinParts("", topology, regions, labels), nil // topology/regions/labels/
	} else if regionid != "*" && clusterid == "*" {
		return JoinParts("", topology, regions, labels, regionid), nil // topology/regions/labels/regionid/
	} else if regionid != "*" && clusterid != "*" { //topology/regions/labels/regionid/clusterid/
		return JoinParts("", topology, regions, labels, regionid, clusterid), nil
	}
	return "", errors.New("Request not valid")
}
