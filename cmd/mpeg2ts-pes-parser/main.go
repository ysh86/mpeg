package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/32bitkid/bitreader"
	"github.com/ysh86/mpeg/pes"
	"github.com/ysh86/mpeg/ts"
)

func main() {
	// args
	var (
		srcFile = flag.String("i", "source.ts", "source file")
		pid     = flag.Uint("p", 0x0100, "PID")
	)
	flag.Parse()

	// Open the file
	fileReader, err := os.Open(*srcFile)
	if err != nil {
		log.Fatal(err)
	}

	// Demux to PES
	pesReader := ts.NewPayloadUnitReader(fileReader, ts.IsPID(uint32(*pid)))
	bits := bitreader.NewReader(pesReader)
	eop := make(chan error)
	go func() {
		var err error
		var packet *pes.Packet
		for i := 0; true; i++ {
			packet, err = pes.NewPacket(bits)
			if err != nil {
				break
			}
			if packet.Header != nil {
				var pts uint32
				var dts uint32
				if packet.Header.PtsDtsFlags == 2 {
					pts = packet.Header.PresentationTimeStamp
				}
				if packet.Header.PtsDtsFlags == 3 {
					pts = packet.Header.PresentationTimeStamp
					dts = packet.Header.DecodingTimeStamp
				}
				log.Printf("PES[ %5d ]: PID = 0x%04X, StreamID = 0x%02X, PacketLength = %6d, PtsDtsFlags = 0x%01X, PTS = %d, DTS = %d\n",
					i,
					*pid,
					packet.StreamID,
					packet.PacketLength,
					packet.Header.PtsDtsFlags,
					pts,
					dts)
			} else {
				log.Printf("PES[ %5d ]: PID = 0x%04X, StreamID = 0x%02X, PacketLength = %6d\n",
					i,
					*pid,
					packet.StreamID,
					packet.PacketLength)
			}
		}

		eop <- err
	}()

	// Wait for finish
	log.Println("Demuxer: start")
	donep := <-eop

	if donep == nil || donep == io.EOF {
		log.Println("Demuxer: done")
	} else {
		log.Fatal("Demuxer: Error = ", donep)
	}
}
