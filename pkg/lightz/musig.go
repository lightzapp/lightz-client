package lightz

import (
	"bytes"
	"errors"
	"fmt"
	"slices"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/flokiorg/go-flokicoin/crypto/schnorr/musig2"
	"github.com/flokiorg/go-flokicoin/wire"
	liquidtx "github.com/vulpemventures/go-elements/transaction"
)

type MusigSession struct {
	*musig2.Session
	tree *SwapTree
}

func NewSigningSession(tree *SwapTree) (*MusigSession, error) {
	ctx, err := musig2.NewContext(
		tree.ourKey,
		false,
		musig2.WithTweakedContext(tree.taprootTweak),
		musig2.WithKnownSigners([]*secp256k1.PublicKey{tree.lightzKey, tree.ourKey.PubKey()}),
	)
	if err != nil {
		return nil, err
	}

	session, err := ctx.NewSession()
	if err != nil {
		return nil, err
	}

	return &MusigSession{session, tree}, nil
}

func (session *MusigSession) Sign(hash []byte, nonce []byte) (*PartialSignature, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("invalid hash length %d", len(hash))
	}

	if len(nonce) != 66 {
		return nil, fmt.Errorf("invalid nonce lenth %d", len(nonce))
	}

	all, err := session.RegisterPubNonce([66]byte(nonce))
	if err != nil {
		return nil, err
	}
	if !all {
		return nil, errors.New("could not combine nonces")
	}

	ourNonce := session.PublicNonce()

	partial, err := session.Session.Sign([32]byte(hash))
	if err != nil {
		return nil, err
	}

	b := bytes.NewBuffer(nil)
	if err := partial.Encode(b); err != nil {
		return nil, err
	}

	return &PartialSignature{
		PubNonce:         HexString(ourNonce[:]),
		PartialSignature: HexString(b.Bytes()),
	}, nil
}

func (session *MusigSession) Finalize(transaction Transaction, outputs []OutputDetails, network *Network, zSignature *PartialSignature) (err error) {
	var hash []byte
	isLiquid := session.tree.isLiquid
	idx := slices.IndexFunc(outputs, func(output OutputDetails) bool {
		return output.SwapTree == session.tree
	})
	if idx == -1 {
		return errors.New("outputs do not contain session swap tree")
	}
	if isLiquid {
		hash = liquidTaprootHash(&transaction.(*LiquidTransaction).Transaction, network, outputs, idx, true)
	} else {
		hash, err = btcTaprootHash(transaction, outputs, idx)
	}
	if err != nil {
		return err
	}

	_, err = session.Sign(hash, zSignature.PubNonce)
	if err != nil {
		return err
	}

	s := &secp256k1.ModNScalar{}
	s.SetByteSlice(zSignature.PartialSignature)
	partial := musig2.NewPartialSignature(s, nil)
	haveFinal, err := session.CombineSig(&partial)
	if err != nil {
		return fmt.Errorf("could not combine signatures: %w", err)
	}
	if !haveFinal {
		return errors.New("could not combine signatures")
	}

	signature := session.FinalSig().Serialize()
	if isLiquid {
		tx := transaction.(*LiquidTransaction)
		tx.Transaction.Inputs[idx].Witness = liquidtx.TxWitness{signature}
	} else {
		tx := transaction.(*BtcTransaction)
		tx.MsgTx().TxIn[idx].Witness = wire.TxWitness{signature}
	}
	return nil
}
