package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	btcec "github.com/flokiorg/go-flokicoin/crypto"
	"github.com/lightzapp/lightz-client/pkg/lightzd"
)

const endpoint = "<Lightz API endpoint>"
const invoiceAmount = 100000
const destinationAddress = "<address to which the swap should be claimed>"

// Swap from Lightning to BTC mainchain
var toCurrency = lightz.CurrencyBtc

var network = lightz.Regtest

func printJson(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}

func reverseSwap() error {
	ourKeys, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}

	preimage := make([]byte, 32)
	_, err = rand.Read(preimage)
	if err != nil {
		return err
	}
	preimageHash := sha256.Sum256(preimage)

	lightzApi := &lightz.Api{URL: endpoint}

	reversePairs, err := lightzApi.GetReversePairs()
	if err != nil {
		return fmt.Errorf("could not get reverse pairs: %s", err)
	}

	pair := lightz.Pair{From: lightz.CurrencyBtc, To: toCurrency}
	pairInfo, err := lightz.FindPair(pair, reversePairs)
	if err != nil {
		return fmt.Errorf("could not find reverse pair: %s", err)
	}

	fees := pairInfo.Fees
	serviceFee := lightz.Percentage(fees.Percentage)
	fmt.Printf("Service Fee: %dsat\n", lightz.CalculatePercentage(serviceFee, invoiceAmount))
	fmt.Printf("Network Fee: %dsat\n", fees.MinerFees.Lockup+fees.MinerFees.Claim)

	swap, err := lightzApi.CreateReverseSwap(lightz.CreateReverseSwapRequest{
		From:           lightz.CurrencyBtc,
		To:             toCurrency,
		ClaimPublicKey: ourKeys.PubKey().SerializeCompressed(),
		PreimageHash:   preimageHash[:],
		InvoiceAmount:  invoiceAmount,
		PairHash:       pairInfo.Hash,
	})
	if err != nil {
		return fmt.Errorf("Could not create swap: %s", err)
	}

	pubKey, err := btcec.ParsePubKey(swap.RefundPublicKey)
	if err != nil {
		return err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(toCurrency, true, ourKeys, pubKey); err != nil {
		return err
	}

	if err := tree.Check(lightz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
		return err
	}

	fmt.Println("Swap created")
	printJson(swap)

	ws := lightzApi.NewWebsocket()
	if err := ws.Connect(); err != nil {
		return fmt.Errorf("Could not connect to Lightz websocket: %w", err)
	}

	if err := ws.Subscribe([]string{swap.Id}); err != nil {
		return err
	}

	for update := range ws.Updates {
		parsedStatus := lightz.ParseEvent(update.Status)

		printJson(update)

		switch parsedStatus {
		case lightz.SwapCreated:
			fmt.Println("Waiting for invoice to be paid")
			break

		case lightz.TransactionMempool:
			lockupTransaction, err := lightz.NewTxFromHex(toCurrency, update.Transaction.Hex, nil)
			if err != nil {
				return err
			}

			vout, _, err := lockupTransaction.FindVout(network, swap.LockupAddress)
			if err != nil {
				return err
			}

			satsPerVbyte := float64(2)
			claimTransaction, _, err := lightz.ConstructTransaction(
				network,
				lightz.CurrencyBtc,
				[]lightz.OutputDetails{
					{
						SwapId:            swap.Id,
						SwapType:          lightz.ReverseSwap,
						Address:           destinationAddress,
						LockupTransaction: lockupTransaction,
						Vout:              vout,
						Preimage:          preimage,
						PrivateKey:        ourKeys,
						SwapTree:          tree,
						Cooperative:       true,
					},
				},
				lightz.Fee{SatsPerVbyte: &satsPerVbyte},
				lightzApi,
			)
			if err != nil {
				return fmt.Errorf("could not create claim transaction: %w", err)
			}

			txHex, err := claimTransaction.Serialize()
			if err != nil {
				return fmt.Errorf("could not serialize claim transaction: %w", err)
			}

			txId, err := lightzApi.BroadcastTransaction(toCurrency, txHex)
			if err != nil {
				return fmt.Errorf("could not broadcast transaction: %w", err)
			}

			fmt.Printf("Broadcast claim transaction: %s\n", txId)
			break

		case lightz.InvoiceSettled:
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
	if err := reverseSwap(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
