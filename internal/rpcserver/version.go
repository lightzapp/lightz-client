package rpcserver

import (
	"github.com/lightzapp/lightz-client/internal/lightning"
	"github.com/lightzapp/lightz-client/internal/logger"
	"github.com/lightzapp/lightz-client/internal/utils"
	"github.com/lightzapp/lightz-client/pkg/lightz"
)

const minLndVersion = "0.15.0"
const minClnVersion = "25.05"
const minLightzVersion = "3.5.0"

func checkLndVersion(info *lightning.LightningInfo) {
	if err := utils.CheckVersion("LND", info.Version, minLndVersion); err != nil {
		logger.Fatal(err.Error())
	}
}

func checkClnVersion(info *lightning.LightningInfo) {
	if err := utils.CheckVersion("CLN", info.Version, minClnVersion); err != nil {
		logger.Fatal(err.Error())
	}
}

func checkLightzVersion(response *lightz.GetVersionResponse) error {
	return utils.CheckVersion("Lightz", response.Version, minLightzVersion)
}
