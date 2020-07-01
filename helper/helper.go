package helper

import (
	"context"
	"errors"
	"google.golang.org/grpc/metadata"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	labels = "labels"
	status = "status"

	topology = "topology"
	regions  = "regions"
	nodes    = "nodes"
	actions  = "actions"
	configs  = "configs"
	secrets  = "secrets"
	undone   = "undone"
	watcher  = "watcher"

	tasks     = "tasks"
	allocated = "allocated"
)

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
topology/regions/labels/regionid/clusterid/nodeid -> [k:v, k:v]
topology/regions/regionid/clusterid/nodes/nodeid -> {stats}

topology/regions/regionid/clusterid/nodeid/userid:namespace:configs -> {config list with status}
topology/regions/regionid/clusterid/nodeid/userid:namespace:secrets -> {secrets list with status}
topology/regions/userid:namespace:actions/regionid/clusterid/nodeid/timestamp -> {actions list history with status}

topology/regions/tasks/timestamp -> {tasks submited to particular cluster} holds data until all changes are commited! Append log
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

func SearchKey(regionid, clusterid, userid, namespace string) (string, error) {
	user := strings.Join([]string{userid, namespace}, ":")
	if regionid == "*" && clusterid == "*" {
		return JoinParts("", topology, regions, user, labels), nil // topology/regions/userid:namespace/labels/
	} else if regionid != "*" && clusterid == "*" {
		return JoinParts("", topology, regions, user, labels, regionid), nil // topology/regions/userid:namespace/labels/regionid/
	} else if regionid != "*" && clusterid != "*" { //topology/regions/userId:namespace/labels/regionid/clusterid/
		return JoinParts("", topology, regions, user, labels, regionid, clusterid), nil
	}
	return "", errors.New("Request not valid")
}

// topology/regions/regionid/clusterid/nodeid/userid:namespace:artifact {configs | secrets | adtions}
func NewNSArtifact(userid, namespace, artifact string) string {
	return strings.Join([]string{userid, namespace, artifact}, ":")
}

func NewKey(path, artifact string) string {
	keyPart := strings.Join(strings.Split(path, "/labels/"), "/")
	newKey := Join(keyPart, artifact)
	return newKey
}

func TSToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func TasksKey() string {
	ts := TSToString(Timestamp())
	return JoinFull(topology, regions, tasks, ts)
}

func NodeKey(key string) string {
	return strings.TrimSuffix(NewKey(key, ""), "/")
}

func ConstructKey(node, kind string) string {
	dotted := strings.Join([]string{node, kind}, ".")
	return strings.ReplaceAll(dotted, ".", "/")
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func AppendToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "c12stoken", token)
}

func ExtractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("No token in the request")
	}

	if _, ok := md["c12stoken"]; !ok {
		return "", errors.New("No token in the request")
	}

	return md["c12stoken"][0], nil
}

// region.cluster.nodeid
func ReserveKey(regionid, clusterid, nodeid string) string {
	return strings.Join([]string{regionid, clusterid, nodeid}, ".")
}

func template(user, kind string) string {
	return strings.Join([]string{user, kind}, ":")
}

func GetParts(id string) (string, string, string) {
	parts := strings.Split(id, ".")
	return parts[0], parts[1], parts[2]
}

/*
topology/regions/userid:namespace/labels/regionid/clusterid/nodeid
topology/regions/regionid/clusterid/nodes/nodeid

topology/regions/regionid/clusterid/nodeid/userid:namespace:configs
topology/regions/regionid/clusterid/nodeid/userid:namespace:secrets
topology/regions/userid:n ffamespace:actions/regionid/clusterid/nodeid
allocated/topology/regions/regionid/clusterid/nodeid
*/
func TKeys(regionid, clusterid, nodeid, userid, ns string) map[string]string {
	prefix := strings.Join([]string{topology, regions}, "/")
	id := strings.Join([]string{regionid, clusterid, nodeid}, "/")
	user := strings.Join([]string{userid, ns}, ":")
	return map[string]string{
		"watch":   strings.Join([]string{allocated, prefix, regionid, clusterid}, "/"),
		"nodeid":  strings.Join([]string{prefix, id}, "/"),
		"labels":  strings.Join([]string{prefix, user, labels, id}, "/"),
		"configs": strings.Join([]string{prefix, id, template(user, "configs")}, "/"),
		"secrets": strings.Join([]string{prefix, id, template(user, "secrets")}, "/"),
		"actions": strings.Join([]string{prefix, template(user, "actions"), id}, "/"),
	}
}

// watch for nodes in some region.cluster
// and this can be used on restart of service to start watchers again
// watcher/topology/regions/region/cluster -> topology/regions/region/cluster
func WatcherKey(key string) string {
	return strings.Join([]string{watcher, key}, "/")
}

func Trim(key string) string {
	return strings.ReplaceAll(key, "allocated/", "")
}
