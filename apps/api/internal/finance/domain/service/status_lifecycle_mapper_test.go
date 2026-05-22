package service

import (
	"testing"

	financeModels "github.com/gilabs/gims/api/internal/finance/data/models"
	"github.com/stretchr/testify/require"
)

func TestResolveLifecycleStage_DefaultMapping(t *testing.T) {
	stage, err := ResolveLifecycleStage(financeModels.AssetStatusUnderMaintenance, nil)
	require.NoError(t, err)
	require.Equal(t, financeModels.AssetLifecycleUnderMaintenance, stage)
}

func TestResolveLifecycleStage_PendingApprovalMapping(t *testing.T) {
	stage, err := ResolveLifecycleStage(financeModels.AssetStatusPendingApproval, nil)
	require.NoError(t, err)
	require.Equal(t, financeModels.AssetLifecyclePending, stage)
}

func TestResolveLifecycleStage_RejectsInvalidCombination(t *testing.T) {
	preferred := financeModels.AssetLifecycleUnderMaintenance
	_, err := ResolveLifecycleStage(financeModels.AssetStatusDisposed, &preferred)
	require.Error(t, err)
}

func TestApplyStatusLifecycle_UpdatesBothFields(t *testing.T) {
	asset := &financeModels.Asset{}
	preferred := financeModels.AssetLifecycleInUse

	err := ApplyStatusLifecycle(asset, financeModels.AssetStatusActive, &preferred)
	require.NoError(t, err)
	require.Equal(t, financeModels.AssetStatusActive, asset.Status)
	require.Equal(t, financeModels.AssetLifecycleInUse, asset.LifecycleStage)
}
