// x/nameservice/types/codec.go
package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgBuyName{}, "nameservice/BuyName", nil)
	cdc.RegisterConcrete(&MsgSetName{}, "nameservice/SetName", nil)
	cdc.RegisterConcrete(&MsgDeleteName{}, "nameservice/DeleteName", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgBuyName{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetName{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteName{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
