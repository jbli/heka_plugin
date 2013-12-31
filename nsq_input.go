package examples

import (
	"bytes"
	"errors"
	"fmt"
	nsq "github.com/bitly/go-nsq"
	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
	"time"
)

type Message struct {
	msg           *nsq.Message
	returnChannel chan *nsq.FinishedMessage
}

type NsqInputConfig struct {
	Address   string `toml:"address"`
	Topic     string `toml:"topic"`
	Channel   string `toml:"channel"`
	Serialize bool   `toml:"serialize"`
	Decoder   string `toml:"decoder"`
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

func (h *MyTestHandler) HandleMessage(m *nsq.Message, responseChannel chan *nsq.FinishedMessage) {
        fmt.Printf("nsq.Message.body %+v", m.Body)
	h.logChan <- &Message{m, responseChannel}
}

func (ni *NsqInput) ConfigStruct() interface{} {
	return &NsqInputConfig{
		Address:   "192.168.1.44:4161",
		Topic:     "test",
		Channel:   "test",
		Serialize: true,
		Decoder:   "ProtobufDecoder",
	}
}

func (ni *NsqInput) Init(config interface{}) error {
	ni.conf = config.(*NsqInputConfig)
	ni.stopChan = make(chan bool)
	var err error
	ni.nsqReader, err = nsq.NewReader(ni.conf.Topic, ni.conf.Channel)
	if err != nil {
		//log.Fatalf(err.Error())
		panic(err)
	}
	ni.nsqReader.SetMaxInFlight(1000)
	ni.handler = &MyTestHandler{logChan: make(chan *Message)}
	ni.nsqReader.AddAsyncHandler(ni.handler)
	return nil
}

//func findMessage(buf []byte, header *message.Header, msg *[]byte) (pos int, ok bool) {
func findMessage(buf []byte, msg *[]byte) (pos int, ok bool) {
        header := message.Header{}
	pos = bytes.IndexByte(buf, message.RECORD_SEPARATOR)
	fmt.Println("pos:", pos)
	if pos != -1 {
		if len(buf)-pos > 1 {
			headerLength := int(buf[pos+1])
			headerEnd := pos + headerLength + 3 // recsep+len+header+unitsep
			fmt.Println("headerEnd:", headerEnd, "len(buf):", len(buf))
			if len(buf) >= headerEnd {
				//if header.MessageLength != nil || pipeline.DecodeHeader(buf[pos+2:headerEnd], &header) {
				if  pipeline.DecodeHeader(buf[pos+2:headerEnd], header) {	
					messageEnd := headerEnd + int(header.GetMessageLength())
					//messageEnd := len(buf)
					fmt.Println("messageEnd:", messageEnd)
					if len(buf) >= messageEnd {
					        fmt.Println("find msg success")
						*msg = (*msg)[:messageEnd-headerEnd]
						copy(*msg, buf[headerEnd:messageEnd])
						pos = messageEnd
						ok = true
					} else {
					        fmt.Println("find msg failed 1")
						*msg = (*msg)[:0]
						//*msg = (*msg)[:len(buf)-headerEnd]
						//copy(*msg, buf[headerEnd:len(buf)])
						//pos = messageEnd
						//ok = true
					}
				} else {
				        fmt.Println("find msg failed 2")
					//pos, ok = findMessage(buf[pos+1:], header, msg)
					pos, ok = findMessage(buf[pos+1:], msg)
				}
				
			}
			
		}
	} else {
	        fmt.Println("find msg failed 3")
		pos = len(buf)
	}
	return
}

func (ni *NsqInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) error {
	// Get the InputRunner's chan to receive empty PipelinePacks
	var pack *pipeline.PipelinePack
	var err error
	var dRunner pipeline.DecoderRunner
	var decoder pipeline.Decoder
	var ok bool
	var e error

	//pos := 0
	//output := make([]*Message, 2)
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
	err = ni.nsqReader.ConnectToLookupd(ni.conf.Address)
	if err != nil {
		ir.LogError(errors.New("ConnectToLookupd failed."))
	}

	//header := &message.Header{}

	stopped := false
	//readLoop:
	for !stopped {
	        //stopped = true
		select {
		case <-ni.stopChan:
                        ir.LogError(errors.New("get ni.stopChan, set stopped=true"))	
			stopped = true
		default:
			pack = <-packSupply
			m := <-ni.handler.logChan

			if ni.conf.Serialize {
				if dRunner == nil {
					pack.Recycle()
					ir.LogError(errors.New("Serialize messages require a decoder."))
				}
				//header := &message.Header{}
				//_, msgOk := findMessage(m.msg.Body, header, &(pack.MsgBytes))
				_, msgOk := findMessage(m.msg.Body,  &(pack.MsgBytes))
				if msgOk {
					dRunner.InChan() <- pack
				} else {
					pack.Recycle()
					ir.LogError(errors.New("Can't find Heka message."))
				}
				//header.Reset()
			} else {

				//ir.LogError(fmt.Errorf("message body: %s", m.msg.Body))
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
			}
			m.returnChannel <- &nsq.FinishedMessage{m.msg.Id, 0, true}
			/*
			   output[pos] = m
			   pos++
			   if pos == 2 {
			           for pos > 0 {
			                   pos--
			                   m1 := output[pos]
			                   m1.returnChannel <- &nsq.FinishedMessage{m1.msg.Id, 0, true}
			                   output[pos] = nil
			           }
			   }
			*/
		}
	}
	return nil
}

func (ni *NsqInput) Stop() {
	fmt.Println("enter func Stop()")
	close(ni.stopChan)
}

func init() {
	pipeline.RegisterPlugin("NsqInput", func() interface{} {
		return new(NsqInput)
	})
}
