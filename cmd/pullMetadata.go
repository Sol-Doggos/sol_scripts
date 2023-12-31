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
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	jsonFile       string
	collectionName string
)

type JsonType struct {
	Array []string
}
type HeliusTokenRequestBody struct {
	MintAccounts    []string `json:"mintAccounts"`
	IncludeOffChain bool     `json:"includeOffChain"`
	DisableCache    bool     `json:"disableCache"`
}

type HeliusTokenResponse struct {
	Account            string             `json:"account"`
	OnChainAccountInfo onChainAccountInfo `json:"onChainAccountInfo"`
	OnChainMetadata    onChainMetadata    `json:"onChainMetadata"`
	OffChainMetadata   offChainMetadata   `json:"offChainMetadata"`
}

type onChainMetadata struct {
	Metadata struct {
		TokenStandard   string `json:"tokenStandard"`
		Key             string `json:"key"`
		UpdateAuthority string `json:"updateAuthority"`
		Mint            string `json:"mint"`
		Data            struct {
			Name                 string `json:"name"`
			Symbol               string `json:"symbol"`
			Uri                  string `json:"uri"`
			SellerFeeBasisPoints int    `json:"sellerFeeBasisPoints"`
			Creators             []struct {
				Address  string `json:"address"`
				Verified bool   `json:"verified"`
				Share    int    `json:"share"`
			} `json:"creators"`
		} `json:"data"`
		PrimarySaleHappened bool `json:"primarySaleHappened"`
		IsMutable           bool `json:"isMutable"`
		EditionNonce        int  `json:"editionNonce"`
		Uses                struct {
			UseMethod string `json:"useMethod"`
			Remaining int    `json:"remaining"`
			Total     int    `json:"total"`
		} `json:"uses"`
		Collection struct {
			Key      string `json:"key"`
			Verified bool   `json:"verified"`
		} `json:"collection"`
	} `json:"metadata"`
}

type onChainAccountInfo struct {
	AccountInfo struct {
		Key        string `json:"key"`
		IsSigner   bool   `json:"isSigner"`
		IsWritable bool   `json:"isWritable"`
		Lamports   int    `json:"lamports"`
		Data       struct {
			Parsed struct {
				Info struct {
					Decimals        int    `json:"decimals"`
					FreezeAuthority string `json:"freezeAuthority"`
					IsInitialized   bool   `json:"isInitialized"`
					MintAuthority   string `json:"mintAuthority"`
					Supply          string `json:"supply"`
				} `json:"info"`
				MintType string `json:"type"`
			} `json:"parsed"`
			Program string `json:"program"`
			Space   int    `json:"space"`
		} `json:"data"`
		Owner      string `json:"owner"`
		Executable bool   `json:"executable"`
		RentEpoch  int    `json:"rentEpoch"`
	} `json:"accountInfo"`
	Error string `json:"error"`
}
type offChainMetadata struct {
	Metadata struct {
		Attributes []struct {
			TraitType string `json:"trait_type"`
			Value     string `json:"value"`
		} `json:"attributes"`
		Description string `json:"description"`
		Image       string `json:"image"`
		Name        string `json:"name"`
		Properties  struct {
			Category string `json:"category"`
			Creators []struct {
				Address string `json:"address"`
				Share   int    `json:"share"`
			} `json:"creators"`
			Files []struct {
				Type string `json:"type"`
				Uri  string `json:"uri"`
			} `json:"files"`
		} `json:"properties"`
		SellerFeeBasisPoints int    `json:"seller_fee_basis_points"`
		Symbol               string `json:"symbol"`
	} `json:"metadata"`
	Uri   string `json:"uri"`
	Error string `json:"error"`
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

		jsonContents, err := os.ReadFile(jsonFile)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(jsonContents, &mintList)
		if err != nil {
			fmt.Println(err)
		}
		metadataPath := filepath.Join(".", "downloads", collectionName, "metadata")
		err = os.MkdirAll(metadataPath, os.ModePerm)
		if err != nil {
			os.Exit(1)
		}

		imagePath := filepath.Join(".", "downloads", collectionName, "images")
		err = os.MkdirAll(imagePath, os.ModePerm)
		if err != nil {
			os.Exit(1)
		}

		// TODO - Split by 100 Mints
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
		// fmt.Println(result)
		for _, data := range result {
			if data.OffChainMetadata.Error != "" {
				// add error to error log
				fmt.Println(data.Account + ": error pulling offchain metadata " + data.OffChainMetadata.Error)
				continue
			}
			fmt.Println(metadataPath + data.Account + ".json")
			file, err := os.Create(metadataPath + "/" + data.Account + ".json")
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.Encode(data.OffChainMetadata.Metadata)

			fmt.Println("Pulling image for " + data.Account + " from " + data.OffChainMetadata.Metadata.Image)
			res, err := http.Get(data.OffChainMetadata.Metadata.Image)
			if err != nil {
				fmt.Printf("client: could not create suest: %s\n", err)
				os.Exit(1)
			}
			defer res.Body.Close()

			extension := determineExtension(data.OffChainMetadata.Metadata.Properties.Files[0].Type)

			file, err = os.Create(imagePath + "/" + data.Account + extension)
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			_, err = io.Copy(file, res.Body)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Successfully pulled image for " + data.Account)
		}
	},
}

func init() {
	rootCmd.AddCommand(pullMetadataCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pullMetadataCmd.PersistentFlags().String("foo", "", "A help for foo")
	pullMetadataCmd.PersistentFlags().StringVar(&jsonFile, "mintList", "", "Where mint list json lives")
	pullMetadataCmd.PersistentFlags().StringVar(&collectionName, "collectionName", "collection", "Collection name for which we pull data")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pullMetadataCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func determineExtension(fileType string) string {
	switch fileType {
	case "image/gif":
		return ".gif"
	case "image/jpeg":
		return ".jpeg"
	default:
		return ".png"
	}
}
