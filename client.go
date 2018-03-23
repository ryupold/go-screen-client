package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"

	"github.com/saljam/mjpeg"
	"github.com/vova616/screenshot"
)

func startStreaming() {
	stream := mjpeg.NewStream()
	stream.FrameInterval = 10

	go http.ListenAndServe("0.0.0.0:4545", stream)

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

			stream.UpdateJPEG(buffer.Bytes())
		}
	}()
}
