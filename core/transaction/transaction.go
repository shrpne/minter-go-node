package transaction

import (
	"minter/core/types"
	"math/big"
	"minter/rlp"
	"bytes"
	"minter/crypto/sha3"
	"minter/crypto"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"minter/hexutil"
	tCrypto "github.com/tendermint/go-crypto"
	"minter/core/commissions"
)

var (
	ErrInvalidSig = errors.New("invalid transaction v, r, s values")
)

const (
	TypeSend                byte = 0x01
	TypeConvert             byte = 0x02
	TypeCreateCoin          byte = 0x03
	TypeDeclareCandidacy    byte = 0x04
	TypeDelegate            byte = 0x05
	TypeUnbond              byte = 0x06
	TypeRedeemCheck         byte = 0x07
	TypeSetCandidateOnline  byte = 0x08
	TypeSetCandidateOffline byte = 0x09
)

type Transaction struct {
	Nonce       uint64
	GasPrice    *big.Int
	Type        byte
	Data        RawData
	Payload     []byte
	ServiceData []byte
	V           *big.Int
	R           *big.Int
	S           *big.Int

	decodedData Data
}

type RawData []byte

type Data interface{}

type SendData struct {
	Coin  types.CoinSymbol `json:"coin,string"`
	To    types.Address    `json:"to"`
	Value *big.Int         `json:"value"`
}

type SetCandidateOnData struct {
	PubKey     []byte
}

type SetCandidateOffData struct {
	PubKey     []byte
}

type ConvertData struct {
	FromCoinSymbol types.CoinSymbol
	ToCoinSymbol   types.CoinSymbol
	Value          *big.Int
}

type CreateCoinData struct {
	Name                 string
	Symbol               types.CoinSymbol
	InitialAmount        *big.Int
	InitialReserve       *big.Int
	ConstantReserveRatio uint
}

type DeclareCandidacyData struct {
	Address    types.Address
	PubKey     []byte
	Commission uint
	Stake      *big.Int
}

type DelegateData struct {
	PubKey []byte
	Stake  *big.Int
}

type RedeemCheckData struct {
	RawCheck []byte
	Proof    [65]byte
}

type UnbondData struct {
	Address types.Address
}

func (tx *Transaction) Serialize() ([]byte, error) {

	buf, err := rlp.EncodeToBytes(tx)

	return buf, err
}

func (tx *Transaction) Gas() int64 {

	gas := int64(0)

	switch tx.Type {
	case TypeSend:
		gas = commissions.SendTx
	case TypeConvert:
		gas = commissions.ConvertTx
	case TypeCreateCoin:
		gas = commissions.CreateTx
	case TypeDeclareCandidacy:
		gas = commissions.DeclareCandidacyTx
	case TypeDelegate:
		gas = commissions.DelegateTx
	case TypeRedeemCheck:
		gas = commissions.RedeemCheckTx
	case TypeSetCandidateOnline:
		gas = commissions.ToggleCandidateStatus
	case TypeSetCandidateOffline:
		gas = commissions.ToggleCandidateStatus
	}

	gas = gas + int64(len(tx.Payload))*commissions.PayloadByte

	return gas
}

func (tx *Transaction) String() string {
	sender, _ := tx.Sender()

	switch tx.Type {
	case TypeSend:
		{
			txData := tx.decodedData.(SendData)
			return fmt.Sprintf("SEND TX nonce:%d from:%s to:%s coin:%s value:%s payload: %s",
				tx.Nonce, sender.String(), txData.To.String(), txData.Coin.String(), txData.Value.String(), tx.Payload)
		}
	case TypeConvert:
		{
			txData := tx.decodedData.(ConvertData)
			return fmt.Sprintf("CONVERT TX nonce:%d from:%s to:%s coin:%s value:%s payload: %s",
				tx.Nonce, sender.String(), txData.FromCoinSymbol.String(), txData.ToCoinSymbol.String(), txData.Value.String(), tx.Payload)
		}
	case TypeCreateCoin:
		{
			txData := tx.decodedData.(CreateCoinData)
			return fmt.Sprintf("CREATE COIN TX nonce:%d from:%s symbol:%s reserve:%s amount:%s crr:%d payload: %s",
				tx.Nonce, sender.String(), txData.Symbol.String(), txData.InitialReserve, txData.InitialAmount, txData.ConstantReserveRatio, tx.Payload)
		}
	case TypeDeclareCandidacy:
		{
			txData := tx.decodedData.(DeclareCandidacyData)
			return fmt.Sprintf("DECLARE CANDIDACY TX nonce:%d address:%s pubkey:%s commission: %d payload: %s",
				tx.Nonce, txData.Address.String(), hexutil.Encode(txData.PubKey[:]), txData.Commission, tx.Payload)
		}
	case TypeDelegate:
		{
			txData := tx.decodedData.(DelegateData)
			return fmt.Sprintf("DELEGATE CANDIDACY TX nonce:%d pubkey:%s payload: %s",
				tx.Nonce, hexutil.Encode(txData.PubKey[:]), tx.Payload)
		}
	case TypeRedeemCheck:
		{
			txData := tx.decodedData.(RedeemCheckData)
			return fmt.Sprintf("REDEEM CHECK TX nonce:%d proof: %x",
				tx.Nonce, txData.Proof)
		}
	}

	return "err"
}

func (tx *Transaction) Sign(prv *ecdsa.PrivateKey) error {

	h := tx.Hash()
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return err
	}

	tx.SetSignature(sig)

	return nil
}

func (tx *Transaction) SetSignature(sig []byte) {
	tx.R = new(big.Int).SetBytes(sig[:32])
	tx.S = new(big.Int).SetBytes(sig[32:64])
	tx.V = new(big.Int).SetBytes([]byte{sig[64] + 27})
}

func (tx *Transaction) Sender() (types.Address, error) {
	return recoverPlain(tx.Hash(), tx.R, tx.S, tx.V, true)
}

func (tx *Transaction) Hash() types.Hash {
	return rlpHash([]interface{}{
		tx.Nonce,
		tx.GasPrice,
		tx.Type,
		tx.Data,
		tx.Payload,
		tx.ServiceData,
	})
}

func (tx *Transaction) SetDecodedData(data Data) {
	tx.decodedData = data
}

func (tx *Transaction) GetDecodedData() Data {
	return tx.decodedData
}

func recoverPlain(sighash types.Hash, R, S, Vb *big.Int, homestead bool) (types.Address, error) {
	if Vb.BitLen() > 8 {
		return types.Address{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return types.Address{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return types.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return types.Address{}, errors.New("invalid public key")
	}
	var addr types.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

func rlpHash(x interface{}) (h types.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func DecodeFromBytes(buf []byte) (*Transaction, error) {

	var tx Transaction
	rlp.Decode(bytes.NewReader(buf), &tx)

	switch tx.Type {
	case TypeSend:
		{
			data := SendData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)
		}
	case TypeRedeemCheck:
		{
			data := RedeemCheckData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)
		}
	case TypeConvert:
		{
			data := ConvertData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)
		}
	case TypeCreateCoin:
		{
			data := CreateCoinData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)

			if data.InitialReserve == nil || data.InitialAmount == nil {
				fmt.Printf("%s\n", tx.String())
				return nil, errors.New("incorrect tx data")
			}
		}
	case TypeDeclareCandidacy:
		{
			data := DeclareCandidacyData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)

			var key tCrypto.PubKeyEd25519
			copy(key[:], data.PubKey)

			data.PubKey = key.Bytes()
			tx.SetDecodedData(data)
		}
	case TypeDelegate:
		{
			data := DelegateData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)

			var key tCrypto.PubKeyEd25519
			copy(key[:], data.PubKey)

			data.PubKey = key.Bytes()
			tx.SetDecodedData(data)
		}
	case TypeSetCandidateOnline:
		{
			data := SetCandidateOnData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)
		}
	case TypeSetCandidateOffline:
		{
			data := SetCandidateOffData{}
			rlp.Decode(bytes.NewReader(tx.Data), &data)
			tx.SetDecodedData(data)
		}
	default:
		return nil, errors.New("incorrect tx data")
	}

	if tx.S == nil || tx.R == nil || tx.V == nil {
		return nil, errors.New("incorrect tx signature")
	}

	if tx.GasPrice == nil || tx.Data == nil {
		return nil, errors.New("incorrect tx data")
	}

	return &tx, nil
}
