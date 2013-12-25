在 /data/heka/cmake/externals.cmake 中

if (INCLUDE_MOZSVC)

    add_external_plugin(git https://github.com/mozilla-services/heka-mozsvc-plugins 6cb4a1610579c02bb25a8c0aaf835b05c3214d532)
    
endif()

加入

add_external_plugin(git https://github.com/jbli/heka_plugin b26829487b668837b2964d81917bfde446fd4ff4)


git_clone(https://github.com/adeven/redismq ec92d9cf876da73ed9659011d2a19c5ca325d2e7)
git_clone(https://github.com/adeven/redis 6a7dfb6ac870f9bf9cece7fb7181dd31cf59f7a8)
git_clone(https://github.com/matttproud/gocheck ecced547db7c1ed7223d400ae8b21820eacc85f3)
git_clone(https://github.com/vmihailenco/bufio 77549187b2c18cc26f0127a8afd40c379dd99ab2)

重新编译
source build.sh
