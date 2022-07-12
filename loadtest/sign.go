package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/sei-protocol/sei-chain/app"
)

const NODE_URI = "tcp://localhost:26657"

type AccountInfo struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

func GetKey(accountIdx uint64) cryptotypes.PrivKey {
	userHomeDir, _ := os.UserHomeDir()
	accountKeyFilePath := filepath.Join(userHomeDir, "test_accounts", fmt.Sprintf("ta%d.json", accountIdx))
	jsonFile, err := os.Open(accountKeyFilePath)
	if err != nil {
		panic(err)
	}
	var accountInfo AccountInfo
	byteVal, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	jsonFile.Close()
	json.Unmarshal(byteVal, &accountInfo)
	kr, _ := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, filepath.Join(userHomeDir, ".sei-chain"), nil, app.MakeEncodingConfig().Marshaler)

	keyringAlgos, _ := kr.SupportedAlgorithms()
	algoStr := string(hd.Secp256k1Type)
	algo, _ := keyring.NewSigningAlgoFromString(algoStr, keyringAlgos)
	hdpath := hd.CreateHDPath(sdk.GetConfig().GetCoinType(), 0, 0).String()
	derivedPriv, _ := algo.Derive()(accountInfo.Mnemonic, "", hdpath)
	return algo.Generate()(derivedPriv)
}

func SignTx(txBuilder *client.TxBuilder, privKey cryptotypes.PrivKey, seqDelta uint64) {
	var sigsV2 []signing.SignatureV2
	accountNum, seqNum := GetAccountNumberSequenceNumber(privKey)
	seqNum += seqDelta
	sigV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  TEST_CONFIG.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seqNum,
	}
	sigsV2 = append(sigsV2, sigV2)
	_ = (*txBuilder).SetSignatures(sigsV2...)
	sigsV2 = []signing.SignatureV2{}
	signerData := xauthsigning.SignerData{
		ChainID:       CHAIN_ID,
		AccountNumber: accountNum,
		Sequence:      seqNum,
	}
	sigV2, _ = clienttx.SignWithPrivKey(
		TEST_CONFIG.TxConfig.SignModeHandler().DefaultMode(),
		signerData,
		*txBuilder,
		privKey,
		TEST_CONFIG.TxConfig,
		seqNum,
	)
	sigsV2 = append(sigsV2, sigV2)
	_ = (*txBuilder).SetSignatures(sigsV2...)
}

func GetAccountNumberSequenceNumber(privKey cryptotypes.PrivKey) (uint64, uint64) {
	hexAccount := privKey.PubKey().Address()
	address, err := sdk.AccAddressFromHexUnsafe(hexAccount.String())
	if err != nil {
		panic(err)
	}
	accountRetriever := authtypes.AccountRetriever{}
	cl, err := client.NewClientFromNode(NODE_URI)
	if err != nil {
		panic(err)
	}
	context := client.Context{}
	context = context.WithNodeURI(NODE_URI)
	context = context.WithClient(cl)
	context = context.WithInterfaceRegistry(TEST_CONFIG.InterfaceRegistry)
	account, seq, err := accountRetriever.GetAccountNumberSequence(context, address)
	if err != nil {
		time.Sleep(5 * time.Second)
		// retry once after 5 seconds
		account, seq, err = accountRetriever.GetAccountNumberSequence(context, address)
		if err != nil {
			panic(err)
		}
	}
	return account, seq
}
