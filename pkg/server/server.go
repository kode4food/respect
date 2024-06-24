package server

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/kode4food/respect/pkg/command"
	"github.com/kode4food/respect/pkg/resp"
)

type (
	Server struct {
		*Config
	}

	Config struct {
		MakeReader ReaderMaker
		Handler    command.Handler
		Port       int
	}

	Option func(*Config)

	ReaderMaker func(*bufio.Reader, ...resp.ReaderOption) *resp.Reader

	socketContext struct {
		*Server

		conn   net.Conn
		reader *resp.Reader
		writer *bufio.Writer

		input  chan resp.Value
		output chan resp.Value
		closed chan struct{}

		close sync.Once
	}
)

const DefaultPort = 6379

var defaultOptions = []Option{
	func(c *Config) {
		c.MakeReader = resp.NewReader
		c.Handler = command.NewHandler(command.Handlers{})
		c.Port = DefaultPort
	},
}

func NewServer(opts ...Option) *Server {
	res := &Server{}
	for _, opt := range append(defaultOptions, opts...) {
		opt(res.Config)
	}
	return res
}

func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithReaderMaker(m ReaderMaker) Option {
	return func(c *Config) {
		c.MakeReader = m
	}
}

func WithHandler(h command.Handler) Option {
	return func(c *Config) {
		c.Handler = h
	}
}

func WithEnvPort() Option {
	if port := os.Getenv("PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			panic(err)
		}
		return WithPort(p)
	}
	return func(*Config) {}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}

	defer func() {
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	ctx := s.makeContext(conn)
	defer func() { _ = ctx.Close() }()
	go ctx.readLoop()
	go ctx.writeLoop()
	ctx.handleLoop()
}

func (s *Server) makeContext(conn net.Conn) *socketContext {
	return &socketContext{
		Server: s,

		conn:   conn,
		reader: s.MakeReader(bufio.NewReader(conn)),
		writer: bufio.NewWriter(conn),

		input:  make(chan resp.Value),
		output: make(chan resp.Value),
		closed: make(chan struct{}),
	}
}

func (c *socketContext) handleLoop() {
	for {
		select {
		case <-c.closed:
			return
		default:
			if err := command.HandleNext(c, c.Handler); err != nil {
				c.forwardError(err)
			}
		}
	}
}

func (c *socketContext) readLoop() {
	for {
		select {
		case <-c.closed:
			return
		default:
			value, err := c.reader.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					_ = c.Close()
					return
				}
				c.forwardError(err)
			}
			c.input <- value
		}
	}
}

func (c *socketContext) writeLoop() {
	for {
		select {
		case <-c.closed:
			return
		case value := <-c.output:
			err := value.Marshal(c.writer)
			if err == nil {
				err = c.writer.Flush()
			}
			if err != nil {
				c.forwardError(err)
			}
		}
	}
}

func (c *socketContext) forwardError(err error) {
	respErr, ok := err.(resp.Value)
	if !ok {
		respErr = resp.MakeError(err.Error())
	}
	select {
	case <-c.closed:
	case c.output <- respErr:
	}
}

func (c *socketContext) Close() error {
	c.close.Do(func() {
		_ = c.conn.Close()
		close(c.closed)
		close(c.input)
		close(c.output)
	})
	return nil
}

func (c *socketContext) Accept() <-chan resp.Value {
	return c.input
}

func (c *socketContext) Emit() chan<- resp.Value {
	return c.output
}

func (c *socketContext) Closed() <-chan struct{} {
	return c.closed
}
