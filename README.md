在 /data/heka/cmake/externals.cmake 中

if (INCLUDE_MOZSVC)

    add_external_plugin(git https://github.com/mozilla-services/heka-mozsvc-plugins 6cb4a1610579c02bb25a8c0aaf835b05c3214d532)
    
endif()

加入

add_external_plugin(git https://github.com/jbli/heka_plugin dbecddeeb49f18a976b4efd3b9bdb4d57b2adc8a)


git_clone(https://github.com/adeven/redismq ec92d9cf876da73ed9659011d2a19c5ca325d2e7)
git_clone(https://github.com/adeven/redis 6a7dfb6ac870f9bf9cece7fb7181dd31cf59f7a8)
git_clone(https://github.com/matttproud/gocheck ecced547db7c1ed7223d400ae8b21820eacc85f3)
git_clone(https://github.com/vmihailenco/bufio 77549187b2c18cc26f0127a8afd40c379dd99ab2)
git_clone(https://github.com/bitly/go-nsq ac1dc8a491c8a37a88cb425bbd52fb3568f85dbe)
git_clone(https://github.com/mreiferson/go-snappystream 97c96e6648e99c2ce4fe7d169aa3f7368204e04d)

重新编译
source build.sh
