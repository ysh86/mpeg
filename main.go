package main

import (
	"flag"
	"log"
	"os"

	"github.com/ysh86/mpeg/ts"
)

func main() {
	// args
	var (
		srcFile = flag.String("i", "source.ts", "source file")
	)
	flag.Parse()

	// Open the file
	fileReader, err := os.Open(*srcFile)
	if err != nil {
		log.Fatal(err)
	}

	// Demux the TS stream
	demuxer := ts.NewDemuxer(fileReader)
	packets := demuxer.Where(func(p *ts.Packet) bool { return true })
	eop := make(chan bool)
	go func() {
		for i := 0; true; i++ {
			packet := <-packets
			if packet == nil {
				// no more packets
				break
			}
			var adaptationLen = 0
			if packet.AdaptationFieldControl == ts.FieldOnly || packet.AdaptationFieldControl == ts.FieldThenPayload {
				if packet.AdaptationField != nil {
					adaptationLen = 1 + int(packet.AdaptationField.Length)
				} else {
					// Error!
				}
			}
			var pusi = 0
			if packet.PayloadUnitStartIndicator {
				pusi = 1
			}
			log.Printf("Packet[ %5d ]: PID = 0x%04X, PayloadUnitStart = %d, ContinuityCounter = %2d, length(%3d) = 4+AdaptationField(%3d)+payload(%3d)\n",
				i,
				packet.PID,
				pusi,
				packet.ContinuityCounter,
				4+adaptationLen+len(packet.Payload),
				adaptationLen,
				len(packet.Payload))
			if packet.PID == 0 {
				log.Printf(" Program Association Table(PAT):")
				log.Printf("  pointer_field:            %d", packet.Payload[0])
				log.Printf("  table_id:                 0x%02X", packet.Payload[1])
				log.Printf("  section_syntax_indicator: %d", (packet.Payload[2]>>7)&1)
				log.Printf("  section_length:           %d", (int(packet.Payload[2]&0xf)<<8)+int(packet.Payload[3]))
				log.Printf("  transport_stream_id:      0x%04X", (int(packet.Payload[4])<<8)+int(packet.Payload[5]))
				log.Printf("  version_number:           %d", (packet.Payload[6]>>1)&0x1f)
				log.Printf("  current_next_indicator:   %d", packet.Payload[6]&1)
				log.Printf("  section_number:           0x%02X", packet.Payload[7])
				log.Printf("  last_section_number:      0x%02X", packet.Payload[8])
				log.Printf("   program_number:  0x%04X", (int(packet.Payload[9])<<8)+int(packet.Payload[10]))
				log.Printf("   program_map_PID: 0x%04X", (int(packet.Payload[11]&0x1f)<<8)+int(packet.Payload[12]))
			}
			if packet.AdaptationFieldControl == ts.FieldOnly || packet.AdaptationFieldControl == ts.FieldThenPayload {
				if packet.AdaptationField != nil {
					var flags byte
					if packet.AdaptationField.Length > 0 {
						flags = packet.AdaptationField.Junk[0]
					}
					var pcr uint64
					if packet.AdaptationField.PCRFlag {
						pcr = packet.AdaptationField.PCR
					}
					log.Printf(" AdaptationField: Control = %d, length = %d, flags = 0x%02X, PCR = %d",
						packet.AdaptationFieldControl,
						packet.AdaptationField.Length,
						flags,
						pcr)
				} else {
					log.Fatal(" AdaptationField is nil!")
				}
			}
		}

		eop <- true
	}()

	// Wait for finish
	log.Println("Demuxer: start")
	done := <-demuxer.Go()
	donep := <-eop

	if done && donep {
		log.Println("Demuxer: done")
	} else {
		log.Fatal("Demuxer: Error = ", demuxer.Err())
	}
}
