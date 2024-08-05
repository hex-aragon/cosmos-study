package keeper

import (
	"github.com/Jeongseup/ludiumapp/x/nameservice/types"
)

var _ types.QueryServer = Keeper{}
