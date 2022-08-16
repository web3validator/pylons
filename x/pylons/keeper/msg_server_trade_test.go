package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Pylons-tech/pylons/x/pylons/keeper"
	"github.com/Pylons-tech/pylons/x/pylons/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestTradeMsgServerCreate1() {
	k := suite.k
	ctx := suite.ctx
	require := suite.Require()
	app := suite.pylonsApp

	enabled := true
	params := banktypes.DefaultParams()

	wctx := sdk.WrapSDKContext(ctx)
	srv := keeper.NewMsgServerImpl(k)

	creator := types.GenTestBech32FromString("creator")

	fooCoin := sdk.NewCoin("testPylons", sdk.OneInt())

	// Set default send_enabled to !enabled, add a foodenom that overrides default as enabled
	params.DefaultSendEnabled = !enabled
	params = params.SetSendEnabledParam(fooCoin.Denom, !enabled)
	app.BankKeeper.SetParams(ctx, params)

	coinInputs := make([]types.CoinInput, 0)
	coinInputs = append(coinInputs, types.CoinInput{Coins: sdk.Coins{sdk.Coin{Denom: "ustripeusd", Amount: sdk.NewInt(0)}}})
	items := createNItem(k, ctx, 1, true)

	coinOutputs := sdk.NewCoins()
	coinOutputs = append(coinOutputs, *&fooCoin)

	for _, tc := range []struct {
		desc          string
		request       types.MsgCreateTrade
		itemTradeable bool
		err           error
	}{
		{
			desc: "Completed",
			request: types.MsgCreateTrade{
				Creator:     creator,
				CoinInputs:  nil,
				ItemInputs:  nil,
				CoinOutputs: nil,
				ItemOutputs: nil,
				ExtraInfo:   "extrainfo",
			},
			itemTradeable: false,
		},
		{
			desc: "Send enabled is disabled",
			request: types.MsgCreateTrade{
				Creator:     creator,
				CoinInputs:  nil,
				ItemInputs:  nil,
				CoinOutputs: coinOutputs,
				ItemOutputs: nil,
				ExtraInfo:   "",
			},
			itemTradeable: false,
			err:           sdkerrors.ErrInvalidRequest,
		},
		{
			desc: "Coin cannot be traded",
			request: types.MsgCreateTrade{
				Creator:     creator,
				CoinInputs:  append(coinInputs, types.CoinInput{Coins: sdk.Coins{sdk.Coin{Denom: "test", Amount: sdk.NewInt(0)}}}),
				ItemInputs:  nil,
				CoinOutputs: nil,
				ItemOutputs: nil,
				ExtraInfo:   "",
			},
			itemTradeable: false,
			err:           sdkerrors.ErrInvalidRequest,
		},
		{
			desc: "Item and Cookbook ID not found",
			request: types.MsgCreateTrade{
				Creator:     items[0].Owner,
				CoinInputs:  coinInputs,
				ItemInputs:  nil,
				CoinOutputs: sdk.Coins{},
				ItemOutputs: []types.ItemRef{{CookbookId: "testCookbookId", ItemId: "testItemId"}},
				ExtraInfo:   "extraInfo",
			},
			itemTradeable: false,
			err:           sdkerrors.ErrInvalidRequest,
		},
		{
			desc: "Creator and item owner difference",
			request: types.MsgCreateTrade{
				Creator:     creator,
				CoinInputs:  coinInputs,
				ItemInputs:  nil,
				CoinOutputs: sdk.Coins{},
				ItemOutputs: []types.ItemRef{{CookbookId: items[0].CookbookId, ItemId: items[0].Id}},
				ExtraInfo:   "extraInfo",
			},
			itemTradeable: false,
			err:           sdkerrors.ErrInvalidRequest,
		},
		{
			desc: "Item and Cookbook ID can not be trade",
			request: types.MsgCreateTrade{
				Creator:     items[0].Owner,
				CoinInputs:  coinInputs,
				ItemInputs:  nil,
				CoinOutputs: sdk.Coins{},
				ItemOutputs: []types.ItemRef{{CookbookId: items[0].CookbookId, ItemId: items[0].Id}},
				ExtraInfo:   "extraInfo",
			},
			itemTradeable: true,
			err:           sdkerrors.ErrInvalidRequest,
		},
		{
			desc: "Invalid Coin Inputs",
			request: types.MsgCreateTrade{
				Creator:     items[0].Owner,
				CoinInputs:  coinInputs,
				ItemInputs:  nil,
				CoinOutputs: sdk.Coins{},
				ItemOutputs: []types.ItemRef{{CookbookId: items[0].CookbookId, ItemId: items[0].Id}},
				ExtraInfo:   "extraInfo",
			},
			itemTradeable: false,
			err:           sdkerrors.ErrInvalidCoins,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			if tc.itemTradeable {
				items[0].Tradeable = false
			} else {
				items[0].Tradeable = true
			}
			k.SetItem(ctx, items[0])
			_, err := srv.CreateTrade(wctx, &tc.request)
			if tc.err != nil {
				require.ErrorIs(err, tc.err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestTradeMsgServerCancel() {
	k := suite.k
	ctx := suite.ctx
	require := suite.Require()

	wctx := sdk.WrapSDKContext(ctx)
	srv := keeper.NewMsgServerImpl(k)
	creator := types.GenTestBech32FromString("creator")

	for _, tc := range []struct {
		desc    string
		request *types.MsgCancelTrade
		err     error
	}{
		{
			desc:    "Completed",
			request: &types.MsgCancelTrade{Creator: creator, Id: 0},
		},
		{
			desc:    "Unauthorized",
			request: &types.MsgCancelTrade{Creator: "B", Id: 1},
			err:     sdkerrors.ErrUnauthorized,
		},
		{
			desc:    "KeyNotFound",
			request: &types.MsgCancelTrade{Creator: creator, Id: 10},
			err:     sdkerrors.ErrKeyNotFound,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			_, err := srv.CreateTrade(wctx, &types.MsgCreateTrade{
				Creator:     creator,
				CoinInputs:  nil,
				ItemInputs:  nil,
				CoinOutputs: sdk.Coins{},
				ItemOutputs: nil,
				ExtraInfo:   "",
			})
			require.NoError(err)
			_, err = srv.CancelTrade(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(err, tc.err)
			} else {
				require.NoError(err)
			}
		})
	}
}
