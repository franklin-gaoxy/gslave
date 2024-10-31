# File mode



#### File.LocalFiles

```yaml
      describe:
        hostConcurrentMode: concurrent
        from: /data/file # from host path
        to: /data/ # to remote host path
```

> hostConcurrentMode: serial and concurrent

#### File.RemoteFile

```yaml
      describe:
        hostConcurrentMode: concurrent
        fromNetwork: "https://xxx.com/files/xxx" # from network
        to: /data/
        sslVerify: false
```

> sslVerify: Whether to enable SSL authentication, default enabled, is true.