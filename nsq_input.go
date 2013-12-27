package examples

import (
	"fmt"
	"bytes"
	nsq "github.com/bitly/go-nsq"
	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
)

type Message struct {
	msg           *nsq.Message
	returnChannel chan *nsq.FinishedMessage
}

type NsqInputConfig struct {
	Address string `toml:"address"`
	Decoder string `toml:"decoder"`
}

type NsqInput struct {
	conf      *NsqInputConfig
	nsqReader *nsq.Reader
	stopChan  chan bool
	handler   *MyTestHandler
}

type MyTestHandler struct {
	logChan chan *Message
}

//func (h *MyTestHandler) HandleMessage(message *nsq.Message) error {

func (h *MyTestHandler) HandleMessage(m *nsq.Message, responseChannel chan *nsq.FinishedMessage) {
	h.logChan <- &Message{m, responseChannel}
}

func (ni *NsqInput) ConfigStruct() interface{} {
	return &NsqInputConfig{"192.168.1.44:4161", ""}
}

func (ni *NsqInput) Init(config interface{}) error {
	ni.conf = config.(*NsqInputConfig)
	ni.stopChan = make(chan bool)
	var err error
	ni.nsqReader, err = nsq.NewReader("test123", "test")
	if err != nil {
		//log.Fatalf(err.Error())
		panic(err)
	}
	ni.nsqReader.SetMaxInFlight(1000)
	ni.handler = &MyTestHandler{logChan: make(chan *Message)}
	ni.nsqReader.AddAsyncHandler(ni.handler)
	return nil
}

func findMessage(buf []byte, header *message.Header, msg *[]byte) (pos int, ok bool) {
	pos = bytes.IndexByte(buf, message.RECORD_SEPARATOR)
	if pos != -1 {
		if len(buf)-pos > 1 {
			headerLength := int(buf[pos+1])
			headerEnd := pos + headerLength + 3 // recsep+len+header+unitsep
			if len(buf) >= headerEnd {
				if header.MessageLength != nil || DecodeHeader(buf[pos+2:headerEnd], header) {
					messageEnd := headerEnd + int(header.GetMessageLength())
					if len(buf) >= messageEnd {
						*msg = (*msg)[:messageEnd-headerEnd]
						copy(*msg, buf[headerEnd:messageEnd])
						pos = messageEnd
						ok = true
					} else {
						*msg = (*msg)[:0]
					}
				} else {
					pos, ok = findMessage(buf[pos+1:], header, msg)
				}
			}
		}
	} else {
		pos = len(buf)
	}
	return
}

func (ni *NsqInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) error {
	// Get the InputRunner's chan to receive empty PipelinePacks
	var pack *pipeline.PipelinePack
	var err error

	packs := ir.InChan()

	var decoding chan<- *pipeline.PipelinePack
	if ni.conf.Decoder != "" {
		// Fetch specified decoder
		decoder, ok := h.DecoderRunner(ni.conf.Decoder)
		if !ok {
			err := fmt.Errorf("Could not find decoder", ni.conf.Decoder)
			return err
		}

		// Get the decoder's receiving chan
		decoding = decoder.InChan()
	}
	err = ni.nsqReader.ConnectToLookupd("192.168.1.44:4161")
	if err != nil {
		fmt.Errorf("ConnectToLookupd failed")
	}

	header := &message.Header{}

	//readLoop:
	for {
		pack = <-packs
		if decoder == nil {
			pack.Recycle()
			ir.LogError(errors.New("require a decoder."))
		}
		m := <-ni.handler.logChan
		_, msgOk := findMessage(m.msg.Body, header, &(pack.MsgBytes))
		if msgOk {
			decoding <- pack
		} else {
			pack.Recycle()
			ir.LogError(errors.New("Can't find Heka message."))
		}
		header.Reset()
	}
	return nil
}

func (ni *NsqInput) Stop() {
	close(ni.stopChan)
}

func init() {
	pipeline.RegisterPlugin("NsqInput", func() interface{} {
		return new(NsqInput)
	})
}

func (ni *NsqInput) Stop() {
	close(ni.stopChan)
}

func init() {
	pipeline.RegisterPlugin("NsqInput", func() interface{} {
		return new(NsqInput)
	})
}
