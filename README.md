在 /data/heka/cmake/externals.cmake 中

if (INCLUDE_MOZSVC)

    add_external_plugin(git https://github.com/mozilla-services/heka-mozsvc-plugins 6fe574dbd32a21f5d5583608a9d2339925edd2a7)
    
endif()

加入
add_external_plugin(git https://github.com/jbli/heka_plugin b26829487b668837b2964d81917bfde446fd4ff4)

重新编译
source build.sh
