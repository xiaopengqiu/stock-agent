package data

import (
	"context"
	"testing"
)

func TestNewPositionAnalysisService(t *testing.T) {
	ctx := context.Background()
	service := NewPositionAnalysisService(ctx)

	if service == nil {
		t.Fatal("NewPositionAnalysisService returned nil")
	}
}
