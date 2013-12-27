package examples

import (
	"fmt"
	nsq "github.com/bitly/go-nsq"
	"github.com/mozilla-services/heka/pipeline"
)

type Message struct {
        *nsq.Message
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
	logChan chan *nsq.Message
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
	ni.handler = &MyTestHandler{ logChan:         make(chan *nsq.Message)}
	ni.nsqReader.AddAsyncHandler(ni.handler)
	return nil
}

func (ni *NsqInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) error {
	// Get the InputRunner's chan to receive empty PipelinePacks
	var pack *pipeline.PipelinePack
	var count int
	var b []byte
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



//readLoop:
	for {
		pack = <-packs
                m := <-ni.handler.logChan
		b = []byte(m.Body)
		// Grab an empty PipelinePack from the InputRunner

		// Trim the excess empty bytes
		count = len(b)
		pack.MsgBytes = pack.MsgBytes[:count]

		// Copy ws bytes into pack's bytes
		copy(pack.MsgBytes, b)

		if decoding != nil {
			// Send pack onto decoder
			decoding <- pack
		} else {
			// Send pack into Heka pipeline
			ir.Inject(pack)
		}
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
