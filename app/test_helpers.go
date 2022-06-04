package app

import (
	"encoding/json"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/tendermint/starport/starport/pkg/cosmoscmd"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
)

var (
	EnableSpecificWasmProposals = ""
	WasmProposalsEnabled        = "true"
	EmptyWasmOpts               []wasm.Option
)

// GetWasmEnabledProposals parses the WasmProposalsEnabled and
// EnableSpecificWasmProposals values to produce a list of enabled proposals to
// pass into the application.
func GetWasmEnabledProposals() []wasm.ProposalType {
	if EnableSpecificWasmProposals == "" {
		if WasmProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}

		return wasm.DisableAllProposals
	}

	chunks := strings.Split(EnableSpecificWasmProposals, ",")

	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}

	return proposals
}

// Setup initializes a new OsmosisApp.
func Setup(isCheckTx bool) *App {
	db := dbm.NewMemDB()
	app := New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, 5, cosmoscmd.MakeEncodingConfig(ModuleBasics), GetWasmEnabledProposals(), simapp.EmptyAppOptions{}, EmptyWasmOpts).(*App)
	if !isCheckTx {
		genesisState := NewDefaultGenesisState(app.appCodec)
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		if err != nil {
			panic(err)
		}

		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: simapp.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}
