# Command mode



#### Command.ExecuteCommand

```yaml
      describe:
        command: "apt-get install apt-file -y" # execute command
        hostConcurrentMode: concurrent
        stepMode: "serial"
        betchNum: 10
```

> hostConcurrentMode: serial, concurrent, batch.
>
> If the mode is not batch, then the betchNum parameter does not have to exist

