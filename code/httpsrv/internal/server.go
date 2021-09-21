package internal

import (
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Service http.Handler

// Options ...
type Options struct {
	addr     string
	services map[string]Service
}

// Option Options helper
type Option func(*Options)

// WithAddr ...
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.addr = addr
	}
}

// WithService ...
func WithService(name string, service Service) Option {
	return func(o *Options) {
		o.services[GenServiceKey(name)] = service
	}
}

// Server ...
type Server struct {
	opts Options
	eg   *errgroup.Group
}

func (s *Server) loadConf() error {
	for k, v := range s.opts.services {
		k, v := k, v
		s.eg.Go(func() error {
			http.Handle(k, v)
			return nil
		})

	}
	if err := s.eg.Wait(); err != nil {
		return errors.Wrapf(err, "loadConf error, err:%+v", err)
	}
	return nil
}

// Serve ...
func (s *Server) Serve() error {
	//return errors.New("test sig") // Output1 test sig
	if err := s.loadConf(); err != nil {
		return err
	}
	return http.ListenAndServe(s.opts.addr, nil)
}

// NewServer ...
func NewServer(opt ...Option) *Server {
	opts := &Options{services: make(map[string]Service)}
	for i := range opt {
		opt[i](opts)
	}

	return &Server{opts: *opts, eg: &errgroup.Group{}}
}
