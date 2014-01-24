在 /data/heka/cmake/externals.cmake 中

```
if (INCLUDE_MOZSVC)

    add_external_plugin(git https://github.com/mozilla-services/heka-mozsvc-plugins 6cb4a1610579c02bb25a8c0aaf835b05c3214d532)
    
endif()
```
加入

```
add_external_plugin(git https://github.com/jbli/heka_plugin 5058ade894076a4ecd9efb467a2af270f6181dde)

plugins:

- 5058ade894076a4ecd9efb467a2af270f6181dde  udpout


git_clone(https://github.com/adeven/redismq ec92d9cf876da73ed9659011d2a19c5ca325d2e7) 
git_clone(https://github.com/adeven/redis 6a7dfb6ac870f9bf9cece7fb7181dd31cf59f7a8) 
git_clone(https://github.com/matttproud/gocheck ecced547db7c1ed7223d400ae8b21820eacc85f3) 
git_clone(https://github.com/vmihailenco/bufio 5a6706bb583bf168d886202cb376c877e9c85521)
git_clone(https://github.com/bitly/go-nsq ac1dc8a491c8a37a88cb425bbd52fb3568f85dbe)
git_clone(https://github.com/mreiferson/go-snappystream 97c96e6648e99c2ce4fe7d169aa3f7368204e04d)
```
重新编译
source build.sh


tag1: 61cfe8611b418803ca51304be1f62d041533d87c   发布端同步， 需要手工杀进程


=================
```
[aabb]
type = "LogfileInput"
logfile = "/tmp/aabb.log"
#decoder = "apache_transform_decoder"
logger = "apache_test"

[apache_transform_decoder]
type = "PayloadRegexDecoder"
match_regex = '^(?P<RemoteIP>\S+) \S+ \S+ \[(?P<Timestamp>[^\]]+)\] "(?P<Method>[A-Z]+) (?P<Url>[^\s^\?]+)[^"]*" (?P<StatusCode>\d+) (?P<RequestSize>\d+) "(?P<Referer>[^\s^\?]+)[^"]*"' 
timestamp_layout = "02/Jan/2006:15:04:05 +0800"

[apache_transform_decoder.message_fields]
Type = "ApacheLogfile"
Logger = "apache"
RemoteIP = "%RemoteIP%"
Timestamp = "%Timestamp%"
Url|uri = "%Url%"
Status = "%StatusCode%"
RequestSize|B = "%RequestSize%"
Referer = "%Referer%"
```