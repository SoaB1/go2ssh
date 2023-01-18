/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
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

		srcpath, err := cmd.Flags().GetString("src")
		if err != nil {
			fmt.Println(err)
		}
		dstpath, err := cmd.Flags().GetString("dst")
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
	},
}

func init() {
	rootCmd.AddCommand(transCmd)

}

func conSSH(sshUser, sshKey, sshHost, sshPort string) {
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
}
