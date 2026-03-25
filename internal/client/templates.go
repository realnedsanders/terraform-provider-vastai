package client

import (
	"context"
	"fmt"
	"net/url"
)

// TemplateService handles template-related API operations.
type TemplateService struct {
	client *VastAIClient
}

// CreateTemplateRequest is the JSON body for POST /template/ (create) and PUT /template/ (update).
type CreateTemplateRequest struct {
	Name                  string                 `json:"name,omitempty"`
	Image                 string                 `json:"image,omitempty"`
	Tag                   string                 `json:"tag,omitempty"`
	Env                   string                 `json:"env,omitempty"`
	Onstart               string                 `json:"onstart,omitempty"`
	SSHDirect             bool                   `json:"ssh_direct,omitempty"`
	JupDirect             bool                   `json:"jup_direct,omitempty"`
	UseJupyterLab         bool                   `json:"use_jupyter_lab,omitempty"`
	UseSSH                bool                   `json:"use_ssh,omitempty"`
	Runtype               string                 `json:"runtype,omitempty"`
	Readme                string                 `json:"readme,omitempty"`
	ReadmeVisible         bool                   `json:"readme_visible,omitempty"`
	Desc                  string                 `json:"desc,omitempty"`
	Private               bool                   `json:"private,omitempty"`
	RecommendedDiskSpace  string                 `json:"recommended_disk_space,omitempty"`
	DockerLoginRepo       string                 `json:"docker_login_repo,omitempty"`
	Href                  string                 `json:"href,omitempty"`
	Repo                  string                 `json:"repo,omitempty"`
	ExtraFilters          map[string]interface{} `json:"extra_filters,omitempty"`
}

// Template represents a template object from the Vast.ai API.
type Template struct {
	ID                    int    `json:"id"`
	HashID                string `json:"hash_id"`
	Name                  string `json:"name"`
	Image                 string `json:"image"`
	Tag                   string `json:"tag"`
	CreatedAt             string `json:"created_at"`
	CountCreated          int    `json:"count_created"`
	Env                   string `json:"env"`
	Onstart               string `json:"onstart"`
	SSHDirect             bool   `json:"ssh_direct"`
	JupDirect             bool   `json:"jup_direct"`
	UseSSH                bool   `json:"use_ssh"`
	UseJupyterLab         bool   `json:"use_jupyter_lab"`
	Runtype               string `json:"runtype"`
	Private               bool   `json:"private"`
	Readme                string `json:"readme"`
	ReadmeVisible         bool   `json:"readme_visible"`
	Desc                  string `json:"desc"`
	RecommendedDiskSpace  string `json:"recommended_disk_space"`
	DockerLoginRepo       string `json:"docker_login_repo"`
	Href                  string `json:"href"`
	Repo                  string `json:"repo"`
	CreatorID             int    `json:"creator_id"`
}

// templateSearchResponse wraps the template search API response.
type templateSearchResponse struct {
	Templates []Template `json:"templates"`
}

// Create creates a new template.
// Sends POST /template/ with template configuration.
func (s *TemplateService) Create(ctx context.Context, req *CreateTemplateRequest) (*Template, error) {
	var resp Template
	if err := s.client.Post(ctx, "/template/", req, &resp); err != nil {
		return nil, fmt.Errorf("creating template: %w", err)
	}
	return &resp, nil
}

// Update updates an existing template by hash_id.
// Sends PUT /template/ with hash_id included in the body.
func (s *TemplateService) Update(ctx context.Context, hashID string, req *CreateTemplateRequest) (*Template, error) {
	// Build body with hash_id included
	body := map[string]interface{}{
		"hash_id": hashID,
	}

	// Copy non-zero fields from request
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Image != "" {
		body["image"] = req.Image
	}
	if req.Tag != "" {
		body["tag"] = req.Tag
	}
	if req.Env != "" {
		body["env"] = req.Env
	}
	if req.Onstart != "" {
		body["onstart"] = req.Onstart
	}
	if req.SSHDirect {
		body["ssh_direct"] = req.SSHDirect
	}
	if req.JupDirect {
		body["jup_direct"] = req.JupDirect
	}
	if req.UseJupyterLab {
		body["use_jupyter_lab"] = req.UseJupyterLab
	}
	if req.UseSSH {
		body["use_ssh"] = req.UseSSH
	}
	if req.Runtype != "" {
		body["runtype"] = req.Runtype
	}
	if req.Readme != "" {
		body["readme"] = req.Readme
	}
	if req.ReadmeVisible {
		body["readme_visible"] = req.ReadmeVisible
	}
	if req.Desc != "" {
		body["desc"] = req.Desc
	}
	if req.Private {
		body["private"] = req.Private
	}
	if req.RecommendedDiskSpace != "" {
		body["recommended_disk_space"] = req.RecommendedDiskSpace
	}
	if req.DockerLoginRepo != "" {
		body["docker_login_repo"] = req.DockerLoginRepo
	}
	if req.Href != "" {
		body["href"] = req.Href
	}
	if req.Repo != "" {
		body["repo"] = req.Repo
	}
	if req.ExtraFilters != nil {
		body["extra_filters"] = req.ExtraFilters
	}

	var resp Template
	if err := s.client.Put(ctx, "/template/", body, &resp); err != nil {
		return nil, fmt.Errorf("updating template %s: %w", hashID, err)
	}
	return &resp, nil
}

// Delete deletes a template by hash_id.
// Sends DELETE /template/ with {"hash_id": hashID} in the request body (Pitfall 5).
func (s *TemplateService) Delete(ctx context.Context, hashID string) error {
	body := map[string]string{"hash_id": hashID}
	if err := s.client.DeleteWithBody(ctx, "/template/", body, nil); err != nil {
		return fmt.Errorf("deleting template %s: %w", hashID, err)
	}
	return nil
}

// Search searches for templates matching the given query.
// Sends GET /template/?select_cols=[*]&select_filters={query}.
func (s *TemplateService) Search(ctx context.Context, query string) ([]Template, error) {
	path := fmt.Sprintf("/template/?select_cols=[*]&select_filters=%s", url.QueryEscape(query))
	var resp templateSearchResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("searching templates: %w", err)
	}
	return resp.Templates, nil
}
