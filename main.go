package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	broker   string
	port     int
	topic    string
	log2file bool
	client   mqtt.Client
)

type WoL struct {
	IP string `json:"ip"`
	HW string `json:"hw"`
}

func processArgs() {
	flag.BoolVar(&log2file, "log2file", false, "log to wol-mqtt.log in app directory instead of syslog (default false)")
	flag.StringVar(&broker, "b", "127.0.0.1", "mqtt broker to subscribe to")
	flag.IntVar(&port, "p", 1883, "TCP port where the mqtt broker process is listening")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("usage: wol-mqtt [-b=127.0.0.1] [-p=1883] [-log2file=false] [mqtt topic]")
	} else {
		topic = flag.CommandLine.Arg(0)
	}
}

var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
	log.Println("msg received: " + string(payload))

	var wol_data WoL
	err := json.Unmarshal(payload, &wol_data)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("running wakeonlan -i " + wol_data.IP + " " + wol_data.HW)
		cmd := exec.Command("wakeonlan", "-i", wol_data.IP, wol_data.HW)
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
		}
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}
}

func subscribeMQTT() {
	opts := mqtt.NewClientOptions().AddBroker(broker + ":" + strconv.Itoa(port)).SetClientID("wol-mqtt")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(handler)
	opts.SetPingTimeout(1 * time.Second)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
	}
	log.Println("client connected to broker " + broker)

	if token := client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}
	log.Println("subscribed to topic " + topic)
}

func main() {

	processArgs()
	if log2file {
		f, err := os.OpenFile("wol-mqtt.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		log.SetOutput(f)
	} else {
		sysLog, err := syslog.New(syslog.LOG_ALERT|syslog.LOG_DAEMON, "wol-mqtt")
		if err != nil {
			panic(err)
		}
		log.SetOutput(sysLog)
	}

	subscribeMQTT()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		log.Println("process entered running state")
		defer wg.Done()
		<-ctx.Done()
		log.Println("process received stop signal")
	}()

	wg.Wait()
	client.Disconnect(500)
	log.Println("process stopped successfully")

}
