# Gas and Fees

개요 이 문서에서는 코스모스 SDK 애플리케이션 내에서 가스 및 수수료를 처리하는 기본 전략을 설명합니다.

# Introduction to Gas and Fees

Cosmos SDK에서 가스는 실행 중 리소스 소비를 추적하는 데 사용되는 특수 단위입니다. 가스는 일반적으로 스토어에 읽고 쓸 때마다 소비되지만, 비용이 많이 드는 연산을 수행해야 할 때에도 소비될 수 있습니다. 가스는 두 가지 주요 목적으로 사용됩니다.

- 블록이 너무 많은 리소스를 소비하지 않고 마무리되는지 확인하는 것입니다. 이는 기본적으로 블록 가스 미터를 통해 Cosmos SDK에서 구현됩니다.

- 최종 사용자의 스팸 및 악용을 방지합니다. 이를 위해 메시지 실행 중에 소비되는 가스에 일반적으로 요금이 책정되어 수수료가 발생합니다(수수료 = 가스 \* 가스 가격). 수수료는 일반적으로 메시지 발신자가 지불해야 합니다. 스팸을 방지하는 다른 방법(예: 대역폭 체계)이 있을 수 있으므로 Cosmos SDK는 기본적으로 가스 가격 책정을 강제하지 않습니다. 하지만 대부분의 애플리케이션은 안티핸들러를 사용하여 스팸을 방지하기 위한 요금 메커니즘을 구현합니다.

## Gas Meter

코스모스 SDK에서 가스는 uint64의 간단한 별칭이며, 가스 미터라는 객체에 의해 관리됩니다. 가스 미터는 GasMeter 인터페이스를 구현합니다.

store/types/gas.go

```
// GasMeter interface to track gas consumption
type GasMeter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	GasRemaining() Gas
	Limit() Gas
	ConsumeGas(amount Gas, descriptor string)
	RefundGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool
	String() string
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/store/types/gas.go#L40-L51

참고:

- GasConsumed()는 가스 미터 인스턴스에서 소비한 가스 양을 반환합니다.
- GasConsumedToLimit()는 가스 미터 인스턴스에서 소비한 가스 양 또는 한계에 도달한 경우 한도를 반환합니다.
- GasRemaining()은 가스 미터에 남은 가스 양을 반환합니다.
- Limit()는 가스 미터 인스턴스의 한도를 반환합니다. 가스 미터가 무한대인 경우 0입니다.
- ConsumeGas(금액 가스, 설명자 문자열)는 제공된 가스의 양을 소비합니다. 가스가 넘치면 설명자 메시지와 함께 패닉 상태가 됩니다. 가스 미터가 무한하지 않은 경우, 소비된 가스가 한도를 초과하면 패닉 상태가 됩니다.
- RefundGas()는 소비된 가스에서 주어진 양을 차감합니다. 이 기능을 사용하면 트랜잭션 또는 블록 가스 풀에 가스를 환불할 수 있으므로 EVM 호환 체인이 go-ethereum StateDB 인터페이스를 완벽하게 지원할 수 있습니다.
- IsPastLimit()는 가스 미터 인스턴스가 소비한 가스 양이 한도를 완전히 초과하면 참을 반환하고 그렇지 않으면 거짓을 반환합니다.
- IsOutOfGas()는 가스 미터 인스턴스가 소비한 가스 양이 한도를 초과하거나 같으면 참을 반환하고 그렇지 않으면 거짓을 반환합니다.

가스 계량기는 일반적으로 CTX에 보관되며, 가스 소비는 다음과 같은 패턴으로 이루어집니다:

```
ctx.GasMeter().ConsumeGas(amount, "description")
```

기본적으로 Cosmos SDK는 메인 가스 계량기와 블록 가스 계량기라는 두 가지 가스 계량기를 사용합니다.

## Main Gas Meter

ctx.GasMeter()는 애플리케이션의 메인 가스 미터입니다. 메인 가스 미터는 setFinalizeBlockState를 통해 FinalizeBlock에서 초기화된 다음, 상태 전환으로 이어지는 실행 시퀀스, 즉 원래 FinalizeBlock에 의해 트리거된 실행 시퀀스 동안 가스 소비량을 추적합니다. 각 트랜잭션 실행이 시작될 때, 트랜잭션별 가스 소비량을 추적할 수 있도록 AnteHandler에서 메인 가스 미터를 0으로 설정해야 합니다. 가스 소비는 일반적으로 모듈 개발자가 BeginBlocker, EndBlocker 또는 Msg 서비스에서 수동으로 수행할 수 있지만 대부분의 경우 스토어에 읽기 또는 쓰기가 있을 때마다 자동으로 수행됩니다. 이 자동 가스 소비 로직은 GasKv라는 특수 스토어에서 구현됩니다.

## Block Gas Meter

ctx.BlockGasMeter()는 블록당 가스 소비량을 추적하고 특정 한도를 초과하지 않도록 하는 데 사용되는 가스 미터입니다.
제네시스 단계에서는 초기화 트랜잭션을 수용하기 위해 가스 소비량이 무제한으로 허용됩니다.

```
app.finalizeBlockState.SetContext(app.finalizeBlockState.Context().WithBlockGasMeter(storetypes.NewInfiniteGasMeter()))
```

제네시스 블록에 이어 블록 가스 미터는 SDK에 의해 유한한 값으로 설정됩니다. 이러한 전환은 컨센서스 엔진(예: CometBFT)이 RequestFinalizeBlock 함수를 호출하고, 이는 다시 SDK의 FinalizeBlock 메서드를 트리거함으로써 촉진됩니다. FinalizeBlock 내에서 내부FinalizeBlock이 실행되어 필요한 상태 업데이트와 함수 실행을 수행합니다. 그런 다음 유한한 한도로 각각 초기화된 블록 가스 미터가 트랜잭션 실행 컨텍스트에 통합되어 가스 소비가 블록의 가스 한도를 초과하지 않도록 보장하고 각 블록이 끝날 때 리셋됩니다.

코스모스 SDK 내의 모듈은 실행 중 언제든 ctx를 활용하여 블록 가스를 소비할 수 있습니다. 이 가스 소비는 주로 상태 읽기/쓰기 작업과 트랜잭션 처리 중에 발생합니다. ctx.BlockGasMeter()를 통해 액세스할 수 있는 블록 가스 미터는 블록 내 총 가스 사용량을 모니터링하여 과도한 계산을 방지하기 위해 가스 제한을 시행합니다. 이렇게 하면 생성 후 첫 번째 블록부터 시작하여 블록별로 가스 한도를 준수할 수 있습니다.

```
gasMeter := app.getBlockGasMeter(app.finalizeBlockState.Context())
app.finalizeBlockState.SetContext(app.finalizeBlockState.Context().WithBlockGasMeter(gasMeter))
```

## AnteHandler

안테핸들러는 트랜잭션의 각 sdk.Msg에 대한 Protobuf Msg 서비스 메서드 전에 CheckTx 및 FinalizeBlock 동안 모든 트랜잭션에 대해 실행됩니다. 안테핸들러는 핵심 Cosmos SDK가 아니라 모듈에 구현되어 있습니다. 즉, 오늘날 대부분의 애플리케이션은 인증 모듈에 정의된 기본 구현을 사용합니다. 일반적인 코스모스 SDK 애플리케이션에서 anteHandler가 수행하는 작업은 다음과 같습니다:

- 트랜잭션이 올바른 유형인지 확인합니다. 트랜잭션 유형은 anteHandler를 구현하는 모듈에 정의되어 있으며 트랜잭션 인터페이스를 따릅니다:

types/tx_msg.go

```
Tx interface {
	HasMsgs

	// GetMsgsV2 gets the transaction's messages as google.golang.org/protobuf/proto.Message's.
	GetMsgsV2() ([]protov2.Message, error)
}
```

이를 통해 개발자는 애플리케이션의 트랜잭션에 다양한 유형을 사용할 수 있습니다. 기본 인증 모듈에서 기본 트랜잭션 유형은 Tx입니다:
proto/cosmos/tx/v1beta1/tx.proto

```
// Tx is the standard type used for broadcasting transactions.
message Tx {
  // body is the processable content of the transaction
  TxBody body = 1;

  // auth_info is the authorization related content of the transaction,
  // specifically signers, signer modes and fee
  AuthInfo auth_info = 2;

  // signatures is a list of signatures that matches the length and order of
  // AuthInfo's signer_infos to allow connecting signature meta information like
  // public key and signing mode by position.
  repeated bytes signatures = 3;
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/proto/cosmos/tx/v1beta1/tx.proto#L14-L27

* 트랜잭션에 포함된 각 메시지의 서명을 확인합니다. 각 메시지는 한 명 또는 여러 명의 발신자가 서명해야 하며, 이러한 서명은 anteHandler에서 확인되어야 합니다. 

* CheckTx 중에 트랜잭션에 제공된 가스 가격이 로컬 최소 가스 가격보다 큰지 확인합니다(참고로 가스 가격은 다음 공식에서 공제할 수 있습니다: 수수료 = 가스 \* 가스 가격). 최소 가스 가격은 각 풀 노드에 로컬한 파라미터이며 CheckTx 중에 최소 수수료가 없는 거래를 삭제하는 데 사용됩니다. 이렇게 하면 멤풀에 가비지 트랜잭션이 스팸으로 유입되지 않습니다. 

* 트랜잭션 발신자에게 수수료를 감당할 수 있는 충분한 자금이 있는지 확인합니다. 최종 사용자가 트랜잭션을 생성할 때 수수료, 가스, 가스 가격의 3가지 매개변수 중 2가지를 표시해야 합니다(세 번째 매개변수는 암시적). 이는 트랜잭션을 실행하기 위해 노드에 얼마를 지불할 의향이 있는지를 나타냅니다. 제공된 가스 값은 나중에 사용할 수 있도록 GasWanted라는 파라미터에 저장됩니다. 

* GasWanted의 한도를 0으로 설정하여 newCtx.GasMeter를 0으로 설정합니다. 이 단계는 트랜잭션이 무한 가스를 소비할 수 없도록 할 뿐만 아니라 각 트랜잭션 사이에 ctx.GasMeter가 재설정되도록 하기 때문에 매우 중요합니다(anteHandler가 실행된 후 ctx가 newCtx로 설정되고 트랜잭션이 실행될 때마다 anteHandler가 실행됩니다).


