package base

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/cors"
	"go.uber.org/fx"
	"golang.org/x/net/http2"
	"golang.org/x/net/webdav"

	"git.pnhub.ru/core/libs/log"
)

const DefaultShutdownTimeout = time.Second * 15

var DefaultCORS = CorsConfig{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "PROPFIND", "MKCOL"},
	AllowedHeaders:   []string{"*"},
	AllowCredentials: false,
}

type HTTPServerConfig struct {
	Host           string       `json:"host" yaml:"host"`
	Port           int          `json:"port" yaml:"port"`
	HTTP2          *HTTP2Config `json:"http2" yaml:"http2"`
	MaxHeaderBytes int          `json:"max_header_bytes" yaml:"max_header_bytes"`

	ReadTimeout       time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout       time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" yaml:"read_header_timeout"`

	TLS    *TLSConfig    `json:"tls" yaml:"tls"`
	WebDav *WebDavConfig `json:"web_dav" yaml:"web_dav"`
	CORS   *CorsConfig   `json:"cors" yaml:"cors"`
}

type HTTP2Config struct {
	MaxHandlers                  int           `json:"max_handlers" yaml:"max_handlers"`
	MaxConcurrentStreams         uint32        `json:"max_concurrent_streams" yaml:"max_concurrent_streams"`
	MaxReadFrameSize             uint32        `json:"max_read_frame_size" yaml:"max_read_frame_size"`
	PermitProhibitedCipherSuites bool          `json:"permit_prohibited_cipher_suites" yaml:"permit_prohibited_cipher_suites"`
	IdleTimeout                  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
	MaxUploadBufferPerConnection int32         `json:"max_upload_buffer_per_connection" yaml:"max_upload_buffer_per_connection"`
	MaxUploadBufferPerStream     int32         `json:"max_upload_buffer_per_stream" yaml:"max_upload_buffer_per_stream"`
}

type TLSConfig struct {
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}

type WebDavConfig struct {
	Prefix string `json:"prefix" yaml:"prefix"`
	Dir    string `json:"dir" yaml:"dir"`
}

type CorsConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
}

type HTTPServer struct {
	*http.Server

	Ctx    context.Context
	Logger log.Logger
	Cfg    *HTTPServerConfig
}

func NewHTTPServer(ctx context.Context, logger log.Logger, cfg *HTTPServerConfig, lc fx.Lifecycle) (*HTTPServer, error) {
	if cfg == nil {
		return nil, ErrNilDependency(cfg)
	}
	var httpServer = &http.Server{
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.ReadHeaderTimeout,
		Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
	}
	if cfg.HTTP2 != nil {
		err := http2.ConfigureServer(httpServer, &http2.Server{
			MaxHandlers:                  cfg.HTTP2.MaxHandlers,
			MaxConcurrentStreams:         cfg.HTTP2.MaxConcurrentStreams,
			MaxReadFrameSize:             cfg.HTTP2.MaxReadFrameSize,
			PermitProhibitedCipherSuites: cfg.HTTP2.PermitProhibitedCipherSuites,
			IdleTimeout:                  cfg.HTTP2.IdleTimeout,
			MaxUploadBufferPerConnection: cfg.HTTP2.MaxUploadBufferPerConnection,
			MaxUploadBufferPerStream:     cfg.HTTP2.MaxUploadBufferPerStream,
		})
		if err != nil {
			return nil, err
		}
	}

	var hs = &HTTPServer{
		Ctx:    ctx,
		Logger: log.ForkLogger(logger),
		Cfg:    cfg,
		Server: httpServer,
	}

	if lc != nil {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					err := hs.Start()
					if err != nil && err != http.ErrServerClosed {
						logger.Fatal(err)
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				hs.Logger.Info("stopping HTTP server.")
				err := hs.Stop()
				if err != nil {
					return err
				}
				hs.Logger.Info("stopped HTTP server.")
				return nil
			},
		})
	}
	return hs, nil
}

// CreateWebDavHandler create webdav handler(prefix must end on '/', example /fs/)
func (h *HTTPServer) CreateWebDavHandler() (string, *webdav.Handler) {
	wdHandler := &webdav.Handler{
		Prefix:     h.Cfg.WebDav.Prefix,
		LockSystem: webdav.NewMemLS(),
		FileSystem: webdav.Dir(h.Cfg.WebDav.Dir),
	}
	return h.Cfg.WebDav.Prefix, wdHandler
}

func (h *HTTPServer) EnableCORS() {
	if h.Cfg.CORS == nil {
		h.Cfg.CORS = &DefaultCORS
	}
	c := h.Cfg.CORS
	if h.Handler != nil {
		h.Handler = cors.New(cors.Options{
			AllowedOrigins:   c.AllowedOrigins,
			AllowedMethods:   c.AllowedMethods,
			AllowedHeaders:   c.AllowedHeaders,
			AllowCredentials: c.AllowCredentials,
		}).Handler(h.Handler)
	}
}

func (h *HTTPServer) SetHandler(handler http.Handler) {
	if h.Cfg.CORS != nil {
		c := h.Cfg.CORS
		h.Handler = cors.New(cors.Options{
			AllowedOrigins:   c.AllowedOrigins,
			AllowedMethods:   c.AllowedMethods,
			AllowedHeaders:   c.AllowedHeaders,
			AllowCredentials: c.AllowCredentials,
		}).Handler(handler)
	} else {
		h.Handler = handler
	}
}

func (h *HTTPServer) Start() error {
	var err error
	if h.Cfg != nil && h.Cfg.TLS != nil {
		err = h.Server.ListenAndServeTLS(h.Cfg.TLS.CertFile, h.Cfg.TLS.KeyFile)
	} else {
		err = h.Server.ListenAndServe()
	}
	return err
}

func (h *HTTPServer) Stop() error {
	stopCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()
	err := h.Server.Shutdown(stopCtx)
	if err != nil {
		return err
	}
	return nil
}
