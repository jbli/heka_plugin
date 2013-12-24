package examples

import (
        "fmt"
        "github.com/adeven/redismq"
        "github.com/mozilla-services/heka/pipeline"
)

type RedisMQInputConfig struct {
        Address string `toml:"address"`
        Decoder string `toml:"decoder"`
}

type RedisMQInput struct {
        conf            *RedisMQInputConfig
        rdqueue         *redismq.Queue
        rdconsumer      *redismq.Consumer
}

func (ri *RedisMQInput) ConfigStruct() interface{} {
        return &RedisMQInputConfig{"192.168.1.44", ""}
}

func (ri *RedisMQInput) Init(config interface{}) error {
        ri.conf = config.(*RedisMQInputConfig)

        var err error
        ri.conf = config.(*RedisMQOutputConfig)
        ri.rdqueue = redismq.CreateQueue(ro.conf.Address, "6379", "", 9, "clicks")
        ri.rdconsumer, err = ri.rdqueue.AddConsumer("testconsumer")
        if err != nil {
                panic(err)
        }
        return nil
}

func (ri *RedisMQInput) Run(ir pipeline.InputRunner, h pipeline.PluginHelper) error {
        // Get the InputRunner's chan to receive empty PipelinePacks
        packs := ir.InChan()

        var decoding chan<- *pipeline.PipelinePack
        if ri.conf.Decoder != "" {
                // Fetch specified decoder
                decoder, ok := h.DecoderSet().ByName(ri.conf.Decoder)
                if !ok {
                        err := fmt.Errorf("Could not find decoder", ri.conf.Decoder)
                        return err
                }

                // Get the decoder's receiving chan
                decoding = decoder.InChan()
        }

        var pack *pipeline.PipelinePack
        var p *redismq.Package
        var count int
        var b []byte
        var err error

        // Read data from websocket broadcast chan
        for {
                p, err = ri.rdconsumer.Get()
                if err != nil {
                        ir.LogError(err)
                        continue
                }
                err = p.Ack()
                if err != nil {
                        ir.LogError(err)
                }
                b = p.Payload
                // Grab an empty PipelinePack from the InputRunner
                pack = <-packs

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

func init() {
        pipeline.RegisterPlugin("RedisMQInput", func() interface{} {
                return new(RedisMQInput)
        })
}
