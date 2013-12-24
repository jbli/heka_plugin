package examples

import (
        "fmt"
        "github.com/adeven/redismq"
        "github.com/mozilla-services/heka/pipeline"
)

type RedisMQOutputConfig struct {
        Address string `toml:"address"`
}

type RedisMQOutput struct {
        conf    *RedisMQOutputConfig
        rdqueue *redismq.Queue
}

func (ro *RedisMQOutput) ConfigStruct() interface{} {
        return &RedisMQOutputConfig{"192.168.1.44"}
}

func (ro *RedisMQOutput) Init(config interface{}) error {
        ro.conf = config.(*RedisMQOutputConfig)
        ro.rdqueue = redismq.CreateQueue(ro.conf.Address, "6379", "", 9, "clicks")
        return nil
}

func (ro *RedisMQOutput) Run(or pipeline.OutputRunner, h pipeline.PluginHelper) error {
        defer func() {
        }()

        var b []byte
        var p [][]byte
        for pc := range or.InChan() {
                b = pc.Pack.MsgBytes
                p = [][]byte{nil, b}
                ro.rdqueue.Put(p)
                pc.Pack.Recycle()
        }

        return nil
}

func init() {
        pipeline.RegisterPlugin("RedisMQOutput", func() interface{} {
                return new(RedisMQOutput)
        })
}
