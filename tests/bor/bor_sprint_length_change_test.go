package bor

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/fdlimit"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/miner"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"gotest.tools/assert"
)

var (
	// addr1 = 0x71562b71999873DB5b286dF957af199Ec94617F7
	pkey12, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	// addr2 = 0x9fB29AAc15b9A4B7F17c3385939b007540f4d791
	pkey22, _ = crypto.HexToECDSA("9b28f36fbd67381120752d6172ecdcf10e06ab2d9a1367aac00cdcd6ac7855d3")
	keys2     = []*ecdsa.PrivateKey{pkey12, pkey22}
)

// Sprint length change tests
func TestValidatorsBlockProduction(t *testing.T) {

	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Generate a batch of accounts to seal and fund with
	faucets := make([]*ecdsa.PrivateKey, 128)
	for i := 0; i < len(faucets); i++ {
		faucets[i], _ = crypto.GenerateKey()
	}

	// Create an Ethash network based off of the Ropsten config
	genesis := InitGenesis1(t, faucets, "./testdata/genesis_sprint_length_change.json")

	var (
		nodes  []*eth.Ethereum
		enodes []*enode.Node
	)
	for i := 0; i < 2; i++ {
		// Start the node and wait until it's up
		stack, ethBackend, err := InitMiner1(genesis, keys2[i], true)
		if err != nil {
			panic(err)
		}
		defer stack.Close()

		for stack.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to all the previous ones
		for _, n := range enodes {
			stack.Server().AddPeer(n)
		}
		// Start tracking the node and its enode
		nodes = append(nodes, ethBackend)
		enodes = append(enodes, stack.Server().Self())
	}

	// Iterate over all the nodes and start mining
	time.Sleep(3 * time.Second)
	for _, node := range nodes {
		if err := node.StartMining(1); err != nil {
			panic(err)
		}
	}

	for {

		// for block 0 to 7, the primary validator is node0
		// for block 8 to 15, the primary validator is node1
		// for block 16 to 19, the primary validator is node0
		// for block 20 to 23, the primary validator is node1
		blockHeaderVal0 := nodes[0].BlockChain().CurrentHeader()

		if blockHeaderVal0.Number.Uint64() == 24 {
			break
		}

	}

	// check block 7 miner ; expected author is node0 signer
	blockHeaderVal0 := nodes[0].BlockChain().GetHeaderByNumber(7)
	blockHeaderVal1 := nodes[1].BlockChain().GetHeaderByNumber(7)
	authorVal0, err := nodes[0].Engine().Author(blockHeaderVal0)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}
	authorVal1, err := nodes[1].Engine().Author(blockHeaderVal1)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}

	// check both nodes have the same block 7
	assert.Equal(t, authorVal0, authorVal1)

	// check block mined by node0
	assert.Equal(t, authorVal0, nodes[0].AccountManager().Accounts()[0])

	// check block 15 miner ; expected author is node1 signer
	blockHeaderVal0 = nodes[0].BlockChain().GetHeaderByNumber(15)
	blockHeaderVal1 = nodes[1].BlockChain().GetHeaderByNumber(15)
	authorVal0, err = nodes[0].Engine().Author(blockHeaderVal0)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}
	authorVal1, err = nodes[1].Engine().Author(blockHeaderVal1)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}

	// check both nodes have the same block 15
	assert.Equal(t, authorVal0, authorVal1)

	// check block mined by node1
	assert.Equal(t, authorVal0, nodes[1].AccountManager().Accounts()[0])

	// check block 19 miner ; expected author is node0 signer
	blockHeaderVal0 = nodes[0].BlockChain().GetHeaderByNumber(19)
	blockHeaderVal1 = nodes[1].BlockChain().GetHeaderByNumber(19)
	authorVal0, err = nodes[0].Engine().Author(blockHeaderVal0)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}
	authorVal1, err = nodes[1].Engine().Author(blockHeaderVal1)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}

	// check both nodes have the same block 19
	assert.Equal(t, authorVal0, authorVal1)

	// check block mined by node0
	assert.Equal(t, authorVal0, nodes[0].AccountManager().Accounts()[0])

	// check block 23 miner ; expected author is node1 signer
	blockHeaderVal0 = nodes[0].BlockChain().GetHeaderByNumber(23)
	blockHeaderVal1 = nodes[1].BlockChain().GetHeaderByNumber(23)
	authorVal0, err = nodes[0].Engine().Author(blockHeaderVal0)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}
	authorVal1, err = nodes[1].Engine().Author(blockHeaderVal1)
	if err != nil {
		log.Error("Error in getting author", "err", err)
	}

	// check both nodes have the same block 23
	assert.Equal(t, authorVal0, authorVal1)

	// check block mined by node1
	assert.Equal(t, authorVal0, nodes[1].AccountManager().Accounts()[0])

}

func TestSprintLengths(t *testing.T) {
	testBorConfig := params.TestChainConfig.Bor
	testBorConfig.Sprint = map[string]uint64{
		"0": 16,
		"8": 4,
	}
	assert.Equal(t, testBorConfig.CalculateSprint(0), uint64(16))
	assert.Equal(t, testBorConfig.CalculateSprint(8), uint64(4))
	assert.Equal(t, testBorConfig.CalculateSprint(9), uint64(4))

}

var keys_21val = []map[string]string{
	{
		"address":  "0x73E033779C9030D4528d86FbceF5B02e97488921",
		"priv_key": "61eb51cf8936309151ab7b931841ea033b6a09931f6a100b464fbbd74f3e0bd7",
		"pub_key":  "0x04f9a5e9bf76b45ac58f1b018ccba4b83b3531010cdadf42174c18a9db9879ef1dcb5d1254ce834bc108b110cd8d0186ed69a0387528a142bdb5936faf58bf98c9",
	},
	{
		"address":  "0x5C3E1B893B9315a968fcC6bce9EB9F7d8E22edB3",
		"priv_key": "c19fac8e538447124ad2408d9fbaeda2bb686fee763dca7a6bab58ea12442413",
		"pub_key":  "0x0495421933eda03dcc37f9186c24e255b569513aefae71e96d55d0db3df17502e24e86297b01a167fab9ce1174f06ee3110510ac242e39218bd964de5b345edbd6",
	},
	{
		"address":  "0xb005bc07015170266Bd430f3EC1322938603be20",
		"priv_key": "17cd9b38c2b3a639c7d97ccbf2bb6c7140ab8f625aec4c249bc8e4cfd3bf9a96",
		"pub_key":  "0x04435a70d343aa569e6f3386c73e39a1aa6f88c30e5943baedda9618b55cc944a2de1d114aff6d0e9fa002bebc780b04ef6c1b8a06bbf0d41c10d1efa55390f198",
	},
	{
		"address":  "0xA464DC4810Bc79B956810759e314d85BcE35cD1c",
		"priv_key": "3efcf3f7014a6257f4a443119851414111820c681b27525dab3f35e72e28e51e",
		"pub_key":  "0x040180920306bf598ea050e258f2c7e50804a77a64f5a11705e08d18ee71eb0a80fafc95d0a42b92371ded042edda16c1f0b5f2fef7c4113ad66c59a71c29d977e",
	},
	{
		"address":  "0xE8d02Da3dFeeB3e755472D95D666BD6821D92129",
		"priv_key": "45c9ef66361a2283cef14184f128c41949103b791aa622ead3c0bc844648b835",
		"pub_key":  "0x04a14651ddc80467eb589d72d95153fa695e4cb2e4bb99edeb912e840d309d61313b6f4676081b099f29e6598ecf98cb7b44bb862d019920718b558f27ba94ca51",
	},
	{
		"address":  "0xF93B54Cf36E917f625B48e1e3C9F93BC2344Fb06",
		"priv_key": "93788a1305605808df1f9a96b5e1157da191680cf08bc15e077138f517563cd5",
		"pub_key":  "0x045eee11dceccd9cccc371ca3d96d74c848e785223f1e5df4d1a7f08efdfeb90bd8f0035342a9c26068cf6c7ab395ca3ceea555541325067fc187c375390efa57d",
	},
	{
		"address":  "0x751eC4877450B8a4D652d0D70197337FC38a42e6",
		"priv_key": "6e7f48d012c9c0baadbdc88af32521e2e477fd6898a9b65e6abe19fd6652cb2e",
		"pub_key":  "0x0479db4c0b757bf0e5d9b8954b078ab7c0e91d6c19697904d23d07ea4853c8584382de91174929ba5c598214b8a991471ae051458ea787cdc15a4e435a55ef8059",
	},
}

func TestSprintLengthReorg(t *testing.T) {
	reorgsLengthTests := []map[string]uint64{
		{
			"reorgLength": 10,
			"startBlock":  23,
		},
		// {
		// 	"reorgLength": 20,
		// 	"validator":   0,
		// 	"startBlock":  16,
		// },
		// {
		// 	"reorgLength": 30,
		// 	"validator":   0,
		// 	"startBlock":  16,
		// },
		// {
		// 	"reorgLength": 10,
		// 	"validator":   0,
		// 	"startBlock":  196,
		// },
	}
	for _, tt := range reorgsLengthTests {
		SetupValidatorsAndTest(t, tt)
	}
}

func SetupValidatorsAndTest(t *testing.T, tt map[string]uint64) {
	log.Root().SetHandler(log.LvlFilterHandler(3, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	fdlimit.Raise(2048)

	// Create an Ethash network based off of the Ropsten config
	genesis := InitGenesis1(t, nil, "./testdata/genesis_7val.json")

	var (
		nodes  []*eth.Ethereum
		enodes []*enode.Node
		stacks []*node.Node
	)

	pkeys_21val := make([]*ecdsa.PrivateKey, len(keys_21val))
	for index, signerdata := range keys_21val {
		pkeys_21val[index], _ = crypto.HexToECDSA(signerdata["priv_key"])
	}

	for i := 0; i < len(keys_21val); i++ {
		// Start the node and wait until it's up
		stack, ethBackend, err := InitMiner1(genesis, pkeys_21val[i], true)
		if err != nil {
			panic(err)
		}
		defer stack.Close()

		for stack.Server().NodeInfo().Ports.Listener == 0 {
			time.Sleep(250 * time.Millisecond)
		}
		// Connect the node to all the previous ones
		for _, n := range enodes {
			stack.Server().AddPeer(n)
		}
		// Start tracking the node and its enode
		stacks = append(stacks, stack)
		nodes = append(nodes, ethBackend)
		enodes = append(enodes, stack.Server().Self())
	}

	// Iterate over all the nodes and start mining
	time.Sleep(3 * time.Second)
	for _, node := range nodes {
		if err := node.StartMining(1); err != nil {
			panic(err)
		}
	}
	chain2HeadCh := make(chan core.Chain2HeadEvent, 64)
	primaryProducerIndex := 0
	subscribedNodeIndex := 2
	nodes[subscribedNodeIndex].BlockChain().SubscribeChain2HeadEvent(chain2HeadCh)
	stacks[1].Server().NoDiscovery = true

	for {

		blockHeaderVal0 := nodes[0].BlockChain().CurrentHeader()
		peers := stacks[1].Server().Peers()
		log.Warn("Peers", "peers length", len(peers))
		log.Warn("Current block", "number", blockHeaderVal0.Number, "hash", blockHeaderVal0.Hash())

		if blockHeaderVal0.Number.Uint64() == tt["startBlock"] {
			author, _ := nodes[0].Engine().Author(blockHeaderVal0)

			log.Warn("Current block", "number", blockHeaderVal0.Number, "hash", blockHeaderVal0.Hash(), "author", author)
			fmt.Printf("\n------%+v, %+v-----\n", blockHeaderVal0.Number.Uint64(), tt["startBlock"])

			for index, signerdata := range keys_21val {
				if strings.EqualFold(signerdata["address"], author.String()) {
					primaryProducerIndex = index
					log.Warn("Primary producer", "index", primaryProducerIndex)
				}
			}
			stacks[1].Server().MaxPeers = 0
			for _, enode := range enodes {
				stacks[1].Server().RemovePeer(enode)
			}
			if primaryProducerIndex == 0 {
				subscribedNodeIndex = 1
			}

			fmt.Println("----------------- startBlock", tt["startBlock"])

		}

		if blockHeaderVal0.Number.Uint64() >= tt["startBlock"] && blockHeaderVal0.Number.Uint64() < tt["startBlock"]+tt["reorgLength"] {
			// stacks[1].Server().NoDiscovery = true
			// stacks[1].Server().MaxPeers = 0
			for _, enode := range enodes {
				stacks[1].Server().RemovePeer(enode)
			}
		}

		if blockHeaderVal0.Number.Uint64() == tt["startBlock"]+tt["reorgLength"]+1 {
			stacks[1].Server().NoDiscovery = false
			stacks[1].Server().MaxPeers = 100

			for _, enode := range enodes {
				stacks[primaryProducerIndex].Server().AddPeer(enode)
			}
			fmt.Println("----------------- endblock", tt["startBlock"]+tt["reorgLength"]+1)
		}

		if blockHeaderVal0.Number.Uint64() == tt["startBlock"]+tt["reorgLength"]+50 {
			fmt.Println("----------------- endblock", tt["startBlock"]+tt["reorgLength"]+50)
			break
		}

		select {
		case ev := <-chain2HeadCh:
			fmt.Println("##############")
			fmt.Printf("\n---------------\n%+v\n---------------\n", ev.NewChain[0].Header().Number.Uint64())
			var newAuthor common.Address
			var oldAuthor common.Address
			if len(ev.NewChain) > 0 {
				newAuthor, _ = nodes[0].Engine().Author(ev.NewChain[0].Header())
			}
			if len(ev.OldChain) > 0 {
				oldAuthor, _ = nodes[0].Engine().Author(ev.OldChain[0].Header())
			}
			fmt.Printf("\n---------------\nOld Author : %+v :::: New Author : %+v\n---------------\n", oldAuthor, newAuthor)
			fmt.Printf("\n---------------\n%+v\n---------------\n", ev)
			fmt.Println("##############")

		default:
			time.Sleep(500 * time.Millisecond)
			// if len(ev.NewChain) != len(expect.Added) {
			// 	t.Fatal("Newchain and Added Array Size don't match")
			// }
			// if len(ev.OldChain) != len(expect.Removed) {
			// 	t.Fatal("Oldchain and Removed Array Size don't match")
			// }

			// for j := 0; j < len(ev.OldChain); j++ {
			// 	if ev.OldChain[j].Hash() != expect.Removed[j] {
			// 		t.Fatal("Oldchain hashes Do Not Match")
			// 	}
			// }
			// for j := 0; j < len(ev.NewChain); j++ {
			// 	if ev.NewChain[j].Hash() != expect.Added[j] {
			// 		t.Fatalf("Newchain hashes Do Not Match %s %s", ev.NewChain[j].Hash(), expect.Added[j])
			// 	}
			// }
		}
	}

}

func InitMiner1(genesis *core.Genesis, privKey *ecdsa.PrivateKey, withoutHeimdall bool) (*node.Node, *eth.Ethereum, error) {
	// Define the basic configurations for the Ethereum node
	datadir, _ := ioutil.TempDir("", "")

	config := &node.Config{
		Name:    "geth",
		Version: params.Version,
		DataDir: datadir,
		P2P: p2p.Config{
			ListenAddr:  "0.0.0.0:0",
			NoDiscovery: true,
			MaxPeers:    25,
		},
		UseLightweightKDF: true,
	}
	// Create the node and configure a full Ethereum node on it
	stack, err := node.New(config)
	if err != nil {
		return nil, nil, err
	}
	ethBackend, err := eth.New(stack, &ethconfig.Config{
		Genesis:         genesis,
		NetworkId:       genesis.Config.ChainID.Uint64(),
		SyncMode:        downloader.FullSync,
		DatabaseCache:   256,
		DatabaseHandles: 256,
		TxPool:          core.DefaultTxPoolConfig,
		GPO:             ethconfig.Defaults.GPO,
		Ethash:          ethconfig.Defaults.Ethash,
		Miner: miner.Config{
			Etherbase: crypto.PubkeyToAddress(privKey.PublicKey),
			GasCeil:   genesis.GasLimit * 11 / 10,
			GasPrice:  big.NewInt(1),
			Recommit:  time.Second,
		},
		WithoutHeimdall: withoutHeimdall,
	})
	if err != nil {
		return nil, nil, err
	}

	// register backend to account manager with keystore for signing
	keydir := stack.KeyStoreDir()

	n, p := keystore.StandardScryptN, keystore.StandardScryptP
	kStore := keystore.NewKeyStore(keydir, n, p)

	kStore.ImportECDSA(privKey, "")
	acc := kStore.Accounts()[0]
	kStore.Unlock(acc, "")
	// proceed to authorize the local account manager in any case
	ethBackend.AccountManager().AddBackend(kStore)

	// ethBackend.AccountManager().AddBackend()
	err = stack.Start()
	return stack, ethBackend, err
}

func InitGenesis1(t *testing.T, faucets []*ecdsa.PrivateKey, fileLocation string) *core.Genesis {

	// sprint size = 8 in genesis
	genesisData, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		t.Fatalf("%s", err)
	}

	genesis := &core.Genesis{}

	if err := json.Unmarshal(genesisData, genesis); err != nil {
		t.Fatalf("%s", err)
	}

	genesis.Config.ChainID = big.NewInt(15001)
	genesis.Config.EIP150Hash = common.Hash{}

	return genesis
}
