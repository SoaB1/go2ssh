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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
