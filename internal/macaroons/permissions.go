package macaroons

import (
	"github.com/lightzapp/lightz-client/pkg/lightzrpc"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

var (
	TenantReadPermissions = []bakery.Op{
		{
			Entity: "info",
			Action: "read",
		},
		{
			Entity: "swap",
			Action: "read",
		},
		{
			Entity: "wallet",
			Action: "read",
		},
		{
			Entity: "autoswap",
			Action: "read",
		},
	}
	TenantWritePermissons = []bakery.Op{
		{
			Entity: "info",
			Action: "write",
		},
		{
			Entity: "swap",
			Action: "write",
		},
		{
			Entity: "wallet",
			Action: "write",
		},
		{
			Entity: "autoswap",
			Action: "write",
		},
	}
	ReadPermissions = append([]bakery.Op{
		{
			Entity: "admin",
			Action: "read",
		},
	}, TenantReadPermissions...)

	WritePermissions = append([]bakery.Op{
		{
			Entity: "admin",
			Action: "write",
		},
	}, TenantWritePermissons...)

	RPCServerPermissions = map[string][]bakery.Op{
		"/lightzrpc.Lightz/GetInfo": {{
			Entity: "info",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetServiceInfo": {{
			Entity: "info",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetPairInfo": {{
			Entity: "info",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetPairs": {{
			Entity: "info",
			Action: "read",
		}},
		"/lightzrpc.Lightz/ListSwaps": {{
			Entity: "swap",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetStats": {{
			Entity: "swap",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetSwapInfo": {{
			Entity: "swap",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetSwapInfoStream": {{
			Entity: "swap",
			Action: "read",
		}},
		"/lightzrpc.Lightz/Deposit": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/CreateSwap": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/CreateChainSwap": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/RefundSwap": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/ClaimSwaps": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/CreateChannel": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/CreateReverseSwap": {{
			Entity: "swap",
			Action: "write",
		}},
		"/lightzrpc.Lightz/CreateWallet": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/ImportWallet": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/SetSubaccount": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/GetSubaccounts": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetWalletSendFee": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/RemoveWallet": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/WalletSend": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/WalletReceive": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetWalletCredentials": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/GetWallets": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/ListWalletTransactions": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/BumpTransaction": {{
			Entity: "wallet",
			Action: "write",
		}},
		"/lightzrpc.Lightz/GetWallet": {{
			Entity: "wallet",
			Action: "read",
		}},
		"/lightzrpc.Lightz/Stop": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/Unlock": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/ChangeWalletPassword": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/VerifyWalletPassword": {{
			Entity: "admin",
			Action: "read",
		}},
		"/lightzrpc.Lightz/CreateTenant": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/ListTenants": {{
			Entity: "admin",
			Action: "read",
		}},
		"/lightzrpc.Lightz/GetTenant": {{
			Entity: "admin",
			Action: "read",
		}},
		"/lightzrpc.Lightz/RemoveTenant": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/BakeMacaroon": {{
			Entity: "admin",
			Action: "write",
		}},
		"/lightzrpc.Lightz/GetSwapMnemonic": {{
			Entity: "admin",
			Action: "read",
		}},
		"/lightzrpc.Lightz/SetSwapMnemonic": {{
			Entity: "admin",
			Action: "write",
		}},
		"/autoswaprpc.AutoSwap/GetRecommendations": {{
			Entity: "autoswap",
			Action: "read",
		}},
		"/autoswaprpc.AutoSwap/ExecuteRecommendations": {{
			Entity: "autoswap",
			Action: "write",
		}},
		"/autoswaprpc.AutoSwap/GetStatus": {{
			Entity: "autoswap",
			Action: "read",
		}},
		"/autoswaprpc.AutoSwap/GetConfig": {{
			Entity: "autoswap",
			Action: "read",
		}},
		"/autoswaprpc.AutoSwap/ReloadConfig": {{
			Entity: "autoswap",
			Action: "write",
		}},
		"/autoswaprpc.AutoSwap/UpdateChainConfig": {{
			Entity: "autoswap",
			Action: "write",
		}},
		"/autoswaprpc.AutoSwap/UpdateLightningConfig": {{
			Entity: "autoswap",
			Action: "write",
		}},
	}
)

func AdminPermissions() []bakery.Op {
	admin := make([]bakery.Op, len(ReadPermissions)+len(WritePermissions))
	copy(admin, ReadPermissions)
	copy(admin[len(ReadPermissions):], WritePermissions)

	return admin
}

func GetPermissions(isTenant bool, permissions []*lightzrpc.MacaroonPermissions) (result []bakery.Op) {
	for _, permission := range permissions {
		switch permission.Action {
		case lightzrpc.MacaroonAction_READ:
			if isTenant {
				result = append(result, TenantReadPermissions...)
			} else {
				result = append(result, ReadPermissions...)
			}
		case lightzrpc.MacaroonAction_WRITE:
			if isTenant {
				result = append(result, TenantWritePermissons...)
			} else {
				result = append(result, WritePermissions...)
			}
		}
	}
	return result
}
