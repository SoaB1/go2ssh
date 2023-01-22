/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

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
		cuser, err := user.Current()
		if err != nil {
			fmt.Println(err)
		}

		csv, err := cmd.Flags().GetString("csv")
		if err != nil {
			fmt.Println(err)
		}

		sshHost, _ := cmd.Flags().GetString("host")
		sshPort, _ := cmd.Flags().GetString("port")

		sshUser, _ := cmd.Flags().GetString("login")
		if sshUser == "" {
			sshUser = cuser.Username
		}
		sshKey, _ := cmd.Flags().GetString("key")
		if sshKey == "" {
			sshKey = cuser.HomeDir + "/.ssh/id_rsa"
		}
		copyFile(sshUser, sshKey, sshHost, sshPort, csv)
	},
}

func init() {
	rootCmd.AddCommand(transCmd)

	transCmd.Flags().StringP("host", "H", "localhost", "SSH Host")
	transCmd.Flags().StringP("port", "P", "22", "SSH Port")
	transCmd.Flags().StringP("login", "l", "", "SSH Login")
	transCmd.Flags().StringP("identity", "i", "", "SSH Identity File")
	transCmd.Flags().StringP("csv", "f", "", "CSV File")

}

func copyFile(sshUser, sshKey, sshHost, sshPort, csv string) {
	// Parse CSV
	f, err := os.Open(csv)
	if err != nil {
		fmt.Println(err)
	}

	r := csv.NewReader(f)
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

	key, err := os.ReadFile(sshKey)
	if err != nil {
		fmt.Println(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println(err)
	}

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshConfig.SetDefaults()
	sshConnection, err := ssh.Dial("tcp", sshHost+":"+sshPort, sshConfig)
	if err != nil {
		fmt.Println(err)
	}
	defer sshConnection.Close()

	sftpClient, err := sftp.NewClient(sshConnection)
	if err != nil {
		fmt.Println(err)
	}
	defer sftpClient.Close()

	for i := range srcFile {
		destFile, err := sftpClient.Create(dstFile[i])
		if err != nil {
			fmt.Println(err)
		}
		defer destFile.Close()

		srcFile, err := os.Open(srcFile[i])
		if err != nil {
			fmt.Println(err)
		}
		defer srcFile.Close()

		bytes, err := io.Copy(destFile, srcFile)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%d bytes copied\n", bytes)
	}
}
