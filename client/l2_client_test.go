package client

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas-go-sdk/accounts"
	"github.com/bnb-chain/zkbas-go-sdk/txutils"
	"github.com/bnb-chain/zkbas-go-sdk/types"
)

var testEndpoint = "http://172.22.41.148:8888"
var seed = "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"

func getSdkClient() *l2Client {
	c := &l2Client{
		endpoint: testEndpoint,
	}
	keyManager, _ := accounts.NewSeedKeyManager(seed)
	c.SetKeyManager(keyManager)
	return c
}

func TestGetGasAccount(t *testing.T) {
	sdkClient := getSdkClient()
	account, err := sdkClient.GetGasAccount()
	if err != nil {
		println(err.Error())
		return
	}

	println("gas account index: ", account.AccountIndex)
}

func TestGetAccountNftList(t *testing.T) {
	sdkClient := getSdkClient()
	account, err := sdkClient.GetAccountNftList(2, 0, 10)
	if err != nil {
		println(err.Error())
		return
	}

	println("account total nft count: ", account.Total)
	bz, _ := json.MarshalIndent(account.Nfts, "", "  ")
	println(string(bz))
}

func TestGetAvailablePairs(t *testing.T) {
	sdkClient := getSdkClient()
	pairs, err := sdkClient.GetAvailablePairs()
	if err != nil {
		print(err.Error())
		return
	}
	bz, _ := json.MarshalIndent(pairs, "", "  ")
	println(string(bz))
}

func TestGetAssetList(t *testing.T) {
	sdkClient := getSdkClient()
	assetList, err := sdkClient.GetAssetsList()
	if err != nil {
		println(err.Error())
		return
	}

	bz, _ := json.MarshalIndent(assetList, "", "  ")
	println(string(bz))
}

func TestGetBalanceByAssetIdAndAccountName(t *testing.T) {
	sdkClient := getSdkClient()
	balance, err := sdkClient.GetBalanceByAssetIdAndAccountName(0, "sher.legend")
	if err != nil {
		println(err.Error())
		return
	}
	println("balance of asset 0: ", balance)

	balance, err = sdkClient.GetBalanceByAssetIdAndAccountName(1, "sher.legend")
	if err != nil {
		println(err.Error())
		return
	}
	println("balance of asset 1: ", balance)

	balance, err = sdkClient.GetBalanceByAssetIdAndAccountName(2, "sher.legend")
	if err != nil {
		println(err.Error())
		return
	}
	println("balance of asset 2: ", balance)
}

func TestCreateCollection(t *testing.T) {
	sdkClient := getSdkClient()
	txInfo := &types.CreateCollectionReq{
		Name:         fmt.Sprintf("Nft Collection - my collection"),
		Introduction: "Great Nft!",
	}

	collectionId, err := sdkClient.CreateCollection(txInfo, nil)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("create collection success, collection_id=%d \n", collectionId)
}

func TestMintNft(t *testing.T) {
	sdkClient := getSdkClient()

	bz := crypto.Keccak256Hash([]byte("contend_hash1"))
	txInfo := &types.MintNftTxReq{
		To:                  "sher.legend",
		NftContentHash:      common.Bytes2Hex(bz[:]),
		NftCollectionId:     1,
		CreatorTreasuryRate: 0,
	}

	nftId, err := sdkClient.MintNft(txInfo, nil)
	assert.NoError(t, err)
	fmt.Printf("mint nft success, assetId=%d \n", nftId)

}

func TestAtomicMatchTx(t *testing.T) {
	sellerSeed := "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"
	sellerName := "sher.legend"

	buyerSeed := "17673b9a9fdec6dc90c7cc1eb1c939134dfb659d2f08edbe071e5c45f343d008"
	buyerName := "gavin.legend"

	sdkClient := getSdkClient()

	buyer, err := sdkClient.GetAccountInfoByAccountName(buyerName)
	if err != nil {
		println(err.Error())
		return
	}

	seller, err := sdkClient.GetAccountInfoByAccountName(sellerName)
	if err != nil {
		println(err.Error())
		return
	}

	buyerOfferId, err := sdkClient.GetMaxOfferId(buyer.Index)
	if err != nil {
		println(err.Error())
		return
	}

	sellerOfferId, err := sdkClient.GetMaxOfferId(seller.Index)
	if err != nil {
		println(err.Error())
		return
	}

	nftIndex := int64(0)

	txInfo := PrepareAtomicMatchInfo(buyerSeed, sellerSeed, nftIndex, int64(buyer.Index), int64(buyerOfferId), int64(seller.Index), int64(sellerOfferId), seller.Nonce)

	txId, err := sdkClient.SendRawTx(types.TxTypeAtomicMatch, txInfo)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("send atomic match tx success, tx_id=%s \n", txId)
}

// TODO, test all transaction type.

func PrepareAtomicMatchInfo(buyerSeed, sellerSeed string, nftIndex, buyerIndex, buyerOfferId, sellerIndex, sellerOfferId, sellerNonce int64) string {
	buyerKey, err := accounts.NewSeedKeyManager(buyerSeed)
	if err != nil {
		panic(err)
	}

	listedAt := time.Now().UnixMilli()
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	buyOffer := &types.OfferTxInfo{
		Type:         types.BuyOfferType,
		OfferId:      buyerOfferId,
		AccountIndex: buyerIndex,
		NftIndex:     nftIndex,
		AssetId:      0,
		AssetAmount:  big.NewInt(10000),
		ListedAt:     listedAt,
		ExpiredAt:    expiredAt,
		TreasuryRate: 200,
		Sig:          nil,
	}

	buyTx, err := txutils.ConstructOfferTx(buyerKey, buyOffer)
	if err != nil {
		panic(err)
	}

	sellerKey, err := accounts.NewSeedKeyManager(sellerSeed)
	if err != nil {
		panic(err)
	}
	sellOffer := &types.OfferTxInfo{
		Type:         types.SellOfferType,
		OfferId:      sellerOfferId,
		AccountIndex: sellerIndex,
		NftIndex:     nftIndex,
		AssetId:      0,
		AssetAmount:  big.NewInt(10000),
		ListedAt:     listedAt,
		ExpiredAt:    expiredAt,
		TreasuryRate: 200,
		Sig:          nil,
	}

	sellTx, err := txutils.ConstructOfferTx(sellerKey, sellOffer)
	if err != nil {
		panic(err)
	}

	signedBuyOffer, _ := types.ParseOfferTxInfo(buyTx)
	signedSellOffer, _ := types.ParseOfferTxInfo(sellTx)

	txInfo := &types.AtomicMatchTxReq{
		BuyOffer:       signedBuyOffer,
		SellOffer:      signedSellOffer,
		TreasuryAmount: big.NewInt(5000),
	}

	tx, err := txutils.ConstructAtomicMatchTx(sellerKey, txInfo, nil)
	if err != nil {
		panic(err)
	}
	return tx
}

func TestTransferNft(t *testing.T) {
	toAccountName := "gavin.legend"

	sdkClient := getSdkClient()

	nftIndex := int64(3)
	txInfo := PrepareTransferNftTxInfo(sdkClient, nftIndex, toAccountName)

	txId, err := sdkClient.SendRawTx(types.TxTypeTransferNft, txInfo)
	if err != nil {
		fmt.Println("err: ", err.Error())
		return
	}
	fmt.Printf("send transfer nft tx success, tx_id=%s \n", txId)
}

func PrepareTransferNftTxInfo(c *l2Client, nftIndex int64, toAccountName string) string {
	txInfo := &types.TransferNftTxReq{
		NftIndex: nftIndex,
		To:       toAccountName,
	}
	ops := new(types.TransactOpts)
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		panic(err)
	}

	ops, err = c.fullFillToAddrOps(ops, toAccountName)
	if err != nil {
		panic(err)
	}
	tx, err := txutils.ConstructTransferNftTx(c.KeyManager(), txInfo, ops)
	if err != nil {
		panic(err)
	}
	return tx
}

func TestCancelOfferTx(t *testing.T) {
	sdkClient := getSdkClient()

	account, err := sdkClient.GetAccountInfoByPubKey(hex.EncodeToString(sdkClient.KeyManager().PubKey().Bytes()))
	if err != nil {
		println(err.Error())
		return
	}

	offerId, err := sdkClient.GetMaxOfferId(account.AccountIndex)
	if err != nil {
		println(err.Error())
		return
	}

	txInfo := PrepareCancelOfferTxInfo(sdkClient, int64(offerId))

	txId, err := sdkClient.SendRawTx(types.TxTypeCancelOffer, txInfo)
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("send cancel offer success, tx_id=%s \n", txId)
}

func PrepareCancelOfferTxInfo(c *l2Client, offerId int64) string {
	txInfo := &types.CancelOfferReq{
		OfferId: offerId,
	}

	ops := new(types.TransactOpts)
	ops, err := c.fullFillDefaultOps(ops)
	if err != nil {
		panic(err)
	}

	tx, err := txutils.ConstructCancelOfferTx(c.keyManager, txInfo, ops)
	if err != nil {
		panic(err)
	}
	return tx
}

func TestTransferInLayer2(t *testing.T) {
	l2Client := getSdkClient()

	txInfo := types.TransferTxReq{
		ToAccountName: "sher.legend",
		AssetId:       0,
		AssetAmount:   big.NewInt(1e17),
	}
	hash, err := l2Client.Transfer(&txInfo, nil)
	assert.NoError(t, err)
	fmt.Println("transfer success, tx id=", hash)
}

func TestAddLiquidityTx(t *testing.T) {
	sdkClient := getSdkClient()

	assetAAmount := big.NewInt(100000)
	assetBAmount := big.NewInt(100000)
	lpAmount, err := ComputeEmptyLpAmount(assetAAmount, assetBAmount)
	if err != nil {
		panic(err)
	}

	txReq := types.AddLiquidityReq{
		PairIndex:    0,
		AssetAId:     0,
		AssetAAmount: assetAAmount,
		AssetBId:     1,
		AssetBAmount: assetBAmount,
		LpAmount:     lpAmount,
	}

	txId, err := sdkClient.AddLiquidity(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("send add liquidity success, tx id: ", txId)
}

func TestRemoveLiquidity(t *testing.T) {
	sdkClient := getSdkClient()

	assetAAmount := big.NewInt(96)
	assetBAmount := big.NewInt(94)
	lpAmount := big.NewInt(100)
	txReq := types.RemoveLiquidityReq{
		PairIndex:         0,
		AssetAId:          0,
		AssetAMinAmount:   assetAAmount,
		AssetAAmountDelta: big.NewInt(0),
		AssetBAmountDelta: big.NewInt(0),
		AssetBId:          1,
		AssetBMinAmount:   assetBAmount,
		LpAmount:          lpAmount,
	}

	txId, err := sdkClient.RemoveLiquidity(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("send remove liquidity success, tx id: ", txId)
}

func TestSwap(t *testing.T) {
	sdkClient := getSdkClient()

	assetAMinAmount := big.NewInt(98)
	assetBMinAmount := big.NewInt(90)

	txReq := types.SwapTxReq{
		PairIndex:       0,
		AssetAId:        0,
		AssetAAmount:    assetAMinAmount,
		AssetBId:        1,
		AssetBMinAmount: assetBMinAmount,
	}

	txId, err := sdkClient.Swap(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("swap success, tx id: ", txId)
}

func TestGetPairInfo(t *testing.T) {
	sdkClient := getSdkClient()

	pairInfo, err := sdkClient.GetPairInfo(0)
	if err != nil {
		println(err.Error())
		return
	}
	bz, _ := json.MarshalIndent(pairInfo, "", "  ")
	println(string(bz))
}

func TestWithdrawBNB(t *testing.T) {
	sdkClient := getSdkClient()

	randomAddress := "0x8b2C5A5744F42AA9269BaabDd05933a96D8EF911"

	txReq := types.WithdrawReq{
		AssetId:     0,
		AssetAmount: big.NewInt(100),
		ToAddress:   randomAddress,
	}

	txId, err := sdkClient.Withdraw(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("withdraw success, tx id: ", txId)
}

func TestWithdrawBEP20(t *testing.T) {
	sdkClient := getSdkClient()

	randomAddress := "0x8b2C5A5744F42AA9269BaabDd05933a96D8EF911"

	txReq := types.WithdrawReq{
		AssetId:     1,
		AssetAmount: big.NewInt(100),
		ToAddress:   randomAddress,
	}

	txId, err := sdkClient.Withdraw(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("withdraw success, tx id: ", txId)
}

func TestWithdrawNft(t *testing.T) {
	sdkClient := getSdkClient()

	randomAddress := "0x8b2C5A5744F42AA9269BaabDd05933a96D8EF911"

	txReq := types.WithdrawNftTxReq{
		AccountIndex:        2,
		CreatorAccountIndex: 2,
		CreatorTreasuryRate: 0,
		NftIndex:            1,
		NftL1Address:        "",
		NftL1TokenId:        big.NewInt(0),
		CollectionId:        0,
		ToAddress:           randomAddress,
	}

	txId, err := sdkClient.WithdrawNft(&txReq, nil)
	if err != nil {
		println(err.Error())
		return
	}
	println("withdraw nft success, tx id: ", txId)
}