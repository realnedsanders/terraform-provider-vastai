package client

import (
	"context"
	"fmt"
	"strconv"
)

// ClusterService handles cluster-related API operations.
type ClusterService struct {
	client *VastAIClient
}

// ClusterNode represents a machine node in a cluster.
type ClusterNode struct {
	MachineID        int    `json:"machine_id"`
	IsClusterManager bool   `json:"is_cluster_manager"`
	LocalIP          string `json:"local_ip"`
}

// Cluster represents a cluster object from the Vast.ai API.
type Cluster struct {
	ID     int           `json:"id"`
	Subnet string        `json:"subnet"`
	Nodes  []ClusterNode `json:"nodes"`
}

// ClusterListResponse wraps the cluster list API response.
// The API returns clusters as a map with string IDs as keys.
type ClusterListResponse struct {
	Clusters map[string]Cluster `json:"clusters"`
}

// Create creates a new cluster with the given subnet and manager machine ID.
// Sends POST /cluster/ with {"subnet": subnet, "manager_id": managerID}.
// Uses create-then-read pattern: the create response returns only a message,
// so we list clusters and find the one matching the subnet (Pitfall 4).
func (s *ClusterService) Create(ctx context.Context, subnet string, managerID int) (*Cluster, error) {
	body := map[string]interface{}{
		"subnet":     subnet,
		"manager_id": managerID,
	}
	if err := s.client.Post(ctx, "/cluster/", body, nil); err != nil {
		return nil, fmt.Errorf("creating cluster: %w", err)
	}

	// Create-then-read: list all clusters and find the newly created one by subnet
	clusters, err := s.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading cluster after create: %w", err)
	}

	for _, cluster := range clusters {
		if cluster.Subnet == subnet {
			return &cluster, nil
		}
	}

	return nil, fmt.Errorf("cluster with subnet %q not found after creation", subnet)
}

// List retrieves all clusters.
// Sends GET /clusters/. The API returns clusters as a map with string IDs as keys.
// Cluster.ID is populated from the map key (parsed from string to int).
func (s *ClusterService) List(ctx context.Context) (map[string]Cluster, error) {
	var resp ClusterListResponse
	if err := s.client.Get(ctx, "/clusters/", &resp); err != nil {
		return nil, fmt.Errorf("listing clusters: %w", err)
	}

	// Populate Cluster.ID from the map key (API uses string IDs as map keys)
	result := make(map[string]Cluster, len(resp.Clusters))
	for key, cluster := range resp.Clusters {
		id, err := strconv.Atoi(key)
		if err == nil {
			cluster.ID = id
		}
		result[key] = cluster
	}

	return result, nil
}

// Delete deletes a cluster by ID.
// Sends DELETE /cluster/ with {"cluster_id": clusterID} in the request body (Pitfall 7).
func (s *ClusterService) Delete(ctx context.Context, clusterID int) error {
	body := map[string]interface{}{
		"cluster_id": clusterID,
	}
	if err := s.client.DeleteWithBody(ctx, "/cluster/", body, nil); err != nil {
		return fmt.Errorf("deleting cluster %d: %w", clusterID, err)
	}
	return nil
}

// JoinMachine adds machines to a cluster.
// Sends PUT /cluster/ with {"cluster_id": clusterID, "machine_ids": machineIDs}.
func (s *ClusterService) JoinMachine(ctx context.Context, clusterID int, machineIDs []int) error {
	body := map[string]interface{}{
		"cluster_id":  clusterID,
		"machine_ids": machineIDs,
	}
	if err := s.client.Put(ctx, "/cluster/", body, nil); err != nil {
		return fmt.Errorf("joining machines to cluster %d: %w", clusterID, err)
	}
	return nil
}

// RemoveMachine removes a machine from a cluster.
// Sends DELETE /cluster/remove_machine/ with {"cluster_id": clusterID, "machine_id": machineID}.
// If newManagerID is non-nil, includes "new_manager_id" to designate a new cluster manager.
func (s *ClusterService) RemoveMachine(ctx context.Context, clusterID, machineID int, newManagerID *int) error {
	body := map[string]interface{}{
		"cluster_id": clusterID,
		"machine_id": machineID,
	}
	if newManagerID != nil {
		body["new_manager_id"] = *newManagerID
	}
	if err := s.client.DeleteWithBody(ctx, "/cluster/remove_machine/", body, nil); err != nil {
		return fmt.Errorf("removing machine %d from cluster %d: %w", machineID, clusterID, err)
	}
	return nil
}
