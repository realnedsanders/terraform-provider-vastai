package sshkey

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SSHKeyResourceModel describes the resource data model for vastai_ssh_key.
type SSHKeyResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	SSHKey    types.String   `tfsdk:"ssh_key"`
	CreatedAt types.String   `tfsdk:"created_at"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

// SSHKeysDataSourceModel describes the data source data model for vastai_ssh_keys.
type SSHKeysDataSourceModel struct {
	SSHKeys []SSHKeyModel `tfsdk:"ssh_keys"`
}

// SSHKeyModel describes a single SSH key in the data source list (read-only).
type SSHKeyModel struct {
	ID        types.String `tfsdk:"id"`
	SSHKey    types.String `tfsdk:"ssh_key"`
	CreatedAt types.String `tfsdk:"created_at"`
}
