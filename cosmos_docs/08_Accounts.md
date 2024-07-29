# Accounts

- 개요 이 문서는 코스모스 SDK에 내장된 계정 및 공개 키 시스템에 대해 설명합니다.

## Account Definition

코스모스 SDK에서 계정은 한 쌍의 공개 키 PubKey와 비공개 키 PrivKey를 지정합니다. PubKey는 애플리케이션에서 사용자를 식별하는 데 사용되는 다양한 주소를 생성하기 위해 파생될 수 있습니다(다른 당사자들 사이에서). 주소는 메시지 발신자를 식별하기 위해 메시지와도 연결됩니다. Priv키는 디지털 서명을 생성하는 데 사용되어 Priv키와 연결된 주소가 특정 메시지를 승인했음을 증명합니다. HD 키 도출을 위해 Cosmos SDK는 BIP32라는 표준을 사용합니다. BIP32를 통해 사용자는 초기 비밀 시드에서 파생된 계정 집합인 HD 지갑(BIP44에 명시된 대로)을 만들 수 있습니다. 시드는 보통 12단어 또는 24단어 니모닉으로 생성됩니다. 단일 시드는 단방향 암호화 함수를 사용하여 원하는 수의 PrivKey를 파생할 수 있습니다. 그런 다음 Priv키에서 Pub키를 파생할 수 있습니다. 니모닉이 보존되어 있으면 언제든지 개인키를 다시 생성할 수 있기 때문에 당연히 니모닉이 가장 민감한 정보입니다.

```
     Account 0                         Account 1                         Account 2

+------------------+              +------------------+               +------------------+
|                  |              |                  |               |                  |
|    Address 0     |              |    Address 1     |               |    Address 2     |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Public key 0    |              |  Public key 1    |               |  Public key 2    |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Private key 0   |              |  Private key 1   |               |  Private key 2   |
|        ^         |              |        ^         |               |        ^         |
+------------------+              +------------------+               +------------------+
         |                                 |                                  |
         |                                 |                                  |
         |                                 |                                  |
         +--------------------------------------------------------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Master PrivKey   |
                                 |                   |
                                 +-------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Mnemonic (Seed)  |
                                 |                   |
                                 +-------------------+

```

Cosmos SDK에서 키는 키링이라는 개체를 사용하여 저장하고 관리합니다.

# Keys, accounts, addresses, and signatures

사용자를 인증하는 주요 방법은 디지털 서명을 사용하는 것입니다. 사용자는 자신의 개인 키를 사용하여 거래에 서명합니다. 서명 검증은 연결된 공개 키로 수행됩니다. 온체인 서명 검증을 위해 공개 키는 적절한 트랜잭션 검증에 필요한 다른 데이터와 함께 계정 객체에 저장됩니다. 노드에서 모든 데이터는 프로토콜 버퍼 직렬화를 사용하여 저장됩니다. Cosmos SDK는 디지털 서명 생성을 위해 다음과 같은 디지털 키 체계를 지원합니다:

코스모스 SDK의 crypto/keys/secp256k1 패키지에 구현된 secp256k1,
코스모스 SDK의 crypto/keys/secp256r1 패키지에 구현된 tm-ed25519,
코스모스 SDK crypto/keys/ed25519 패키지에서 구현된 tm-ed25519. 이 체계는 합의 유효성 검사에만 지원됩니다.

|            | Address length in bytes | Public key length in bytes | Used for transaction authentication | Used for consensus (cometbft) |
| ---------- | ----------------------- | -------------------------- | ----------------------------------- | ----------------------------- |
| secp256k1  | 20                      | 33                         | yes                                 | no                            |
| secp256r1  | 32                      | 33                         | yes                                 | no                            |
| tm-ed25519 | -- not used --          | 32                         | no                                  | yes                           |

# Addresses

주소와 퍼블릭 키는 모두 애플리케이션에서 액터를 식별하는 공개 정보입니다. 계정은 인증 정보를 저장하는 데 사용됩니다. 기본 계정 구현은 BaseAccount 객체에 의해 제공됩니다. 각 계정은 공개 키에서 파생된 바이트 시퀀스인 Address를 사용하여 식별됩니다. Cosmos SDK에서는 계정이 사용되는 컨텍스트를 지정하는 세 가지 유형의 주소를 정의합니다:

- AccAddress는 사용자(메시지 발신자)를 식별합니다.
- ValAddress는 검증자 운영자를 식별합니다.
- ConsAddress는 합의에 참여하고 있는 검증자 노드를 식별합니다. 유효성 검사기 노드는 ed25519 곡선을 사용하여 파생됩니다.

이러한 유형은 주소 인터페이스를 구현합니다:
types/address.go

```
type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}
```

link : https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/types/address.go#L126-L134

```
sdk.AccAddress(pub.Address().Bytes())
```

참고로, Marshal() 및 Bytes() 메서드는 모두 동일한 원시 []바이트 형식의 주소를 반환합니다. Marshal()은 Protobuf 호환성을 위해 필요합니다. 사용자 상호작용을 위해 주소는 Bech32를 사용하여 형식이 지정되고 String 메서드로 구현됩니다. Bech32 메서드는 블록체인과 상호작용할 때 사용할 수 있는 유일한 지원 형식입니다. Bech32 사람이 읽을 수 있는 부분(Bech32 접두사)은 주소 유형을 나타내는 데 사용됩니다. 예시:

types/address.go

```
func (aa AccAddress) String() string {
	if aa.Empty() {
		return ""
	}

	key := conv.UnsafeBytesToStr(aa)

	if IsAddrCacheEnabled() {
		accAddrMu.Lock()
		defer accAddrMu.Unlock()

		addr, ok := accAddrCache.Get(key)
		if ok {
			return addr.(string)
		}
	}
	return cacheBech32Addr(GetConfig().GetBech32AccountAddrPrefix(), aa, accAddrCache, key)
}
```

|                    | Address Bech32 Prefix |
| ------------------ | --------------------- |
| Accounts           | cosmos                |
| Validator Operator | cosmosvaloper         |
| Consensus Nodes    | cosmosvalcons         |

## Public Keys

코스모스 SDK의 공개 키는 cryptotypes.PubKey 인터페이스에 의해 정의됩니다. 공개 키는 저장소에 저장되므로 cryptotypes.PubKey는 proto.Message 인터페이스를 확장합니다:

crypto/types/types.go

```
// PubKey defines a public key and extends proto.Message.
type PubKey interface {
	proto.Message

	Address() Address
	Bytes() []byte
	VerifySignature(msg, sig []byte) bool
	Equals(PubKey) bool
	Type() string
}
```

secp256k1 및 secp256r1 직렬화에는 압축 형식이 사용됩니다.

- Y 좌표가 X 좌표와 연결된 두 좌표 중 사전적으로 가장 큰 좌표인 경우 첫 바이트는 0x02 바이트입니다.

- 그렇지 않으면 첫 바이트는 0x03 바이트입니다.

이 접두사 뒤에는 x 좌표가 붙습니다.

공개 키는 계정(또는 사용자)을 참조하는 데 사용되지 않으며 일반적으로 트랜잭션 메시지를 작성할 때 사용되지 않습니다(몇 가지 예외: MsgCreateValidator, Validator 및 Multisig 메시지). 사용자 상호작용을 위해 PubKey는 Protobufs JSON(ProtoMarshalJSON 함수)을 사용하여 형식이 지정됩니다. 예시:

client/keys/output.go

```
// NewKeyOutput creates a default KeyOutput instance without Mnemonic, Threshold and PubKeys
func NewKeyOutput(name string, keyType keyring.KeyType, a sdk.Address, pk cryptotypes.PubKey) (KeyOutput, error) {
	apk, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return KeyOutput{}, err
	}
	bz, err := codec.ProtoMarshalJSON(apk, nil)
	if err != nil {
		return KeyOutput{}, err
	}
	return KeyOutput{
		Name:    name,
		Type:    keyType.String(),
		Address: a.String(),
		PubKey:  string(bz),
	}, nil
}
```

# Keyring

키링은 계정을 저장하고 관리하는 객체입니다. 코스모스 SDK에서 키링 구현은 키링 인터페이스를 따릅니다:

crypto/keyring/keyring.go

```
// Keyring exposes operations over a backend supported by github.com/99designs/keyring.
type Keyring interface {
	// Get the backend type used in the keyring config: "file", "os", "kwallet", "pass", "test", "memory".
	Backend() string
	// List all keys.
	List() ([]*Record, error)

	// Supported signing algorithms for Keyring and Ledger respectively.
	SupportedAlgorithms() (SigningAlgoList, SigningAlgoList)

	// Key and KeyByAddress return keys by uid and address respectively.
	Key(uid string) (*Record, error)
	KeyByAddress(address sdk.Address) (*Record, error)

	// Delete and DeleteByAddress remove keys from the keyring.
	Delete(uid string) error
	DeleteByAddress(address sdk.Address) error

	// Rename an existing key from the Keyring
	Rename(from, to string) error

	// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic key from it, and
	// persists the key to storage. Returns the generated mnemonic and the key Info.
	// It returns an error if it fails to generate a key for the given algo type, or if
	// another key is already stored under the same name or address.
	//
	// A passphrase set to the empty string will set the passphrase to the DefaultBIP39Passphrase value.
	NewMnemonic(uid string, language Language, hdPath, bip39Passphrase string, algo SignatureAlgo) (*Record, string, error)

	// NewAccount converts a mnemonic to a private key and BIP-39 HD Path and persists it.
	// It fails if there is an existing key Info with the same address.
	NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo SignatureAlgo) (*Record, error)

	// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
	SaveLedgerKey(uid string, algo SignatureAlgo, hrp string, coinType, account, index uint32) (*Record, error)

	// SaveOfflineKey stores a public key and returns the persisted Info structure.
	SaveOfflineKey(uid string, pubkey types.PubKey) (*Record, error)

	// SaveMultisig stores and returns a new multsig (offline) key reference.
	SaveMultisig(uid string, pubkey types.PubKey) (*Record, error)

	Signer

	Importer
	Exporter

	Migrator
}
```

키링의 기본 구현은 타사 99designs/keyring 라이브러리에서 제공됩니다.

키링 메서드에 대한 몇 가지 참고 사항:

- 서명(uid 문자열, msg []바이트) ([]바이트, types.PubKey, 오류)는 msg 바이트의 서명만 엄격하게 처리합니다. 트랜잭션을 준비하여 표준 []바이트 형식으로 인코딩해야 합니다. 프로토뷰는 결정론적이지 않기 때문에, 서명할 표준 페이로드는 ADR-020에서 결정론적으로 인코딩된 SignDoc 구조체이며, ADR-027을 사용하여 결정론적으로 인코딩하도록 결정되었습니다. 서명 확인은 기본적으로 코스모스 SDK에서 구현되지 않으며, anteHandler로 이연된다는 점에 유의하세요.

proto/cosmos/tx/v1beta1/tx.proto

```
message SignDoc {
  // body_bytes is protobuf serialization of a TxBody that matches the
  // representation in TxRaw.
  bytes body_bytes = 1;

  // auth_info_bytes is a protobuf serialization of an AuthInfo that matches the
  // representation in TxRaw.
  bytes auth_info_bytes = 2;

  // chain_id is the unique identifier of the chain this transaction targets.
  // It prevents signed transactions from being used on another chain by an
  // attacker
  string chain_id = 3;

  // account_number is the account number of the account in state
  uint64 account_number = 4;
}
```

- NewAccount(uid, 니모닉, bip39Passphrase, hdPath 문자열, algo 서명알고) (\*Record, 오류)는 bip44 경로를 기반으로 새 계정을 생성하고 디스크에 보존합니다. PrivKey는 암호화되지 않은 상태로 저장되지 않고 비밀번호로 암호화된 후 유지됩니다. 이 방법의 맥락에서 키 유형과 시퀀스 번호는 니모닉에서 개인 키와 공개 키를 파생하는 데 사용되는 BIP44 파생 경로의 세그먼트(예: 0, 1, 2, ...)를 참조합니다. 동일한 니모닉과 파생 경로를 사용하면 동일한 PrivKey, PubKey 및 주소가 생성됩니다. 키링에서 지원되는 키는 다음과 같습니다:

* secp256k1

* ed25519

* ExportPrivKeyArmor(uid, encryptPassphrase 문자열) (armor 문자열, 오류)는 주어진 암호문구를 사용하여 개인키를 ASCII로 아머링된 암호화된 형식으로 내보냅니다. 그런 다음 ImportPrivKey(uid, armor, passphrase 문자열) 함수를 사용하여 개인 키를 키링으로 다시 가져오거나 UnarmorDecryptPrivKey(armorStr 문자열, passphrase 문자열) 함수를 사용하여 원시 개인 키로 암호를 해독할 수 있습니다.

### Create New Key Type

키링에서 사용할 새 키 유형을 만들려면 keyring.SignatureAlgo 인터페이스가 충족되어야 합니다.

crypto/keyring/signing_algorithms.go

```
// SignatureAlgo defines the interface for a keyring supported algorithm.
type SignatureAlgo interface {
	Name() hd.PubKeyType
	Derive() hd.DeriveFn
	Generate() hd.GenerateFn


```

이 인터페이스는 세 가지 메서드로 구성되며, Name()은 알고리즘의 이름을 hd.PubKeyType으로 반환하고 Derive() 및 Generate()는 각각 다음 함수를 반환해야 합니다:

crypto/hd/algo.go

```
type (
	DeriveFn   func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error)
	GenerateFn func(bz []byte) types.PrivKey
)
```

keyring.SignatureAlgo를 구현한 후에는 키링의 지원되는 알고리즘 목록에 추가해야 합니다.

간단하게 하기 위해 새 키 유형을 구현하는 작업은 crypto/hd 패키지 내에서 수행해야 합니다. 작동하는 secp256k1 구현의 예는 algo.go에 있습니다.

- secp256r1 알고 구현하기

- 다음은 secp256r1을 구현하는 방법의 예입니다.
- 먼저 비밀 번호로 개인 키를 생성하는 새로운 함수가 secp256r1 패키지에 필요합니다. 이 함수는 다음과 같이 보일 수 있습니다:

cosmos-sdk/crypto/keys/secp256r1/privkey.go

```
// NewPrivKeyFromSecret creates a private key derived for the secret number
// represented in big-endian. The `secret` must be a valid ECDSA field element.
func NewPrivKeyFromSecret(secret []byte) (*PrivKey, error) {
    var d = new(big.Int).SetBytes(secret)
    if d.Cmp(secp256r1.Params().N) >= 1 {
        return nil, errorsmod.Wrap(errors.ErrInvalidRequest, "secret not in the curve base field")
    }
    sk := new(ecdsa.PrivKey)
    return &PrivKey{&ecdsaSK{*sk}}, nil
}
```

그 후 secp256r1Algo를 구현할 수 있습니다.

```
// cosmos-sdk/crypto/hd/secp256r1Algo.go

package hd

import (
    "github.com/cosmos/go-bip39"

    "github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
    "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Secp256r1Type uses the secp256r1 ECDSA parameters.
const Secp256r1Type = PubKeyType("secp256r1")

var Secp256r1 = secp256r1Algo{}

type secp256r1Algo struct{}

func (s secp256r1Algo) Name() PubKeyType {
    return Secp256r1Type
}

// Derive derives and returns the secp256r1 private key for the given seed and HD path.
func (s secp256r1Algo) Derive() DeriveFn {
    return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
        seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
        if err != nil {
            return nil, err
        }

        masterPriv, ch := ComputeMastersFromSeed(seed)
        if len(hdPath) == 0 {
            return masterPriv[:], nil
        }
        derivedKey, err := DerivePrivateKeyForPath(masterPriv, ch, hdPath)

        return derivedKey, err
    }
}

// Generate generates a secp256r1 private key from the given bytes.
func (s secp256r1Algo) Generate() GenerateFn {
    return func(bz []byte) types.PrivKey {
        key, err := secp256r1.NewPrivKeyFromSecret(bz)
        if err != nil {
            panic(err)
        }
        return key
    }
}
```

마지막으로 키링을 통해 지원되는 알고리즘 목록에 알고리즘을 추가해야 합니다.

// cosmos-sdk/crypto/keyring/keyring.go

```
func newKeystore(kr keyring.Keyring, cdc codec.Codec, backend string, opts ...Option) keystore {
    // Default options for keybase, these can be overwritten using the
    // Option function
    options := Options{
        SupportedAlgos:       SigningAlgoList{hd.Secp256k1, hd.Secp256r1}, // added here
        SupportedAlgosLedger: SigningAlgoList{hd.Secp256k1},
    }
...
```

이후 알고리즘을 사용하여 새 키를 생성하려면 --algo 플래그를 사용하여 지정해야 합니다:

```
simd keys add myKey --algo secp256r1
```
