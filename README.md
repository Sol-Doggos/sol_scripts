# sol_scripts

Just a simple script to pull down off-chain metadata, we used it to migrate over to shdwDrive because of an IPFS issue with some of our collections

My migration process was:
1. Pull unaltered metadata as a backup incase the script messes things up
2. Create a ShdwDrive storage account and use that as the prefix in step 3
3. Pull images and metadata allowing the script to change the location for the images in the pulled off-chain metadata
4. Use ShdwDrive UI to upload images and Metadata to shdwDrive (could use cli or sdk script if you want but this felt easier)
5. Run `metaboss update uri-all` with the changeList.json generated by the script

`go run main.go pullMetadata --mintList {pathToMintListJson} --collectionName {nameOfCollection} --newMetadataPrefix {shadowDriveStorageAccount}`

Other options: `--skipImages true` was helpful to create backup of OG metadata in the case that I made a mistake (of which I made a few during this process)

This script will generate several things in the downloads folder under a collection sub directory:
1. images folder with the images pulled from wherever they live now
2. OffChain metadata
3. Change list file used by Metaboss to change offchain metadata url
4. errors a list of mints for which it was not able to pull information


Future improvements:
- [ ] Upload to shdwdrive with wallet that solana-cli is using
- [ ] Improve error handling (currently burnt mints show in this)
- [ ] Pull mintList from collection so that it doesn't need to be entered

This script could possibly not be useful to anyone, I wrote it in golang because it's what I felt like doing at the time.
Javascript would have been better so that I could use ShdwDrive SDK so I might move that way at some point