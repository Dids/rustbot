# Things To Do

- [x] Add support for join/disconnect messages (Webrcon -> Discord)  
- [x] Add support for @user pinging (Webrcon -> Discord) [NOTE: Make sure to disallow @everyone and @here]  
- [x] Add support for "Playing with X players" (use Discord RPC)  
- [ ] Add support for unit tests and code coverage (codecov.io)
- [ ] Add support for commands (including automatic command detection and output parsing)  
- [ ] Add support for emotes (Discord -> Webrcon) [NOTE: Currently Discord emotes are simply dropped, which is not ideal]  
- [ ] Add optional support for death messages   

- [ ] Fix the following bug when shutting down and not connected:  
```
2018/12/24 10:01:35 Shutting down the Webrcon client..
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x65cd37]
2018-12-24T10:01:35.509213308Z 
goroutine 1 [running]:
github.com/gorilla/websocket.(*Conn).WriteMessage(0x0, 0x8, 0xc000168148, 0x2, 0x2, 0x2, 0x6c9760)
/go/pkg/mod/github.com/gorilla/websocket@v1.4.0/conn.go:742 +0x37
github.com/sacOO7/gowebsocket.(*Socket).send(0x9922e0, 0x8, 0xc000168148, 0x2, 0x2, 0x2, 0x80)
/go/pkg/mod/github.com/sac!o!o7/gowebsocket@v0.0.0-20180719182212-1436bb906a4e/gowebsocket.go:171 +0x6f
github.com/sacOO7/gowebsocket.(*Socket).Close(0x9922e0)
/go/pkg/mod/github.com/sac!o!o7/gowebsocket@v0.0.0-20180719182212-1436bb906a4e/gowebsocket.go:177 +0xc3
github.com/Dids/rustbot/webrcon.Close()
/tmp/rustbot/webrcon/root.go:194 +0xa2
main.main()
/tmp/rustbot/main.go:44 +0x2be
```

- [ ] Fix bot not exiting/failing when unable to connect:  
```
2018/12/24 10:06:43 Initializing the Webrcon client..
2018/12/24 10:06:43 Received connect error  dial tcp 172.23.0.2:28016: connect: connection refused
2018/12/24 10:06:57 Initializing the Discord client..
2018/12/24 10:06:57 Successfully created the Discord client
```
