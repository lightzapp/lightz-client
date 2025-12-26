package lightz

import (
	"fmt"

	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chainutil/hdkeychain"
	btcec "github.com/flokiorg/go-flokicoin/crypto"
	"github.com/tyler-smith/go-bip39"
)

func mnemonicToHdKey(mnemonic string) (*hdkeychain.ExtendedKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// lightzd backend and web app also use main net params across all networks
	return hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
}

func deriveKey(hdKey *hdkeychain.ExtendedKey, index uint32) (*hdkeychain.ExtendedKey, error) {
	path := []uint32{44, 0, 0, 0, index}
	extendedKey, err := hdKey.Derive(path[0])
	if err != nil {
		return nil, err
	}

	for _, p := range path[1:] {
		extendedKey, err = extendedKey.Derive(p)
		if err != nil {
			return nil, err
		}
	}
	return extendedKey, nil
}

func DeriveKey(mnemonic string, index uint32) (*btcec.PrivateKey, error) {
	hdKey, err := mnemonicToHdKey(mnemonic)
	if err != nil {
		return nil, err
	}

	extendedKey, err := deriveKey(hdKey, index)
	if err != nil {
		return nil, err
	}

	return extendedKey.ECPrivKey()
}
