package csilvm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

// IdentityService RPCs

func TestGetPluginInfoMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetPluginInfoRequest()
	req.Version = nil
	resp, err := client.GetPluginInfo(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestGetPluginInfoUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetPluginInfoRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.GetPluginInfo(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

// ControllerService RPCs

func TestControllerProbeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testControllerProbeRequest()
	req.Version = nil
	resp, err := client.ControllerProbe(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestControllerProbeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testControllerProbeRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.ControllerProbe(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestCreateVolumeRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testCreateVolumeRequest()
	_, err := client.CreateVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestCreateVolumeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.Version = nil
	resp, err := client.CreateVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestCreateVolumeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.CreateVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestCreateVolumeMissingName(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.Name = ""
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The name field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestCreateVolumeMissingVolumeCapabilities(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.VolumeCapabilities = nil
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The volume_capabilities field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestCreateVolumeEmptyVolumeCapabilities(t *testing.T) {
	t.Skip("gRPC apparently unmarshals an empty list as nil.")
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.VolumeCapabilities = req.VolumeCapabilities[:0]
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "One or more volume_capabilities must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestCreateVolumeMissingVolumeCapabilitiesAccessType(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.VolumeCapabilities[0].AccessType = nil
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The volume_capability.access_type field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestCreateVolumeMissingVolumeCapabilitiesAccessMode(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.VolumeCapabilities[0].AccessMode = nil
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The volume_capability.access_mode field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestCreateVolumeVolumeCapabilitiesAccessModeUNKNOWN(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testCreateVolumeRequest()
	req.VolumeCapabilities[0].AccessMode.Mode = csi.VolumeCapability_AccessMode_UNKNOWN
	resp, err := client.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The volume_capability.access_mode.mode field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestDeleteVolumeRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testDeleteVolumeRequest("test-volume")
	resp, err := client.DeleteVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestDeleteVolumeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testDeleteVolumeRequest("test-volume")
	req.Version = nil
	resp, err := client.DeleteVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestDeleteVolumeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testDeleteVolumeRequest("test-volume")
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.DeleteVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestDeleteVolumeMissingVolumeId(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testDeleteVolumeRequest("test-volume")
	req.VolumeId = nil
	resp, err := client.DeleteVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeId) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.Version = nil
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesMissingVolumeId(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.VolumeId = ""
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeId) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesMissingVolumeCapabilities(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.VolumeCapabilities = nil
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeCapabilities) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesMissingVolumeCapabilitiesAccessType(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.VolumeCapabilities[0].AccessType = nil
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessType) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesNodeUnpublishVolume_MountVolume_BadFilesystem(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "ext4", nil)
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedFilesystem) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesMissingVolumeCapabilitiesAccessMode(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.VolumeCapabilities[0].AccessMode = nil
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessMode) {
		t.Fatal(err)
	}
}

func TestValidateVolumeCapabilitiesVolumeCapabilitiesAccessModeUNKNOWN(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testValidateVolumeCapabilitiesRequest("fake_volume_id", "", nil)
	req.VolumeCapabilities[0].AccessMode.Mode = csi.VolumeCapability_AccessMode_UNKNOWN
	resp, err := client.ValidateVolumeCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessModeMode) {
		t.Fatal(err)
	}
}

func TestListVolumesRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testListVolumesRequest()
	resp, err := client.ListVolumes(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestListVolumesMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testListVolumesRequest()
	req.Version = nil
	resp, err := client.ListVolumes(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestListVolumesUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testListVolumesRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.ListVolumes(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestGetCapacityMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("xfs")
	req.Version = nil
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestGetCapacityUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("xfs")
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestGetCapacityMissingVolumeCapabilitiesAccessType(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("xfs")
	req.VolumeCapabilities[0].AccessType = nil
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessType) {
		t.Fatal(err)
	}
}

func TestGetCapacity_BadFilesystem(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("ext4")
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedFilesystem) {
		t.Fatal(err)
	}
}

func TestGetCapacityMissingVolumeCapabilitiesAccessMode(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("xfs")
	req.VolumeCapabilities[0].AccessMode = nil
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessMode) {
		t.Fatal(err)
	}
}

func TestGetCapacityVolumeCapabilitiesAccessModeUNKNOWN(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetCapacityRequest("xfs")
	req.VolumeCapabilities[0].AccessMode.Mode = csi.VolumeCapability_AccessMode_UNKNOWN
	resp, err := client.GetCapacity(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessModeMode) {
		t.Fatal(err)
	}
}

func TestControllerGetCapabilitiesRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := &csi.ControllerGetCapabilitiesRequest{}
	resp, err := client.ControllerGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestControllerGetCapabilitiesInfoMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := &csi.ControllerGetCapabilitiesRequest{}
	resp, err := client.ControllerGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestControllerGetCapabilitiesInfoUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := &csi.ControllerGetCapabilitiesRequest{Version: &csi.Version{0, 2, 0}}
	resp, err := client.ControllerGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

// NodeService RPCs

var fakeMountDir = "/run/dcos/csilvm/mnt"

func TestNodePublishVolumeRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.Version = nil
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingVolumeId(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.VolumeId = nil
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeId) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumePresentPublishVolumeInfo(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.PublishVolumeInfo = map[string]string{"foo": "bar"}
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrSpecifiedPublishVolumeInfo) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingTargetPath(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.TargetPath = ""
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingTargetPath) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingVolumeCapability(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.VolumeCapability = nil
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeCapability) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingVolumeCapabilityAccessType(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.VolumeCapability.AccessType = nil
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessType) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeMissingVolumeCapabilityAccessMode(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.VolumeCapability.AccessMode = nil
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessMode) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeVolumeCapabilityAccessModeUNKNOWN(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "", nil)
	req.VolumeCapability.AccessMode.Mode = csi.VolumeCapability_AccessMode_UNKNOWN
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingAccessModeMode) {
		t.Fatal(err)
	}
}

func TestNodePublishVolumeNodeUnpublishVolume_MountVolume_BadFilesystem(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodePublishVolumeRequest("fake_volume_id", fakeMountDir, "ext4", nil)
	resp, err := client.NodePublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedFilesystem) {
		t.Fatal(err)
	}
}

var fakeTargetPath = filepath.Join(fakeMountDir, "fake_volume_id".GetId())

func TestNodeUnpublishVolumeRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testNodeUnpublishVolumeRequest("fake_volume_id", fakeTargetPath)
	resp, err := client.NodeUnpublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestNodeUnpublishVolumeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeUnpublishVolumeRequest("fake_volume_id", fakeTargetPath)
	req.Version = nil
	resp, err := client.NodeUnpublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestNodeUnpublishVolumeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeUnpublishVolumeRequest("fake_volume_id", fakeTargetPath)
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.NodeUnpublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestNodeUnpublishVolumeMissingVolumeId(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeUnpublishVolumeRequest("fake_volume_id", fakeTargetPath)
	req.VolumeId = nil
	resp, err := client.NodeUnpublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVolumeId) {
		t.Fatal(err)
	}
}

func TestNodeUnpublishVolumeMissingTargetPath(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeUnpublishVolumeRequest("fake_volume_id", fakeTargetPath)
	req.TargetPath = ""
	resp, err := client.NodeUnpublishVolume(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingTargetPath) {
		t.Fatal(err)
	}
}

func TestGetNodeID_RemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testGetNodeIDRequest()
	resp, err := client.GetNodeID(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestGetNodeIDMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetNodeIDRequest()
	req.Version = nil
	resp, err := client.GetNodeID(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestGetNodeIDUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testGetNodeIDRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.GetNodeID(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func TestNodeProbeMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeProbeRequest()
	req.Version = nil
	resp, err := client.NodeProbe(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_MISSING_REQUIRED_FIELD
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != false {
		t.Fatal("Expected CallerMustNotRetry to be false")
	}
	expdesc := "The version field must be specified."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestNodeProbeUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeProbeRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.NodeProbe(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result := resp.GetResult()
	if result != nil {
		t.Fatalf("Expected Result to be nil but was: %+v", resp.GetResult())
	}
	error := resp.GetError().GetGeneralError()
	expcode := csi.Error_GeneralError_UNSUPPORTED_REQUEST_VERSION
	if error.GetErrorCode() != expcode {
		t.Fatalf("Expected error code %d but got %d", expcode, error.GetErrorCode())
	}
	if error.GetCallerMustNotRetry() != true {
		t.Fatal("Expected CallerMustNotRetry to be true")
	}
	expdesc := "The requested version is not supported."
	if error.GetErrorDescription() != expdesc {
		t.Fatalf("Expected ErrorDescription to be '%s' but was '%s'", expdesc, error.GetErrorDescription())
	}
}

func TestNodeGetCapabilitiesRemoveVolumeGroup(t *testing.T) {
	client, cleanup := startTest(RemoveVolumeGroup())
	defer cleanup()
	req := testNodeGetCapabilitiesRequest()
	resp, err := client.NodeGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrRemovingMode) {
		t.Fatal(err)
	}
}

func TestNodeGetCapabilitiesRequestMissingVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeGetCapabilitiesRequest()
	req.Version = nil
	resp, err := client.NodeGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrMissingVersion) {
		t.Fatal(err)
	}
}

func TestNodeGetCapabilitiesRequestUnsupportedVersion(t *testing.T) {
	client, cleanup := startTest()
	defer cleanup()
	req := testNodeGetCapabilitiesRequest()
	req.Version = &csi.Version{0, 2, 0}
	resp, err := client.NodeGetCapabilities(context.Background(), req)
	if !grpcErrorEqual(err, ErrUnsupportedVersion) {
		t.Fatal(err)
	}
}

func grpcErrorEqual(gotErr, expErr error) bool {
	got, ok := status.FromError(gotErr)
	if !ok {
		return false
	}
	exp, ok := status.FromError(expErr)
	if !ok {
		return false
	}
	return got.Code() == exp.Code() && got.Message() == exp.Message()
}
