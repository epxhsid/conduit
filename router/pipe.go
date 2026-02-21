package router

import (
	"fmt"
	"io"
)

type Pipe struct{}

func NewPipe() *Pipe {
	return &Pipe{}
}

func (p *Pipe) Pipe(src, dst io.ReadWriteCloser) {
	done := make(chan struct{}, 2)

	go func() {
		_, err := io.Copy(dst, src)
		if err != nil {
			fmt.Println("pipe error:", err)
		}
		done <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(src, dst)
		if err != nil {
			fmt.Println("pipe error:", err)
		}
		done <- struct{}{}
	}()
	<-done

	src.Close()
	dst.Close()
}
