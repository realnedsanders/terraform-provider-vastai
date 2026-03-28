package client

import (
	"context"
	"fmt"
)

// OverlayService handles overlay network-related API operations.
type OverlayService struct {
	client *VastAIClient
}

// Overlay represents an overlay network object from the Vast.ai API.
type Overlay struct {
	OverlayID      int    `json:"overlay_id"`
	Name           string `json:"name"`
	InternalSubnet string `json:"internal_subnet"`
	ClusterID      int    `json:"cluster_id"`
	Instances      []int  `json:"instances"`
}

// Create creates a new overlay network on the given cluster.
// Sends POST /overlay/ with {"cluster_id": clusterID, "name": name}.
// Uses create-then-read pattern: the create response returns only a message,
// so we list overlays and find the one matching the name (Pitfall 4).
func (s *OverlayService) Create(ctx context.Context, clusterID int, name string) (*Overlay, error) {
	body := map[string]interface{}{
		"cluster_id": clusterID,
		"name":       name,
	}
	if err := s.client.Post(ctx, "/overlay/", body, nil); err != nil {
		return nil, fmt.Errorf("creating overlay: %w", err)
	}

	// Create-then-read: list all overlays and find the newly created one by name
	overlays, err := s.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading overlay after create: %w", err)
	}

	// Find by name, iterating backwards for the most recently created match
	for i := len(overlays) - 1; i >= 0; i-- {
		if overlays[i].Name == name {
			return &overlays[i], nil
		}
	}

	return nil, fmt.Errorf("overlay %q not found after creation", name)
}

// List retrieves all overlay networks.
// Sends GET /overlay/.
func (s *OverlayService) List(ctx context.Context) ([]Overlay, error) {
	var resp []Overlay
	if err := s.client.Get(ctx, "/overlay/", &resp); err != nil {
		return nil, fmt.Errorf("listing overlays: %w", err)
	}
	return resp, nil
}

// Delete deletes an overlay network by ID.
// Sends DELETE /overlay/ with {"overlay_id": overlayID} in the request body.
func (s *OverlayService) Delete(ctx context.Context, overlayID int) error {
	body := map[string]interface{}{
		"overlay_id": overlayID,
	}
	if err := s.client.DeleteWithBody(ctx, "/overlay/", body, nil); err != nil {
		return fmt.Errorf("deleting overlay %d: %w", overlayID, err)
	}
	return nil
}

// JoinInstance adds an instance to an overlay network.
// Sends PUT /overlay/ with {"name": name, "instance_id": instanceID}.
func (s *OverlayService) JoinInstance(ctx context.Context, name string, instanceID int) error {
	body := map[string]interface{}{
		"name":        name,
		"instance_id": instanceID,
	}
	if err := s.client.Put(ctx, "/overlay/", body, nil); err != nil {
		return fmt.Errorf("joining instance %d to overlay %q: %w", instanceID, name, err)
	}
	return nil
}
