package node

// A Node represents a physical machine within a Cluster.
type Node struct {
	Name            string
	IP              string
	Memory          int
	MemoryAllocated int
	Cores           int
	Disk            int
	DiskAllocated   int
	Role            string
	TaskCount       int
	Api             string
}

// NewNode returns a reference to a new Node entity.
func NewNode(name, api, role string) *Node {
	return &Node{
		Name: name,
		Api:  api,
		Role: role,
	}
}
