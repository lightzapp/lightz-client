package nursery

import (
	"errors"
	"testing"
	"time"

	"github.com/lightzapp/lightz-client/internal/database"
	"github.com/lightzapp/lightz-client/internal/lightning"
	lnmock "github.com/lightzapp/lightz-client/internal/mocks/lightning"
	onchainmock "github.com/lightzapp/lightz-client/internal/mocks/onchain"
	"github.com/lightzapp/lightz-client/internal/onchain"
	"github.com/lightzapp/lightz-client/internal/test"
	"github.com/lightzapp/lightz-client/pkg/lightz"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const defaultFeeLimitPpm = uint64(1000)

func setup(t *testing.T) *Nursery {
	chain := &onchain.Onchain{
		Btc:     &onchain.Currency{},
		Liquid:  &onchain.Currency{},
		Network: lightz.Regtest,
	}
	chain.Init()

	db := database.Database{Path: ":memory:"}
	require.NoError(t, db.Connect())

	nursery := New(
		nil,
		defaultFeeLimitPpm,
		lightz.Regtest,
		nil,
		chain,
		&lightz.Api{URL: lightz.Regtest.DefaultLightzUrl},
		&db,
	)

	return nursery
}

func TestPayReverseSwap(t *testing.T) {
	t.Run("MaxRoutingFee", func(t *testing.T) {
		testInvoice := "lnbcrt10m1p5y4z9epp5hh09qu0605hcjvc5r6dv3ma0z45h7pxjcp4xv383avzxk4yf0tlsdqqcqzzsxqyz5vqsp5nzsy8g59gvlp694x7rc7gxfllk0wswl95vvk5eguc30jrvcqeuws9qxpqysgqmfdaryxsaze7s26ew6y4zu3hk8p9sj8ezcpcvt6rchjuxva5zvwyq7897ffw4mjmsg6efugt5k7qhfy04j6wxnlzpfu48r5mjsruzugqjp04ec"
		testInvoiceAmount := uint64(1_000_000)
		maxRoutingFeePpm := uint64(100)
		expectedLimit := uint(maxRoutingFeePpm) // 1 million sat invoice

		setup := func(t *testing.T) *Nursery {
			nursery := setup(t)
			mockLightning := lnmock.NewMockLightningNode(t)
			mockLightning.EXPECT().
				PayInvoice(mock.Anything, testInvoice, expectedLimit, mock.Anything, mock.Anything).
				Return(&lightning.PayInvoiceResponse{
					FeeMsat: 1,
				}, nil)
			mockLightning.EXPECT().
				PaymentStatus(mock.Anything).
				Return(nil, errors.New("invoice not found"))
			nursery.lightning = mockLightning
			return nursery
		}
		testSwap := &database.ReverseSwap{
			Id:            "test-swap",
			Invoice:       testInvoice,
			InvoiceAmount: testInvoiceAmount,
		}

		t.Run("Custom", func(t *testing.T) {
			nursery := setup(t)
			swap := testSwap
			swap.RoutingFeeLimitPpm = &maxRoutingFeePpm

			err := nursery.payReverseSwap(swap)
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // wait for the pay call to run in the goroutine
		})

		t.Run("Default", func(t *testing.T) {
			nursery := setup(t)
			nursery.maxRoutingFeePpm = maxRoutingFeePpm
			err := nursery.payReverseSwap(testSwap)
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // wait for the pay call to run in the goroutine
		})

	})
	t.Run("ExternalPay", func(t *testing.T) {
		nursery := setup(t)
		swap := &database.ReverseSwap{
			ExternalPay: true,
		}

		err := nursery.payReverseSwap(swap)
		require.NoError(t, err) // Should not attempt to pay external swaps
	})

	t.Run("NoLightning", func(t *testing.T) {
		nursery := setup(t)
		nursery.lightning = nil
		err := nursery.payReverseSwap(&database.ReverseSwap{
			ExternalPay: false,
		})
		require.Error(t, err)
		require.Equal(t, "no lightning node available to pay invoice", err.Error())
	})
}

func TestChooseDirectOutput(t *testing.T) {
	test.InitLogger()
	nursery := setup(t)
	tests := []struct {
		name           string
		outputs        []*onchain.Output
		expected       *onchain.Output
		swap           *database.ReverseSwap
		feeEstimations lightz.FeeEstimations
		err            require.ErrorAssertionFunc
	}{
		{
			name: "Match",
			outputs: []*onchain.Output{
				{
					TxId:  "tx1",
					Value: 1000,
				},
				{
					TxId:  "tx3",
					Value: 15000,
				},
				{
					TxId:  "tx2",
					Value: 9000,
				},
			},
			expected: &onchain.Output{
				TxId:  "tx2",
				Value: 9000,
			},
			swap: &database.ReverseSwap{
				Pair: lightz.Pair{
					From: lightz.CurrencyBtc,
					To:   lightz.CurrencyLiquid,
				},
				ServiceFeePercent: 1,
				OnchainAmount:     9500,
				InvoiceAmount:     10000,
			},
			feeEstimations: lightz.FeeEstimations{
				lightz.CurrencyLiquid: 0.1,
				lightz.CurrencyBtc:    1,
			},
			err: require.NoError,
		},
		{
			name: "NoMatch",
			outputs: []*onchain.Output{
				{
					TxId:  "tx1",
					Value: 500,
				},
				{
					TxId:  "tx2",
					Value: 600,
				},
			},
			swap: &database.ReverseSwap{
				Pair: lightz.Pair{
					From: lightz.CurrencyBtc,
					To:   lightz.CurrencyBtc,
				},
				InvoiceAmount:     10000,
				OnchainAmount:     9500,
				ServiceFeePercent: 1,
			},
			feeEstimations: lightz.FeeEstimations{
				lightz.CurrencyLiquid: 0.1,
				lightz.CurrencyBtc:    1,
			},
			expected: nil,
			err:      require.NoError,
		},
		{
			name: "Overpaid",
			outputs: []*onchain.Output{
				{
					TxId:  "tx1",
					Value: 10500,
				},
			},
			expected: &onchain.Output{
				TxId:  "tx1",
				Value: 10500,
			},
			swap: &database.ReverseSwap{
				Pair: lightz.Pair{
					From: lightz.CurrencyBtc,
					To:   lightz.CurrencyLiquid,
				},
				InvoiceAmount:     10000,
				OnchainAmount:     9500,
				ServiceFeePercent: 1,
			},
			feeEstimations: lightz.FeeEstimations{
				lightz.CurrencyLiquid: 0.1,
				lightz.CurrencyBtc:    1,
			},
			err: require.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := nursery.chooseDirectOutput(test.swap, test.feeEstimations, test.outputs)
			test.err(t, err)
			require.Equal(t, test.expected, output)
		})
	}
}

func TestGetFeeEstimations(t *testing.T) {
	chainProvider := func(t *testing.T, fees float64, err error) *onchainmock.MockChainProvider {
		mock := onchainmock.NewMockChainProvider(t)
		mock.EXPECT().EstimateFee().Return(fees, err)
		return mock
	}
	tests := []struct {
		swapType lightz.SwapType
		pair     lightz.Pair
		fees     lightz.FeeEstimations
		err      require.ErrorAssertionFunc
		setup    func(t *testing.T, nursery *Nursery)
	}{
		{
			swapType: lightz.NormalSwap,
			pair:     lightz.Pair{From: lightz.CurrencyBtc, To: lightz.CurrencyLiquid},
			fees:     lightz.FeeEstimations{lightz.CurrencyBtc: 10},
			setup: func(t *testing.T, nursery *Nursery) {
				nursery.onchain.Btc.Chain = chainProvider(t, 10, nil)
			},
			err: require.NoError,
		},
		{
			swapType: lightz.NormalSwap,
			pair:     lightz.Pair{From: lightz.CurrencyBtc, To: lightz.CurrencyLiquid},
			setup: func(t *testing.T, nursery *Nursery) {
				nursery.onchain.Btc.Chain = chainProvider(t, 10, errors.New("error"))
			},
			err: require.Error,
		},
		{
			swapType: lightz.ReverseSwap,
			pair:     lightz.Pair{From: lightz.CurrencyBtc, To: lightz.CurrencyLiquid},
			fees:     lightz.FeeEstimations{lightz.CurrencyLiquid: 1},
			setup: func(t *testing.T, nursery *Nursery) {
				nursery.onchain.Liquid.Chain = chainProvider(t, 1, nil)
			},
			err: require.NoError,
		},
		{
			swapType: lightz.ChainSwap,
			pair:     lightz.Pair{From: lightz.CurrencyBtc, To: lightz.CurrencyLiquid},
			fees:     lightz.FeeEstimations{lightz.CurrencyBtc: 10, lightz.CurrencyLiquid: 1},
			setup: func(t *testing.T, nursery *Nursery) {
				nursery.onchain.Btc.Chain = chainProvider(t, 10, nil)
				nursery.onchain.Liquid.Chain = chainProvider(t, 1, nil)
			},
			err: require.NoError,
		},
	}
	for _, test := range tests {
		t.Run(string(test.swapType), func(t *testing.T) {
			nursery := setup(t)
			test.setup(t, nursery)
			fees, err := nursery.GetFeeEstimations(test.swapType, test.pair)
			test.err(t, err)
			require.Equal(t, test.fees, fees)
		})
	}
}
