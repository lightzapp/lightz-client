package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	btcec "github.com/flokiorg/go-flokicoin/crypto"

	"github.com/flokiorg/flnd/zpay32"
	"github.com/lightzapp/lightz-client/pkg/lightzd"
)

const endpoint = "<Lightz API endpoint>"
const invoice = "<the invoice you want to pay"

var network = lightz.Regtest

func printJson(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}

func submarineSwap() error {
	keys, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}

	lightzApi := &lightz.Api{URL: endpoint}

	submarinePairs, err := lightzApi.GetSubmarinePairs()
	if err != nil {
		return fmt.Errorf("Could not get submarine pairs: %s", err)
	}

	pair, err := lightz.FindPair(lightz.Pair{From: lightz.CurrencyBtc, To: lightz.CurrencyBtc}, submarinePairs)
	if err != nil {
		return fmt.Errorf("Could not find submarine pair: %s", err)
	}

	decodedInvoice, err := zpay32.Decode(invoice, network.Btc)
	if err != nil {
		return fmt.Errorf("could not decode invoice: %s", err)
	}
	invoiceAmount := *decodedInvoice.MilliSat / 1000

	fees := pair.Fees
	serviceFee := lightz.Percentage(fees.Percentage)
	fmt.Printf("Service Fee: %dsat\n", lightz.CalculatePercentage(serviceFee, invoiceAmount))
	fmt.Printf("Network Fee: %dsat\n", fees.MinerFees)

	swap, err := lightzApi.CreateSwap(lightz.CreateSwapRequest{
		From:            lightz.CurrencyBtc,
		To:              lightz.CurrencyBtc,
		RefundPublicKey: keys.PubKey().SerializeCompressed(),
		Invoice:         invoice,
		PairHash:        pair.Hash,
	})
	if err != nil {
		return fmt.Errorf("Could not create swap: %s", err)
	}

	pubKey, err := btcec.ParsePubKey(swap.ClaimPublicKey)
	if err != nil {
		return err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(lightz.CurrencyBtc, false, keys, pubKey); err != nil {
		return err
	}

	// Check the scripts of the Taptree to make sure Lightz is not cheating
	if err := tree.Check(lightz.NormalSwap, swap.TimeoutBlockHeight, decodedInvoice.PaymentHash[:]); err != nil {
		return err
	}

	// Verify that Lightz is giving us the correct address
	if err := tree.CheckAddress(swap.Address, network, nil); err != nil {
		return err
	}

	fmt.Println("Swap created")
	printJson(swap)

	ws := lightzApi.NewWebsocket()
	if err := ws.Connect(); err != nil {
		return fmt.Errorf("Could not connect to Lightz websocket: %s", err)
	}

	if err := ws.Subscribe([]string{swap.Id}); err != nil {
		return err
	}

	for update := range ws.Updates {
		parsedStatus := lightz.ParseEvent(update.Status)

		printJson(update)

		switch parsedStatus {
		case lightz.InvoiceSet:
			fmt.Println("Waiting for onchain transaction")
			break

		case lightz.TransactionMempool:
			fmt.Println("Lightz found transaction in mempool")
			break

		case lightz.TransactionConfirmed:
			fmt.Println("Lightz found transaction in blockchain")
			break

		case lightz.TransactionClaimPending:
			// Create a partial signature to allow Lightz to do a key path spend to claim the onchain coins
			claimDetails, err := lightzApi.GetSwapClaimDetails(swap.Id)
			if err != nil {
				return fmt.Errorf("Could not get claim details from Lightz: %s", err)
			}

			// Verify that the invoice was actually paid
			preimageHash := sha256.Sum256(claimDetails.Preimage)
			if !bytes.Equal(decodedInvoice.PaymentHash[:], preimageHash[:]) {
				return fmt.Errorf("Lightz returned wrong preimage: %x", claimDetails.Preimage)
			}

			session, err := lightz.NewSigningSession(tree)
			partial, err := session.Sign(claimDetails.TransactionHash, claimDetails.PubNonce)
			if err != nil {
				return fmt.Errorf("could not create partial signature: %s", err)
			}

			if err := lightzApi.SendSwapClaimSignature(swap.Id, partial); err != nil {
				return fmt.Errorf("could not send partial signature to Lightz: %s", err)
			}
			break

		case lightz.TransactionClaimed:
			fmt.Println("Swap succeeded", swap.Id)
			if err := ws.Close(); err != nil {
				return err
			}
			break
		}

	}

	return nil
}

func main() {
	if err := submarineSwap(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
