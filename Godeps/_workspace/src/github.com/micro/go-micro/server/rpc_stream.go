package server

import (
	"log"
	"sync"

	"github.com/micro/go-plugins/Godeps/_workspace/src/golang.org/x/net/context"
)

// Implements the Streamer interface
type rpcStream struct {
	sync.RWMutex
	seq     uint64
	closed  bool
	err     error
	request Request
	codec   serverCodec
	context context.Context
}

func (r *rpcStream) Context() context.Context {
	return r.context
}

func (r *rpcStream) Request() Request {
	return r.request
}

func (r *rpcStream) Send(msg interface{}) error {
	r.Lock()
	defer r.Unlock()

	seq := r.seq
	r.seq++

	resp := response{
		ServiceMethod: r.request.Method(),
		Seq:           seq,
	}

	err := r.codec.WriteResponse(&resp, msg, false)
	if err != nil {
		log.Println("rpc: writing response:", err)
	}
	return err
}

func (r *rpcStream) Recv(msg interface{}) error {
	r.Lock()
	defer r.Unlock()

	req := request{}

	if err := r.codec.ReadRequestHeader(&req, false); err != nil {
		// discard body
		r.codec.ReadRequestBody(nil)
		return err
	}

	if err := r.codec.ReadRequestBody(msg); err != nil {
		return err
	}

	return nil
}

func (r *rpcStream) Error() error {
	r.RLock()
	defer r.RUnlock()
	return r.err
}

func (r *rpcStream) Close() error {
	r.Lock()
	defer r.Unlock()
	r.closed = true
	return r.codec.Close()
}
