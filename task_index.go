// Copyright 2013-2014 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"regexp"
	"strconv"
	"strings"
)

// Task used to poll for long running create index completion.
type IndexTask struct {
	BaseTask

	namespace string
	indexName string
}

// Initialize task with fields needed to query server nodes.

func NewIndexTask(cluster *Cluster, namespace string, indexName string) *IndexTask {
	return &IndexTask{
		BaseTask:  *NewTask(cluster, false),
		namespace: namespace,
		indexName: indexName,
	}
}

// Query all nodes for task completion status.
func (this *IndexTask) IsDone() (bool, error) {
	command := "sindex/" + this.namespace + "/" + this.indexName
	nodes := this.cluster.GetNodes()
	complete := false

	r := regexp.MustCompile(`\.*load_pct=(\d+)\.*`)

	for _, node := range nodes {
		responseMap, err := RequestInfoForNode(node, command)
		if err != nil {
			return true, err
		}

		response := responseMap["statistics"]
		find := "load_pct="
		index := strings.Index(response, find)

		if index < 0 {
			complete = true
			continue
		}

		matchRes := r.FindStringSubmatch(response)
		// we know it exists and is a valid number
		pct, _ := strconv.Atoi(matchRes[1])

		if pct >= 0 && pct < 100 {
			return false, nil
		}
		complete = true
	}
	return complete, nil
}

func (this *IndexTask) OnComplete() chan error {
	return this.onComplete(this)
}
