package examples

import (
	//"bytes"
	//"errors"
	"fmt"
	"time"
	nsq "github.com/bitly/go-nsq"
	//"github.com/mozilla-services/heka/message"
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

/*
func findMessage(buf []byte, header *message.Header, msg *[]byte) (pos int, ok bool) {
	pos = bytes.IndexByte(buf, message.RECORD_SEPARATOR)
	if pos != -1 {
		if len(buf)-pos > 1 {
			headerLength := int(buf[pos+1])
			headerEnd := pos + headerLength + 3 // recsep+len+header+unitsep
			if len(buf) >= headerEnd {
				if header.MessageLength != nil || pipeline.DecodeHeader(buf[pos+2:headerEnd], header) {
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
*/

func (ni *NsqInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) error {
	// Get the InputRunner's chan to receive empty PipelinePacks
	var pack *pipeline.PipelinePack
	var err error
	var dRunner pipeline.DecoderRunner
	var decoder pipeline.Decoder
	var ok bool
	var e error

	pos := 0
	output := make([]*Message, 1000)
	packSupply := ir.InChan()

	//var decoding chan<- *pipeline.PipelinePack
	if ni.conf.Decoder != "" {
		if dRunner, ok = h.DecoderRunner(ni.conf.Decoder); !ok {
			return fmt.Errorf("Decoder not found: %s", ni.conf.Decoder)
		}
		decoder = dRunner.Decoder()
	}
	/*
			// Fetch specified decoder
			decoder, ok = h.DecoderRunner(ni.conf.Decoder)
			if !ok {
				err := fmt.Errorf("Could not find decoder", ni.conf.Decoder)
				return err
			}

			// Get the decoder's receiving chan
			//decoding = decoder.InChan()
		}
	*/
	err = ni.nsqReader.ConnectToLookupd("192.168.1.44:4161")
	if err != nil {
		fmt.Errorf("ConnectToLookupd failed")
	}

	//header := &message.Header{}

	//readLoop:
	for {
		pack = <-packSupply
		m := <-ni.handler.logChan
		ir.LogError(fmt.Errorf("message body: %s", m.msg.Body))
		pack.Message.SetType("nsq")
		pack.Message.SetPayload(string(m.msg.Body))
		pack.Message.SetTimestamp(time.Now().UnixNano())
		var packs []*pipeline.PipelinePack
		if decoder == nil {
			packs = []*pipeline.PipelinePack{pack}
		} else {
			packs, e = decoder.Decode(pack)
		}
		if packs != nil {
			for _, p := range packs {
				ir.Inject(p)
			}
		} else {
			if e != nil {
				ir.LogError(fmt.Errorf("Couldn't parse AMQP message: %s", m.msg.Body))
			}
			pack.Recycle()
		}

                output[pos] = m
		pos++
		if pos == 1000 {
			for pos > 0 {
				pos--
				m1 := output[pos]
				m1.returnChannel <- &nsq.FinishedMessage{m1.msg.Id, 0, true}
				output[pos] = nil
			}
		}
		/*
			_, msgOk := findMessage(m.msg.Body, header, &(pack.MsgBytes))
			if msgOk {
				decoding <- pack
			} else {
				pack.Recycle()
				ir.LogError(errors.New("Can't find Heka message."))
			}
			header.Reset()
		*/
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
