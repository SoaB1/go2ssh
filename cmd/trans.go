/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"go2ssh/config"
	"io"
	"os"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var csvReader = csv.NewReader

// transCmd represents the trans command
var transCmd = &cobra.Command{
	Use:   "trans",
	Short: "Transfer file to SFTP Server",
	Long: `Transfer file to SFTP Server using a CSV file.

Example:
  # Transfer file to SFTP Server using a CSV file
  go2ssh trans -f /path/to/file.csv

  # Transfer file to SFTP Server using a CSV file,Using a non-default configuration file
  go2ssh trans -f /path/to/file.csv --config /path/to/config.yaml

# CSV file format has need 3 columns.
# Column 1: Source file path
# Column 2: Destination file path
# Column 3: File permission
# Permission must be expressed in octal(3 digits) 
# docs: https://en.wikipedia.org/wiki/File_system_permissions#Numeric_notation

CSV file format:
Column1,Column2,Column3
# Server1 to Server2 (Commnet line are skipped)
/home/user1/server1_file.txt,/home/user2/server2_file1.txt,644
/home/user1/server1_file.txt,/home/user2/server2_file2.txt,644

`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgs := config.Conf.SSHConfigs
		csv, err := cmd.Flags().GetString("csv")
		if err != nil {
			fmt.Println(err)
		}

		f, err := os.Open(csv)
		if err != nil {
			fmt.Println(err)
		}
		r := csvReader(f)
		r.Comma = ','
		r.Comment = '#'
		r.FieldsPerRecord = 3
		r.TrimLeadingSpace = true
		r.ReuseRecord = true

		// Skip header
		_, err = r.Read()
		if err == io.EOF {
			fmt.Println("No records found")
		}

		srcFile := []string{}
		dstFile := []string{}
		filePerm := []string{}

		// Read records
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			srcFile = append(srcFile, record[0])
			dstFile = append(dstFile, record[1])
			filePerm = append(filePerm, record[2])
		}

		key, err := os.ReadFile(cfgs.KeyPath)
		if err != nil {
			fmt.Println(err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			fmt.Println(err)
		}

		// Set SSH Config
		sshConfig := &ssh.ClientConfig{
			User: cfgs.UserName,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		// Connect to SSH Server
		sshConfig.SetDefaults()
		sshConnection, err := ssh.Dial("tcp", cfgs.Server+":"+cfgs.Port, sshConfig)
		if err != nil {
			fmt.Println(err)
		}
		defer sshConnection.Close()

		// Create SFTP Client
		sftpClient, err := sftp.NewClient(sshConnection)
		if err != nil {
			fmt.Println(err)
		}
		defer sftpClient.Close()

		// Copy files
		for i, v := range srcFile {
			src, err := os.Open(v)
			if err != nil {
				fmt.Println(err)
			}
			defer src.Close()

			dst, err := sftpClient.Create(dstFile[i])
			if err != nil {
				fmt.Println(err)
			}
			defer dst.Close()

			bytes, err := io.Copy(dst, src)
			if err != nil {
				fmt.Println(err)
			}
			s, err := sshConnection.NewSession()
			if err != nil {
				fmt.Println(err)
			}
			if err := s.Run("chmod " + filePerm[i] + " " + dstFile[i]); err != nil {
				fmt.Println(err)
			}
			fmt.Printf("%d bytes copied\n", bytes)
		}

	},
}

func init() {
	rootCmd.AddCommand(transCmd)

	transCmd.Flags().StringP("csv", "f", "", "CSV File")

}
