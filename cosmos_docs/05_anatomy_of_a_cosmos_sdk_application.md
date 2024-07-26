# Anatomy of a Cosmos SDK Application

SYNOPSIS

이 문서는 코스모스 SDK 애플리케이션의 핵심 부분을 설명하며, 이 문서 전체에서 앱이라는 이름의 플레이스홀더 애플리케이션으로 표시됩니다.

## Node Client

데몬 또는 풀노드 클라이언트는 코스모스 SDK 기반 블록체인의 핵심 프로세스입니다. 네트워크 참여자는 이 프로세스를 실행하여 스테이트 머신을 초기화하고, 다른 풀노드와 연결하고, 새로운 블록이 들어올 때 스테이트 머신을 업데이트합니다.

```
                ^  +-------------------------------+  ^
                |  |                               |  |
                |  |  State-machine = Application  |  |
                |  |                               |  |   Built with Cosmos SDK
                |  |            ^      +           |  |
                |  +----------- | ABCI | ----------+  v
                |  |            +      v           |  ^
                |  |                               |  |
Blockchain Node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   CometBFT
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```

블록체인 풀노드는 바이너리로 표시되며, 일반적으로 "데몬"을 나타내는 -d가 접미사로 붙습니다(예: 앱의 경우 앱드, 가이아의 경우 가이아드). 이 바이너리는 ./cmd/appd/에 있는 간단한 main.go 함수를 실행하여 빌드됩니다. 이 작업은 일반적으로 메이크파일을 통해 이루어집니다. 메인 바이너리가 빌드되면 start 명령을 실행하여 노드를 시작할 수 있습니다. 이 명령은 주로 세 가지 작업을 수행합니다.

- app.go에 정의된 상태 머신의 인스턴스를 생성합니다.

- ~/.app/data 폴더에 저장된 db에서 추출한 최신 상태로 상태 머신을 초기화합니다.
  이 시점에서 상태 머신의 높이는 앱블록높이입니다.

- 새 CometBFT 인스턴스를 생성하고 시작합니다. 무엇보다도 노드는 피어와 핸드셰이크를 수행합니다. 피어로부터 최신 블록 높이를 가져와 로컬 앱 블록 높이보다 큰 경우 이 높이로 동기화하기 위해 블록을 재생합니다. 노드는 제네시스에서 시작되고 CometBFT는 ABCI를 통해 앱에 InitChain 메시지를 전송하여 InitChainer를 트리거합니다.

```
참고 CometBFT 인스턴스를 시작할 때 제네시스 파일은 0 높이이며 제네시스 파일 내의 상태는 블록 높이 1에 커밋됩니다. 노드의 상태를 쿼리할 때 블록 높이 0을 쿼리하면 오류가 반환됩니다.
```

# Core Application File

일반적으로 스테이트 머신의 핵심은 app.go라는 파일에 정의되어 있습니다. 이 파일에는 주로 애플리케이션의 유형 정의와 애플리케이션을 생성하고 초기화하는 함수가 포함되어 있습니다.

## Type Definition of the Application

app.go에서 가장 먼저 정의되는 것은 애플리케이션의 유형입니다. 일반적으로 다음과 같은 부분으로 구성됩니다.

- baseapp에 대한 참조. app.go에 정의된 사용자 정의 애플리케이션은 baseapp의 확장입니다. CometBFT가 애플리케이션에 트랜잭션을 전달하면 app은 baseapp의 메서드를 사용하여 해당 모듈로 트랜잭션을 라우팅합니다. baseapp은 모든 ABCI 메서드와 라우팅 로직을 포함하여 애플리케이션의 핵심 로직 대부분을 구현합니다. \*

- 스토어 키 목록. 전체 상태를 포함하는 스토어는 코스모스 SDK에서 멀티스토어(즉, 스토어들의 스토어)로 구현됩니다. 각 모듈은 멀티스토어에서 하나 또는 여러 개의 스토어를 사용하여 상태의 일부를 유지합니다. 이러한 스토어는 앱 유형에 선언된 특정 키로 액세스할 수 있습니다. 이러한 키는 키퍼와 함께 코스모스 SDK의 객체 기능 모델의 핵심입니다.

- 모듈의 키퍼 목록입니다. 각 모듈은 이 모듈의 저장소에 대한 읽기 및 쓰기를 처리하는 키퍼라는 추상화를 정의합니다. 한 모듈의 키퍼 메서드는 권한이 있는 경우 다른 모듈에서 호출할 수 있으므로 애플리케이션의 유형에 선언되고 다른 모듈에 인터페이스로 내보내어 후자가 권한이 있는 함수에만 액세스할 수 있도록 합니다.

- 앱코덱에 대한 참조입니다. 저장소는 []바이트만 유지할 수 있으므로 애플리케이션의 appCodec은 데이터 구조를 직렬화 및 역직렬화하여 저장하는 데 사용됩니다. 기본 코덱은 프로토콜 버퍼입니다. 레거시Amino 코덱에 대한 참조입니다. 코스모스 SDK의 일부 부분은 위의 앱코덱을 사용하도록 마이그레이션되지 않았으며, 여전히 아미노를 사용하도록 하드코딩되어 있습니다. 다른 부분은 이전 버전과의 호환성을 위해 명시적으로 Amino를 사용합니다. 이러한 이유로 애플리케이션은 여전히 레거시 Amino 코덱에 대한 참조를 보유하고 있습니다.

- 아미노 코덱은 향후 릴리스에서 SDK에서 제거될 예정입니다. 모듈 관리자 및 기본 모듈 관리자에 대한 참조. 모듈 매니저는 애플리케이션의 모듈 목록이 포함된 객체입니다.

- 모듈 관리자는 해당 모듈의 메시지 서비스 및 gRPC 쿼리 서비스를 등록하거나 InitChainer, PreBlocker, BeginBlocker, EndBlocker 등 다양한 함수에 대한 모듈 간 실행 순서를 설정하는 등 해당 모듈과 관련된 작업을 용이하게 합니다.

데모 및 테스트 목적으로 사용되는 코스모스 SDK의 자체 앱인 simapp의 애플리케이션 유형 정의 예시를 참조하세요:

```
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

## 생성자 함수

또한 app.go에는 이전 섹션에서 정의한 유형의 새 애플리케이션을 생성하는 생성자 함수가 정의되어 있습니다. 이 함수가 애플리케이션의 데몬 명령의 시작 명령에 사용되려면 AppCreator 서명을 충족해야 합니다.

```
server/types/app.go
// AppCreator는 다양한 설정을 사용하여
// 애플리케이션을 느리게 초기화할 수 있는 함수입니다.

AppCreator func(log.Logger, dbm.DB, io.Writer, AppOptions) Application
```

link: https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/server/types/app.go#L66-L68

- 새 코덱을 인스턴스화하고 기본 관리자를 사용하여 애플리케이션의 각 모듈의 코덱을 초기화합니다.

- baseapp 인스턴스, 코덱 및 모든 적절한 저장소 키에 대한 참조를 사용하여 새 애플리케이션을 인스턴스화합니다.

- 애플리케이션의 각 모듈의 NewKeeper 함수를 사용하여 애플리케이션 유형에 정의된 모든 키퍼 객체를 인스턴스화합니다. 한 모듈의 NewKeeper는 다른 모듈의 키퍼에 대한 참조가 필요할 수 있으므로 키퍼는 올바른 순서로 인스턴스화해야 합니다.

- 애플리케이션의 각 모듈의 AppModule 객체로 애플리케이션의 모듈 매니저를 인스턴스화합니다.

- 모듈 매니저를 사용하여 애플리케이션의 메시지 서비스, gRPC 쿼리 서비스, 레거시 메시지 경로 및 레거시 쿼리 경로를 초기화합니다. 트랜잭션이 ABCI를 통해 CometBFT에 의해 애플리케이션으로 전달되면 여기에 정의된 경로를 사용하여 해당 모듈의 Msg 서비스로 라우팅됩니다. 마찬가지로 애플리케이션에서 gRPC 쿼리 요청이 수신되면 여기에 정의된 gRPC 경로를 사용하여 해당 모듈의 gRPC 쿼리 서비스로 라우팅됩니다.

- 코스모스 SDK는 레거시 메시지 경로와 레거시 쿼리 경로를 사용하여 각각 라우팅되는 레거시 메시지 및 레거시 CometBFT 쿼리를 계속 지원합니다. 모듈 매니저를 사용하여 애플리케이션의 모듈 불변성을 등록합니다. 불변값은 각 블록이 끝날 때마다 평가되는 변수(예: 토큰의 총 공급량)입니다. 불변값을 확인하는 과정은 불변값 레지스트리라는 특수 모듈을 통해 이루어집니다. 불변값의 값은 모듈에 정의된 예측 값과 같아야 합니다. 값이 예측된 값과 다른 경우, 불변 레지스트리에 정의된 특수 로직이 트리거됩니다(일반적으로 체인이 중지됨). 이 기능은 수정하기 어렵고 오래 지속되는 치명적인 버그를 발견하지 못하도록 하는 데 유용합니다.

* 모듈 관리자를 사용하여 각 애플리케이션 모듈의 InitGenesis, PreBlocker, BeginBlocker, EndBlocker 함수 간의 실행 순서를 설정합니다.
* 모든 모듈이 이러한 함수를 구현하는 것은 아닙니다. 나머지 애플리케이션 파라미터를 설정합니다:

* InitChainer: 애플리케이션을 처음 시작할 때 초기화하는 데 사용됩니다.
* PreBlocker: BeginBlock 전에 호출됩니다.
* BeginBlocker, EndBlocker: 모든 블록의 시작과 끝에 호출됩니다.
* anteHandler: 수수료와 서명 확인을 처리하는 데 사용됩니다.

* 스토어를 마운트합니다.
* 애플리케이션을 반환합니다.

생성자 함수는 앱의 인스턴스만 생성하며, 실제 상태는 노드가 다시 시작되면 ~/.app/data 폴더에서 이월되거나 노드가 처음 시작되면 제네시스 파일에서 생성됩니다. simapp의 애플리케이션 생성자 예시를 참조하세요:

```
// NewSimApp returns a reference to an initialized SimApp.
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// bApp := baseapp.NewBaseApp(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, bApp)
	//
	// bApp.SetMempool(nonceMempool)
	// bApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// bApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to NewBaseApp.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)

	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, crisistypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, consensusparamtypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, circuittypes.StoreKey,
		authzkeeper.StoreKey, nftkeeper.StoreKey, group.StoreKey,
	)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	app := &SimApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress(govtypes.ModuleName).String(), runtime.EventService{})
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(appCodec, runtime.NewKVStoreService(keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	app.MintKeeper = mintkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[minttypes.StoreKey]), app.StakingKeeper, app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.DistrKeeper = distrkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), app.AccountKeeper, app.BankKeeper, app.StakingKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, runtime.NewKVStoreService(keys[slashingtypes.StoreKey]), app.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	invCheckPeriod := cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))
	app.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[crisistypes.StoreKey]), invCheckPeriod,
		app.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(), app.AccountKeeper.AddressCodec())

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.CircuitKeeper = circuitkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[circuittypes.StoreKey]), authtypes.NewModuleAddress(govtypes.ModuleName).String(), app.AccountKeeper.AddressCodec())
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(runtime.NewKVStoreService(keys[authzkeeper.StoreKey]), appCodec, app.MsgServiceRouter(), app.AccountKeeper)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.MsgServiceRouter(), app.AccountKeeper, groupConfig)

	// get skipUpgradeHeights from the app options
	skipUpgradeHeights := map[int64]bool{}
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(keys[upgradetypes.StoreKey]), appCodec, homePath, app.BaseApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper))
	govConfig := govtypes.DefaultConfig()
	/*
		Example of setting gov params:
		govConfig.MaxMetadataLen = 10000
	*/
	govKeeper := govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[govtypes.StoreKey]), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.DistrKeeper, app.MsgServiceRouter(), govConfig, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	app.NFTKeeper = nftkeeper.NewKeeper(runtime.NewKVStoreService(keys[nftkeeper.StoreKey]), appCodec, app.AccountKeeper, app.BankKeeper)

	// create evidence keeper with router
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), app.StakingKeeper, app.SlashingKeeper, app.AccountKeeper.AddressCodec(), runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.ModuleManager = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitKeeper),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			govtypes.ModuleName: gov.NewAppModuleBasic(
				[]govclient.ProposalHandler{
					paramsclient.ProposalHandler,
				},
			),
		})
	app.BasicModuleManager.RegisterLegacyAminoCodec(legacyAmino)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.ModuleManager.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
	)
	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	genesisModuleOrder := []string{
		authtypes.ModuleName, banktypes.ModuleName,
		distrtypes.ModuleName, stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName,
		minttypes.ModuleName, crisistypes.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName, authz.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName, paramstypes.ModuleName, upgradetypes.ModuleName,
		vestingtypes.ModuleName, consensusparamtypes.ModuleName, circuittypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	// Uncomment if you want to set a custom migration order here.
	// app.ModuleManager.SetOrderMigrations(custom order)

	app.ModuleManager.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.ModuleManager.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.configurator` are set.
	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.ModuleManager.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(txConfig)

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain, which comprises of only
	// one decorator: the Transaction Tips decorator. However, some chains do
	// not need it by default, so feel free to comment the next line if you do
	// not need tips.
	// To read more about tips:
	// https://docs.cosmos.network/main/core/tips.html
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/simapp/app.go#L223-L575

# InitChainer

- InitChainer는 제네시스 파일에서 애플리케이션의 상태(즉, 제네시스 계정의 토큰 잔액)를 초기화하는 함수입니다. 이 함수는 애플리케이션이 CometBFT 엔진으로부터 InitChain 메시지를 수신할 때 호출되며, 이는 앱블록높이 == 0에서 노드가 시작될 때(즉, 제네시스에서) 발생합니다.

- 애플리케이션은 생성자에서 SetInitChainer 메서드를 통해 InitChainer를 설정해야 합니다. 일반적으로 InitChainer는 대부분 애플리케이션의 각 모듈의 InitGenesis 함수로 구성됩니다. 이는 모듈 매니저의 InitGenesis 함수를 호출하고, 이 함수는 다시 포함된 각 모듈의 InitGenesis 함수를 호출하는 방식으로 이루어집니다. 모듈의 InitGenesis 함수를 호출해야 하는 순서는 모듈 관리자의 SetOrderInitGenesis 메서드를 사용하여 모듈 관리자에서 설정해야 합니다. 이 작업은 애플리케이션의 생성자에서 수행되며 SetOrderInitGenesis는 SetInitChainer보다 먼저 호출되어야 합니다. simapp의 InitChainer 예시를 참조하세요.

```
simapp/app.go
// InitChainer application update at chain initialization
func (app *SimApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	return app.ModuleManager.InitGenesis(ctx, app.appCodec, genesisState)
}
```

link : https://docs.cosmos.network/v0.50/learn/beginner/app-anatomy

# PreBlocker

There are two semantics around the new lifecycle method:

- It runs before the BeginBlocker of all modules
- It can modify consensus parameters in storage, and signal the caller through the return value.

When it returns ConsensusParamsChanged=true, the caller must refresh the consensus parameter in the finalize context:

```
app.finalizeBlockState.ctx = app.finalizeBlockState.ctx.WithConsensusParams(app.GetConsensusParams())
```

새 ctx는 다른 모든 수명 주기 메서드에 전달되어야 합니다.

# BeginBlocker and EndBlocker

- 코스모스 SDK는 개발자가 애플리케이션의 일부로 코드 자동 실행을 구현할 수 있는 기능을 제공합니다. 이는 BeginBlocker와 EndBlocker라는 두 함수를 통해 구현됩니다. 이 함수는 애플리케이션이 각 블록의 시작과 끝에서 각각 발생하는 CometBFT 합의 엔진으로부터 FinalizeBlock 메시지를 수신할 때 호출됩니다. 애플리케이션은 생성자에서 SetBeginBlocker 및 SetEndBlocker 메서드를 통해 BeginBlocker와 EndBlocker를 설정해야 합니다.

- 일반적으로 BeginBlocker 및 EndBlocker 함수는 대부분 애플리케이션의 각 모듈의 BeginBlock 및 EndBlock 함수로 구성됩니다. 이는 모듈 관리자의 BeginBlock 및 EndBlock 함수를 호출하고, 이 함수는 다시 포함된 각 모듈의 BeginBlock 및 EndBlock 함수를 호출하는 방식으로 이루어집니다. 모듈의 BeginBlock 및 EndBlock 함수를 호출해야 하는 순서는 모듈 관리자에서 각각 SetOrderBeginBlockers 및 SetOrderEndBlockers 메서드를 사용하여 설정해야 합니다. 이 작업은 애플리케이션의 생성자에서 모듈 관리자를 통해 수행되며, SetBeginBlocker 및 SetEndBlocker 함수보다 먼저 SetOrderBeginBlockers 및 SetOrderEndBlockers 메서드를 호출해야 합니다.

- 여담으로, 애플리케이션별 블록체인은 결정론적이라는 점을 기억하는 것이 중요합니다. 개발자는 비긴블록커나 엔드블록커에 비결정성을 도입하지 않도록 주의해야 하며, 가스가 비긴블록커와 엔드블록커 실행 비용을 제약하지 않기 때문에 계산 비용이 너무 많이 들지 않도록 주의해야 합니다. simapp의 비긴블록커와 엔드블록커 함수 예시를 참조하세요.

```
simapp/app.go
func (app *SimApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *SimApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/simapp/app.go#L613-L620

# Register Codec

- EncodingConfig 구조는 app.go 파일의 마지막 중요한 부분입니다. 이 구조의 목표는 앱 전체에서 사용될 코덱을 정의하는 것입니다.

```
simapp/params/encoding.go
// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}
```

다음은 네 가지 필드 각각에 대한 설명입니다:

- 인터페이스 레지스트리: InterfaceRegistry는 Protobuf 코덱에서 google.protobuf.Any를 사용하여 인코딩 및 디코딩(우리는 "언패킹"이라고도 함)되는 인터페이스를 처리하는 데 사용됩니다. Any는 type_url(인터페이스를 구현하는 구체적인 타입의 이름)과 value(인코딩된 바이트)를 포함하는 구조체로 생각할 수 있습니다. InterfaceRegistry는 Any에서 안전하게 언패킹할 수 있는 인터페이스와 구현을 등록하는 메커니즘을 제공합니다. 각 애플리케이션 모듈은 모듈의 자체 인터페이스 및 구현을 등록하는 데 사용할 수 있는 RegisterInterfaces 메서드를 구현합니다. Any에 대한 자세한 내용은 ADR-019에서 확인할 수 있습니다. 더 자세히 살펴보기 위해 Cosmos SDK는 gogoprotobuf라는 Protobuf 사양의 구현을 사용합니다. 기본적으로 Any의 gogo protobuf 구현은 전역 유형 등록을 사용하여 Any에 패킹된 값을 구체적인 Go 유형으로 디코딩합니다. 이로 인해 종속성 트리의 악의적인 모듈이 글로벌 protobuf 레지스트리에 유형을 등록하여 type_url 필드에서 해당 유형을 참조한 트랜잭션에 의해 로드 및 마샬링 해제될 수 있는 취약점이 발생하게 됩니다.

* 자세한 내용은 ADR-019를 참조하십시오.
  ADR-019 : https://docs.cosmos.network/v0.50/build/architecture/adr-019-protobuf-state-encoding
  google.protobuf.Any : https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/any.proto
  gogoprotobuf : https://github.com/cosmos/gogoproto
  gogoprotobuf implement : https://pkg.go.dev/github.com/cosmos/gogoproto/types

- 코덱: 코스모스 SDK 전체에서 사용되는 기본 코덱입니다. 상태를 인코딩 및 디코딩하는 데 사용되는 BinaryCodec과 사용자에게 데이터를 출력하는 데 사용되는 JSONCodec으로 구성됩니다(예: CLI에서). 기본적으로 SDK는 Protobuf를 코덱으로 사용합니다.

- TxConfig: TxConfig는 클라이언트가 애플리케이션이 정의한 구체적인 트랜잭션 유형을 생성하는 데 활용할 수 있는 인터페이스를 정의합니다. 현재 SDK는 두 가지 트랜잭션 유형을 처리합니다: SIGN_MODE_DIRECT(무선 인코딩으로 Protobuf 바이너리를 사용)와 SIGN_MODE_LEGACY_AMINO_JSON(Amino에 의존)입니다. 트랜잭션에 대한 자세한 내용은 여기를 참조하세요.

- 아미노: 이전 버전과의 호환성을 위해 코스모스 SDK의 일부 레거시 부분은 여전히 아미노를 사용합니다. 각 모듈은 Amino 내에서 모듈의 특정 유형을 등록하기 위해 RegisterLegacyAmino 메서드를 노출합니다. 이 아미노 코덱은 앱 개발자가 더 이상 사용해서는 안 되며, 향후 릴리스에서 제거될 예정입니다.

애플리케이션은 자체 인코딩 구성을 생성해야 합니다. 심앱의 심앱파람스.인코딩설정 예시를 참조하세요:

```
simapp/params/encoding.go
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}


```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/simapp/params/encoding.go#L11-L16

# Modules

- 모듈은 코스모스 SDK 애플리케이션의 핵심이자 영혼입니다. 모듈은 스테이트 머신 내에 중첩된 스테이트 머신으로 간주할 수 있습니다. 트랜잭션이 기본 CometBFT 엔진에서 ABCI를 통해 애플리케이션으로 전달되면, 베이스앱은 이를 처리하기 위해 적절한 모듈로 라우팅합니다. 이 패러다임에서는 개발자가 필요로 하는 대부분의 모듈이 이미 존재하는 경우가 많기 때문에 복잡한 상태 머신을 쉽게 구축할 수 있습니다. 개발자의 경우 코스모스 SDK 애플리케이션 구축과 관련된 대부분의 작업은 아직 존재하지 않는 애플리케이션에 필요한 사용자 지정 모듈을 구축하고 이미 존재하는 모듈과 통합하여 하나의 일관된 애플리케이션으로 만드는 것입니다. 애플리케이션 디렉토리에서 표준 관행은 x/ 폴더에 모듈을 저장하는 것입니다(이미 빌드된 모듈이 있는 Cosmos SDK의 x/ 폴더와 혼동하지 마세요).

## Application Module Interface

- 모듈은 코스모스 SDK에 정의된 인터페이스인 AppModuleBasic과 AppModule을 구현해야 합니다. 전자는 코덱과 같이 모듈의 기본 비종속 요소를 구현하고, 후자는 대부분의 모듈 메서드(다른 모듈의 키퍼에 대한 참조가 필요한 메서드 포함)를 처리합니다. AppModule과 AppModuleBasic 유형은 모두 관례에 따라 module.go라는 파일에 정의됩니다. AppModule은 모듈을 일관된 애플리케이션으로 쉽게 구성할 수 있도록 모듈에 유용한 메서드 모음을 노출합니다. 이러한 메서드는 애플리케이션의 모듈 컬렉션을 관리하는 모듈 관리자에서 호출됩니다.

## Msg Services

- 각 애플리케이션 모듈은 메시지를 처리하는 Msg 서비스 1개와 쿼리를 처리하는 gRPC 쿼리 서비스 1개 등 두 개의 Protobuf 서비스를 정의합니다. 모듈을 상태 머신으로 간주하면 Msg 서비스는 상태 전환 RPC 메서드의 집합입니다. 각 Protobuf Msg 서비스 메서드는 Protobuf 요청 유형과 1:1로 연결되며, 이 요청 유형은 반드시 sdk.Msg 인터페이스를 구현해야 합니다. sdk.Msgs는 트랜잭션에 번들로 제공되며 각 트랜잭션에는 하나 또는 여러 개의 메시지가 포함됩니다. 전체 노드에서 유효한 트랜잭션 블록을 수신하면 CometBFT는 DeliverTx를 통해 각 트랜잭션을 애플리케이션에 릴레이합니다. 그러면 애플리케이션이 트랜잭션을 처리합니다:

* 트랜잭션을 수신하면 애플리케이션은 먼저 []바이트에서 마샬링을 해제합니다.

* 그런 다음 수수료 지불 및 서명과 같은 트랜잭션에 대한 몇 가지 사항을 확인한 후 트랜잭션에 포함된 메시지를 추출합니다.

* sdk.Msgs는 Protobuf Anys를 사용하여 인코딩됩니다. 각 Any의 type_url을 분석하여 baseapp의 msgServiceRouter는 sdk.Msg를 해당 모듈의 Msg 서비스로 라우팅합니다.

* 메시지가 성공적으로 처리되면 상태가 업데이트됩니다.

- 자세한 내용은 트랜잭션 수명 주기를 참조하세요. 모듈 개발자는 자체 모듈을 빌드할 때 사용자 정의 메시지 서비스를 만듭니다. 일반적인 방법은 tx.proto 파일에 Msg Protobuf 서비스를 정의하는 것입니다. 예를 들어, x/bank 모듈은 토큰을 전송하는 두 가지 메서드가 있는 서비스를 정의합니다:

proto/cosmos/bank/v1beta1/tx.proto

```
// Msg defines the bank Msg service.
service Msg {
  option (cosmos.msg.v1.service) = true;

  // Send defines a method for sending coins from one account to another account.
  rpc Send(MsgSend) returns (MsgSendResponse);

  // MultiSend defines a method for sending coins from some accounts to other accounts.
  rpc MultiSend(MsgMultiSend) returns (MsgMultiSendResponse);

  // UpdateParams defines a governance operation for updating the x/bank module parameters.
  // The authority is defined in the keeper.
  //
  // Since: cosmos-sdk 0.47
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);

  // SetSendEnabled is a governance operation for setting the SendEnabled flag
  // on any number of Denoms. Only the entries to add or update should be
  // included. Entries that already exist in the store, but that aren't
  // included in this message, will be left unchanged.
  //
  // Since: cosmos-sdk 0.47
  rpc SetSendEnabled(MsgSetSendEnabled) returns (MsgSetSendEnabledResponse);
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/proto/cosmos/bank/v1beta1/tx.proto#L13-L36

서비스 메서드는 모듈 상태를 업데이트하기 위해 키퍼를 사용합니다. 각 모듈은 AppModule 인터페이스의 일부로 RegisterServices 메서드도 구현해야 합니다. 이 메서드는 생성된 Protobuf 코드에서 제공하는 RegisterMsgServer 함수를 호출해야 합니다.

## gRPC Query Services

- gRPC 쿼리 서비스를 사용하면 사용자가 gRPC를 사용하여 상태를 쿼리할 수 있습니다. 기본적으로 활성화되어 있으며 app.toml 내의 grpc.enable 및 grpc.address 필드에서 구성할 수 있습니다.

* gRPC 쿼리 서비스는 모듈의 Protobuf 정의 파일, 특히 query.proto 내에 정의됩니다. query.proto 정의 파일은 단일 쿼리 Protobuf 서비스를 노출합니다. 각 gRPC 쿼리 엔드포인트는 쿼리 서비스 내부의 rpc 키워드로 시작하는 서비스 메서드에 해당합니다.

* Protobuf는 모든 서비스 메서드를 포함하는 각 모듈에 대한 QueryServer 인터페이스를 생성합니다. 그런 다음 모듈의 키퍼는 각 서비스 메서드의 구체적인 구현을 제공함으로써 이 QueryServer 인터페이스를 구현해야 합니다. 이 구체적인 구현은 해당 gRPC 쿼리 엔드포인트의 핸들러입니다.

* 마지막으로, 각 모듈은 AppModule 인터페이스의 일부로 RegisterServices 메서드도 구현해야 합니다. 이 메서드는 생성된 Protobuf 코드에서 제공하는 RegisterQueryServer 함수를 호출해야 합니다.

## Keeper

- 키퍼는 해당 모듈의 저장소를 관리하는 게이트키퍼입니다.

- 모듈의 저장소를 읽거나 쓰려면 반드시 키퍼의 메서드 중 하나를 거쳐야 합니다. 이는 코스모스 SDK의 객체 기능 모델에 의해 보장됩니다. 스토어에 대한 키를 보유한 객체만 스토어에 액세스할 수 있으며, 모듈의 키퍼만 모듈의 스토어에 대한 키를 보유해야 합니다.

- 키퍼는 일반적으로 keeper.go라는 파일에 정의됩니다. 이 파일에는 키퍼의 타입 정의와 메서드가 포함되어 있습니다. 키퍼 타입 정의는 일반적으로 다음과 같이 구성됩니다:

유형 정의와 함께 keeper.go 파일의 다음 중요한 구성 요소는 키퍼의 생성자 함수인 NewKeeper입니다. 이 함수는 코덱을 사용하여 위에 정의된 유형의 새 키퍼를 인스턴스화하고 키를 저장하며 다른 모듈의 키퍼를 매개변수로 참조할 수 있습니다. NewKeeper 함수는 애플리케이션의 생성자에서 호출됩니다. 파일의 나머지 부분은 키퍼의 메서드를 정의하며, 주로 게터와 세터입니다.

# Command-Line, gRPC Services and REST Interfaces

- 각 모듈은 애플리케이션의 인터페이스를 통해 최종 사용자에게 노출되는 명령줄 명령, gRPC 서비스 및 REST 경로를 정의합니다. 이를 통해 최종 사용자는 모듈에 정의된 유형의 메시지를 만들거나 모듈에서 관리하는 상태의 하위 집합을 쿼리할 수 있습니다.

## CLI

- 일반적으로 모듈과 관련된 명령은 모듈의 폴더에 있는 client/cli라는 폴더에 정의됩니다. CLI는 명령을 트랜잭션과 쿼리라는 두 가지 범주로 나누며, 각각 client/cli/tx.go 및 client/cli/query.go에 정의되어 있습니다. 두 명령 모두 Cobra 라이브러리 위에 구축됩니다:

* 트랜잭션 명령을 사용하면 사용자가 새 트랜잭션을 생성하여 블록에 포함시키고 최종적으로 상태를 업데이트할 수 있습니다. 모듈에 정의된 각 메시지 유형에 대해 하나의 명령을 만들어야 합니다. 이 명령은 최종 사용자가 제공한 매개변수로 메시지의 생성자를 호출하고 트랜잭션으로 래핑합니다. Cosmos SDK는 서명 및 기타 트랜잭션 메타데이터 추가를 처리합니다.

* 쿼리는 사용자가 모듈에서 정의한 상태의 하위 집합을 쿼리할 수 있도록 합니다. 쿼리 명령은 애플리케이션의 쿼리 라우터로 쿼리를 전달하고, 쿼리 라우터는 제공된 쿼리 경로 매개변수에 따라 적절한 쿼리어로 쿼리를 라우팅합니다.

## gRPC

- gRPC는 여러 언어로 지원되는 최신 오픈소스 고성능 RPC 프레임워크입니다. 외부 클라이언트(예: 지갑, 브라우저, 기타 백엔드 서비스)가 노드와 상호 작용하는 데 권장되는 방식입니다. 각 모듈은 모듈의 Protobuf query.proto 파일에 정의된 서비스 메서드라고 하는 gRPC 엔드포인트를 노출할 수 있습니다. 서비스 메서드는 이름, 입력 인수, 출력 응답으로 정의됩니다. 그런 다음 모듈은 다음 작업을 수행해야 합니다:

* 클라이언트 gRPC 요청을 모듈 내부의 올바른 핸들러에 연결하기 위해 AppModuleBasic에 RegisterGRPCGatewayRoutes 메서드를 정의합니다.

* 각 서비스 메서드에 해당하는 핸들러를 정의합니다. 핸들러는 gRPC 요청을 처리하는 데 필요한 핵심 로직을 구현하며, keeper/grpc_query.go 파일에 위치합니다.

## gRPC-gateway REST Endpoints

- 일부 외부 클라이언트는 gRPC 사용을 원하지 않을 수 있습니다. 이 경우 코스모스 SDK는 각 gRPC 서비스를 해당 REST 엔드포인트로 노출하는 gRPC 게이트웨이 서비스를 제공합니다. 자세한 내용은 grpc-gateway 설명서를 참조하세요.

- REST 엔드포인트는 Protobuf 어노테이션을 사용하여 gRPC 서비스와 함께 Protobuf 파일에 정의됩니다. REST 쿼리를 노출하려는 모듈은 rpc 메서드에 google.api.http 어노테이션을 추가해야 합니다. 기본적으로 SDK에 정의된 모든 REST 엔드포인트는 /cosmos/ 접두사로 시작하는 URL을 갖습니다.

- Cosmos SDK는 이러한 REST 엔드포인트에 대한 Swagger 정의 파일을 생성하는 개발 엔드포인트도 제공합니다. 이 엔드포인트는 app.toml 구성 파일의 api.swagger 키 아래에서 활성화할 수 있습니다.

# Application Interface

- 인터페이스를 통해 최종 사용자는 풀노드 클라이언트와 상호작용할 수 있습니다. 즉, 풀노드에서 데이터를 쿼리하거나 풀노드가 릴레이하고 최종적으로 블록에 포함될 새 트랜잭션을 생성하여 전송할 수 있습니다. 주요 인터페이스는 명령줄 인터페이스입니다. 코스모스 SDK 애플리케이션의 CLI는 애플리케이션에서 사용하는 각 모듈에 정의된 CLI 명령을 모아 구축됩니다. 애플리케이션의 CLI는 데몬(예: 앱드)과 동일하며 앱드/main.go라는 파일에 정의됩니다. 이 파일에는 다음이 포함됩니다:

* appd 인터페이스 클라이언트를 빌드하기 위해 실행되는 main() 함수입니다. 이 함수는 각 명령을 준비하여 빌드하기 전에 rootCmd에 추가합니다. 이 함수는 appd의 루트에서 status, keys, config와 같은 일반 명령, 쿼리 명령, tx 명령, rest-server와 같은 쿼리 명령을 추가합니다.

* 쿼리 명령은 queryCmd 함수를 호출하여 추가합니다. 이 함수는 애플리케이션의 각 모듈에 정의된 쿼리 명령(main() 함수에서 sdk.ModuleClients 배열로 전달됨)과 블록 또는 유효성 검사기 쿼리와 같은 다른 하위 수준의 쿼리 명령이 포함된 Cobra 명령을 반환합니다. 쿼리 명령은 CLI의 앱드 쿼리 [쿼리] 명령을 사용하여 호출합니다.

* 트랜잭션 명령은 txCmd 함수를 호출하여 추가합니다. 이 함수는 쿼리Cmd와 마찬가지로 각 애플리케이션의 모듈에 정의된 tx 명령과 트랜잭션 서명 또는 브로드캐스팅과 같은 하위 수준의 tx 명령이 포함된 코브라 명령을 반환합니다. tx 명령은 CLI의 appd tx [tx] 명령을 사용하여 호출합니다.

코스모스 허브에서 애플리케이션의 기본 명령줄 파일 예시를 참조하세요.
link : https://github.com/cosmos/gaia

- cmd/gaiad/cmd/root.go

```
// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := gaia.MakeTestEncodingConfig()
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(gaia.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "gaiad",
		Short: "Stargate Cosmos Hub App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err = client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customTemplate, customGaiaConfig := initAppConfig()
			customTMConfig := initTendermintConfig()
			return server.InterceptConfigsPreRunHandler(cmd, customTemplate, customGaiaConfig, customTMConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd, encodingConfig
}
```

# Dependencies and Makefile

- 이 섹션은 개발자가 종속성 관리자와 프로젝트 빌드 방법을 자유롭게 선택할 수 있으므로 선택 사항입니다. 하지만 현재 버전 관리를 위해 가장 많이 사용되는 프레임워크는 go.mod입니다. 이 프레임워크는 애플리케이션 전체에서 사용되는 각 라이브러리를 올바른 버전으로 가져오도록 보장합니다. 다음은 예시로 제공되는 Cosmos Hub의 go.mod입니다.



* 애플리케이션 빌드에는 일반적으로 메이크파일이 사용됩니다. 메이크파일은 주로 애플리케이션의 두 가지 엔트리포인트인 노드 클라이언트 및 애플리케이션 인터페이스를 빌드하기 전에 go.mod가 실행되도록 합니다. 다음은 코스모스 허브 메이크파일의 예시입니다.