# Query Lifecycle

- 개요
  이 문서에서는 사용자 인터페이스에서 애플리케이션 스토어에 이르는 코스모스 SDK 애플리케이션의 쿼리 수명 주기에 대해 설명합니다. 쿼리를 MyQuery라고 합니다.

# Query Creation

쿼리는 애플리케이션의 최종 사용자가 인터페이스를 통해 전체 노드에서 처리하는 정보 요청입니다. 사용자는 애플리케이션의 스토어 또는 모듈에서 직접 네트워크, 애플리케이션 자체 및 애플리케이션 상태에 대한 정보를 쿼리할 수 있습니다. 쿼리는 트랜잭션(여기에서 수명 주기 보기)과 다르며, 특히 상태 전환을 트리거하지 않기 때문에 컨센서스가 필요하지 않고 하나의 풀 노드에서 완전히 처리할 수 있다는 점에서 다릅니다. 쿼리 수명 주기를 설명하기 위해 MyQuery라는 쿼리가 simapp이라는 애플리케이션의 특정 위임자 주소가 만든 위임 목록을 요청한다고 가정해 보겠습니다. 예상대로 스테이킹 모듈이 이 쿼리를 처리합니다. 하지만 먼저 사용자가 MyQuery를 만들 수 있는 몇 가지 방법이 있습니다.

## CLI

애플리케이션의 기본 인터페이스는 명령줄 인터페이스입니다. 사용자는 전체 노드에 연결하여 자신의 컴퓨터에서 직접 CLI를 실행하며, CLI는 전체 노드와 직접 상호 작용합니다. 터미널에서 MyQuery를 만들려면 다음 명령을 입력합니다:

```
simd query staking delegations <delegatorAddress>
```

이 쿼리 명령은 스테이킹 모듈 개발자가 정의하고 애플리케이션 개발자가 CLI를 만들 때 하위 명령 목록에 추가한 것입니다. 일반적인 형식은 다음과 같습니다:

```
simd query [moduleName] [command] <arguments> --flag <flagArg>
```

노드(CLI가 연결되는 전체 노드)와 같은 값을 제공하려면 app.toml 구성 파일을 사용하여 설정하거나 플래그로 제공할 수 있습니다. CLI는 애플리케이션 개발자가 계층 구조로 정의한 특정 명령 집합을 이해합니다: 루트 명령(simd), 명령 유형(Myquery), 명령이 포함된 모듈(스테이킹), 명령 자체(위임)로부터 명령합니다. 따라서 CLI는 이 명령을 처리하는 모듈을 정확히 파악하고 해당 모듈로 직접 호출을 전달합니다.

## gRPC

사용자가 쿼리를 수행할 수 있는 또 다른 인터페이스는 gRPC 서버에 대한 gRPC 요청입니다. 엔드포인트는 프로토콜 버퍼 서비스 메서드로 정의된 .proto 파일 안에 Protobuf의 자체 언어에 구애받지 않는 인터페이스 정의 언어(IDL)로 작성됩니다. Protobuf 에코시스템은 \*.proto 파일에서 다양한 언어로 코드를 생성하기 위한 도구를 개발했습니다. 이러한 도구를 사용하면 gRPC 클라이언트를 쉽게 빌드할 수 있습니다. 이러한 도구 중 하나가 grpcurl이며, 이 클라이언트를 사용하는 MyQuery에 대한 gRPC 요청은 다음과 같습니다:

```
grpcurl \
    -plaintext                                           # We want results in plain test
    -import-path ./proto \                               # Import these .proto files
    -proto ./proto/cosmos/staking/v1beta1/query.proto \  # Look into this .proto file for the Query protobuf service
    -d '{"address":"$MY_DELEGATOR"}' \                   # Query arguments
    localhost:9090 \                                     # gRPC server endpoint
    cosmos.staking.v1beta1.Query/Delegations             # Fully-qualified service method name
```

## REST

사용자가 쿼리를 수행할 수 있는 또 다른 인터페이스는 REST 서버에 대한 HTTP 요청을 이용하는 것입니다. REST 서버는 gRPC 게이트웨이를 사용하여 Protobuf 서비스에서 완전히 자동 생성됩니다. MyQuery에 대한 HTTP 요청의 예는 다음과 같습니다:

```
GET http://localhost:1317/cosmos/staking/v1beta1/delegators/{delegatorAddr}/delegations
```

# How Queries are Handled by the CLI

앞의 예는 외부 사용자가 노드의 상태를 쿼리하여 노드와 상호 작용하는 방법을 보여줍니다. 쿼리의 정확한 수명 주기를 더 자세히 이해하기 위해 CLI가 쿼리를 준비하는 방법과 노드가 쿼리를 처리하는 방법을 자세히 살펴봅시다. 사용자 관점에서의 상호작용은 약간 다르지만 모듈 개발자가 정의한 동일한 명령을 구현한 것이므로 기본 기능은 거의 동일합니다. 이 처리 단계는 CLI, gRPC 또는 REST 서버 내에서 이루어지며 클라이언트.Context와 크게 관련되어 있습니다.

## Context

CLI 명령을 실행할 때 가장 먼저 생성되는 것은 클라이언트.컨텍스트입니다. client.Context는 사용자 측에서 요청을 처리하는 데 필요한 모든 데이터를 저장하는 객체입니다. 특히 클라이언트.컨텍스트는 다음을 저장합니다:

코덱: 애플리케이션에서 사용하는 인코더/디코더로, CometBFT RPC 요청을 하기 전에 매개변수와 쿼리를 마샬링하고 반환된 응답을 JSON 객체로 마샬링 해제하는 데 사용됩니다. CLI에서 사용하는 기본 코덱은 Protobuf.

계정 디코더입니다: 인증 모듈의 계정 디코더로, []바이트를 계정으로 변환합니다.

RPC 클라이언트: 요청이 릴레이되는 CometBFT RPC 클라이언트 또는 노드입니다.

키링: 키 관리자]../beginner/03-accounts.md#키링)는 트랜잭션에 서명하고 키를 사용하여 다른 작업을 처리하는 데 사용됩니다.

출력 작성자: 응답을 출력하는 데 사용되는 Writer입니다.

구성: 구성: 사용자가 이 명령에 대해 구성한 플래그로, 쿼리할 블록체인의 높이를 지정하는 --height와 JSON 응답에 들여쓰기를 추가하도록 지정하는 --indent를 포함합니다.

client/context.go

```
// Context implements a typical context created in SDK modules for transaction
// handling and queries.
type Context struct {
	FromAddress       sdk.AccAddress
	Client            CometRPC
	GRPCClient        *grpc.ClientConn
	ChainID           string
	Codec             codec.Codec
	InterfaceRegistry codectypes.InterfaceRegistry
	Input             io.Reader
	Keyring           keyring.Keyring
	KeyringOptions    []keyring.Option
	Output            io.Writer
	OutputFormat      string
	Height            int64
	HomeDir           string
	KeyringDir        string
	From              string
	BroadcastMode     string
	FromName          string
	SignModeStr       string
	UseLedger         bool
	Simulate          bool
	GenerateOnly      bool
	Offline           bool
	SkipConfirm       bool
	TxConfig          TxConfig
	AccountRetriever  AccountRetriever
	NodeURI           string
	FeePayer          sdk.AccAddress
	FeeGranter        sdk.AccAddress
	Viper             *viper.Viper
	LedgerHasProtobuf bool
	PreprocessTxHook  PreprocessTxFn

	// IsAux is true when the signer is an auxiliary signer (e.g. the tipper).
	IsAux bool

	// TODO: Deprecated (remove).
	LegacyAmino *codec.LegacyAmino

	// CmdContext is the context.Context from the Cobra command.
	CmdContext context.Context
}
```

클라이언트.컨텍스트의 주요 역할은 최종 사용자와 상호작용하는 동안 사용되는 데이터를 저장하고 이 데이터와 상호작용하는 메서드를 제공하는 것으로, 쿼리가 풀노드에서 처리되기 전후에 사용됩니다. 특히 MyQuery를 처리할 때 client.Context는 쿼리 매개변수를 인코딩하고, 풀 노드를 검색하고, 출력을 작성하는 데 활용됩니다. 풀 노드는 애플리케이션에 구애받지 않고 특정 유형을 이해하지 못하므로 풀 노드로 전달되기 전에 쿼리를 []바이트 형식으로 인코딩해야 합니다. 풀노드(RPC 클라이언트) 자체는 사용자 CLI가 연결된 노드를 알고 있는 client.Context를 사용하여 검색됩니다. 쿼리는 이 풀 노드로 전달되어 처리됩니다. 마지막으로, 클라이언트.컨텍스트에는 응답이 반환될 때 출력을 기록하는 Writer가 포함됩니다. 이러한 단계는 이후 섹션에서 자세히 설명합니다.

## Arguments and Route Creation

수명 주기의 이 시점에서 사용자는 쿼리에 포함하려는 모든 데이터가 포함된 CLI 명령을 만들었습니다. MyQuery의 나머지 여정을 지원하기 위해 client.Context가 존재합니다. 이제 다음 단계는 명령 또는 요청을 구문 분석하고 인수를 추출한 다음 모든 것을 인코딩하는 것입니다. 이러한 단계는 모두 상호작용하는 인터페이스 내에서 사용자 측에서 이루어집니다.

### Encoding

이 경우(주소의 위임을 쿼리하는 경우)에는 MyQuery에 주소 delegatorAddress가 유일한 인자로 포함됩니다. 그러나 요청은 애플리케이션 유형에 대한 고유한 지식이 없는 풀 노드의 합의 엔진(예: CometBFT)으로 최종적으로 전달되기 때문에 []바이트만 포함할 수 있습니다. 따라서 클라이언트.컨텍스트의 코덱은 주소를 마샬링하는 데 사용됩니다. 다음은 CLI 명령의 코드입니다:

x/staking/client/cli/query.go

```
_, err = ac.StringToBytes(args[0])
if err != nil {
	return err
}
```

### gRPC Query Client Creation

코스모스 SDK는 프로토부프 서비스에서 생성된 코드를 활용하여 쿼리를 생성합니다. 스테이킹 모듈의 MyQuery 서비스는 CLI가 쿼리를 생성하는 데 사용하는 쿼리클라이언트를 생성합니다. 관련 코드는 다음과 같습니다:

x/staking/client/cli/query.go

```
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			_, err = ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorDelegationsRequest{
				DelegatorAddr: args[0],
				Pagination:    pageReq,
			}

			res, err := queryClient.DelegatorDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegations")

	return cmd
}
```

내부적으로 client.Context에는 미리 구성된 노드를 검색하고 쿼리를 전달하는 데 사용되는 Query() 함수가 있으며, 이 함수는 쿼리 정규화된 서비스 메서드 이름을 경로(이 경우: /cosmos.staking.v1beta1.Query/Delegations)로, 인수를 파라미터로 받습니다. 먼저 사용자가 이 쿼리를 전달하도록 구성한 RPC 클라이언트(노드라고 함)를 검색하고 ABCIQueryOptions(ABCI 호출을 위해 형식이 지정된 매개변수)를 생성합니다. 그런 다음 이 노드를 사용하여 ABCI 호출인 ABCIQueryWithOptions()를 실행합니다. 코드의 모습은 다음과 같습니다:

client/query.go

```
func (ctx Context) queryABCI(req abci.RequestQuery) (abci.ResponseQuery, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	var queryHeight int64
	if req.Height != 0 {
		queryHeight = req.Height
	} else {
		// fallback on the context height
		queryHeight = ctx.Height
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: queryHeight,
		Prove:  req.Prove,
	}

	result, err := node.ABCIQueryWithOptions(context.Background(), req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		return abci.ResponseQuery{}, sdkErrorToGRPCError(result.Response)
	}

	// data from trusted node or subspace query doesn't need verification
	if !opts.Prove || !isQueryStoreWithProof(req.Path) {
		return result.Response, nil
	}

	return result.Response, nil
}
```

## RPC

ABCIQueryWithOptions()를 호출하면 풀 노드가 MyQuery를 수신한 다음 요청을 처리합니다. RPC는 풀노드의 합의 엔진(예: CometBFT)으로 이루어지지만 쿼리는 합의의 일부가 아니므로 네트워크가 합의할 필요가 없으므로 나머지 네트워크에 브로드캐스트되지 않습니다. ABCI 클라이언트 및 CometBFT RPC에 대한 자세한 내용은 CometBFT 문서를 참조하세요.

## Application Query Handling

쿼리가 기본 합의 엔진에서 릴레이된 후 풀 노드에서 수신되면, 그 시점에서 애플리케이션별 유형을 이해하고 상태의 복사본을 가진 환경 내에서 처리됩니다. baseapp은 ABCI Query() 함수를 구현하고 gRPC 쿼리를 처리합니다. 쿼리 경로가 구문 분석되고 기존 서비스 메서드의 정규화된 서비스 메서드 이름(대부분 모듈 중 하나에 있음)과 일치하면 baseapp은 요청을 관련 모듈로 전달합니다.

MyQuery에는 스테이킹 모듈의 Protobuf 정규화된 서비스 메서드 이름이 있으므로(/cosmos.staking.v1beta1.Query/Delegations 호출), baseapp은 먼저 경로를 파싱한 다음 자체 내부 GRPCQueryRouter를 사용하여 해당 gRPC 핸들러를 검색하고 해당 모듈로 쿼리를 라우팅합니다. gRPC 처리기는 이 쿼리를 인식하고 애플리케이션의 스토어에서 적절한 값을 검색하여 응답을 반환하는 역할을 담당합니다. 쿼리 서비스에 대한 자세한 내용은 여기를 참조하세요. 쿼리자로부터 결과를 받으면 baseapp은 사용자에게 응답을 반환하는 프로세스를 시작합니다.

## Response

Query()는 ABCI 함수이므로 baseapp은 응답을 abci.ResponseQuery 유형으로 반환합니다. 클라이언트.컨텍스트 쿼리() 루틴은 응답을 수신하고.

### CLI Response

애플리케이션 코덱은 응답을 JSON으로 마샬링 해제하는 데 사용되며, client.Context는 출력 유형(텍스트, JSON 또는 YAML)과 같은 구성을 적용하여 명령줄에 출력을 인쇄합니다.

client/context.go

```
func (ctx Context) printOutput(out []byte) error {
	var err error
	if ctx.OutputFormat == "text" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return err
		}
	}
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/client/context.go#L341-L349
