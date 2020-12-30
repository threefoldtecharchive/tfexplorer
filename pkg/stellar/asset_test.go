package stellar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetValidation(t *testing.T) {
	assets := []Asset{
		"",                                // empty asset -> invalid amount of parts
		"TFT:24:1243",                     // too many parts
		":1fjdspsjafo",                    // missing code
		"TFT:",                            // missing issuer
		"TFT:SomethingSomethingSomething", // valid
	}

	assert.Error(t, assets[0].validate(), "invalid amount of parts in asset string, got 1, expected 2")
	assert.Error(t, assets[1].validate(), "invalid amount of parts in asset string, got 3, expected 2")
	assert.Error(t, assets[2].validate(), "missing code in asset")
	assert.Error(t, assets[3].validate(), "missing issuer in asset")
	assert.NoError(t, assets[4].validate())
	assert.Equal(t, assets[4].Code(), "TFT")
	assert.Equal(t, assets[4].Issuer(), "SomethingSomethingSomething")
}

func TestTFTMainnetAsset(t *testing.T) {
	assert.Equal(t, TFTMainnet.Code(), "TFT")
	assert.Equal(t, TFTMainnet.Issuer(), "GBOVQKJYHXRR3DX6NOX2RRYFRCUMSADGDESTDNBDS6CDVLGVESRTAC47")
}

func TestTFTAMainnetAsset(t *testing.T) {
	assert.Equal(t, TFTAMainnet.Code(), "TFTA")
	assert.Equal(t, TFTAMainnet.Issuer(), "GBUT4GP5GJ6B3XW5PXENHQA7TXJI5GOPW3NF4W3ZIW6OOO4ISY6WNLN2")
}

func TestFreeTFTMainnetAsset(t *testing.T) {
	assert.Equal(t, FreeTFTMainnet.Code(), "FreeTFT")
	assert.Equal(t, FreeTFTMainnet.Issuer(), "GCBGS5TFE2BPPUVY55ZPEMWWGR6CLQ7T6P46SOFGHXEBJ34MSP6HVEUT")
}

func TestMainnetAssetsCodeUniqueness(t *testing.T) {
	knownCodes := make(map[string]struct{})

	for asset := range mainnetAssets {
		if _, exists := knownCodes[asset.Code()]; exists {
			t.Fatal("Code ", asset.Code(), " registered twice on mainnet")
		}
	}
}
