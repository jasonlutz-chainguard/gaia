package ante_test

import (
    cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
    "github.com/cosmos/cosmos-sdk/testutil/testdata"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/gaia/v8/ante"
    globfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"
    ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
    ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// test global fees with bypass msg types.  min_gas_price = 0, global_fee=[msg_types]
func (s *IntegrationTestSuite) TestGlobalMinimumChainFeeAnteHandler() {
    // setup test
    s.SetupTest()
    s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
    priv1, _, addr1 := testdata.KeyTestPubAddr()
    privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
    
    globalfeeParamsEmpty := &globfeetypes.Params{MinimumGasPrices: []sdk.DecCoin{}}
    minGasPriceEmpty := []sdk.DecCoin{}
    
    globalfeeParamsHigh := &globfeetypes.Params{
        MinimumGasPrices: []sdk.DecCoin{
            sdk.NewDecCoinFromDec("uatom", sdk.NewDec(400).Quo(sdk.NewDec(100000))),
        },
    }
    minGasPrice := []sdk.DecCoin{
        sdk.NewDecCoinFromDec("uatom", sdk.NewDec(200).Quo(sdk.NewDec(100000))),
        sdk.NewDecCoinFromDec("stake", sdk.NewDec(200).Quo(sdk.NewDec(100000))),
    }
    globalfeeParamsLow := &globfeetypes.Params{
        MinimumGasPrices: []sdk.DecCoin{
            sdk.NewDecCoinFromDec("uatom", sdk.NewDec(100).Quo(sdk.NewDec(100000))),
        },
    }
    globalfeeParamsNewDenom := &globfeetypes.Params{
        MinimumGasPrices: []sdk.DecCoin{
            sdk.NewDecCoinFromDec("photon", sdk.NewDec(400).Quo(sdk.NewDec(100000))),
            sdk.NewDecCoinFromDec("quark", sdk.NewDec(400).Quo(sdk.NewDec(100000))),
        },
    }
    testCases := map[string]struct {
        minGasPrice     []sdk.DecCoin
        globalFeeParams *globfeetypes.Params
        feeAmount       sdk.Coins
        gasLimit        sdk.Gas
        txMsg           sdk.Msg
        txCheck         bool
        expErr          bool
    }{
        // test fees
        "empty min_gas_price, nonempty global fee, fee higher/equal than global_fee": {
            minGasPrice:     minGasPriceEmpty,
            globalFeeParams: globalfeeParamsHigh,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 800)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          false,
        },
        "empty min_gas_price, nonempty global fee, fee lower than global_fee": {
            minGasPrice:     minGasPriceEmpty,
            globalFeeParams: globalfeeParamsHigh,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 150)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        "nonempty min_gas_price, empty global fee, fee higher/equal than min_gas_price": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsEmpty,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 400)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          false,
        },
        "nonempty min_gas_price, empty global fee, fee lower than min_gas_price": {
            minGasPrice:     minGasPrice, // 0.002
            globalFeeParams: globalfeeParamsEmpty,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 50)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        "fee higher than globalfee and min_gas_price": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsHigh,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 800)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          false,
        },
        "fee lower than globalfee and min_gas_price": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsHigh,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 150)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        "globalfee > min_gas_price, fee higher than min_gas_price, lower than globalfee": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsHigh,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 500)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        "globalfee < min_gas_price, fee higher than globalfee and lower than min_gas_price": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsLow,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 150)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        "min_gas_price denom is not subset of global fee denom , fee paying in global fee denom": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsNewDenom,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("photon", 800)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          false,
        },
        "min_gas_price denom is not subset of global fee denom, fee paying in min_gas_price denom": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsNewDenom,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("stake", 800)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         true,
            expErr:          true,
        },
        
        // test bypass msg
        "msg type ibc, zero fee": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsLow,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg: ibcchanneltypes.NewMsgRecvPacket(
                ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
            txCheck: true,
            expErr:  false,
        },
        "msg type ibc, empty fee": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsLow,
            feeAmount:       sdk.Coins{},
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg: ibcchanneltypes.NewMsgRecvPacket(
                ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
            txCheck: true,
            expErr:  false,
        },
        "disable checkTx: no fee check": {
            minGasPrice:     minGasPrice,
            globalFeeParams: globalfeeParamsLow,
            feeAmount:       sdk.NewCoins(sdk.NewInt64Coin("uatom", 0)),
            gasLimit:        testdata.NewTestGasLimit(),
            txMsg:           testdata.NewTestMsg(addr1),
            txCheck:         false,
            expErr:          false,
        },
    }
    
    for name, testCase := range testCases {
        s.Run(name, func() {
            // set globalfees and min gas price
            subspace := s.setupTestGlobalFeeStoreAndMinGasPrice(testCase.minGasPrice, testCase.globalFeeParams)
            // setup antehandler
            mfd := ante.NewBypassMinFeeDecorator([]string{
                sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
                sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
                sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
            }, subspace)
            antehandler := sdk.ChainAnteDecorators(mfd)
            
            s.Require().NoError(s.txBuilder.SetMsgs(testCase.txMsg))
            s.txBuilder.SetFeeAmount(testCase.feeAmount)
            s.txBuilder.SetGasLimit(testCase.gasLimit)
            tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
            s.Require().NoError(err)
            
            s.ctx = s.ctx.WithIsCheckTx(testCase.txCheck)
            _, err = antehandler(s.ctx, tx, false)
            if !testCase.expErr {
                s.Require().NoError(err)
            } else {
                s.Require().Error(err)
            }
        })
    }
}
