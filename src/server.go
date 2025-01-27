package main

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
	"github.com/moby/term"
	"github.com/spf13/pflag"
)

type Type uint8

type Channel struct {
	io   net.Conn
	argv []string
}

func (s *Channel) sendArgvEnd() error {
	err := binary.Write(s.io, binary.LittleEndian, uint32(0))
	if err != nil {
		return err
	}
	return nil
}
func (s *Channel) sendArgv(argv string) error {
	data := []byte(argv)
	err := binary.Write(s.io, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return err
	}
	err = binary.Write(s.io, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Channel) run() {
	defer s.io.Close()
	if term.IsTerminal(os.Stdin.Fd()) {
		state, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			log.Printf("Failed to set raw terminal: %v", err)
		} else {
			defer term.RestoreTerminal(os.Stdin.Fd(), state)
		}
	}

	go func() {
		_, _ = io.Copy(s.io, os.Stdin)
	}()
	_, _ = io.Copy(os.Stdout, s.io)
}
func (s *Channel) ReadArg() ([]byte, error) {
	var len uint32
	if err := binary.Read(s.io, binary.LittleEndian, &len); err != nil {
		return nil, err
	}
	if len == 0 {
		return nil, nil
	}
	var err error
	buffer := make([]byte, len)
	_, err = io.ReadFull(s.io, buffer)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
func (s *Channel) ParseArgs() ([]string, error) {
	var args []string

	for {
		item, err := s.ReadArg()
		if err != nil {
			return nil, err
		}
		if item == nil {
			break
		}
		args = append(args, string(item))
	}
	if len(args) == 0 {
		args = append(args, DefaultShell())
	}
	return args, nil
}

func (s *Channel) Serve() {
	defer s.io.Close()
	args, err := s.ParseArgs()
	if err != nil {
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("Error starting PTY: %v\n", err)
		return
	}
	defer ptmx.Close()
	go func() {
		_, _ = ptmx.ReadFrom(s.io) // 将客户端输入写入伪终端
	}()
	go func() {
		_, _ = ptmx.WriteTo(s.io) // 将伪终端输出发送到客户端
	}()

	_ = cmd.Wait()
}
func CreateChannel(conn net.Conn) *Channel {
	channel := &Channel{
		io: conn,
	}
	return channel
}

type ChannelFlags struct {
	Unix            string
	Server          bool
	AutoExit        bool
	Timeout         time.Duration
	connectionCount int32
	connectionTimes int32
	signal          chan int
}

func (flags *ChannelFlags) HandleConnection(conn net.Conn) {
	atomic.AddInt32(&flags.connectionCount, 1)
	channel := CreateChannel(conn)
	channel.Serve()
	atomic.AddInt32(&flags.connectionCount, -1)
	flags.signal <- 1
}
func (flags *ChannelFlags) CreateServer() {
	socketPath := flags.Unix
	if _, err := os.Stat(socketPath); err == nil {
		_ = os.Remove(socketPath)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to create socket:%v", err)
	}
	defer os.Remove(socketPath)
	defer listener.Close()
	log.Println("Server is listening on", socketPath)

	if flags.AutoExit {
		go func() {
			for {
				<-flags.signal
				if atomic.LoadInt32(&flags.connectionTimes) > 0 && atomic.LoadInt32(&flags.connectionCount) == 0 {
					// log.Println("No active sessions. Exiting...")
					os.Exit(0)
				}
			}
		}()
	}
	if flags.Timeout != 0 {
		go func() {
			time.Sleep(flags.Timeout)
			if atomic.LoadInt32(&flags.connectionTimes) == 0 {
				log.Println("No client connected within the specified timeout. Exiting...")
				os.Exit(0)
			}
		}()
	}
	for {
		conn, err := listener.Accept()
		atomic.AddInt32(&flags.connectionTimes, 1)
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go flags.HandleConnection(conn)
	}
}
func (flags *ChannelFlags) WaitForConnection() net.Conn {
	startTime := time.Now()
	printed := false
	for {
		conn, err := net.Dial("unix", flags.Unix)
		if err != nil {
			if flags.Timeout != 0 && time.Since(startTime) < flags.Timeout {
				if !printed {
					log.Println("Waiting for connection:", flags.Unix)
					printed = true
				}
				time.Sleep(1 * time.Second)
			} else {
				return nil
			}
		} else {
			return conn
		}
	}
}
func (flags *ChannelFlags) CreateClient(args []string) {
	conn := flags.WaitForConnection()
	if conn == nil {
		log.Fatalln("Failed to connect server:", flags.Unix)
		return
	}

	defer conn.Close()

	channel := CreateChannel(conn)
	for _, item := range args {
		channel.sendArgv(item)
	}
	channel.sendArgvEnd()
	channel.run()
}

func NewChannelFlags() *ChannelFlags {
	return &ChannelFlags{
		Unix:     "",
		Server:   false,
		AutoExit: true,
		Timeout:  time.Second * 30,
		signal:   make(chan int),
	}
}
func (flags *ChannelFlags) TestConnectivity() bool {
	conn, err := net.Dial("unix", flags.Unix)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
func (opts *ChannelFlags) Parse(argv []string) ([]string, error) {
	fs := pflag.NewFlagSet("channel", pflag.ContinueOnError)
	fs.StringVar(&opts.Unix, "unix", opts.Unix, "Singleton mode. --unix unix://path")
	fs.BoolVar(&opts.Server, "server", opts.Server, "Server mode.")
	fs.DurationVar(&opts.Timeout, "timeout", opts.Timeout, "Connect timeout.")
	fs.BoolVar(&opts.AutoExit, "exit", opts.AutoExit, "Auto exit when no active sessions.")
	if err := fs.Parse(argv); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}

// func DefaultShell() string {
// 	shell := os.Getenv("SHELL")
// 	if shell == "" {
// 		return "/bin/sh"
// 	}
// 	return shell
// }
// func main() {
// 	flags := NewChannelFlags()
// 	args, err := flags.Parse(os.Args)
// 	if err != nil {
// 		log.Fatalln("Failed to Parse:", err)
// 	}
// 	if flags.Server {
// 		flags.CreateServer()
// 	} else {
// 		flags.CreateClient(args[1:])
// 	}
// }
