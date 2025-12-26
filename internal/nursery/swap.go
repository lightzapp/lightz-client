package nursery

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/lightzapp/lightz-client/internal/database"
	"github.com/lightzapp/lightz-client/internal/lightning"
	"github.com/lightzapp/lightz-client/internal/logger"
	"github.com/lightzapp/lightz-client/internal/onchain"
	"github.com/lightzapp/lightz-client/internal/utils"
	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
)

func (nursery *Nursery) sendSwapUpdate(swap database.Swap) {
	isFinal := swap.State == lightzrpc.SwapState_SUCCESSFUL || swap.State == lightzrpc.SwapState_REFUNDED
	if swap.LockupTransactionId == "" && swap.State != lightzrpc.SwapState_PENDING {
		isFinal = false
	}

	nursery.sendUpdate(swap.Id, SwapUpdate{
		Swap:    &swap,
		IsFinal: isFinal,
	})
}

// TODO: abstract interactions with chain (querying and broadcasting transactions) into interface to be able to switch between Lightz API and bitcoin core

type Output struct {
	*lightz.OutputDetails
	walletId   *database.Id
	outputArgs onchain.OutputArgs

	setTransaction func(transactionId string, fee uint64) error
	setError       func(err error)
}

func swapOutputArgs(swap *database.Swap) onchain.OutputArgs {
	return onchain.OutputArgs{
		TransactionId: swap.LockupTransactionId,
		Currency:      swap.Pair.From,
		Address:       swap.Address,
		BlindingKey:   swap.BlindingKey,
	}
}

func (nursery *Nursery) getRefundOutput(swap *database.Swap) *Output {
	return &Output{
		OutputDetails: &lightz.OutputDetails{
			SwapId:             swap.Id,
			SwapType:           lightz.NormalSwap,
			Address:            swap.RefundAddress,
			PrivateKey:         swap.PrivateKey,
			Preimage:           []byte{},
			TimeoutBlockHeight: swap.TimoutBlockHeight,
			SwapTree:           swap.SwapTree,
			// TODO: remember if cooperative fails and set this to false
			Cooperative: true,
		},
		walletId:   swap.WalletId,
		outputArgs: swapOutputArgs(swap),
		setTransaction: func(transactionId string, fee uint64) error {
			if err := nursery.database.SetSwapRefundTransactionId(swap, transactionId, fee); err != nil {
				return err
			}

			nursery.sendSwapUpdate(*swap)

			return nil
		},
		setError: func(err error) {
			nursery.handleSwapError(swap, err)
		},
	}
}

func (nursery *Nursery) RegisterSwap(swap database.Swap) error {
	if err := nursery.registerSwaps([]string{swap.Id}); err != nil {
		return err
	}
	nursery.sendSwapUpdate(swap)
	return nil
}

func validatePreimage(preimage []byte, invoice *lightning.DecodedInvoice) error {
	preimageHash := sha256.Sum256(preimage)
	if !bytes.Equal(invoice.PaymentHash[:], preimageHash[:]) {
		return fmt.Errorf("lightzd returned wrong preimage: %x", preimage)
	}
	return nil
}

func (nursery *Nursery) cooperativeSwapClaim(swap *database.Swap, status lightz.SwapStatusResponse) error {
	logger.Debugf("Trying to claim swap %s cooperatively", swap.Id)

	claimDetails, err := nursery.lightz.GetSwapClaimDetails(swap.Id)
	if err != nil {
		return fmt.Errorf("could not get claim details from lightzd: %w", err)
	}

	// Verify that the invoice was actually paid
	decodedInvoice, err := lightning.DecodeInvoice(swap.Invoice, nursery.network.Btc)
	if err != nil {
		return fmt.Errorf("could not decode swap invoice: %w", err)
	}

	if err := validatePreimage(claimDetails.Preimage, decodedInvoice); err != nil {
		return err
	}

	// save the preimage if we dont have it yet
	if swap.Preimage == nil {
		if err := nursery.database.SetSwapPreimage(swap, claimDetails.Preimage); err != nil {
			return fmt.Errorf("could not save preimage: %w", err)
		}
	}

	session, err := lightz.NewSigningSession(swap.SwapTree)
	if err != nil {
		return fmt.Errorf("could not create signing session: %w", err)
	}

	partial, err := session.Sign(claimDetails.TransactionHash, claimDetails.PubNonce)
	if err != nil {
		return fmt.Errorf("could not create partial signature: %w", err)
	}

	if err := nursery.lightz.SendSwapClaimSignature(swap.Id, partial); err != nil {
		return fmt.Errorf("could not send partial signature to lightzd: %w", err)
	}
	return nil
}

func (nursery *Nursery) handleSwapError(swap *database.Swap, err error) {
	if dbErr := nursery.database.UpdateSwapState(swap, lightzrpc.SwapState_ERROR, err.Error()); dbErr != nil {
		logger.Error(dbErr.Error())
	}
	logger.Errorf("Swap %s error: %v", swap.Id, err)
	nursery.sendSwapUpdate(*swap)
}

func (nursery *Nursery) handleSwapStatus(swap *database.Swap, status lightz.SwapStatusResponse) {
	parsedStatus := lightz.ParseEvent(status.Status)

	// transaction mempool can be sent multiple times in case of RBF
	if parsedStatus == swap.Status && parsedStatus != lightz.TransactionMempool {
		logger.Debugf("Status of Swap %s is %s already", swap.Id, parsedStatus)
		return
	}

	logger.Infof("Status of Swap %s changed to: %s", swap.Id, parsedStatus)

	handleError := func(err string) {
		nursery.handleSwapError(swap, errors.New(err))
	}

	// if we're at invoice.set, there is no lockup transaction yet.
	// since there is the possibility that `transaction.mempool` is transitioned through while the client is offline
	// we have to check if the transaction is empty aswell
	if parsedStatus != lightz.InvoiceSet && (parsedStatus == lightz.TransactionMempool || parsedStatus == lightz.TransactionConfirmed || swap.LockupTransactionId == "") {
		swapTransactionResponse, err := nursery.lightz.GetSwapTransaction(swap.Id)
		if err != nil {
			var err lightz.Error
			if !errors.As(err, &err) {
				handleError("Could not get lockup tx from lightzd: " + err.Error())
				return
			}
		} else {
			lockupTransaction, err := lightz.NewTxFromHex(swap.Pair.From, swapTransactionResponse.Hex, swap.BlindingKey)
			if err != nil {
				handleError("Could not decode lockup transaction: " + err.Error())
				return
			}

			if err := nursery.database.SetSwapLockupTransactionId(swap, lockupTransaction.Hash()); err != nil {
				handleError("Could not set lockup transaction in database: " + err.Error())
				return
			}

			result, err := nursery.onchain.FindOutput(swapOutputArgs(swap))
			if err != nil {
				handleError(err.Error())
				return
			}

			logger.Infof("Got lockup transaction of Swap %s: %s", swap.Id, lockupTransaction.Hash())

			if err := nursery.database.SetSwapExpectedAmount(swap, result.Value); err != nil {
				handleError("Could not set expected amount in database: " + err.Error())
				return
			}

			// dont add onchain fee if the swap was paid externally as it might have been part of a larger transaction
			if swap.WalletId != nil {
				fee, err := nursery.onchain.GetTransactionFee(result.Transaction)
				if err != nil {
					handleError("could not get lockup transaction fee: " + err.Error())
					return
				}
				if err := nursery.database.SetSwapOnchainFee(swap, fee); err != nil {
					handleError("could not set lockup transaction fee in database: " + err.Error())
					return
				}
			}
		}
	}

	// batchOnly indicates whether the backend is willing to do a
	// cooperative claim transaction or will batch claim
	batchOnly := false

	switch parsedStatus {
	case lightz.TransactionMempool:
		fallthrough

	case lightz.TransactionConfirmed:
		// Set the invoice of Swaps that were created with only a preimage hash
		if swap.Invoice != "" {
			break
		}

		swapRates, err := nursery.lightz.GetInvoiceAmount(swap.Id)
		if err != nil {
			handleError("Could not query Swap rates of Swap " + swap.Id + ": " + err.Error())
			return
		}

		if err := nursery.CheckAmounts(lightz.NormalSwap, swap.Pair, swap.ExpectedAmount, swapRates.InvoiceAmount, swap.ServiceFeePercent); err != nil {
			handleError(fmt.Sprintf("not accepting invoice amount %d from lightzd: %s", swapRates.InvoiceAmount, err))
			return
		}

		blockHeight, err := nursery.onchain.GetBlockHeight(swap.Pair.From)

		if err != nil {
			handleError("Could not get block height: " + err.Error())
			return
		}

		if nursery.lightning == nil {
			handleError("No lightning node available, can not create invoice for Swap " + swap.Id)
			return
		}

		invoice, err := nursery.lightning.CreateInvoice(
			swapRates.InvoiceAmount,
			swap.Preimage,
			lightz.CalculateInvoiceExpiry(swap.TimoutBlockHeight-blockHeight, swap.Pair.From),
			utils.GetSwapMemo(string(swap.Pair.From)),
		)

		if err != nil {
			handleError("Could not get new invoice for Swap " + swap.Id + ": " + err.Error())
			return
		}

		logger.Infof("Generated new invoice for Swap %s for %d saothis", swap.Id, swapRates.InvoiceAmount)

		_, err = nursery.lightz.SetInvoice(swap.Id, invoice.PaymentRequest)

		if err != nil {
			handleError("Could not set invoice of Swap: " + err.Error())
			return
		}

		err = nursery.database.SetSwapInvoice(swap, invoice.PaymentRequest)

		if err != nil {
			handleError("Could not set invoice of Swap in database: " + err.Error())
			return
		}

	case lightz.TransactionClaimPending, lightz.TransactionClaimed:
		logger.Infof("Swap %s succeeded", swap.Id)

		if parsedStatus == lightz.TransactionClaimPending {
			submarinePairs, err := nursery.lightz.GetSubmarinePairs()
			if err != nil {
				handleError("Could not get submarine pairs: " + err.Error())
				return
			}
			submarinePair, err := lightz.FindPair(swap.Pair, submarinePairs)
			if err != nil {
				handleError("Could not find submarine pair: " + err.Error())
				return
			}

			decodedInvoice, err := lightning.DecodeInvoice(swap.Invoice, nursery.network.Btc)
			if err != nil {
				handleError("Could not decode invoice: " + err.Error())
				return
			}
			batchOnly = decodedInvoice.AmountSat < submarinePair.Limits.Minimal

			if !batchOnly {
				if err := nursery.cooperativeSwapClaim(swap, status); err != nil {
					logger.Warnf("Could not claim swap %s cooperatively: %s", swap.Id, err)
				}
			}
		}
	}

	err := nursery.database.UpdateSwapStatus(swap, parsedStatus)

	if err != nil {
		handleError(fmt.Sprintf("Could not update status of Swap %s to %s: %s", swap.Id, parsedStatus, err))
		return
	}

	// Don't wait for lightzd to claim the swap in case of batchOnly
	// in which case it will stay at transaction.claim.pending for longer
	if parsedStatus.IsCompletedStatus() || batchOnly {
		decodedInvoice, err := lightning.DecodeInvoice(swap.Invoice, nursery.network.Btc)
		if err != nil {
			handleError(fmt.Sprintf("Could not decode invoice: %s", err))
			return
		}
		if swap.Preimage == nil {
			preimage, err := nursery.lightz.GetSwapPreimage(swap.Id)
			if err != nil {
				handleError(fmt.Sprintf("Could not get preimage from lightzd: %s", err))
				return
			}
			if err := validatePreimage(preimage, decodedInvoice); err != nil {
				handleError(err.Error())
				return
			}

			if err := nursery.database.SetSwapPreimage(swap, preimage); err != nil {
				handleError(fmt.Sprintf("Could not set preimage in database: %s", err))
				return
			}
		}

		invoiceAmount := int64(decodedInvoice.AmountSat)
		serviceFee := lightz.CalculatePercentage(swap.ServiceFeePercent, invoiceAmount)
		onchainFee := int64(swap.ExpectedAmount) - invoiceAmount - serviceFee
		if onchainFee < 0 {
			logger.Warnf("Swap %s has negative lightzd onchain fee: %dsat", swap.Id, onchainFee)
			onchainFee = 0
		}

		logger.Infof("Swap service fee: %dsat onchain fee: %dsat", serviceFee, onchainFee)

		if err := nursery.database.SetSwapServiceFee(swap, serviceFee, uint64(onchainFee)); err != nil {
			handleError("Could not set swap service fee in database: " + err.Error())
			return
		}

		if err := nursery.database.UpdateSwapState(swap, lightzrpc.SwapState_SUCCESSFUL, ""); err != nil {
			handleError(err.Error())
			return
		}
	} else if parsedStatus.IsFailedStatus() {
		if swap.State == lightzrpc.SwapState_PENDING {
			if err := nursery.database.UpdateSwapState(swap, lightzrpc.SwapState_SERVER_ERROR, ""); err != nil {
				handleError(err.Error())
				return
			}

			if swap.LockupTransactionId != "" {
				logger.Infof("Swap %s failed, trying to refund cooperatively", swap.Id)
				if _, err := nursery.RefundSwaps(swap.Pair.From, []*database.Swap{swap}, nil); err != nil {
					handleError("Could not refund Swap " + swap.Id + ": " + err.Error())
					return
				}
			}
		}

	}
	nursery.sendSwapUpdate(*swap)
}
