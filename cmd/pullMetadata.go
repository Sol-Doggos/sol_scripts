/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
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
	skipImages     bool
	imageURLprefix string
	change         bool
	changeList     ChangeList
)

type ChangeList []Change

type Change struct {
	MintAccount string `json:"mint_account"`
	NewURI      string `json:"new_uri"`
}

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
		SellerFeeBasisPoints int    `json:"sellerFeeBasisPoints"`
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
		mintList, err := readJsonFile(jsonFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = createFolder(collectionName, "metadata")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if !skipImages {
			err = createFolder(collectionName, "images")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		createFileIfNotExist(filepath.Join(".", "downloads", collectionName, "errors.txt"))

		batch := 99
		for i := 0; i < len(mintList); i += batch {
			j := i + batch
			if j > len(mintList) {
				j = len(mintList)
			}
			mintsInBatch := mintList[i:j]
			result, err := pullMetadata(mintsInBatch)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, data := range result {
				var fileName string
				var err error
				if data.OffChainMetadata.Error != "" {
					fmt.Println(data.Account + ": error pulling offchain metadata " + data.OffChainMetadata.Error)
					addErrorToFile(data.Account, data.OffChainMetadata.Error)
					continue
				}
				if !skipImages {
					fileName, err = downloadImage(data.OffChainMetadata.Metadata.Image, data.Account)
					if err != nil {
						addErrorToFile(data.Account, err.Error())
						continue
					}
				}

				err = saveMetadata(data, fileName)
				if err != nil {
					addErrorToFile(data.Account, err.Error())
					continue
				}

				if err != nil {
					addErrorToFile(data.Account, err.Error())
					continue
				}

				if change {
					// fmt.Println(data.Account)
					// fmt.Println(imageURLprefix + fileName)
					changeList = append(changeList, Change{
						MintAccount: data.Account,
						NewURI:      imageURLprefix + fileName,
					})
				}
			}
			if change {
				fmt.Println(changeList)
				jsonData, err := json.MarshalIndent(changeList, "", "  ")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println(jsonData)

				file, err := os.Create(filepath.Join(".", "downloads", collectionName, "changeList.json"))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				defer file.Close()

				_, err = file.Write(jsonData)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
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
	pullMetadataCmd.PersistentFlags().BoolVar(&change, "change", false, "Change image path in metadata")
	pullMetadataCmd.PersistentFlags().StringVar(&imageURLprefix, "imageURLprefix", "", "Prefix for image URL")
	pullMetadataCmd.PersistentFlags().BoolVar(&skipImages, "skipImages", false, "Skip downloading images")
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

func readJsonFile(jsonFile string) ([]string, error) {
	jsonContents, err := os.ReadFile(jsonFile)
	if err != nil {
		return []string{}, err
	}
	var mintList []string
	err = json.Unmarshal(jsonContents, &mintList)
	if err != nil {
		return []string{}, err
	}
	return mintList, nil
}

func createFolder(collectionName, folderName string) error {
	metadataPath := filepath.Join(".", "downloads", collectionName, folderName)
	err := os.MkdirAll(metadataPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func pullMetadata(mintList []string) ([]HeliusTokenResponse, error) {
	var reqURL = "https://api.helius.xyz/v0/token-metadata?api-key=" + os.Getenv("HELIUS_API_KEY")
	jsonBody := &HeliusTokenRequestBody{
		MintAccounts:    mintList,
		IncludeOffChain: true,
		DisableCache:    false,
	}

	jsonData, err := json.MarshalIndent(jsonBody, "", "  ")
	if err != nil {
		return []HeliusTokenResponse{}, err
	}

	bodyReader := bytes.NewReader(jsonData)
	req, err := http.NewRequest(http.MethodPost, reqURL, bodyReader)
	if err != nil {
		return []HeliusTokenResponse{}, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return []HeliusTokenResponse{}, err
	}

	if res.StatusCode != 200 {
		return []HeliusTokenResponse{}, errors.New("request failed with status code: " + string(res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []HeliusTokenResponse{}, err
	}

	var result []HeliusTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return []HeliusTokenResponse{}, err
	}
	return result, nil
}

func downloadImage(url string, mint string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	fileType := response.Header.Get("Content-Type")
	fileExtension := determineExtension(fileType)
	fileName := mint + fileExtension
	filePath := filepath.Join(".", "downloads", collectionName, "images", fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func createFileIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func addErrorToFile(mint, error_ string) error {
	file, err := os.OpenFile(filepath.Join(".", "downloads", collectionName, "errors.txt"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(mint + ", " + error_ + "\n")
	if err != nil {
		return err
	}
	return nil
}

func saveMetadata(data HeliusTokenResponse, imageFileName string) error {
	metadataPath := filepath.Join(".", "downloads", collectionName, "metadata")
	file, err := os.Create(metadataPath + "/" + data.Account + ".json")
	if err != nil {
		return err
	}
	defer file.Close()
	if !skipImages && change && imageURLprefix != "" {
		newImageURL := imageURLprefix + imageFileName
		data.OffChainMetadata.Metadata.Image = newImageURL
		data.OffChainMetadata.Metadata.Properties.Files[0].Uri = newImageURL
	}

	jsonData, err := json.MarshalIndent(data.OffChainMetadata.Metadata, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}
