package mqtt

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
	"github.com/fatih/color"
	"github.com/maxpowel/dislet"
	"log"
)

type Config struct {
	Hostname string `default:"valor"`
	Port int `default:"PUERTOOO"`
	Topic string
}

type TopicManager struct {
	Config *Config
	Client mqtt.Client
}

type Topic interface {
	Name() string
	OnMessage(client mqtt.Client, msg mqtt.Message)
}

func (tm *TopicManager) Subscribe(topic Topic) (error){

	if token := tm.Client.Subscribe(topic.Name(), 0, topic.OnMessage); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
	//color.Green("Subscribed to %v", conf.Topic)

	/*defer func() {
		color.Green("Unsubscribing")
		unsubscribeToken := c.Unsubscribe(conf.Topic)
		unsubscribeToken.Wait()
	}()*/
}

func (tm *TopicManager) Unsubscribe(topic string) (error){
	if token := tm.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
//	defer func() {
//		unsubscribeToken := tm.Client.Unsubscribe(topic)
//		unsubscribeToken.Wait()
//	}()
}

func (tm *TopicManager) Publish(topic string, payload interface{}) (error) {
	if token := tm.Client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func NewTopicManager(k *dislet.Kernel, conf *Config) *TopicManager {
	tm := &TopicManager{
		Client: k.Container.MustGet("mqtt").(mqtt.Client),
		Config: conf,
	}
	return tm

}
func NewClient(conf *Config) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", conf.Hostname, conf.Port))

	//opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", "a", "b"))
	opts.SetKeepAlive(2 * time.Second)
	/*var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		color.Blue("TOPIC: %s\n", msg.Topic())
	}

	opts.SetDefaultPublishHandler(f)*/
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf(token.Error().Error())
	}

	return c
}

func Bootstrap(k *dislet.Kernel) {
	mapping := k.Config.Mapping
	mapping["mqtt"] = &Config{}

	var baz dislet.OnKernelReady = func(k *dislet.Kernel){
		color.Green("Booting mqtt")
		conf := k.Config.Mapping["mqtt"].(*Config)
		//conf = k.Config.mapping["mqtt"]
		// Start mqtt connection
		//opts := mqtt.NewClientOptions().AddBroker("tcp://iot.eclipse.org:1883").SetClientID("gotrivial")
		//fmt.Println(fmt.Sprintf("tcp://%v:%v", conf.Hostname, conf.Port))

		k.Container.RegisterType("mqtt", NewClient, conf)
		k.Container.RegisterType("topic_manager", NewTopicManager, k, conf)


		service := func() {
			c := k.Container.MustGet("mqtt").(mqtt.Client)
			color.Green("MQTT connection established with %v", conf.Hostname)
			defer func() {
				color.Green("Disconnecting")
				c.Disconnect(250)
			}()

			dislet.Daemonize()
		}
		go service()
	}
	k.Subscribe(baz)
}
