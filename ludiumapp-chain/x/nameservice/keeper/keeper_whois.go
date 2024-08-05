package keeper

import (
	"github.com/Jeongseup/ludiumapp/x/nameservice/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetWhois set a specific whois in the store from its index
func (k Keeper) SetWhois(ctx sdk.Context, whois types.Whois) {
	// NOTE: 각 키퍼 레벨에서 이제 대부분 구현된 코드를 바탕으로 이해를 해봅시다.
	// 우린 전달받은 컨텍스트를 이용해서, 모듈의 키퍼 스토어와 메세지 프리픽스를 통해서 스토어란 디비를 만들어냅니다.
	// 그리고 전달받은 whois란 데이터를 압축시키기 위해서 protobuf를 통해서 marshal하고
	// 디비의 Set함수를 이용해서 whois데이터를 디비에 저장합니다.
	// GetWhois, RemoveWhois 또한 비슷하므로 생략하겠습니다.
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhoisKeyPrefix))
	b := k.cdc.MustMarshal(&whois)
	store.Set(types.WhoisKey(
		whois.Index,
	), b)
}

// GetWhois returns a whois from its index
func (k Keeper) GetWhois(
	ctx sdk.Context,
	index string,
) (val types.Whois, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhoisKeyPrefix))

	b := store.Get(types.WhoisKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveWhois removes a whois from the store
func (k Keeper) RemoveWhois(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhoisKeyPrefix))
	store.Delete(types.WhoisKey(
		index,
	))
}

// GetAllWhois returns all whois
func (k Keeper) GetAllWhois(ctx sdk.Context) (list []types.Whois) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.WhoisKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Whois
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
