package helpers

import (
	"context"
	"fmt"
	"io"
)

type readerCtx struct {
	ctx    context.Context
	r      io.Reader
	length int64
	eofCh  chan struct{}
}

func (r *readerCtx) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		fmt.Printf("readerCtx.Ctx.Err: %v\n", err)
		return 0, err
	}
	n, errR := r.r.Read(p)
	if errR == io.EOF {
		fmt.Println("readerCtx.Read: end of file")
		r.eofCh <- struct{}{}

		remainBytes := int(r.length % 188)
		if remainBytes > 0 {
			n = 188 - remainBytes
			copy(p, make([]byte, n))
		}
	} else if errR != nil {
		fmt.Printf("readerCtx.Read: %v\n", errR)
	}

	return n, errR
}

func NewReader(ctx context.Context, r io.Reader, contentLength int64, eofCh chan struct{}) io.Reader {
	return &readerCtx{
		ctx:    ctx,
		r:      r,
		eofCh:  eofCh,
		length: contentLength,
	}
}
