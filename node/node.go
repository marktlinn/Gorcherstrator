package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/marktlinn/Gorcherstrator/stats"
	"github.com/marktlinn/Gorcherstrator/utils"
)

// A Node represents a physical machine within a Cluster.
type Node struct {
	Name            string
	IP              string
	Memory          int64
	MemoryAllocated int64
	Cores           int
	Disk            int64
	DiskAllocated   int64
	Role            string
	TaskCount       int
	Api             string
	Stats           stats.Stats
}

// NewNode returns a reference to a new Node entity.
func NewNode(name, api, role string) *Node {
	return &Node{
		Name: name,
		Api:  api,
		Role: role,
	}
}

// getNodeStats is a helper function that makes the requests to the Node's "/stats/" endpoint to retrieve the stats from the Node.
func (n *Node) GetStats() (*stats.Stats, error) {
	var res *http.Response
	var err error

	url := fmt.Sprintf("%s/stats", n.Api)
	res, err = utils.HTTPWithRetry(http.Get, url)
	if err != nil {
		errMsg := fmt.Sprintf("failed to connect to %s: %s\n", n.Api, err)
		return nil, errors.New(errMsg)
	}

	if res.StatusCode != 200 {
		errMsg := fmt.Sprintf("failed to retrieve stats from node: %s; response StatusCode: %d\n",
			n.Api,
			res.StatusCode)
		return nil, errors.New(errMsg)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var status stats.Stats
	err = json.Unmarshal(body, &status)
	if err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal JSON data into Status object: %s\n", err)
		return nil, errors.New(errMsg)
	}

	n.Disk = int64(status.DiskTotal())
	n.Memory = int64(status.MemTotalKB())

	return &status, nil
}
