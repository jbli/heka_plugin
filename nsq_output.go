package examples

import (
        "fmt"
        nsq "github.com/bitly/go-nsq"
        "github.com/mozilla-services/heka/pipeline"
)

type NsqOutputConfig struct {
        Address string `toml:"address"`
}

type NsqOutput struct {
        conf    *NsqOutputConfig
        nsqwriter *nsq.Writer
}

func (no *NsqOutput) ConfigStruct() interface{} {
        return &NsqOutputConfig{"192.168.1.44:4160"}
}

func (no *NsqOutput) Init(config interface{}) error {
        no.conf = config.(*NsqOutputConfig)
        no.nsqwriter = nsq.NewWriter(no.conf.Address)
        return nil
}

func (no *NsqOutput) Run(or pipeline.OutputRunner, h pipeline.PluginHelper) error {
        var outgoing string
        for pack := range or.InChan() {
                err := writer.PublishAsync("test", []byte(pack.Message.GetPayload()),nil)
                if err != nil{
                  or.LogError(fmt.Errorf("error in writer.PublishAsync"))
                }
                pack.Recycle()
        }

        return nil
}

func init() {
        pipeline.RegisterPlugin("NsqOutput", func() interface{} {
                return new(NsqOutput)
        })
}
