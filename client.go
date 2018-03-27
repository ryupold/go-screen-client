package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"net"

	"github.com/vova616/screenshot"
)

func startStreaming(ip string, port uint16) error {
	ctx, cancel := context.WithCancel(context.Background())
	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		cancel()
		return err
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		defer cancel()
		defer func() {
			pancake := recover()
			if pancake == nil {
				return
			}
			paniC, ok := pancake.(error)
			if ok {
				errChan <- paniC
			}
		}()

		for {
			img, err := screenshot.CaptureScreen()
			if err != nil || img == nil {
				errChan <- err
				return
			}
			log("screen captured")

			buffer := &bytes.Buffer{}
			if err := jpeg.Encode(buffer, image.Image(img), &jpeg.Options{Quality: 100 /*default: 75*/}); err != nil {
				errChan <- err
				return
			}
			log("jpeg encoded")

			sizeBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(sizeBytes, uint32(buffer.Len()))
			log("size calculated")

			n, err := con.Write(append(sizeBytes, buffer.Bytes()...))
			if err != nil {
				errChan <- err
				return
			}

			log("bytes sent")
			if n != buffer.Len()+4 /*4 bytes -> img size*/ {
				errChan <- fmt.Errorf("%d != %d written bytes differ", n, buffer.Len())
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func log(s string) {
	fmt.Println(s)
}
