/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonFile string
)

type HeliusTokenRequestBody struct {
	MintAccounts    []string `json:"mintAccounts"`
	IncludeOffChain bool     `json:"includeOffChain"`
	DisableCache    bool     `json:"disableCache"`
}

type HeliusTokenResponse struct {
	Account            string `json:"account"`
	onChainAccountInfo struct {
		accountInfo struct {
			key        string `json:"key"`
			isSigner   bool   `json:"isSigner"`
			isWritable bool   `json:"isWritable"`
			lamports   int    `json:"lamports"`
			data       struct {
				parsed struct {
					info struct {
						decimals        int    `json:"decimals"`
						freezeAuthority string `json:"freezeAuthority"`
						isInitialized   bool   `json:"isInitialized"`
						mintAuthority   string `json:"mintAuthority"`
						supply          string `json:"supply"`
					} `json:"info"`
					mintType string `json:"type"`
				} `json:"parsed"`
				program string `json:"program"`
				space   int    `json:"space"`
			} `json:"data"`
			owner      string `json:"owner"`
			executable bool   `json:"executable"`
			rentEpoch  int    `json:"rentEpoch"`
		} `json:"accountInfo"`
		error_ string `json:"error"`
	} `json:"onChainAccountInfo"`
	onChainMetadata struct {
		metadata struct {
			tokenStandard   string `json:"tokenStandard"`
			key             string `json:"key"`
			updateAuthority string `json:"updateAuthority"`
			mint            string `json:"mint"`
			data            struct {
				name                 string `json:"name"`
				symbol               string `json:"symbol"`
				uri                  string `json:"uri"`
				sellerFeeBasisPoints int    `json:"sellerFeeBasisPoints"`
				creators             []struct {
					address  string `json:"address"`
					verified bool   `json:"verified"`
					share    int    `json:"share"`
				} `json:"creators"`
			} `json:"data"`
			primarySaleHappened bool `json:"primarySaleHappened"`
			isMutable           bool `json:"isMutable"`
			editionNonce        int  `json:"editionNonce"`
			uses                struct {
				useMethod string `json:"useMethod"`
				remaining int    `json:"remaining"`
				total     int    `json:"total"`
			} `json:"uses"`
			collection struct {
				key      string `json:"key"`
				verified bool   `json:"verified"`
			} `json:"collection"`
		} `json:"metadata"`
	} `json:"onChainMetadata"`
	offChainMetadata offChainMetadata `json:"offChainMetadata"`
}

type offChainMetadata struct {
	metadata struct {
		attributes []struct {
			traitType string `json:"trait_type"`
			value     string `json:"value"`
		} `json:"attributes"`
		description string `json:"description"`
		image       string `json:"image"`
		name        string `json:"name"`
		properties  struct {
			category string `json:"category"`
			creators []struct {
				address string `json:"address"`
				share   int    `json:"share"`
			} `json:"creators"`
			files []struct {
				type_ string `json:"type"`
				uri   string `json:"uri"`
			} `json:"files"`
		} `json:"properties"`
		sellerFeeBasisPoints int    `json:"seller_fee_basis_points"`
		symbol               string `json:"symbol"`
	} `json:"metadata"`
	uri    string `json:"uri"`
	error_ string `json:"error"`
}

// pullMetadataCmd represents the pullMetadata command
var pullMetadataCmd = &cobra.Command{
	Use:   "pullMetadata",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var mintList []string
		// heliusUrl := "https://api.helius.xyz/v0/token-metadata?api-key=" + os.Getenv("HELIUS_API_KEY")
		jsonContents, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(jsonContents, &mintList)
		if err != nil {
			fmt.Println(err)
		}
		// jsonData, err := json.Marshal(mintList)
		// rl := rate.NewLimiter(rate.Every(4*time.Second), 50)
		reqURL := "https://api.helius.xyz/v0/token-metadata?api-key=" + os.Getenv("HELIUS_API_KEY")
		jsonBody := &HeliusTokenRequestBody{
			MintAccounts:    mintList,
			IncludeOffChain: true,
			DisableCache:    false,
		}
		// c := http.Client{}
		// jsonBody := []byte(`{"mintAccounts": ` + jsonData + `, "includeOffChain": true, "disableCache": false }`)
		jsonData, _ := json.MarshalIndent(jsonBody, "", "  ")
		bodyReader := bytes.NewReader(jsonData)
		req, err := http.NewRequest(http.MethodPost, reqURL, bodyReader)
		if err != nil {
			fmt.Printf("client: could not create request: %s\n", err)
			os.Exit(1)
		}

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			fmt.Printf("client: error making http request: %s\n", err)
			os.Exit(1)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("response body: %s\n", body)
		var result []HeliusTokenResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("error unmarshalling response body: %s\n", err)
		}
		for _, data := range result {
			fmt.Println(data)
			fmt.Println(data.offChainMetadata)
			fmt.Println(data.onChainMetadata)
			// fmt.Println(data.onChainMetadata.metadata.mint)
			// fmt.Println(data.Account)
		}
	},
}

func init() {
	rootCmd.AddCommand(pullMetadataCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullMetadataCmd.PersistentFlags().String("foo", "", "A help for foo")
	pullMetadataCmd.PersistentFlags().StringVar(&jsonFile, "jsonFile", "", "Where mint list json lives")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullMetadataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
