package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate genesis command takes
func ValidateGenesisCmd(ctx *server.Context, cdc *codec.Codec, mbm sdk.ModuleBasicManager) *cobra.Command {
	return &cobra.Command{
		Use:   "validate-genesis [file]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "validates the genesis file at the default location or at the location passed as an arg",
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// Load default if passed no args, otherwise load passed file
			var genesis string
			if len(args) == 0 {
				genesis = ctx.Config.GenesisFile()
			} else {
				genesis = args[0]
			}

			fmt.Fprintf(os.Stderr, "validating genesis file at %s\n", genesis)

			var genDoc *tmtypes.GenesisDoc
			if genDoc, err = tmtypes.GenesisDocFromFile(genesis); err != nil {
				return fmt.Errorf("error loading genesis doc from %s: %s", genesis, err.Error())
			}

			var genState map[string]json.RawMessage
			if err = cdc.UnmarshalJSON(genDoc.AppState, &genState); err != nil {
				return fmt.Errorf("error unmarshaling genesis doc %s: %s", genesis, err.Error())
			}

			if err = mbm.ValidateGenesis(genState); err != nil {
				return fmt.Errorf("error validating genesis file %s: %s", genesis, err.Error())
			}

			// make sure initchain doesn't panic
			db := db.NewMemDB()
			gapp := app.NewGaiaApp(log.NewNopLogger(), db, nil, true, 0)
			defaultGenesisBytes := cdc.MustMarshalJSON(mbm.DefaultGenesis())
			cms := store.NewCommitMultiStore(db)
			ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
			req := abci.RequestInitChain{AppStateBytes: defaultGenesisBytes}
			//defer func() {
			//if r := recover(); r != nil {
			//err = fmt.Errorf("InitChainer panics %s", r)
			//}
			//}()
			_ = gapp.InitChainer(ctx, req)

			fmt.Printf("File at %s is a valid genesis file for gaiad\n", genesis)
			return nil
		},
	}
}
