package networkvolume

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// TestNetworkVolumeOffersDataSource_Metadata verifies the data source type name.
func TestNetworkVolumeOffersDataSource_Metadata(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_network_volume_offers" {
		t.Errorf("expected type name vastai_network_volume_offers, got %s", resp.TypeName)
	}
}

// TestNetworkVolumeOffersDataSource_Schema_HasFilterAttributes verifies filter attributes exist as Optional.
func TestNetworkVolumeOffersDataSource_Schema_HasFilterAttributes(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	filterAttrs := []string{
		"disk_space", "max_storage_cost", "inet_up", "inet_down",
		"reliability", "geolocation", "verified", "static_ip",
		"disk_bw", "order_by", "limit", "allocated_storage", "raw_query",
	}

	for _, name := range filterAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing filter attribute: %s", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", name)
		}
	}
}

// TestNetworkVolumeOffersDataSource_Schema_HasResultAttributes verifies offers (List) and most_affordable (Object) exist as Computed.
func TestNetworkVolumeOffersDataSource_Schema_HasResultAttributes(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Verify offers is a computed ListNestedAttribute
	offersAttr, ok := s.Attributes["offers"]
	if !ok {
		t.Fatal("missing offers attribute")
	}
	if !offersAttr.IsComputed() {
		t.Error("offers should be computed")
	}
	_, ok = offersAttr.(schema.ListNestedAttribute)
	if !ok {
		t.Error("offers should be ListNestedAttribute")
	}

	// Verify most_affordable is a computed SingleNestedAttribute
	mostAffordableAttr, ok := s.Attributes["most_affordable"]
	if !ok {
		t.Fatal("missing most_affordable attribute")
	}
	if !mostAffordableAttr.IsComputed() {
		t.Error("most_affordable should be computed")
	}
	_, ok = mostAffordableAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Error("most_affordable should be SingleNestedAttribute")
	}
}

// TestNetworkVolumeOffersDataSource_Schema_Validators verifies validators on constrained attributes.
func TestNetworkVolumeOffersDataSource_Schema_Validators(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Check limit has validators (Between 1, 1000)
	limitAttr, ok := s.Attributes["limit"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("limit is not Int64Attribute")
	}
	if len(limitAttr.Validators) == 0 {
		t.Error("limit should have validators (Between(1, 1000))")
	}

	// Check reliability has validators (Between 0, 1)
	reliabilityAttr, ok := s.Attributes["reliability"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("reliability is not Float64Attribute")
	}
	if len(reliabilityAttr.Validators) == 0 {
		t.Error("reliability should have validators (Between(0, 1))")
	}

	// Check disk_space has validators (AtLeast 0)
	diskSpaceAttr, ok := s.Attributes["disk_space"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("disk_space is not Float64Attribute")
	}
	if len(diskSpaceAttr.Validators) == 0 {
		t.Error("disk_space should have validators (AtLeast(0))")
	}

	// Check max_storage_cost has validators
	maxStorageCostAttr, ok := s.Attributes["max_storage_cost"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("max_storage_cost is not Float64Attribute")
	}
	if len(maxStorageCostAttr.Validators) == 0 {
		t.Error("max_storage_cost should have validators (AtLeast(0))")
	}

	// Check order_by has validators (OneOf)
	orderByAttr, ok := s.Attributes["order_by"].(schema.StringAttribute)
	if !ok {
		t.Fatal("order_by is not StringAttribute")
	}
	if len(orderByAttr.Validators) == 0 {
		t.Error("order_by should have validators (OneOf)")
	}

	// Check disk_bw has validators
	diskBWAttr, ok := s.Attributes["disk_bw"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("disk_bw is not Float64Attribute")
	}
	if len(diskBWAttr.Validators) == 0 {
		t.Error("disk_bw should have validators (AtLeast(0))")
	}
}

// TestNetworkVolumeOffersDataSource_Schema_NetworkSpecificFields verifies
// network-volume-specific fields (nw_disk_min_bw, nw_disk_max_bw, nw_disk_avg_bw, cluster_id)
// exist in the nested offer attributes.
func TestNetworkVolumeOffersDataSource_Schema_NetworkSpecificFields(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	offersAttr, ok := s.Attributes["offers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("offers attribute is not ListNestedAttribute")
	}

	// Network-volume-specific fields that must exist
	networkSpecificFields := []string{
		"nw_disk_min_bw",
		"nw_disk_max_bw",
		"nw_disk_avg_bw",
		"cluster_id",
	}

	for _, name := range networkSpecificFields {
		attr, ok := offersAttr.NestedObject.Attributes[name]
		if !ok {
			t.Errorf("missing network-specific offer attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("network-specific offer attribute %s should be computed", name)
		}
	}

	// Verify nw_disk_* fields are Float64
	for _, name := range []string{"nw_disk_min_bw", "nw_disk_max_bw", "nw_disk_avg_bw"} {
		_, ok := offersAttr.NestedObject.Attributes[name].(schema.Float64Attribute)
		if !ok {
			t.Errorf("attribute %s should be Float64Attribute", name)
		}
	}

	// Verify cluster_id is Int64
	_, ok = offersAttr.NestedObject.Attributes["cluster_id"].(schema.Int64Attribute)
	if !ok {
		t.Error("attribute cluster_id should be Int64Attribute")
	}
}

// TestNetworkVolumeOffersDataSource_Schema_NoLocalVolumeFields verifies that
// local-volume-only fields are NOT present in network volume offer attributes.
func TestNetworkVolumeOffersDataSource_Schema_NoLocalVolumeFields(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	offersAttr, ok := s.Attributes["offers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("offers attribute is not ListNestedAttribute")
	}

	// Local-volume-only fields that should NOT exist on network volume offers
	localOnlyFields := []string{
		"cuda_max_good",
		"cpu_ghz",
		"disk_bw",
		"disk_name",
		"driver_version",
		"machine_id",
	}

	for _, name := range localOnlyFields {
		if _, ok := offersAttr.NestedObject.Attributes[name]; ok {
			t.Errorf("network volume offers should NOT have local-volume-only field: %s", name)
		}
	}
}

// TestNetworkVolumeOffersDataSource_Schema_NestedOfferAttributes verifies the nested offer attributes
// match the expected NetworkVolumeOfferModel fields.
func TestNetworkVolumeOffersDataSource_Schema_NestedOfferAttributes(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	offersAttr, ok := s.Attributes["offers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("offers attribute is not ListNestedAttribute")
	}

	expectedNestedAttrs := []string{
		"id", "disk_space", "storage_cost", "inet_up", "inet_down",
		"reliability", "duration", "verification", "host_id",
		"cluster_id", "geolocation", "nw_disk_min_bw", "nw_disk_max_bw",
		"nw_disk_avg_bw",
	}

	for _, name := range expectedNestedAttrs {
		attr, ok := offersAttr.NestedObject.Attributes[name]
		if !ok {
			t.Errorf("missing nested offer attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("nested offer attribute %s should be computed", name)
		}
	}
}

// TestNetworkVolumeOffersDataSource_Schema_Descriptions verifies all attributes have non-empty descriptions.
func TestNetworkVolumeOffersDataSource_Schema_Descriptions(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	for name, attr := range s.Attributes {
		desc := getDSAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}

	// Verify nested offer attributes have descriptions
	offersAttr, ok := s.Attributes["offers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("offers attribute is not ListNestedAttribute")
	}
	for name, attr := range offersAttr.NestedObject.Attributes {
		desc := getDSAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested offer attribute %s has empty description", name)
		}
	}
}

// TestNetworkVolumeOffersDataSource_Schema_AllocatedStorage verifies allocated_storage attribute
// exists with description mentioning pricing.
func TestNetworkVolumeOffersDataSource_Schema_AllocatedStorage(t *testing.T) {
	ds := NewNetworkVolumeOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	allocStorageAttr, ok := s.Attributes["allocated_storage"]
	if !ok {
		t.Fatal("missing allocated_storage attribute")
	}
	if !allocStorageAttr.IsOptional() {
		t.Error("allocated_storage should be optional")
	}

	// Verify description mentions pricing
	float64Attr, ok := allocStorageAttr.(schema.Float64Attribute)
	if !ok {
		t.Fatal("allocated_storage is not Float64Attribute")
	}
	if float64Attr.Description == "" {
		t.Error("allocated_storage should have a description")
	}

	// Check for pricing-related content in description
	desc := float64Attr.Description
	hasPricingRef := false
	for _, keyword := range []string{"pricing", "cost", "storage_cost"} {
		if containsSubstring(desc, keyword) {
			hasPricingRef = true
			break
		}
	}
	if !hasPricingRef {
		t.Error("allocated_storage description should mention pricing calculations")
	}
}

// getDSAttributeDescription extracts the description from a data source schema attribute.
func getDSAttributeDescription(attr schema.Attribute) string {
	switch a := attr.(type) {
	case schema.StringAttribute:
		return a.Description
	case schema.Int64Attribute:
		return a.Description
	case schema.Float64Attribute:
		return a.Description
	case schema.BoolAttribute:
		return a.Description
	case schema.ListNestedAttribute:
		return a.Description
	case schema.SingleNestedAttribute:
		return a.Description
	default:
		return ""
	}
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

// searchString is a simple substring search.
func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
