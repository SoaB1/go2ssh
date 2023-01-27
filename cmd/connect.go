package cmd

import (
	"fmt"
	"go2ssh/config"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connet to SSH Server",
	Long: `Connet to SSH Server

Example:
  # Connect to SSH Server
  go2ssh connect

  # Connect to SSH Server,Using a non-default configuration file
  go2ssh connect --config /path/to/config.yaml
`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgs := config.Conf.SSHConfigs
		cuser, err := user.Current()
		if err != nil {
			fmt.Println(err)
		}

		sshUser := cfgs.UserName
		if sshUser == "" {
			sshUser = cuser.Username
		}

		sshKey := cfgs.KeyPath
		if sshKey == "" {
			fmt.Println("Please set key file path")
		}

		key, err := os.ReadFile(sshKey)
		if err != nil {
			fmt.Printf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			fmt.Printf("unable to parse private key: %v", err)
		}

		sshConfig := &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		client, err := ssh.Dial("tcp", cfgs.Server+":"+cfgs.Port, sshConfig)
		if err != nil {
			fmt.Printf("unable to connect: %v", err)
		}
		session, err := client.NewSession()
		if err != nil {
			fmt.Printf("unable to create session: %v", err)
		}
		defer session.Close()

		fd := int(os.Stdin.Fd())
		state, err := terminal.MakeRaw(fd)
		if err != nil {
			fmt.Println(err)
		}
		defer terminal.Restore(fd, state)

		w, h, err := terminal.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}

		err = session.RequestPty("xterm", h, w, modes)
		if err != nil {
			fmt.Println(err)
		}

		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		session.Stdin = os.Stdin

		err = session.Shell()
		if err != nil {
			fmt.Println(err)
		}

		signal_chan := make(chan os.Signal, 1)
		signal.Notify(signal_chan, syscall.SIGWINCH)
		go func() {
			for {
				s := <-signal_chan
				switch s {
				case syscall.SIGWINCH:
					fd := int(os.Stdout.Fd())
					w, h, _ = terminal.GetSize(fd)
					session.WindowChange(h, w)
				}
			}
		}()

		err = session.Wait()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
