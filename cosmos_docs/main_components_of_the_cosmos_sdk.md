# Main Components of the Cosmos SDK

Cosmos SDK는 CometBFT를 기반으로 보안 상태 머신을 쉽게 개발할 수 있는 프레임워크입니다. Cosmos SDK의 핵심은 Golang에서 ABCI를 상용구로 구현한 것입니다. 데이터를 보존하는 멀티스토어와 트랜잭션을 처리하는 라우터가 함께 제공됩니다. 다음은 DeliverTx를 통해 CometBFT에서 전송된 트랜잭션이 Cosmos SDK 위에 구축된 애플리케이션에서 어떻게 처리되는지 간략하게 보여줍니다.

- CometBFT 합의 엔진에서 받은 트랜잭션을 디코딩합니다(CometBFT는 []바이트만 처리한다는 점을 기억하세요).
- 거래에서 메시지를 추출하고 기본적인 건전성 검사를 수행합니다.
- 각 메시지를 적절한 모듈로 라우팅하여 처리할 수 있도록 합니다.
- 상태 변경 사항을 커밋합니다.

# baseapp

baseapp은 코스모스 SDK 애플리케이션의 상용구 구현입니다. 여기에는 기본 합의 엔진과의 연결을 처리하기 위한 ABCI 구현이 함께 제공됩니다. 일반적으로 코스모스 SDK 애플리케이션은 app.go에 임베드하여 베이스앱을 확장합니다. 다음은 코스모스 SDK 데모 앱인 simapp의 예시입니다:

```
// SimApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type SimApp struct {
	*baseapp.BaseApp
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys  map[string]*storetypes.KVStoreKey
	tkeys map[string]*storetypes.TransientStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper

	// the module manager
	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}
```

simapp : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/simapp/app.go#L170-L212

baseapp의 목표는 스토어와 확장 가능한 상태 머신 사이에 안전한 인터페이스를 제공하는 동시에 상태 머신에 대해 가능한 한 적게 정의하는 것입니다(ABCI에 충실). baseapp에 대해 자세히 알아보려면 여기를 클릭하세요.

baseapp : https://docs.cosmos.network/v0.50/learn/advanced/baseapp

# Multistore

코스모스 SDK는 상태 지속을 위한 멀티스토어를 제공합니다. 멀티스토어를 사용하면 개발자는 원하는 수의 KV스토어를 선언할 수 있습니다. 이러한 KVStore는 []바이트 유형만 값으로 허용하므로 모든 사용자 정의 구조는 저장하기 전에 코덱을 사용하여 마샬링해야 합니다. 멀티스토어 추상화는 상태를 각각 자체 모듈에서 관리하는 별개의 구획으로 나누는 데 사용됩니다. 멀티스토어에 대해 자세히 알아보려면 여기를 클릭하세요.

# Modules

코스모스 SDK의 강점은 모듈성에 있습니다. Cosmos SDK 애플리케이션은 상호 운용 가능한 모듈 모음을 모아 구축됩니다. 각 모듈은 상태의 하위 집합을 정의하고 자체 메시지/트랜잭션 프로세서를 포함하며, Cosmos SDK는 각 메시지를 해당 모듈로 라우팅합니다. 다음은 유효한 블록에서 트랜잭션이 수신될 때 각 풀 노드의 애플리케이션에서 트랜잭션을 처리하는 방법을 단순화한 그림입니다:

```
                                      +
                                      |
                                      |  Transaction relayed from the full-node's
                                      |  CometBFT engine to the node's application
                                      |  via DeliverTx
                                      |
                                      |
                +---------------------v--------------------------+
                |                 APPLICATION                    |
                |                                                |
                |     Using baseapp's methods: Decode the Tx,    |
                |     extract and route the message(s)           |
                |                                                |
                +---------------------+--------------------------+
                                      |
                                      |
                                      |
                                      +---------------------------+
                                                                  |
                                                                  |
                                                                  |  Message routed to
                                                                  |  the correct module
                                                                  |  to be processed
                                                                  |
                                                                  |
+----------------+  +---------------+  +----------------+  +------v----------+
|                |  |               |  |                |  |                 |
|  AUTH MODULE   |  |  BANK MODULE  |  | STAKING MODULE |  |   GOV MODULE    |
|                |  |               |  |                |  |                 |
|                |  |               |  |                |  | Handles message,|
|                |  |               |  |                |  | Updates state   |
|                |  |               |  |                |  |                 |
+----------------+  +---------------+  +----------------+  +------+----------+
                                                                  |
                                                                  |
                                                                  |
                                                                  |
                                       +--------------------------+
                                       |
                                       | Return result to CometBFT
                                       | (0=Ok, 1=Err)
                                       v
```

각 모듈은 작은 상태 머신으로 볼 수 있습니다. 개발자는 모듈이 처리하는 상태의 하위 집합과 상태를 수정하는 사용자 정의 메시지 유형을 정의해야 합니다(참고: 메시지는 baseapp에 의해 트랜잭션에서 추출됩니다). 일반적으로 각 모듈은 멀티스토어에서 자체 KVStore를 선언하여 자신이 정의한 상태의 하위 집합을 유지합니다. 대부분의 개발자는 자체 모듈을 빌드할 때 다른 타사 모듈에 액세스해야 합니다. Cosmos SDK는 개방형 프레임워크이므로 일부 모듈은 악의적일 수 있으므로 모듈 간 상호 작용을 추론할 수 있는 보안 원칙이 필요합니다. 이러한 원칙은 객체 기능을 기반으로 합니다. 실제로는 각 모듈이 다른 모듈에 대한 액세스 제어 목록을 보관하는 대신, 각 모듈은 다른 모듈에 전달하여 미리 정의된 기능 집합을 부여할 수 있는 키퍼라는 특수 객체를 구현한다는 의미입니다.

코스모스 SDK 모듈은 코스모스 SDK의 x/ 폴더에 정의되어 있습니다. 일부 핵심 모듈은 다음과 같습니다:

- x/auth: 계정과 서명을 관리하는 데 사용됩니다.
- x/bank: 토큰과 토큰 전송을 활성화하는 데 사용됩니다.
- x/staking + x/slashing: 지분 증명 블록체인을 구축하는 데 사용됩니다.

누구나 앱에서 사용할 수 있는 x/의 기존 모듈 외에도 코스모스 SDK를 사용하면 자신만의 맞춤형 모듈을 만들 수 있습니다. 튜토리얼에서 그 예시를 확인할 수 있습니다.

link : https://tutorials.cosmos.network/
