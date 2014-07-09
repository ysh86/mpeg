package ts

import "io"

func Demux(source io.Reader) TransportStreamDemuxer {
	reader := NewReader(source)
	return &tsDemuxer{
		reader: reader,
		reg:    make(map[uint32]chan<- *TsPacket),
	}
}

type TransportStreamDemuxer interface {
	PID(uint32) <-chan *TsPacket
	Begin() <-chan bool
	Err() error
}

type tsDemuxer struct {
	reader  TransportStreamReader
	reg     map[uint32]chan<- *TsPacket
	lastErr error
}

func (tsd *tsDemuxer) PID(PID uint32) <-chan *TsPacket {
	channel := make(chan *TsPacket)
	tsd.reg[PID] = channel
	return channel
}

func (tsd *tsDemuxer) Begin() <-chan bool {
	done := make(chan bool)

	go func() {
		for true {
			p, err := tsd.reader.Next()

			if err != nil {
				tsd.lastErr = err
				done <- true
				break
			}

			if targetChannel, ok := tsd.reg[p.PID]; ok == true {
				targetChannel <- p
			}
		}
	}()

	return done
}

func (tsd *tsDemuxer) Err() error {
	return tsd.lastErr
}
