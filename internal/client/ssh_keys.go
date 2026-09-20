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
	ID        int     `json:"id"`
	SSHKey    string  `json:"ssh_key"`
	CreatedAt float64 `json:"created_at"`
	MachineID int     `json:"machine_id"`
	PublicKey string  `json:"public_key"`
}

// sshKeyCreateResponse wraps the SSH key creation response.
// The API returns {"success": true, "key": {...}}.
type sshKeyCreateResponse struct {
	Key SSHKey `json:"key"`
}

// Create creates a new SSH key.
// Sends POST /ssh/ with {"ssh_key": publicKey}.
// The API returns {"success": true, "key": {...}} envelope.
func (s *SSHKeyService) Create(ctx context.Context, publicKey string) (*SSHKey, error) {
	body, err := openAPIJSONBody(map[string]string{"ssh_key": publicKey})
	if err != nil {
		return nil, fmt.Errorf("creating SSH key: %w", err)
	}
	var result sshKeyCreateResponse
	resp, err := s.client.openAPIClient.CreateSshKeyWithBody(ctx, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("creating SSH key: %w", err)
	}
	return &result.Key, nil
}

// List retrieves all SSH keys for the authenticated user.
// Sends GET /ssh/. The API returns a bare JSON array of SSH key objects.
func (s *SSHKeyService) List(ctx context.Context) ([]SSHKey, error) {
	var result []SSHKey
	resp, err := s.client.openAPIClient.GetSshKeysUser(ctx, nil)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("listing SSH keys: %w", err)
	}
	return result, nil
}

// Update updates an existing SSH key.
// Sends PUT /ssh/{id}/ with {"id": id, "ssh_key": publicKey}.
func (s *SSHKeyService) Update(ctx context.Context, id int, publicKey string) (*SSHKey, error) {
	body, err := openAPIJSONBody(map[string]interface{}{
		"id":      id,
		"ssh_key": publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("updating SSH key %d: %w", id, err)
	}
	var result SSHKey
	resp, err := s.client.openAPIClient.UpdateSshKeyWithBody(ctx, id, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, &result); err != nil {
		return nil, fmt.Errorf("updating SSH key %d: %w", id, err)
	}
	return &result, nil
}

// Delete deletes an SSH key by ID.
// Sends DELETE /ssh/{id}/.
func (s *SSHKeyService) Delete(ctx context.Context, id int) error {
	resp, err := s.client.openAPIClient.DeleteSshKey(ctx, int64(id))
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("deleting SSH key %d: %w", id, err)
	}
	return nil
}

// AttachToInstance attaches an SSH key to an instance.
// Sends POST /instances/{instanceID}/ssh/ with {"ssh_key": publicKey}.
func (s *SSHKeyService) AttachToInstance(ctx context.Context, instanceID int, publicKey string) error {
	body, err := openAPIJSONBody(map[string]string{"ssh_key": publicKey})
	if err != nil {
		return fmt.Errorf("attaching SSH key to instance %d: %w", instanceID, err)
	}
	resp, err := s.client.openAPIClient.AttachSshKeyWithBody(ctx, instanceID, applicationJSON, body)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("attaching SSH key to instance %d: %w", instanceID, err)
	}
	return nil
}

// DetachFromInstance detaches an SSH key from an instance.
// Sends DELETE /instances/{instanceID}/ssh/{sshKeyID}/.
func (s *SSHKeyService) DetachFromInstance(ctx context.Context, instanceID int, sshKeyID int) error {
	resp, err := s.client.openAPIClient.DetachSshKey(ctx, instanceID, sshKeyID)
	if err := s.client.doOpenAPIResponse(ctx, resp, err, nil); err != nil {
		return fmt.Errorf("detaching SSH key %d from instance %d: %w", sshKeyID, instanceID, err)
	}
	return nil
}
