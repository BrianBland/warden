package warden

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"sync"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

type Warden struct {
	addr        string
	privateKeys []ssh.Signer
	jailImage   string
}

func New(config Config) (*Warden, error) {
	if len(config.PrivateKeys) == 0 {
		return nil, errors.New("No private keys provided")
	}
	privateKeys := make([]ssh.Signer, len(config.PrivateKeys))
	for i, pkFile := range config.PrivateKeys {
		privateBytes, err := ioutil.ReadFile(expand(pkFile))
		if err != nil {
			return nil, err
		}
		pk, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			return nil, err
		}
		privateKeys[i] = pk
	}
	addr := config.Addr
	if addr == "" {
		addr = ":22"
	}
	jailImage := config.JailImage
	if jailImage == "" {
		jailImage = "ubuntu"
	}
	return &Warden{
		addr:        addr,
		privateKeys: privateKeys,
		jailImage:   jailImage,
	}, nil
}

func (w *Warden) Run() error {
	config := &ssh.ServerConfig{PublicKeyCallback: checkAuth}
	for _, pk := range w.privateKeys {
		config.AddHostKey(pk)
	}
	listener, err := net.Listen("tcp", w.addr)
	if err != nil {
		log.Fatalln("Failed to listen for connections:", err)
	}
	for {
		fmt.Printf("Listening on %s...\n", w.addr)
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept incoming connection:", err)
			continue
		}
		go w.handleConn(conn, config)
	}
}

func checkAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Println("No auth yet! Allowing user:", conn.User())
	return nil, nil
}

func (w *Warden) handleConn(conn net.Conn, conf *ssh.ServerConfig) {
	defer conn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, conf)
	if err != nil {
		log.Println("Failed to handshake:", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	for ch := range chans {
		if ch.ChannelType() != "session" {
			ch.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		go w.handleChannel(sshConn, ch)
	}
}

func (w *Warden) handleChannel(conn *ssh.ServerConn, newChan ssh.NewChannel) {
	ch, reqs, err := newChan.Accept()
	if err != nil {
		log.Println("newChan.Accept failed:", err)
		return
	}

	bash := exec.Command("docker", "run", "-it", "--rm", w.jailImage, "bash", "-c", jailScript(conn.User()))

	close := func() {
		ch.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Println("Failed to exit bash:", err)
		}
		log.Println("Session closed")
	}

	log.Println("Creating pty...")
	bashf, err := pty.Start(bash)
	if err != nil {
		log.Println("Failed to start pty:", err)
		close()
		return
	}

	var once sync.Once
	go func() {
		io.Copy(ch, bashf)
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, ch)
		once.Do(close)
	}()

	go func() {
		for req := range reqs {
			switch req.Type {
			case "shell":
				ok := len(req.Payload) == 0
				if req.WantReply {
					req.Reply(ok, nil)
				}
			case "pty-req":
				termLen := req.Payload[3]
				w, h := parseDimensions(req.Payload[termLen+4:])
				setWindowSize(bashf.Fd(), w, h)
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "window-change":
				w, h := parseDimensions(req.Payload)
				setWindowSize(bashf.Fd(), w, h)
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "env":
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()
}

const jailScriptFmt = `
user=%s
if [ "$user" == root ]; then
  user=r00t
fi
exists=false
(getent passwd $user && exists=true
if ! $exists; then
  adduser --disabled-password --gecos '' $user
fi) > /dev/null 2>&1
su $user
`

func jailScript(username string) string {
	return fmt.Sprintf(jailScriptFmt, username)
}
