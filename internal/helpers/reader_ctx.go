package helpers

import (
	"context"
	"io"
)

type readerCtx struct {
	ctx   context.Context
	r     io.Reader
	eofCh chan struct{}
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, errR := r.r.Read(p)
	if errR == io.EOF {
		r.eofCh <- struct{}{}
	}
	return n, errR
}

func NewReader(ctx context.Context, r io.Reader, eofCh chan struct{}) io.Reader {
	return &readerCtx{ctx: ctx, r: r, eofCh: eofCh}
}
