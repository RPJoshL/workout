# Modifiying Manifest within aab package

Test if the build aab is valid *(aab → apk)*.

```sh
java -jar bundletool-all-1.15.6.jar  build-apks --bundle /home/ubuntugui/Downloads/RPdb_4.3.1-dev.aab --mode universal --output rollback.apks --ks /mnt/entwicklung/git/RPdb/client/android/keystore.jks --ks-pass "pass:xxx" --ks-key-alias key0
```

## Proto buffers

Download from GitHub:

* [Configuration](https://github.com/aosp-mirror/platform_frameworks_base/blob/master/tools/aapt2/Configuration.proto)
* [Ressources](https://github.com/aosp-mirror/platform_frameworks_base/blob/master/tools/aapt2/Resources.proto)