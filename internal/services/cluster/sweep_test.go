package cluster

// Cluster sweeper is intentionally omitted.
//
// The Cluster API does not expose a name or label field that could be used
// to identify test resources by the "tfacc-" prefix convention. The Cluster
// struct only contains: ID, Subnet, and Nodes. Without a name field, there
// is no safe way to distinguish test-created clusters from user-created clusters.
//
// Acceptance tests that create clusters should clean up in their own
// CheckDestroy functions.
