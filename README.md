# Web.go

## Graceful Shutdown

```bash
kill -2 $PID
```

```go
signal.Notify(sigint, os.Interrupt) // kill -2 pid
signal.Notify(sigint, syscall.SIGTERM) // kill pid
```

### Thanks for all open source projects， I learned a lot of from them.
### Special thanks to these two projects：
https://github.com/julienschmidt/httprouter

https://github.com/hoisie/web