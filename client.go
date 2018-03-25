package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/jpeg"
	"net"

	"github.com/vova616/screenshot"
)

func startStreaming(ip string, port uint16) error {

	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}

	go func() {
		for {
			img, err := screenshot.CaptureScreen()
			if err != nil || img == nil {
				panic(err)
			}
			buffer := &bytes.Buffer{}
			if err := jpeg.Encode(buffer, image.Image(img), &jpeg.Options{Quality: 100 /*default*/}); err != nil {
				fmt.Printf("ERROR: %+v", err)
			}

			sizeBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(sizeBytes, uint32(buffer.Len()))

			n, err := con.Write(append(sizeBytes, buffer.Bytes()...))
			if err != nil {
				panic(err)
			}
			if n != buffer.Len() {
				panic(fmt.Errorf("%d != %d written bytes differ", n, buffer.Len()))
			}
		}
	}()

	return nil
}
