# WoL-mqtt
I am having devices in home IoT vLAN I would like to remotely turn on via wake-on-LAN from my OpenHAB home automationm platform. Though, as OpenHAB is running in the main vLAN, its broadcast packets cannot reach devices on other vLANs. 
After unsuccessfuly tinkering with solution alteranatives at network level, I decided to approch the issue differently. 

WoL-MQTT runs on network device connected to both vLANs (router / smart switch / AP ), and is essentially mqtt client listening for wake-on-lan message. 
Once recieved, it is triggers the [wakeonlan](https://openwrt.org/packages/pkgdata/wakeonlan) (must be installed on the device), and sends the necessary magic packet to the destination requested in the mqtt payload.

The executable is statically linked and is meant to be run as background service. On OpenWRT, I run it via [procd init script](https://openwrt.org/docs/guide-developer/procd-init-scripts): ref. [wol-mqtt](wol-mqtt)

Basic configuration is via command line flags and arguments
```
wol-mqtt [-b=127.0.0.1] [-p=1883] [-log2file=false] [mqtt topic]
```

```
~$ wol-mqtt -h
Usage of wol-mqtt:
  -b string
    	mqtt broker to subscribe to (default "127.0.0.1")
  -log2file
    	log to wol-mqtt.log in app directory instead of syslog (default false)
  -p int
    	TCP port where the mqtt broker process is listening (default 1883)
~$ 
```

MQTT message payload is expected in JSON format as follows
```json
{"ip":"destination broadcast address","hw":"hardware address of the receiver NIC"}
```

Example test wake-up message
```
mosquitto_pub -h 192.168.1.10 -t wol -m '{"ip":"192.168.20.255","hw":"2e:12:70:7f:a0:06"}'
```
