package cli

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/depinject"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestGetCommandEncode(t *testing.T) {
	var (
		cdc         codec.Codec
		legacyAmino *codec.LegacyAmino
		txConfig    client.TxConfig
	)

	err := depinject.Inject(testutil.AppConfig,
		&txConfig,
		&legacyAmino,
		&cdc,
	)
	require.NoError(t, err)

	cmd := GetEncodeCommand()
	_ = sdktestutil.ApplyMockIODiscardOutErr(cmd)

	authtypes.RegisterLegacyAminoCodec(legacyAmino)
	sdk.RegisterLegacyAminoCodec(legacyAmino)

	// Build a test transaction
	builder := txConfig.NewTxBuilder()
	builder.SetGasLimit(50000)
	builder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	builder.SetMemo("foomemo")
	jsonEncoded, err := txConfig.TxJSONEncoder()(builder.GetTx())
	require.NoError(t, err)

	txFile := sdktestutil.WriteToNewTempFile(t, string(jsonEncoded))
	txFileName := txFile.Name()

	ctx := context.Background()
	clientCtx := client.Context{}.
		WithTxConfig(txConfig).
		WithCodec(cdc)
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd.SetArgs([]string{txFileName})
	err = cmd.ExecuteContext(ctx)
	require.NoError(t, err)
}

func TestGetCommandDecode(t *testing.T) {
	var (
		cdc         codec.Codec
		legacyAmino *codec.LegacyAmino
		txConfig    client.TxConfig
	)

	err := depinject.Inject(testutil.AppConfig,
		&txConfig,
		&legacyAmino,
		&cdc,
	)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithTxConfig(txConfig).
		WithCodec(cdc)

	cmd := GetDecodeCommand()
	_ = sdktestutil.ApplyMockIODiscardOutErr(cmd)

	sdk.RegisterLegacyAminoCodec(legacyAmino)

	clientCtx = clientCtx.WithTxConfig(txConfig)

	// Build a test transaction
	builder := txConfig.NewTxBuilder()
	builder.SetGasLimit(50000)
	builder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	builder.SetMemo("foomemo")

	// Encode transaction
	txBytes, err := clientCtx.TxConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	// Convert the transaction into base64 encoded string
	base64Encoded := base64.StdEncoding.EncodeToString(txBytes)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	// Execute the command
	cmd.SetArgs([]string{base64Encoded})
	require.NoError(t, cmd.ExecuteContext(ctx))
}
