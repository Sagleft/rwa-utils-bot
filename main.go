package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tonkeeper/tonapi-go"
	"gopkg.in/telebot.v4"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}

	log.Println("app finished")
}

func run() error {
	cfg, err := GetConfig()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	client, err := tonapi.NewClient(tonapi.TonApiURL, &tonapi.Security{})
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	b, err := telebot.NewBot(telebot.Settings{
		Token: cfg.BotToken,
	})
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	b.Handle("/start", func(ctx telebot.Context) error {
		if ctx.Sender().ID != cfg.AdminID {
			return ctx.Send("access denied!")
		}

		return ctx.Send(fmt.Sprintf(
			"Hi! Use command /getcsv to obtain payout data.\n\n"+
				"Example: `/getcsv 1000`\n\n(your telegram id is %v)",
			ctx.Sender().ID,
		), telebot.ModeMarkdown)
	})

	b.Handle("/getcsv", func(ctx telebot.Context) error {
		if ctx.Sender().ID != cfg.AdminID {
			return ctx.Send("access denied!")
		}

		userInput := filterUserInput(ctx.Message().Text)
		if userInput == "" {
			return ctx.Send("please specify project revenue")
		}

		commandParts := strings.Split(userInput, " ")
		if len(commandParts) < 2 {
			return ctx.Send("please specify project revenue")
		}

		projectRevenue, err := strconv.ParseFloat(commandParts[1], 64)
		if err != nil {
			return ctx.Send(err.Error())
		}

		data, err := getPayoutDataCSV(client, cfg, projectRevenue)
		if err != nil {
			return ctx.Send(err.Error())
		}

		msg := &telebot.Document{
			File:     telebot.FromReader(strings.NewReader(data)),
			FileName: "data.csv",
			Caption:  "Payment data for NFT owners",
		}

		return ctx.Send(msg)
	})

	log.Println("app started")
	b.Start() // blocking method
	return nil
}

func getPayoutDataCSV(
	client *tonapi.Client,
	cfg Config,
	projectRevenue float64,
) (string, error) {
	data, err := client.GetItemsFromCollection(
		context.Background(),
		tonapi.GetItemsFromCollectionParams{
			AccountID: cfg.CollectionAddress,
		},
	)
	if err != nil {
		return "", fmt.Errorf("get items: %w", err)
	}

	nftsCount := len(data.NftItems)

	// revenue * yield / nft count
	oneNFTPayout := decimal.NewFromFloat(projectRevenue).
		Mul(decimal.NewFromFloat(cfg.NftYield)).
		Div(decimal.NewFromInt(100)).
		Div(decimal.NewFromInt(int64(nftsCount))).
		Round(4)

	toPayout := map[string]decimal.Decimal{
		// owner address -> payout amount
	}

	for _, item := range data.NftItems {
		if item.Sale.IsSet() {
			continue // ignore NFT on sale
		}

		prevAmount, isSet := toPayout[item.Owner.Value.Address]
		if !isSet {
			toPayout[item.Owner.Value.Address] = decimal.Zero
			prevAmount = decimal.Zero
		}

		toPayout[item.Owner.Value.Address] = prevAmount.Add(oneNFTPayout)
	}

	return mapToCSVData(toPayout), nil
}
