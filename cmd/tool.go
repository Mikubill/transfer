package cmd

import (
	"fmt"
	"github.com/Mikubill/transfer/apis/image"
	"github.com/Mikubill/transfer/crypto"
	"github.com/Mikubill/transfer/hash"
	"github.com/spf13/cobra"
)

var (
	hashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Hash a file",
		Long: `
Hash a file. We will hash it on crc32, md5, sha1 and sha256.

Example:
  transfer hash your-file

Note: Large file may need more time to finish.
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := uploadWalker(args)
			if len(files) != 0 {
				hash.Hash(files)
			} else {
				fmt.Println("Error: no file detected.")
				fmt.Println("Use \"transfer tool hash --help\" for more information.")
			}

		},
	}

	encryptCmd = &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a file",
		Long: `
Encrypt a file (Using AES-ECB Method). You can specify the password or we will generate it for you.

Example:
  transfer encrypt your-file

  # specify path
  transfer encrypt -o output your-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := uploadWalker(args)
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
  transfer decrypt -k your-password your-encrypted-file

  # specify path
  transfer encrypt -o output your-encrypted-file
`,
		Run: func(cmd *cobra.Command, args []string) {
			files := uploadWalker(args)
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
	crypto.InitCmd(encryptCmd)
	crypto.InitCmd(decryptCmd)

	rootCmd.AddCommand(hashCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)

	//rootCmd.AddCommand(toolCmd)
}
