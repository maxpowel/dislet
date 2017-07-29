package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
	"github.com/jinzhu/gorm"
	"github.com/fatih/color"
	"github.com/maxpowel/dislet"
)

type MqttConfig struct {
	Hostname string `default:"valor"`
	Port int `default:"PUERTOOO"`
	Topic string
}

type PlayList struct {
	gorm.Model
	Source string
	SourceType string
}


func Bootstrap(k *dislet.Kernel) {
	mapping := k.Config.Mapping
	mapping["mqtt"] = &MqttConfig{}

	var baz dislet.OnKernelReady = func(k *dislet.Kernel){
		color.Green("Booting mqtt")
		conf := k.Config.Mapping["mqtt"].(*MqttConfig)
		//conf = k.Config.mapping["mqtt"]
		// Start mqtt connection
		//opts := mqtt.NewClientOptions().AddBroker("tcp://iot.eclipse.org:1883").SetClientID("gotrivial")
		//fmt.Println(fmt.Sprintf("tcp://%v:%v", conf.Hostname, conf.Port))
		service := func() {
			opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", conf.Hostname, conf.Port))

			//opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", "a", "b"))
			opts.SetKeepAlive(2 * time.Second)
			var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
				msg.Topic()
				color.Blue("TOPIC: %s\n", msg.Topic())
				switch msg.Topic() {
				case "setup":
				}

			}

			opts.SetDefaultPublishHandler(f)
			opts.SetPingTimeout(1 * time.Second)

			c := mqtt.NewClient(opts)
			if token := c.Connect(); token.Wait() && token.Error() != nil {
				color.Red(token.Error().Error())
				return
			}
			color.Green("MQTT connection established with %v", conf.Hostname)
			if token := c.Subscribe(conf.Topic, 0, nil); token.Wait() && token.Error() != nil {
				color.Red(token.Error().Error())
				return
			}
			color.Green("Subscribed to %v", conf.Topic)

			defer func() {
				color.Green("Disconnecting")
				c.Disconnect(250)
			}()

			defer func() {
				color.Green("Unsubscribing")
				unsubscribeToken := c.Unsubscribe(conf.Topic)
				unsubscribeToken.Wait()
			}()

			dislet.Daemonize()
		}
		go service()
	}
	k.Subscribe(baz)
}
