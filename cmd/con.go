package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// conCmd represents the con command
var conCmd = &cobra.Command{
	Use:   "con",
	Short: "A Connet to SSH Server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		// host := "localhost"
		// port := "22"
		// user := "test-user"
		cuser, err := user.Current()
		if err != nil {
			fmt.Println(err)
		}
		// pass := "Zaq12wsx"

		sshHost, _ := cmd.Flags().GetString("host")
		sshPort, _ := cmd.Flags().GetString("port")

		sshUser, _ := cmd.Flags().GetString("login")
		if sshUser == "" {
			sshUser = cuser.Username
		}
		sshPass, _ := cmd.Flags().GetString("password")
		if sshPass == "" {
			sshPass = ""
		}

		// Create sshClientConfig
		sshConfig := &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				ssh.Password(sshPass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		// SSH connect.
		client, err := ssh.Dial("tcp", sshHost+":"+sshPort, sshConfig)

		// Create Session
		session, err := client.NewSession()
		defer session.Close()

		// キー入力を接続先が認識できる形式に変換する(ここがキモ)
		fd := int(os.Stdin.Fd())
		state, err := terminal.MakeRaw(fd)
		if err != nil {
			fmt.Println(err)
		}
		defer terminal.Restore(fd, state)

		// ターミナルサイズの取得
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

		// ターミナルサイズの変更検知・処理
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
	rootCmd.AddCommand(conCmd)

	conCmd.Flags().StringP("host", "H", "localhost", "SSH Server Host")
	// conCmd.Flags().IntP("port", "p", 22, "SSH Server Port")
	conCmd.Flags().StringP("port", "p", "22", "SSH Server Port")
	conCmd.Flags().StringP("login", "L", "", "SSH Server User")
	conCmd.Flags().StringP("password", "P", "", "SSH Server Password")
}
