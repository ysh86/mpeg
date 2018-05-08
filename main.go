package main

import (
	"flag"
	"log"
	"os"

	"github.com/32bitkid/bitreader"
	"github.com/ysh86/mpeg/ps"
)

func main() {
	// args
	var (
		srcFile = flag.String("i", "source.mpg", "source file")
	)
	flag.Parse()

	// Open the file
	fileReader, err := os.Open(*srcFile)
	if err != nil {
		log.Fatal(err)
	}

	// Decode the PS stream
	bitReader := bitreader.NewBitReader(fileReader)
	decoder := ps.NewDecoder(bitReader)
	eop := make(chan bool)
	go func() {
		for i := 0; true; i++ {
			pack := <-decoder.Packs()
			if pack == nil {
				// no more packs
				break
			}
			scr := pack.PackHeader.SystemClockReferenceBase*300 + pack.PackHeader.SystemClockReferenceExtension
			bps := pack.PackHeader.ProgramMuxRate * 50 /*[bytes/sec]*/ * 8 /*[bits/byte]*/
			log.Println("Pack[", i, "]: SCR[27MHz] =", scr, ", bitrate[bps] =", bps)
			if pack.PackHeader.SystemHeader != nil {
				log.Println(" System header: audio bound =", pack.PackHeader.SystemHeader.AudioBound)
				log.Println(" System header: video bound =", pack.PackHeader.SystemHeader.VideoBound)
				n := len(pack.PackHeader.SystemHeader.Streams)
				log.Println(" System header: streams =", n)
				for i := 0; i < n; i++ {
					id := pack.PackHeader.SystemHeader.Streams[i].StreamID
					stream := "Unknown"
					if id == 0xb8 {
						stream = "all audio streams"
					} else if id == 0xb9 {
						stream = "all video streams"
					} else if id == 0xbc {
						stream = "Program stream map"
					} else if id == 0xbd {
						stream = "Private stream 1"
					} else if id == 0xbe {
						stream = "Padding"
					} else if id == 0xbf {
						stream = "Private stream 2"
					} else if id&0xe0 == 0xc0 {
						stream = "Audio"
					} else if id&0xf0 == 0xe0 {
						stream = "Video"
					}
					log.Printf(" StreamID[ %d ] = 0x%02X (%s)\n", i, id, stream)
				}
			}
			for j := 0; true; j++ {
				packet := <-pack.Packets()
				if packet == nil {
					// no more packets
					break
				}
				stream := "Unknown"
				if packet.StreamID == 0xbc {
					stream = "Program stream map"
				} else if packet.StreamID == 0xbd {
					stream = "Private stream 1"
				} else if packet.StreamID == 0xbe {
					stream = "Padding"
				} else if packet.StreamID == 0xbf {
					stream = "Private stream 2"
				} else if packet.StreamID&0xe0 == 0xc0 {
					stream = "Audio"
				} else if packet.StreamID&0xf0 == 0xe0 {
					stream = "Video"
				}
				log.Printf(" Packet[ %d ]: StreamID = 0x%02X , length(%d) = 3 + header + payload(%d) , %s\n", j, packet.StreamID, packet.PacketLength, len(packet.Payload), stream)
				if packet.Header != nil {
					log.Printf("   Header: PtsDtsFlags = 0x%02X , ExtensionFlag = %t , length = %d\n", packet.Header.PtsDtsFlags, packet.Header.ExtensionFlag, packet.Header.HeaderDataLength)
					if packet.Header.PtsDtsFlags == 3 {
						log.Printf("   Pts = %d , Dts = %d\n", packet.Header.PresentationTimeStamp, packet.Header.DecodingTimeStamp)
					}
					if packet.Header.PtsDtsFlags == 2 {
						log.Printf("   Pts = %d\n", packet.Header.PresentationTimeStamp)
					}
					if packet.Header.ExtensionFlag {
						log.Printf("   Extension %t %t %t %t %t\n",
							packet.Header.Extension.PrivateDataFlag,
							packet.Header.Extension.PackHeaderFieldFlag,
							packet.Header.Extension.ProgramPacketSequenceCounterFlag,
							packet.Header.Extension.P_STD_BufferFlag,
							packet.Header.Extension.ExtensionFlag2)
					}
				}
			}
		}

		eop <- true
	}()

	// Wait for finish
	log.Println("Decoder: start")
	done := <-decoder.Go()
	donep := <-eop

	if done && donep {
		log.Println("Decoder: done")
	} else {
		log.Fatal("Decoder: Error = ", decoder.Err())
	}
}
