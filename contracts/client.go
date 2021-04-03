package contracts

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jongregis/linkPoolBackend/env"
)

// init of the contract instance
func getContract() *AccessControlledAggregator {
	viperenv := env.ViperEnvVariable("HTTPS")
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}
	contract, err := NewAccessControlledAggregator(common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419"), client)
	if err != nil {
		log.Fatal(err)
	}

	return contract
}

// still trying to figure out event watching??
func watchContract() *AccessControlledAggregator {
	viperenv := env.ViperEnvVariable("WSS")
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}

	contract2, err := NewAccessControlledAggregator(common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419"), client)
	if err != nil {
		log.Fatal(err)
	}

	return contract2
}

//subscribing to event
func SubscribeToEventContract() {
	viperenv := env.ViperEnvVariable("WSS")

	fmt.Println("Starting Event Subscription...")
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419")
	query := ethereum.FilterQuery{Addresses: []common.Address{contractAddress}}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println(vLog)
		}
	}
}

// Reading the event log
func ReadEventLogs() {
	viperenv := env.ViperEnvVariable("WSS")

	fmt.Println("Reading Event Logs...")
	client, err := ethclient.Dial(viperenv)
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress("0xF570deEffF684D964dc3E15E1F9414283E3f7419")
	query := ethereum.FilterQuery{
		// FromBlock: big.NewInt(12142943),
		// ToBlock:   big.NewInt(12142943),
		Addresses: []common.Address{contractAddress}}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}
	contractAbi, err := abi.JSON(strings.NewReader(string(AccessControlledAggregatorABI)))
	if err != nil {
		log.Fatal(err)
	}

	for _, vLog := range logs {
		fmt.Println("BlockHash: ", vLog.BlockHash.Hex())
		fmt.Println("BlockNumber: ", vLog.BlockNumber)
		fmt.Println("TxHash: ", vLog.TxHash.Hex())
		event := struct {
			Current   int
			RoundId   uint
			UpdatedAt uint
		}{}
		err := contractAbi.UnpackIntoInterface(&event, "AnswerUpdated", vLog.Data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(event.Current)
		fmt.Println(event.RoundId)
		fmt.Println(event.UpdatedAt)

		var topics [4]string
		for i := range vLog.Topics {
			topics[i] = vLog.Topics[i].Hex()
		}
		fmt.Println(topics[0])

	}
}

// latest price of BTC/USD on contract
func LatestAnswerFunc() string {
	contract := getContract()

	amt, _ := contract.LatestAnswer(&bind.CallOpts{})
	bigStr := amt.String()

	return bigStr
}

// latest round data on contract
func LatestRoundFunc() map[string]string {
	contract := getContract()

	amt, _ := contract.LatestRoundData(&bind.CallOpts{})
	m := make(map[string]string)
	m["answer"] = amt.Answer.String()
	m["round"] = amt.RoundId.String()
	m["updated"] = amt.UpdatedAt.String()

	return m
}

// info from specific round from contract
func RoundData(round string) map[string]string {
	contract := getContract()
	i := new(big.Int)
	_, err := fmt.Sscan(round, i)
	if err != nil {
		log.Println("error scanning value:", err)
	}

	amt, _ := contract.GetRoundData(&bind.CallOpts{}, i)

	m := make(map[string]string)
	m["answer"] = amt.Answer.String()
	m["round"] = amt.RoundId.String()
	m["updated"] = amt.UpdatedAt.String()
	return m
}
