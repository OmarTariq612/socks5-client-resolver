
# socks5 client dns resolver
```
socks5 client <--> socks5-client-resolver <--> socks5 server
```

`socks5-client-resolver` is a simple socks server implementation that resolves dns locally instead of sending it in the request to the specified socks5 server.
## Build

```
go build .
```
## Usage
```
usage: socks5-client-resolver [bind_addr:bind_port] server_addr:server_port
```
`bind_addr:bind_port` is `:5555` by default.

ex: to provide `192.168.1.50:3456` as the socks server address:
```
./socks5-client-resolver 192.168.1.50:3456
```
```
2022/08/11 04:59:42 Serving on :5555
2022/08/11 04:59:42 the provided server address is 192.168.1.50:3456




```
Now the program is ready to accept socks5 client connections on `:5555` and resolve dns requests locally and send the resolved IP addresses instead of sending domain names directly to the provided socks5 server `192.168.1.50:3456`.
## REF
* socks 5 (rfc 1928) : https://datatracker.ietf.org/doc/html/rfc1928