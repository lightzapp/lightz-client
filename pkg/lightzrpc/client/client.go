package client

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
)

type Lightz struct {
	Connection
	Client lightzrpc.LightzClient
}

type AutoSwapType string

const (
	LnAutoSwap    AutoSwapType = "lightning"
	ChainAutoSwap AutoSwapType = "chain"
)

var FullPermissions = []*lightzrpc.MacaroonPermissions{
	{Action: lightzrpc.MacaroonAction_READ},
	{Action: lightzrpc.MacaroonAction_WRITE},
}

var ReadPermissions = []*lightzrpc.MacaroonPermissions{
	{Action: lightzrpc.MacaroonAction_READ},
}

func NewLightzClient(conn Connection) Lightz {
	return Lightz{
		Connection: conn,
		Client:     lightzrpc.NewLightzClient(conn.ClientConn),
	}
}

func (lightz *Lightz) GetInfo() (*lightzrpc.GetInfoResponse, error) {
	return lightz.Client.GetInfo(lightz.Ctx, &lightzrpc.GetInfoRequest{})
}

func (lightz *Lightz) GetPairs() (*lightzrpc.GetPairsResponse, error) {
	return lightz.Client.GetPairs(lightz.Ctx, &empty.Empty{})
}

func (lightz *Lightz) GetPairInfo(swapType lightzrpc.SwapType, pair *lightzrpc.Pair) (*lightzrpc.PairInfo, error) {
	return lightz.Client.GetPairInfo(lightz.Ctx, &lightzrpc.GetPairInfoRequest{Pair: pair, Type: swapType})
}

func (lightz *Lightz) ListSwaps(request *lightzrpc.ListSwapsRequest) (*lightzrpc.ListSwapsResponse, error) {
	return lightz.Client.ListSwaps(lightz.Ctx, request)
}

func (lightz *Lightz) GetStats(request *lightzrpc.GetStatsRequest) (*lightzrpc.GetStatsResponse, error) {
	return lightz.Client.GetStats(lightz.Ctx, request)
}

func (lightz *Lightz) RefundSwap(request *lightzrpc.RefundSwapRequest) (*lightzrpc.GetSwapInfoResponse, error) {
	return lightz.Client.RefundSwap(lightz.Ctx, request)
}

func (lightz *Lightz) ClaimSwaps(request *lightzrpc.ClaimSwapsRequest) (*lightzrpc.ClaimSwapsResponse, error) {
	return lightz.Client.ClaimSwaps(lightz.Ctx, request)
}

func (lightz *Lightz) GetSwapInfo(id string) (*lightzrpc.GetSwapInfoResponse, error) {
	return lightz.Client.GetSwapInfo(lightz.Ctx, &lightzrpc.GetSwapInfoRequest{
		Identifier: &lightzrpc.GetSwapInfoRequest_SwapId{SwapId: id},
	})
}

func (lightz *Lightz) GetSwapInfoByPaymentHash(paymentHash []byte) (*lightzrpc.GetSwapInfoResponse, error) {
	return lightz.Client.GetSwapInfo(lightz.Ctx, &lightzrpc.GetSwapInfoRequest{
		Identifier: &lightzrpc.GetSwapInfoRequest_PaymentHash{PaymentHash: paymentHash},
	})
}

func (lightz *Lightz) GetSwapInfoStream(id string) (lightzrpc.Lightz_GetSwapInfoStreamClient, error) {
	return lightz.Client.GetSwapInfoStream(lightz.Ctx, &lightzrpc.GetSwapInfoRequest{
		Id: id,
	})
}

func (lightz *Lightz) CreateSwap(request *lightzrpc.CreateSwapRequest) (*lightzrpc.CreateSwapResponse, error) {
	return lightz.Client.CreateSwap(lightz.Ctx, request)
}

func (lightz *Lightz) CreateReverseSwap(request *lightzrpc.CreateReverseSwapRequest) (*lightzrpc.CreateReverseSwapResponse, error) {
	return lightz.Client.CreateReverseSwap(lightz.Ctx, request)
}

func (lightz *Lightz) CreateChainSwap(request *lightzrpc.CreateChainSwapRequest) (*lightzrpc.ChainSwapInfo, error) {
	return lightz.Client.CreateChainSwap(lightz.Ctx, request)
}

func (lightz *Lightz) GetWallet(name string) (*lightzrpc.Wallet, error) {
	return lightz.Client.GetWallet(lightz.Ctx, &lightzrpc.GetWalletRequest{Name: &name})
}

func (lightz *Lightz) GetWalletById(id uint64) (*lightzrpc.Wallet, error) {
	return lightz.Client.GetWallet(lightz.Ctx, &lightzrpc.GetWalletRequest{Id: &id})
}

func (lightz *Lightz) GetWallets(currency *lightzrpc.Currency, includeReadonly bool) (*lightzrpc.Wallets, error) {
	return lightz.Client.GetWallets(lightz.Ctx, &lightzrpc.GetWalletsRequest{Currency: currency, IncludeReadonly: &includeReadonly})
}

func (lightz *Lightz) ListWalletTransactions(request *lightzrpc.ListWalletTransactionsRequest) (*lightzrpc.ListWalletTransactionsResponse, error) {
	return lightz.Client.ListWalletTransactions(lightz.Ctx, request)
}

func (lightz *Lightz) BumpTransaction(request *lightzrpc.BumpTransactionRequest) (*lightzrpc.BumpTransactionResponse, error) {
	return lightz.Client.BumpTransaction(lightz.Ctx, request)
}

func (lightz *Lightz) ImportWallet(params *lightzrpc.WalletParams, credentials *lightzrpc.WalletCredentials) (*lightzrpc.Wallet, error) {
	return lightz.Client.ImportWallet(lightz.Ctx, &lightzrpc.ImportWalletRequest{Params: params, Credentials: credentials})
}

//nolint:staticcheck
func (lightz *Lightz) SetSubaccount(walletId uint64, subaccount *uint64) (*lightzrpc.Subaccount, error) {
	return lightz.Client.SetSubaccount(lightz.Ctx, &lightzrpc.SetSubaccountRequest{Subaccount: subaccount, WalletId: walletId})
}

//nolint:staticcheck
func (lightz *Lightz) GetSubaccounts(walletId uint64) (*lightzrpc.GetSubaccountsResponse, error) {
	return lightz.Client.GetSubaccounts(lightz.Ctx, &lightzrpc.GetSubaccountsRequest{WalletId: walletId})
}

func (lightz *Lightz) CreateWallet(params *lightzrpc.WalletParams) (*lightzrpc.CreateWalletResponse, error) {
	return lightz.Client.CreateWallet(lightz.Ctx, &lightzrpc.CreateWalletRequest{
		Params: params,
	})
}

func (lightz *Lightz) GetWalletCredentials(id uint64, password *string) (*lightzrpc.WalletCredentials, error) {
	return lightz.Client.GetWalletCredentials(lightz.Ctx, &lightzrpc.GetWalletCredentialsRequest{Id: id, Password: password})
}

func (lightz *Lightz) RemoveWallet(id uint64) (*lightzrpc.RemoveWalletResponse, error) {
	return lightz.Client.RemoveWallet(lightz.Ctx, &lightzrpc.RemoveWalletRequest{Id: id})
}

func (lightz *Lightz) GetSendFee(request *lightzrpc.WalletSendRequest) (*lightzrpc.WalletSendFee, error) {
	return lightz.Client.GetWalletSendFee(lightz.Ctx, request)
}

func (lightz *Lightz) WalletSend(request *lightzrpc.WalletSendRequest) (*lightzrpc.WalletSendResponse, error) {
	return lightz.Client.WalletSend(lightz.Ctx, request)
}

func (lightz *Lightz) WalletReceive(id uint64) (*lightzrpc.WalletReceiveResponse, error) {
	return lightz.Client.WalletReceive(lightz.Ctx, &lightzrpc.WalletReceiveRequest{Id: id})
}

func (lightz *Lightz) Stop() error {
	_, err := lightz.Client.Stop(lightz.Ctx, &empty.Empty{})
	return err
}

func (lightz *Lightz) Unlock(password string) error {
	_, err := lightz.Client.Unlock(lightz.Ctx, &lightzrpc.UnlockRequest{Password: password})
	return err
}

func (lightz *Lightz) VerifyWalletPassword(password string) (bool, error) {
	response, err := lightz.Client.VerifyWalletPassword(lightz.Ctx, &lightzrpc.VerifyWalletPasswordRequest{Password: password})
	if err != nil {
		return false, err
	}
	return response.Correct, nil
}

func (lightz *Lightz) HasPassword() (bool, error) {
	correct, err := lightz.VerifyWalletPassword("")
	return !correct, err
}

func (lightz *Lightz) ChangeWalletPassword(old string, new string) error {
	_, err := lightz.Client.ChangeWalletPassword(lightz.Ctx, &lightzrpc.ChangeWalletPasswordRequest{Old: old, New: new})
	return err
}

func (lightz *Lightz) CreateTenant(name string) (*lightzrpc.Tenant, error) {
	return lightz.Client.CreateTenant(lightz.Ctx, &lightzrpc.CreateTenantRequest{Name: name})
}

func (lightz *Lightz) GetTenant(name string) (*lightzrpc.Tenant, error) {
	return lightz.Client.GetTenant(lightz.Ctx, &lightzrpc.GetTenantRequest{Name: name})
}

func (lightz *Lightz) ListTenants() (*lightzrpc.ListTenantsResponse, error) {
	return lightz.Client.ListTenants(lightz.Ctx, &lightzrpc.ListTenantsRequest{})
}

func (lightz *Lightz) RemoveTenant(name string) error {
	_, err := lightz.Client.RemoveTenant(lightz.Ctx, &lightzrpc.RemoveTenantRequest{Name: name})
	return err
}

func (lightz *Lightz) BakeMacaroon(request *lightzrpc.BakeMacaroonRequest) (*lightzrpc.BakeMacaroonResponse, error) {
	return lightz.Client.BakeMacaroon(lightz.Ctx, request)
}

func (lightz *Lightz) GetSwapMnemonic() (*lightzrpc.GetSwapMnemonicResponse, error) {
	return lightz.Client.GetSwapMnemonic(lightz.Ctx, &lightzrpc.GetSwapMnemonicRequest{})
}

func (lightz *Lightz) SetSwapMnemonic(request *lightzrpc.SetSwapMnemonicRequest) (*lightzrpc.SetSwapMnemonicResponse, error) {
	return lightz.Client.SetSwapMnemonic(lightz.Ctx, request)
}
