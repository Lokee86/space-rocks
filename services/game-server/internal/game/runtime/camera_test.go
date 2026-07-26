package runtime

import "testing"

func TestCameraViewDefaultsToBaseVisibleWorldSize(t *testing.T) {
	view := &CameraView{}

	if view.VisibleWorldWidth() != BaseVisibleWorldWidth {
		t.Fatalf("expected default width %v, got %v", BaseVisibleWorldWidth, view.VisibleWorldWidth())
	}
	if view.VisibleWorldHeight() != BaseVisibleWorldHeight {
		t.Fatalf("expected default height %v, got %v", BaseVisibleWorldHeight, view.VisibleWorldHeight())
	}
}

func TestClampCameraConfigEnforcesSupportedZoomRange(t *testing.T) {
	zoomedOut := ClampCameraConfig(ClientConfig{
		VisibleWorldWidth:  9999,
		VisibleWorldHeight: 9999,
	})
	if zoomedOut.VisibleWorldWidth != MaxVisibleWorldWidth || zoomedOut.VisibleWorldHeight != MaxVisibleWorldHeight {
		t.Fatalf("expected max visible world %vx%v, got %+v", MaxVisibleWorldWidth, MaxVisibleWorldHeight, zoomedOut)
	}

	zoomedIn := ClampCameraConfig(ClientConfig{
		VisibleWorldWidth:  1,
		VisibleWorldHeight: 1,
	})
	if zoomedIn.VisibleWorldWidth != MinVisibleWorldWidth || zoomedIn.VisibleWorldHeight != MinVisibleWorldHeight {
		t.Fatalf("expected min visible world %vx%v, got %+v", MinVisibleWorldWidth, MinVisibleWorldHeight, zoomedIn)
	}
}

func TestClampCameraConfigPreservesBaseAspectRatio(t *testing.T) {
	config := ClampCameraConfig(ClientConfig{
		VisibleWorldWidth:  9999,
		VisibleWorldHeight: MinVisibleWorldHeight,
	})

	if config.VisibleWorldWidth != MinVisibleWorldWidth || config.VisibleWorldHeight != MinVisibleWorldHeight {
		t.Fatalf("expected narrow dimension to constrain the full view to %vx%v, got %+v", MinVisibleWorldWidth, MinVisibleWorldHeight, config)
	}
}
