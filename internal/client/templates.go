package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// TemplateService handles template-related API operations.
type TemplateService struct {
	client *VastAIClient
}

// CreateTemplateRequest is the JSON body for POST /template/ (create) and PUT /template/ (update).
type CreateTemplateRequest struct {
	Name                 string                 `json:"name,omitempty"`
	Image                string                 `json:"image,omitempty"`
	Tag                  string                 `json:"tag,omitempty"`
	Env                  string                 `json:"env,omitempty"`
	Onstart              string                 `json:"onstart,omitempty"`
	SSHDirect            bool                   `json:"ssh_direct"`
	JupDirect            bool                   `json:"jup_direct"`
	UseJupyterLab        bool                   `json:"use_jupyter_lab"`
	UseSSH               bool                   `json:"use_ssh"`
	Runtype              string                 `json:"runtype,omitempty"`
	Readme               string                 `json:"readme,omitempty"`
	ReadmeVisible        bool                   `json:"readme_visible"`
	Desc                 string                 `json:"desc,omitempty"`
	Private              bool                   `json:"private"`
	RecommendedDiskSpace float64                `json:"recommended_disk_space,omitempty"`
	DockerLoginRepo      string                 `json:"docker_login_repo,omitempty"`
	Href                 string                 `json:"href,omitempty"`
	Repo                 string                 `json:"repo,omitempty"`
	ExtraFilters         map[string]interface{} `json:"extra_filters,omitempty"`
	JupyterDir           string                 `json:"jupyter_dir,omitempty"`
}

// Template represents a template object from the Vast.ai API.
type Template struct {
	ID                   int     `json:"id"`
	HashID               string  `json:"hash_id"`
	Name                 string  `json:"name"`
	Image                string  `json:"image"`
	Tag                  string  `json:"tag"`
	CreatedAt            string  `json:"created_at"`
	CountCreated         int     `json:"count_created"`
	Env                  string  `json:"env"`
	Onstart              string  `json:"onstart"`
	SSHDirect            bool    `json:"ssh_direct"`
	JupDirect            bool    `json:"jup_direct"`
	UseSSH               bool    `json:"use_ssh"`
	UseJupyterLab        bool    `json:"use_jupyter_lab"`
	Runtype              string  `json:"runtype"`
	Private              bool    `json:"private"`
	Readme               string  `json:"readme"`
	ReadmeVisible        bool    `json:"readme_visible"`
	Desc                 string  `json:"desc"`
	RecommendedDiskSpace float64 `json:"recommended_disk_space"`
	DockerLoginRepo      string  `json:"docker_login_repo"`
	Href                 string  `json:"href"`
	Repo                 string  `json:"repo"`
	CreatorID            int     `json:"creator_id"`
	JupyterDir           string  `json:"jupyter_dir"`
}

// RecommendedDiskSpaceString returns the recommended disk space as a string.
func (t *Template) RecommendedDiskSpaceString() string {
	if t.RecommendedDiskSpace == 0 {
		return ""
	}
	return strconv.FormatFloat(t.RecommendedDiskSpace, 'f', -1, 64)
}

// templateSearchResponse wraps the template search API response.
type templateSearchResponse struct {
	Templates []Template `json:"templates"`
}

// templateMutationResponse wraps the Create/Update template API response.
// API returns {"success": true, "template": {...}}.
type templateMutationResponse struct {
	Success  bool     `json:"success"`
	Template Template `json:"template"`
	Message  string   `json:"msg,omitempty"`
}

// Create creates a new template.
// Sends POST /template/ with template configuration.
// API returns {"success": true, "template": {...}} envelope.
func (s *TemplateService) Create(ctx context.Context, req *CreateTemplateRequest) (*Template, error) {
	var resp templateMutationResponse
	if err := s.client.Post(ctx, "/template/", req, &resp); err != nil {
		return nil, fmt.Errorf("creating template: %w", err)
	}
	return &resp.Template, nil
}

// Update updates an existing template by hash_id.
// Sends PUT /template/ with hash_id included in the body.
func (s *TemplateService) Update(ctx context.Context, hashID string, req *CreateTemplateRequest) (*Template, error) {
	// Build body with hash_id included
	body := map[string]interface{}{
		"hash_id": hashID,
	}

	// Copy all fields from request unconditionally so that false/empty
	// values are sent explicitly (fixes zero-value omission for bools).
	body["name"] = req.Name
	body["image"] = req.Image
	body["tag"] = req.Tag
	body["env"] = req.Env
	body["onstart"] = req.Onstart
	body["ssh_direct"] = req.SSHDirect
	body["jup_direct"] = req.JupDirect
	body["use_jupyter_lab"] = req.UseJupyterLab
	body["use_ssh"] = req.UseSSH
	body["runtype"] = req.Runtype
	body["readme"] = req.Readme
	body["readme_visible"] = req.ReadmeVisible
	body["desc"] = req.Desc
	body["private"] = req.Private
	if req.RecommendedDiskSpace != 0 {
		body["recommended_disk_space"] = req.RecommendedDiskSpace
	}
	body["docker_login_repo"] = req.DockerLoginRepo
	body["href"] = req.Href
	body["repo"] = req.Repo
	body["jupyter_dir"] = req.JupyterDir
	if req.ExtraFilters != nil {
		body["extra_filters"] = req.ExtraFilters
	}

	var resp templateMutationResponse
	if err := s.client.Put(ctx, "/template/", body, &resp); err != nil {
		return nil, fmt.Errorf("updating template %s: %w", hashID, err)
	}
	return &resp.Template, nil
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

// DeleteByID deletes a template by numeric template_id.
// Sends DELETE /template/ with {"template_id": templateID} in the request body.
func (s *TemplateService) DeleteByID(ctx context.Context, templateID int) error {
	body := map[string]int{"template_id": templateID}
	if err := s.client.DeleteWithBody(ctx, "/template/", body, nil); err != nil {
		return fmt.Errorf("deleting template by id %d: %w", templateID, err)
	}
	return nil
}

// Search searches for templates matching the given query.
// Sends GET /template/?select_cols=["*"]&select_filters={query}.
func (s *TemplateService) Search(ctx context.Context, query string) ([]Template, error) {
	path := fmt.Sprintf("/template/?select_cols=%s&select_filters=%s",
		url.QueryEscape(`["*"]`),
		url.QueryEscape(query))
	var resp templateSearchResponse
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("searching templates: %w", err)
	}
	return resp.Templates, nil
}
