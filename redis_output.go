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
        context *redismq.Context
        socket  *redismq.Socket
}

func (ro *RedisMQOutput) ConfigStruct() interface{} {
        return &RedisMQOutputConfig{":6379"}
}

func (ro *RedisMQOutput) Init(config interface{}) error {
        ro.conf = config.(*RedisMQOutputConfig)

        var err error
        testQueue := redismq.CreateQueue("192.168.1.44", "6379", "", 9, "clicks")
        

        return nil
}

func (zo *ZeroMQOutput) Run(or pipeline.OutputRunner, h pipeline.PluginHelper) error {
        defer func() {
                zo.socket.Close()
                zo.context.Close()
        }()

        var b []byte
        var p [][]byte
        for pc := range or.InChan() {
                b = pc.Pack.MsgBytes
                p = [][]byte{nil, b}
                zo.socket.SendMultipart(p, 0)
                pc.Pack.Recycle()
        }

        return nil
}

func init() {
        pipeline.RegisterPlugin("ZeroMQOutput", func() interface{} {
                return new(ZeroMQOutput)
        })
}
