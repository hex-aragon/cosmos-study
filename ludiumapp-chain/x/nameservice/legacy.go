package nameservice

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

// RegisterRESTRoutes registers the REST routes for the bank module.
// NOTE: 레거시 메소드라 구현만 하고 넘어갑니다.
// ref; https://github.com/cosmos/cosmos-sdk/blob/v0.45.4/x/bank/module.go#L66
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	// rest.RegisterHandlers(clientCtx, rtr)
}
