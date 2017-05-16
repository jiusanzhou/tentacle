# Tentacle

代理服务有服务端和客户端组成，客户端是真正的网络出口，目前是PC版本独立程序（包含但不限于linux，windows，darwin），
后续会增加手机SDK版本，内置在android和iOS程序中，使运行App的终端作为节点网络出口。

## 服务端

服务端的功能是管理出口节点，提供Socks5或HTTP代理服务供使用者使用，并将网络请求调度至出口节点。

## 客户端

客户端指的就是网络节点出口，将服务端给的请求发出并回传数据给服务端。同时他还能够接受服务端发出的控制指令；
比如，重新拨号等。

## 功能流程

服务端开启4个端口（默认）：
1. `4442`控制通道
2. `4443`数据通道
3. `8888`Socks5代理
4. `8887`HTTP代理

3，4是代理服务无需介绍。

每个客户端程序起来后，会与服务端的`4442`建立连接，并处理心跳包，控制指令等。
当服务端收到代理请求时，会通过控制通道发送`Dial`指令，这个里面包含了目标主机的地址与本机的数据通道端口信息；
客户端收到`Dial`指令后，会同时向需要请求的目标主机和服务端的`4443`建立连接，然后在这两个连接上进行数据转发，
完成后断开这两个连接。

```
+----------------------------------------------------------------------+
|                                                                      |
|                                                                      |
|                       Tentacle                                       |
|                                                                      |
|                                           Tentacled                  |
|                                          +------------+              |
|                                          |            |              |
|                                          |            |              |
|   Tentacler                              |            |              |
|    +------+                              |            |              |
|    |      |                         :4442|Control     |              |
|    |      <----^-------------------^----->            |              |
|    |      |                              |            |Socks5 Proxy  |
|    |      <----------------+             |            ++             |
|    |      |                |             |             |             |
|    +------+                |             |             |             |
|                            |             |            ++             |
|                            |             |            |              |
|                            |        :4443|Tunnel      |              |
|                            +------------->            |              |
|                                          |            |              |
|                                          |            |              |
|                                          |            |              |
|                                          +------------+              |
|                                                                      |
+----------------------------------------------------------------------+
```

## 使用

```bash
make build
```

or

```bash
make build-all
```

### 服务端

```
Usage of tentacled:
  -controlAddr string
        Public address listening for tentacle client (default ":4442")
  -httpAddr string
        Public address listening forhttp proxy (default ":8887")
  -log string
        Write log messages to this file. 'stdout' and 'none' have special meanin
gs (default "stdout")
  -log-level string
        The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR (defau
lt "DEBUG")
  -redial-interval duration
        Redial interval for each tentacler (default 1m0s)
  -socketAddr string
        Public address listening for socket5 proxy (default ":8888")
  -tlsCrt string
        Path to a TLS certificate file
  -tlsKey string
        Path to a TLS key file
  -tunnelAddr string
        Public address listening for tentacle request (default ":4443")
```

### 客户端

```
To honor the memory of fox&rabbit.
  -config string
        Path to ngrok configuration file. (default: $HOME/.tentacler)
  -log string
        Write log messages to this file. 'stdout' and 'none' have special meanin
gs (default "stdout")
  -log-level string
        The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR (defau
lt "DEBUG")
  -pool-size int
        Pool size for connections to tentacle service, 0 for no pool

Advanced usage: tentacler [OPTIONS] <command> [command args] [...]
        tentacler info                          List info from tentacled service
.
        tentacler start [tcp] [...]             Start and regist to tentacled se
rvice.
        tentacler help                          Print help
        tentacler version                       Print tentacle version

Examples:
        tentacler start ilife codertool
        tentacler -log=stdout -config=tentacler.yml start
        tentacler version
```

## TODO

1. tls
2. SDK
3. web admin
4. connection pool