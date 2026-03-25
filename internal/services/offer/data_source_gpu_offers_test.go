package offer

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/realnedsanders/terraform-provider-vastai/internal/client"
)

// TestGpuOffersDataSourceMetadata verifies the data source type name.
func TestGpuOffersDataSourceMetadata(t *testing.T) {
	ds := NewGpuOffersDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "vastai",
	}
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "vastai_gpu_offers" {
		t.Errorf("expected type name vastai_gpu_offers, got %s", resp.TypeName)
	}
}

// TestGpuOffersDataSourceSchema verifies the schema has all expected attributes
// with correct types, descriptions, and validators.
func TestGpuOffersDataSourceSchema(t *testing.T) {
	ds := NewGpuOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("schema returned errors: %v", resp.Diagnostics)
	}

	s := resp.Schema

	// Verify filter attributes exist and have descriptions
	filterAttrs := []struct {
		name     string
		optional bool
		computed bool
	}{
		{"gpu_name", true, false},
		{"num_gpus", true, false},
		{"gpu_ram_gb", true, false},
		{"max_price_per_hour", true, false},
		{"datacenter_only", true, false},
		{"region", true, false},
		{"offer_type", true, false},
		{"order_by", true, true},
		{"limit", true, true},
		{"raw_query", true, false},
	}

	for _, fa := range filterAttrs {
		attr, ok := s.Attributes[fa.name]
		if !ok {
			t.Errorf("missing filter attribute: %s", fa.name)
			continue
		}

		if fa.optional && !attr.IsOptional() {
			t.Errorf("attribute %s should be optional", fa.name)
		}
		if fa.computed && !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", fa.name)
		}
	}

	// Verify result attributes exist
	resultAttrs := []string{"offers", "most_affordable"}
	for _, name := range resultAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("missing result attribute: %s", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("attribute %s should be computed", name)
		}
	}

	// Verify all attributes have non-empty descriptions
	for name, attr := range s.Attributes {
		desc := getAttributeDescription(attr)
		if desc == "" {
			t.Errorf("attribute %s has empty description", name)
		}
	}

	// Verify offers nested attributes have descriptions
	offersAttr, ok := s.Attributes["offers"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("offers attribute is not ListNestedAttribute")
	}
	for name, attr := range offersAttr.NestedObject.Attributes {
		desc := getAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested offer attribute %s has empty description", name)
		}
	}

	// Verify most_affordable nested attributes have descriptions
	mostAffordableAttr, ok := s.Attributes["most_affordable"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("most_affordable attribute is not SingleNestedAttribute")
	}
	for name, attr := range mostAffordableAttr.Attributes {
		desc := getAttributeDescription(attr)
		if desc == "" {
			t.Errorf("nested most_affordable attribute %s has empty description", name)
		}
	}
}

// TestGpuOffersDataSourceSchemaValidators verifies that validators are present
// on constrained filter attributes.
func TestGpuOffersDataSourceSchemaValidators(t *testing.T) {
	ds := NewGpuOffersDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), req, &resp)

	s := resp.Schema

	// Check that num_gpus has validators
	numGPUs, ok := s.Attributes["num_gpus"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("num_gpus is not Int64Attribute")
	}
	if len(numGPUs.Validators) == 0 {
		t.Error("num_gpus should have validators (Between(1, 16))")
	}

	// Check that gpu_ram_gb has validators
	gpuRam, ok := s.Attributes["gpu_ram_gb"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("gpu_ram_gb is not Float64Attribute")
	}
	if len(gpuRam.Validators) == 0 {
		t.Error("gpu_ram_gb should have validators (AtLeast(1.0))")
	}

	// Check that max_price_per_hour has validators
	maxPrice, ok := s.Attributes["max_price_per_hour"].(schema.Float64Attribute)
	if !ok {
		t.Fatal("max_price_per_hour is not Float64Attribute")
	}
	if len(maxPrice.Validators) == 0 {
		t.Error("max_price_per_hour should have validators (AtLeast(0.001))")
	}

	// Check that offer_type has validators (OneOf)
	offerType, ok := s.Attributes["offer_type"].(schema.StringAttribute)
	if !ok {
		t.Fatal("offer_type is not StringAttribute")
	}
	if len(offerType.Validators) == 0 {
		t.Error("offer_type should have validators (OneOf)")
	}

	// Check that limit has validators
	limitAttr, ok := s.Attributes["limit"].(schema.Int64Attribute)
	if !ok {
		t.Fatal("limit is not Int64Attribute")
	}
	if len(limitAttr.Validators) == 0 {
		t.Error("limit should have validators (Between(1, 1000))")
	}

	// Check that gpu_name has validators
	gpuName, ok := s.Attributes["gpu_name"].(schema.StringAttribute)
	if !ok {
		t.Fatal("gpu_name is not StringAttribute")
	}
	if len(gpuName.Validators) == 0 {
		t.Error("gpu_name should have validators (LengthAtLeast(1))")
	}
}

// TestApiOfferToModel verifies the conversion from API Offer to Terraform model.
func TestApiOfferToModel(t *testing.T) {
	apiOffer := struct {
		ID                int
		MachineID         int
		GPUName           string
		NumGPUs           int
		GPURAM            float64
		GPUTotalRAM       float64
		CPUCoresEffective float64
		CPURAM            float64
		DiskSpace         float64
		DPHTotal          float64
		DLPerf            float64
		InetUp            float64
		InetDown          float64
		Reliability       float64
		Geolocation       string
		HostingType       int
		Verification      string
		StaticIP          bool
		DirectPortCount   int
		CUDAMaxGood       float64
		MinBid            float64
		StorageCost       float64
	}{
		ID: 12345, MachineID: 67890, GPUName: "RTX 4090", NumGPUs: 2,
		GPURAM: 24000.0, GPUTotalRAM: 48000.0, CPUCoresEffective: 16.0,
		CPURAM: 64000.0, DiskSpace: 512.0, DPHTotal: 0.55,
		DLPerf: 42.5, InetUp: 500.0, InetDown: 1000.0,
		Reliability: 0.98, Geolocation: "US", HostingType: 1,
		Verification: "verified", StaticIP: true, DirectPortCount: 3,
		CUDAMaxGood: 12.4, MinBid: 0.25, StorageCost: 0.05,
	}

	// Use the client.Offer type directly
	from := clientOfferFromTest(apiOffer.ID, apiOffer.MachineID, apiOffer.GPUName,
		apiOffer.NumGPUs, apiOffer.GPURAM, apiOffer.GPUTotalRAM,
		apiOffer.CPUCoresEffective, apiOffer.CPURAM, apiOffer.DiskSpace,
		apiOffer.DPHTotal, apiOffer.DLPerf, apiOffer.InetUp, apiOffer.InetDown,
		apiOffer.Reliability, apiOffer.Geolocation, apiOffer.HostingType,
		apiOffer.Verification, apiOffer.StaticIP, apiOffer.DirectPortCount,
		apiOffer.CUDAMaxGood, apiOffer.MinBid, apiOffer.StorageCost)

	model := apiOfferToModel(from)

	// Verify MB to GB conversion
	expectedGPURamGB := 24000.0 / 1000.0
	if model.GPURamGB.ValueFloat64() != expectedGPURamGB {
		t.Errorf("expected gpu_ram_gb %f, got %f", expectedGPURamGB, model.GPURamGB.ValueFloat64())
	}

	expectedCPURamGB := 64000.0 / 1000.0
	if model.CPURamGB.ValueFloat64() != expectedCPURamGB {
		t.Errorf("expected cpu_ram_gb %f, got %f", expectedCPURamGB, model.CPURamGB.ValueFloat64())
	}

	// Verify datacenter_hosted derived from hosting_type
	if !model.DatacenterHosted.ValueBool() {
		t.Error("expected datacenter_hosted to be true for hosting_type=1")
	}

	// Verify direct field mappings
	if model.PricePerHour.ValueFloat64() != 0.55 {
		t.Errorf("expected price_per_hour 0.55, got %f", model.PricePerHour.ValueFloat64())
	}
	if model.GPUName.ValueString() != "RTX 4090" {
		t.Errorf("expected gpu_name 'RTX 4090', got %s", model.GPUName.ValueString())
	}
}

// TestApiOfferToModelConsumerHosting verifies datacenter_hosted is false for hosting_type=0.
func TestApiOfferToModelConsumerHosting(t *testing.T) {
	from := clientOfferFromTest(1, 1, "RTX 3090", 1, 24000, 24000, 8, 32000, 256,
		0.3, 20, 100, 200, 0.95, "EU", 0, "verified", false, 1, 11.8, 0.15, 0.03)

	model := apiOfferToModel(from)
	if model.DatacenterHosted.ValueBool() {
		t.Error("expected datacenter_hosted to be false for hosting_type=0")
	}
}

// getAttributeDescription extracts the description from a schema attribute.
func getAttributeDescription(attr schema.Attribute) string {
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

// clientOfferFromTest creates a client.Offer for testing purposes.
func clientOfferFromTest(
	id, machineID int, gpuName string, numGPUs int,
	gpuRAM, gpuTotalRAM, cpuCores, cpuRAM, diskSpace, dphTotal, dlPerf,
	inetUp, inetDown, reliability float64,
	geolocation string, hostingType int, verification string,
	staticIP bool, directPortCount int,
	cudaMaxGood, minBid, storageCost float64,
) clientOffer {
	return clientOffer{
		ID: id, MachineID: machineID, GPUName: gpuName, NumGPUs: numGPUs,
		GPURAM: gpuRAM, GPUTotalRAM: gpuTotalRAM, CPUCoresEffective: cpuCores,
		CPURAM: cpuRAM, DiskSpace: diskSpace, DPHTotal: dphTotal, DLPerf: dlPerf,
		InetUp: inetUp, InetDown: inetDown, Reliability: reliability,
		Geolocation: geolocation, HostingType: hostingType, Verification: verification,
		StaticIP: staticIP, DirectPortCount: directPortCount,
		CUDAMaxGood: cudaMaxGood, MinBid: minBid, StorageCost: storageCost,
	}
}

// clientOffer is a test-local mirror of client.Offer to avoid import cycles.
// We use apiOfferToModel which takes client.Offer, so we need a type alias approach.
type clientOffer = client.Offer
