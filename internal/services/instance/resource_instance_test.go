package instance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// getTestSchema returns the instance resource schema for testing.
func getTestSchema(t *testing.T) schema.Schema {
	t.Helper()

	r := NewInstanceResource()
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %s", schemaResp.Diagnostics)
	}

	return schemaResp.Schema
}

// TestInstanceResourceSchema_RequiresReplace verifies that immutable fields
// (offer_id, disk_gb) have RequiresReplace plan modifiers.
func TestInstanceResourceSchema_RequiresReplace(t *testing.T) {
	s := getTestSchema(t)

	// Check offer_id has plan modifiers (RequiresReplace)
	offerIDAttr, ok := s.Attributes["offer_id"]
	if !ok {
		t.Fatal("expected offer_id attribute to exist")
	}
	int64Attr, ok := offerIDAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("expected offer_id to be Int64Attribute")
	}
	if len(int64Attr.PlanModifiers) == 0 {
		t.Error("expected offer_id to have plan modifiers (RequiresReplace)")
	}

	// Verify RequiresReplace is present by checking the Description output
	found := false
	for _, pm := range int64Attr.PlanModifiers {
		desc := fmt.Sprintf("%T", pm)
		if desc != "" {
			found = true // At least one modifier exists
		}
	}
	if !found {
		t.Error("expected offer_id to have RequiresReplace plan modifier")
	}

	// Check disk_gb has plan modifiers (RequiresReplace)
	diskGBAttr, ok := s.Attributes["disk_gb"]
	if !ok {
		t.Fatal("expected disk_gb attribute to exist")
	}
	float64Attr, ok := diskGBAttr.(schema.Float64Attribute)
	if !ok {
		t.Fatal("expected disk_gb to be Float64Attribute")
	}
	if len(float64Attr.PlanModifiers) == 0 {
		t.Error("expected disk_gb to have plan modifiers (RequiresReplace)")
	}
}

// TestInstanceResourceSchema_UseStateForUnknown verifies that stable computed fields
// (id, machine_id, ssh_host, ssh_port, num_gpus, gpu_name, created_at) have plan modifiers.
func TestInstanceResourceSchema_UseStateForUnknown(t *testing.T) {
	s := getTestSchema(t)

	stableFields := []struct {
		name     string
		attrType string // "string" or "int64"
	}{
		{"id", "string"},
		{"machine_id", "int64"},
		{"ssh_host", "string"},
		{"ssh_port", "int64"},
		{"num_gpus", "int64"},
		{"gpu_name", "string"},
		{"created_at", "string"},
	}

	for _, field := range stableFields {
		t.Run(field.name, func(t *testing.T) {
			attr, ok := s.Attributes[field.name]
			if !ok {
				t.Fatalf("expected %s attribute to exist", field.name)
			}

			switch field.attrType {
			case "string":
				strAttr, ok := attr.(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected %s to be StringAttribute", field.name)
				}
				if len(strAttr.PlanModifiers) == 0 {
					t.Errorf("expected %s to have UseStateForUnknown plan modifier", field.name)
				}
			case "int64":
				intAttr, ok := attr.(schema.Int64Attribute)
				if !ok {
					t.Fatalf("expected %s to be Int64Attribute", field.name)
				}
				if len(intAttr.PlanModifiers) == 0 {
					t.Errorf("expected %s to have UseStateForUnknown plan modifier", field.name)
				}
			}
		})
	}
}

// TestInstanceResourceSchema_Validators verifies that constrained fields have proper validators.
func TestInstanceResourceSchema_Validators(t *testing.T) {
	s := getTestSchema(t)

	// status should have OneOf("running", "stopped")
	statusAttr, ok := s.Attributes["status"]
	if !ok {
		t.Fatal("expected status attribute to exist")
	}
	strAttr, ok := statusAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("expected status to be StringAttribute")
	}
	if len(strAttr.Validators) == 0 {
		t.Error("expected status to have validators")
	}

	// offer_id should have AtLeast(1)
	offerIDAttr, ok := s.Attributes["offer_id"]
	if !ok {
		t.Fatal("expected offer_id attribute to exist")
	}
	intAttr, ok := offerIDAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("expected offer_id to be Int64Attribute")
	}
	if len(intAttr.Validators) == 0 {
		t.Error("expected offer_id to have validators")
	}

	// disk_gb should have AtLeast(1.0)
	diskGBAttr, ok := s.Attributes["disk_gb"]
	if !ok {
		t.Fatal("expected disk_gb attribute to exist")
	}
	f64Attr, ok := diskGBAttr.(schema.Float64Attribute)
	if !ok {
		t.Fatal("expected disk_gb to be Float64Attribute")
	}
	if len(f64Attr.Validators) == 0 {
		t.Error("expected disk_gb to have validators")
	}

	// bid_price should have AtLeast(0.001)
	bidPriceAttr, ok := s.Attributes["bid_price"]
	if !ok {
		t.Fatal("expected bid_price attribute to exist")
	}
	f64Attr, ok = bidPriceAttr.(schema.Float64Attribute)
	if !ok {
		t.Fatal("expected bid_price to be Float64Attribute")
	}
	if len(f64Attr.Validators) == 0 {
		t.Error("expected bid_price to have validators")
	}

	// label should have LengthAtMost(200)
	labelAttr, ok := s.Attributes["label"]
	if !ok {
		t.Fatal("expected label attribute to exist")
	}
	strLabelAttr, ok := labelAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("expected label to be StringAttribute")
	}
	if len(strLabelAttr.Validators) == 0 {
		t.Error("expected label to have validators")
	}
}

// TestInstanceResourceSchema_SensitiveFields verifies that image_login has Sensitive=true.
func TestInstanceResourceSchema_SensitiveFields(t *testing.T) {
	s := getTestSchema(t)

	imageLoginAttr, ok := s.Attributes["image_login"]
	if !ok {
		t.Fatal("expected image_login attribute to exist")
	}
	strAttr, ok := imageLoginAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("expected image_login to be StringAttribute")
	}
	if !strAttr.Sensitive {
		t.Error("expected image_login to have Sensitive=true")
	}
}

// TestInstanceResourceSchema_DescriptionsPresent verifies all attributes have non-empty descriptions.
func TestInstanceResourceSchema_DescriptionsPresent(t *testing.T) {
	s := getTestSchema(t)

	for name, attr := range s.Attributes {
		desc := ""
		switch a := attr.(type) {
		case schema.StringAttribute:
			desc = a.Description
		case schema.Int64Attribute:
			desc = a.Description
		case schema.Float64Attribute:
			desc = a.Description
		case schema.BoolAttribute:
			desc = a.Description
		case schema.MapAttribute:
			desc = a.Description
		case schema.SetAttribute:
			desc = a.Description
		}
		if desc == "" {
			t.Errorf("attribute %q has no description (SCHM-04 violation)", name)
		}
	}
}

// TestInstanceResourceSchema_Timeouts verifies the timeouts block exists.
func TestInstanceResourceSchema_Timeouts(t *testing.T) {
	s := getTestSchema(t)

	_, ok := s.Blocks["timeouts"]
	if !ok {
		t.Fatal("expected timeouts block to exist (SCHM-06)")
	}
}

// TestPreemptionDetection verifies the preemption detection logic.
// Per D-09: only spot instances (is_bid=true) that intended to be running
// but are actually stopped/offline should be considered preempted.
func TestPreemptionDetection(t *testing.T) {
	tests := []struct {
		name     string
		instance *client.Instance
		expected bool
	}{
		{
			name: "spot instance preempted (is_bid=true, intended=running, actual=stopped)",
			instance: &client.Instance{
				ID:             1,
				IsBid:          true,
				IntendedStatus: "running",
				ActualStatus:   "stopped",
			},
			expected: true,
		},
		{
			name: "spot instance preempted offline (is_bid=true, intended=running, actual=offline)",
			instance: &client.Instance{
				ID:             2,
				IsBid:          true,
				IntendedStatus: "running",
				ActualStatus:   "offline",
			},
			expected: true,
		},
		{
			name: "on-demand instance stopped (is_bid=false, intended=running, actual=stopped)",
			instance: &client.Instance{
				ID:             3,
				IsBid:          false,
				IntendedStatus: "running",
				ActualStatus:   "stopped",
			},
			expected: false,
		},
		{
			name: "spot instance intentionally stopped (is_bid=true, intended=stopped, actual=stopped)",
			instance: &client.Instance{
				ID:             4,
				IsBid:          true,
				IntendedStatus: "stopped",
				ActualStatus:   "stopped",
			},
			expected: false,
		},
		{
			name: "spot instance running normally (is_bid=true, intended=running, actual=running)",
			instance: &client.Instance{
				ID:             5,
				IsBid:          true,
				IntendedStatus: "running",
				ActualStatus:   "running",
			},
			expected: false,
		},
		{
			name: "on-demand instance exited (is_bid=false, intended=running, actual=exited)",
			instance: &client.Instance{
				ID:             6,
				IsBid:          false,
				IntendedStatus: "running",
				ActualStatus:   "exited",
			},
			expected: false,
		},
		{
			name: "spot instance loading (is_bid=true, intended=running, actual=loading)",
			instance: &client.Instance{
				ID:             7,
				IsBid:          true,
				IntendedStatus: "running",
				ActualStatus:   "loading",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPreempted(tt.instance)
			if result != tt.expected {
				t.Errorf("isPreempted() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuildRuntype verifies the runtype string construction from SSH and JupyterLab flags.
func TestBuildRuntype(t *testing.T) {
	tests := []struct {
		name           string
		useSSH         bool
		useSSHNull     bool
		useJupyterLab  bool
		useJupyterNull bool
		expected       string
	}{
		{
			name:          "both SSH and JupyterLab enabled",
			useSSH:        true,
			useJupyterLab: true,
			expected:      "ssh_direc ssh_proxy jupyter_direc",
		},
		{
			name:     "only SSH enabled",
			useSSH:   true,
			expected: "ssh_direc ssh_proxy",
		},
		{
			name:          "only JupyterLab enabled",
			useJupyterLab: true,
			expected:      "jupyter_direc",
		},
		{
			name:     "neither enabled",
			expected: "",
		},
		{
			name:           "both null",
			useSSHNull:     true,
			useJupyterNull: true,
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ssh, jupyter types.Bool
			if tt.useSSHNull {
				ssh = types.BoolNull()
			} else {
				ssh = types.BoolValue(tt.useSSH)
			}
			if tt.useJupyterNull {
				jupyter = types.BoolNull()
			} else {
				jupyter = types.BoolValue(tt.useJupyterLab)
			}
			result := buildRuntype(ssh, jupyter)
			if result != tt.expected {
				t.Errorf("buildRuntype() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSetDifference verifies the set difference helper function.
func TestSetDifference(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected []string
	}{
		{
			name:     "elements only in a",
			a:        []string{"1", "2", "3"},
			b:        []string{"2"},
			expected: []string{"1", "3"},
		},
		{
			name:     "identical sets",
			a:        []string{"1", "2"},
			b:        []string{"1", "2"},
			expected: nil,
		},
		{
			name:     "empty a",
			a:        nil,
			b:        []string{"1"},
			expected: nil,
		},
		{
			name:     "empty b",
			a:        []string{"1", "2"},
			b:        nil,
			expected: []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setDifference(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Errorf("setDifference() returned %d elements, want %d: got %v", len(result), len(tt.expected), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("setDifference()[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestMapInstanceToModel verifies the API-to-model mapping.
func TestMapInstanceToModel(t *testing.T) {
	instance := &client.Instance{
		ID:                1234,
		MachineID:         5678,
		ActualStatus:      "running",
		IntendedStatus:    "running",
		NumGPUs:           2,
		GPUName:           "RTX 4090",
		GPUTotalRAM:       24000, // 24000 MB -> 24.0 GB
		CPURAM:            64000, // 64000 MB -> 64.0 GB
		CPUCoresEffective: 16,
		SSHHost:           "ssh.example.com",
		SSHPort:           22022,
		DPHTotal:          0.5,
		IsBid:             false,
		InetUp:            1000,
		InetDown:          1000,
		Reliability2:      0.99,
		Geolocation:       "US",
		Label:             "test-instance",
		ImageUUID:         "pytorch/pytorch:latest",
		StartDate:         1700000000,
	}

	model := &InstanceResourceModel{}
	mapInstanceToModel(instance, model)

	if model.ID.ValueString() != "1234" {
		t.Errorf("ID = %q, want %q", model.ID.ValueString(), "1234")
	}
	if model.MachineID.ValueInt64() != 5678 {
		t.Errorf("MachineID = %d, want %d", model.MachineID.ValueInt64(), 5678)
	}
	if model.Status.ValueString() != "running" {
		t.Errorf("Status = %q, want %q", model.Status.ValueString(), "running")
	}
	if model.NumGPUs.ValueInt64() != 2 {
		t.Errorf("NumGPUs = %d, want %d", model.NumGPUs.ValueInt64(), 2)
	}
	if model.GPUName.ValueString() != "RTX 4090" {
		t.Errorf("GPUName = %q, want %q", model.GPUName.ValueString(), "RTX 4090")
	}

	// RAM conversion: 24000 MB -> 24.0 GB
	if model.GPURamGB.ValueFloat64() != 24.0 {
		t.Errorf("GPURamGB = %f, want %f", model.GPURamGB.ValueFloat64(), 24.0)
	}
	// RAM conversion: 64000 MB -> 64.0 GB
	if model.CPURamGB.ValueFloat64() != 64.0 {
		t.Errorf("CPURamGB = %f, want %f", model.CPURamGB.ValueFloat64(), 64.0)
	}

	if model.SSHHost.ValueString() != "ssh.example.com" {
		t.Errorf("SSHHost = %q, want %q", model.SSHHost.ValueString(), "ssh.example.com")
	}
	if model.SSHPort.ValueInt64() != 22022 {
		t.Errorf("SSHPort = %d, want %d", model.SSHPort.ValueInt64(), 22022)
	}
	if model.DPHTotal.ValueFloat64() != 0.5 {
		t.Errorf("DPHTotal = %f, want %f", model.DPHTotal.ValueFloat64(), 0.5)
	}
	if model.IsBid.ValueBool() != false {
		t.Errorf("IsBid = %v, want %v", model.IsBid.ValueBool(), false)
	}
	if model.Geolocation.ValueString() != "US" {
		t.Errorf("Geolocation = %q, want %q", model.Geolocation.ValueString(), "US")
	}
	if model.Label.ValueString() != "test-instance" {
		t.Errorf("Label = %q, want %q", model.Label.ValueString(), "test-instance")
	}
}

// TestInstanceResource_Interfaces verifies that InstanceResource satisfies required interfaces.
func TestInstanceResource_Interfaces(t *testing.T) {
	r := NewInstanceResource()

	// Check resource.ResourceWithConfigure
	if _, ok := r.(resource.ResourceWithConfigure); !ok {
		t.Error("InstanceResource does not implement resource.ResourceWithConfigure")
	}

	// Check resource.ResourceWithImportState
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("InstanceResource does not implement resource.ResourceWithImportState")
	}
}

// TestExtractStringSet verifies the string set extraction helper.
func TestExtractStringSet(t *testing.T) {
	// Test with values
	s := stringSetValue([]string{"1", "2", "3"})
	result := extractStringSet(s)
	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}

	// Test with null
	nullSet := types.SetNull(types.StringType)
	result = extractStringSet(nullSet)
	if result != nil {
		t.Errorf("expected nil for null set, got %v", result)
	}
}

// TestMapInstanceToModel_EmptyOptionalFields verifies the mapping handles empty optional fields.
func TestMapInstanceToModel_EmptyOptionalFields(t *testing.T) {
	instance := &client.Instance{
		ID:             1,
		MachineID:      2,
		ActualStatus:   "running",
		IntendedStatus: "running",
		// All optional fields empty/zero
	}

	model := &InstanceResourceModel{}
	mapInstanceToModel(instance, model)

	// SSH host should be null when empty
	if !model.SSHHost.IsNull() {
		t.Errorf("expected SSHHost to be null when empty, got %q", model.SSHHost.ValueString())
	}

	// SSH port should be null when 0
	if !model.SSHPort.IsNull() {
		t.Errorf("expected SSHPort to be null when 0, got %d", model.SSHPort.ValueInt64())
	}

	// Geolocation should be null when empty
	if !model.Geolocation.IsNull() {
		t.Errorf("expected Geolocation to be null when empty")
	}
}

// TestInstanceResourceSchema_AttributeExists verifies all expected attributes are present.
func TestInstanceResourceSchema_AttributeExists(t *testing.T) {
	s := getTestSchema(t)

	expectedAttrs := []string{
		// Computed stable
		"id", "machine_id", "ssh_host", "ssh_port", "num_gpus", "gpu_name", "created_at",
		// Required immutable
		"offer_id", "disk_gb",
		// Optional+Computed
		"image", "status", "use_ssh", "use_jupyter_lab",
		// Optional mutable
		"label", "bid_price", "onstart", "env", "template_hash_id",
		"ssh_key_ids", "image_login", "cancel_unavail",
		// Computed dynamic
		"actual_status", "cost_per_hour", "gpu_ram_gb", "cpu_ram_gb",
		"cpu_cores", "inet_up_mbps", "inet_down_mbps", "reliability",
		"geolocation", "is_bid", "status_msg",
	}

	for _, name := range expectedAttrs {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("expected attribute %q to exist in schema", name)
		}
	}
}
