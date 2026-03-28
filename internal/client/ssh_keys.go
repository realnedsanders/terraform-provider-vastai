package client

import (
	"context"
	"fmt"
)

// SSHKeyService handles SSH key-related API operations.
type SSHKeyService struct {
	client *VastAIClient
}

// SSHKey represents an SSH key object from the Vast.ai API.
type SSHKey struct {
	ID        int    `json:"id"`
	SSHKey    string `json:"ssh_key"`
	CreatedAt string `json:"created_at"`
	MachineID int    `json:"machine_id"`
	PublicKey string `json:"public_key"`
}

// Create creates a new SSH key.
// Sends POST /ssh/ with {"ssh_key": publicKey}.
func (s *SSHKeyService) Create(ctx context.Context, publicKey string) (*SSHKey, error) {
	body := map[string]string{"ssh_key": publicKey}
	var resp SSHKey
	if err := s.client.Post(ctx, "/ssh/", body, &resp); err != nil {
		return nil, fmt.Errorf("creating SSH key: %w", err)
	}
	return &resp, nil
}

// List retrieves all SSH keys for the authenticated user.
// Sends GET /ssh/. The API returns a bare JSON array of SSH key objects.
func (s *SSHKeyService) List(ctx context.Context) ([]SSHKey, error) {
	var resp []SSHKey
	if err := s.client.Get(ctx, "/ssh/", &resp); err != nil {
		return nil, fmt.Errorf("listing SSH keys: %w", err)
	}
	return resp, nil
}

// Update updates an existing SSH key.
// Sends PUT /ssh/{id}/ with {"id": id, "ssh_key": publicKey}.
func (s *SSHKeyService) Update(ctx context.Context, id int, publicKey string) (*SSHKey, error) {
	path := fmt.Sprintf("/ssh/%d/", id)
	body := map[string]interface{}{
		"id":      id,
		"ssh_key": publicKey,
	}
	var resp SSHKey
	if err := s.client.Put(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("updating SSH key %d: %w", id, err)
	}
	return &resp, nil
}

// Delete deletes an SSH key by ID.
// Sends DELETE /ssh/{id}/.
func (s *SSHKeyService) Delete(ctx context.Context, id int) error {
	path := fmt.Sprintf("/ssh/%d/", id)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting SSH key %d: %w", id, err)
	}
	return nil
}

// AttachToInstance attaches an SSH key to an instance.
// Sends POST /instances/{instanceID}/ssh/ with {"ssh_key": publicKey}.
func (s *SSHKeyService) AttachToInstance(ctx context.Context, instanceID int, publicKey string) error {
	path := fmt.Sprintf("/instances/%d/ssh/", instanceID)
	body := map[string]string{"ssh_key": publicKey}
	if err := s.client.Post(ctx, path, body, nil); err != nil {
		return fmt.Errorf("attaching SSH key to instance %d: %w", instanceID, err)
	}
	return nil
}

// DetachFromInstance detaches an SSH key from an instance.
// Sends DELETE /instances/{instanceID}/ssh/{sshKeyID}/.
func (s *SSHKeyService) DetachFromInstance(ctx context.Context, instanceID int, sshKeyID int) error {
	path := fmt.Sprintf("/instances/%d/ssh/%d/", instanceID, sshKeyID)
	if err := s.client.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("detaching SSH key %d from instance %d: %w", sshKeyID, instanceID, err)
	}
	return nil
}
