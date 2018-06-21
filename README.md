# mqtt-server
a simple mqtt server support plugins

# MQTT 协议

## 格式

固定头部+可变头部+消息体

### 固定头部

<table>
   <tr>
      <td>Bit</td>
      <td>7</td>
      <td>6</td>
      <td>5</td>
      <td>4</td>
      <td>3</td>
      <td>2</td>
      <td>1</td>
      <td>0</td>
   </tr>
   <tr>
      <td>byte 1</td>
      <td colspan="4">Message Type</td>
      <td>DUP flag</td>
      <td colspan="2">QOS level</td>
      <td>RETAIN</td>
   </tr>
   <tr>
      <td> byte 2 ... </td>
      <td colspan="8">Remaining length （可变长编码）</td>
   </tr>
</table>

1. Message Type

    ```go
    _           Type = iota // 保留
	CONNECT                 // 请求连接
	CONNACK                 // 请求应答
	PUBLISH                 // 发布消息
	PUBACK                  // 发布应答
	PUBREC                  // 发布已接收，保证传递1
	PUBREL                  // 发布释放，保证传递2
	PUBCOMP                 // 发布完成，保证传递3
	SUBSCRIBE               // 订阅请求
	SUBACK                  // 订阅应答
	UNSUBSCRIBE             // 取消订阅
	UNSUBACK                // 取消订阅应答
	PINGREQ                 // ping请求
	PINGRESP                // ping响应
	DISCONNECT              // 断开连接
	_                       // 保留
    ```
2. UDP Flag
    用于保证消息的可靠传输， 设为1，则在下面的边长头部中加入MessageID, 并且需要回复确认。
3. Qos
    主要用于PUBLISH（发布态）消息的，保证消息传递的次数。
    
    * 00表示最多一次 即<=1
    * 01表示至少一次  即>=1
    * 10表示一次，即==1
    * 11保留后用
4. Retain
    主要用于PUBLISH(发布态)的消息，表示服务器要保留这次推送的信息，如果有新的订阅者出现，就把这消息推送给它。如果不设那么推送至当前订阅的就释放了。

5. 固定头部的剩余字节（最多4个字节）
    使用边长数据保存边长头部+消息体的总大小。最多可以实现将近256M的数据。

### 可变头部



### 消息体