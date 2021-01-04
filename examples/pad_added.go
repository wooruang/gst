package main

import (
	"log"

	".."
	"github.com/ziutek/glib"
)
func cbNewpad(decodebin *gst.Element, decoderSrcPad *gst.Pad, data glib.Pointer) {
	log.Printf("In cb_newpad %p", data)
}

func main() {
	loop := glib.NewMainLoop(nil)

	pipeline := gst.NewPipeline("pad-added")
	if pipeline == nil {
		log.Printf("pipeline could not be created. Exiting.\n")
		log.Fatal(pipeline)
	}

	bin := gst.NewBin("testBin")
	if bin == nil {
		log.Fatal(bin)
	}

	uriDecodeBin := gst.ElementFactoryMake("uridecodebin", "uri-decode-bin")
	if uriDecodeBin == nil {
		log.Fatal(uriDecodeBin)
	}

	uriDecodeBin.SetProperty("uri", "rtsp://192.168.0.6:554/media/1/1/Profile1")

	log.Printf("bin %p", bin.GetPtr())
	uriDecodeBin.ConnectPadAdded(cbNewpad, bin.GetPtr())

	bin.Add(uriDecodeBin)
	//bin.AsElement().AddPad(gst.NewGhostPadNoTarget("src", gst.PAD_SRC).AsPad())
	pipeline.AsBin().Add(bin.AsElement())

	pipeline.SetState(gst.STATE_PLAYING)

	loop.Run()
}
