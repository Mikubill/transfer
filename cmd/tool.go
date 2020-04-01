package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"transfer/apis/image"
	"transfer/chunk"
	"transfer/crypto"
	"transfer/hash"
	"transfer/utils"
)

var (
	toolCmd = &cobra.Command{
		Use:   "tool",
		Short: "File process toolbox",
	}
	hashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Hash a file",
		Long: `
Hash a file. We will hash it on crc32, md5, sha1 and sha256.

Example:
  transfer tool hash your-file

Note: Large file may need more time to finish.
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := utils.UploadWalker(args)
			if len(files) != 0 {
				hash.Hash(files)
			} else {
				fmt.Println("Error: no file detected.")
				fmt.Println("Use \"transfer tool hash --help\" for more information.")
			}

		},
	}
	chunkCmd = &cobra.Command{
		Use:   "chunk",
		Short: "Chunked upload to imageBed(beta)",
		Long: `
Chunk a file and upload chunks to image bed.
Cuz it's a unusual application, your ip address will be blocked by some services if you abuse this function.

Example:
  transfer tool chunk your-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				{
					fmt.Println("Error: no file detected.")
					fmt.Println("Use \"transfer tool chunk --help\" for more information.")
				}
			}
			for _, item := range args {
				if utils.IsExist(item) && utils.IsFile(item) {
					chunk.Upload(item)
				} else if chunk.Matcher(item) {
					chunk.Download(item)
				}
			}
		},
	}

	encryptCmd = &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a file",
		Long: `
Encrypt a file (Using AES-ECB Method). You can specify the password or we will generate it for you.

Example:
  transfer tool encrypt your-file

  # specify path
  transfer tool encrypt -o output your-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := utils.UploadWalker(args)
			if len(files) != 0 {
				for _, file := range files {
					err := crypto.Encrypt(file)
					if err != nil {
						fmt.Printf("encrypt failed: %s\n", err)
					}
				}
			} else {
				fmt.Println("Error: no file detected.")
				fmt.Println("Use \"transfer tool encrypt --help\" for more information.")
			}

		},
	}
	decryptCmd = &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt a file",
		Long: `
Decrypt a file. You must specify the password.

Example:
  transfer tool decrypt -k your-password your-encrypted-file

  # specify path
  transfer tool encrypt -o output your-encrypted-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := utils.UploadWalker(args)
			if len(files) != 0 {
				for _, file := range files {
					err := crypto.Decrypt(file)
					if err != nil {
						fmt.Printf("decrypt failed: %s\n", err)
					}
				}
			} else {
				fmt.Println("Error: no file detected.")
				fmt.Println("Use \"transfer tool decrypt --help\" for more information.")
			}

		},
	}
)

func init() {
	image.InitCmd(picCmd)
	chunk.InitCmd(chunkCmd)
	crypto.InitCmd(encryptCmd)
	crypto.InitCmd(decryptCmd)

	toolCmd.AddCommand(hashCmd)
	toolCmd.AddCommand(chunkCmd)
	toolCmd.AddCommand(encryptCmd)
	toolCmd.AddCommand(decryptCmd)

	rootCmd.AddCommand(toolCmd)
}
