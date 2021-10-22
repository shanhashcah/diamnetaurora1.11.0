package actions

import (
	"net/http"
	"net/url"

	"github.com/diamnet/go/protocols/aurora"
	"github.com/diamnet/go/services/aurora/internal/ledger"
	"github.com/diamnet/go/services/aurora/internal/resourceadapter"
)

type CoreSettings struct {
	CurrentProtocolVersion       int32
	CoreSupportedProtocolVersion int32
	CoreVersion                  string
}

type CoreSettingsGetter interface {
	GetCoreSettings() CoreSettings
}

type GetRootHandler struct {
	CoreSettingsGetter
	NetworkPassphrase string
	FriendbotURL      *url.URL
	AuroraVersion    string
}

func (handler GetRootHandler) GetResource(w HeaderWriter, r *http.Request) (interface{}, error) {
	var res aurora.Root
	templates := map[string]string{
		"accounts":           AccountsQuery{}.URITemplate(),
		"claimableBalances":  ClaimableBalancesQuery{}.URITemplate(),
		"offers":             OffersQuery{}.URITemplate(),
		"strictReceivePaths": StrictReceivePathsQuery{}.URITemplate(),
		"strictSendPaths":    FindFixedPathsQuery{}.URITemplate(),
	}
	coreSettings := handler.GetCoreSettings()
	resourceadapter.PopulateRoot(
		r.Context(),
		&res,
		ledger.CurrentState(),
		handler.AuroraVersion,
		coreSettings.CoreVersion,
		handler.NetworkPassphrase,
		coreSettings.CurrentProtocolVersion,
		coreSettings.CoreSupportedProtocolVersion,
		handler.FriendbotURL,
		templates,
	)
	return res, nil
}
