package server

import (
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrServerRunning = errors.New("server already running")
	ErrServerStopped = errors.New("server already stopped")
)

type Server struct {
	config Config

	listener net.Listener
	shutdown chan bool
	players  sync.Map
}

func (server *Server) Start() error {
	if server.listener != nil {
		return ErrServerRunning
	}

	bind := net.JoinHostPort(server.config.Host, strconv.Itoa(server.config.Port))
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}
	server.listener = listener

	log.Info().Stringer("addr", listener.Addr()).Msg("server listening for new connections")

	shutdown := false
	go func() {
		shutdown = <-server.shutdown
	}()

	go func() {
		for !shutdown {
			client, err := listener.Accept()
			if err != nil {
				// See https://github.com/golang/go/issues/4373 for info.
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Warn().Err(err).Msg("error occurred while accepting s new connection")
				}
				continue
			}

			go server.handleClient(client)
		}
	}()

	return nil
}

func (server *Server) Stop() error {
	if server.listener == nil {
		return ErrServerStopped
	}

	log.Info().Msg("stopping server")
	server.ForEachPlayer(func(player *Player) {
		_ = player.Kick("server is shutting down")
	})

	server.shutdown <- true
	if err := server.listener.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close listener")
	}
	server.listener = nil

	return nil
}

func (server *Server) GetPlayer(uniqueID uuid.UUID) *Player {
	if value, ok := server.players.Load(uniqueID); ok {
		return value.(*Player)
	}
	return nil
}

func (server *Server) ForEachPlayer(fn func(player *Player)) {
	server.players.Range(func(key, value interface{}) bool {
		fn(value.(*Player))
		return true
	})
}

func (server *Server) handleClient(conn net.Conn) {
	log.Debug().Stringer("connection", conn.RemoteAddr()).Msg("client connected")

	connection := NewConnection(conn, server)
	for {
		if err := connection.ReadPacket(); err != nil {
			// See https://github.com/golang/go/issues/4373 for info.
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error().Err(err).Stringer("connection", connection.RemoteAddr()).Msg("got error during packet read")
				// todo: should we disconnect?
			}
			break
		}
	}

	if err := connection.Close(); err != nil {
		// See https://github.com/golang/go/issues/4373 for info.
		if !strings.Contains(err.Error(), "use of closed network connection") {
			log.Warn().Err(err).Stringer("connection", connection.RemoteAddr()).Msg("got error while closing connection")
			return
		}
	}

	log.Debug().Stringer("connection", conn.RemoteAddr()).Msg("client disconnected")
}

func (server *Server) createPlayer(conn *Connection) *Player {
	player := NewPlayer(conn)
	server.players.Store(player.GetUniqueID(), player)
	return player
}

func (server *Server) removePlayer(uniqueID uuid.UUID) {
	server.players.Delete(uniqueID)
}

func NewServer(config Config) *Server {
	return &Server{
		config:   config,
		shutdown: make(chan bool),
		players:  sync.Map{},
	}
}
