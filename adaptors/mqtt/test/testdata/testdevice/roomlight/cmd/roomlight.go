package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	waitTimeout = time.Second * 10
)

var roomLight RoomLight

func NewCommand() *cobra.Command {
	var c = &cobra.Command{
		Use:  "roomlight",
		Long: "test device roomlight",
		RunE: func(cmd *cobra.Command, args []string) error {
			roomLight.Run()
			return nil
		},
	}

	c.Flags().StringVarP(&roomLight.cfg.Broker, "broker", "b", "tcp://127.0.0.1:1883", "broker of mqtt")
	c.Flags().StringVarP(&roomLight.cfg.SubTopic, "subtopic", "s", "device/room/light/cmd", "subscription topic")
	c.Flags().StringVarP(&roomLight.cfg.PubTopic, "pubtopic", "p", "device/room/light", "publish topic")
	c.Flags().StringVarP(&roomLight.cfg.Username, "username", "u", "parchk", "username of mqtt broker")
	c.Flags().StringVarP(&roomLight.cfg.Password, "password", "w", "test123", "password of mqtt broker")
	c.Flags().IntVarP(&roomLight.cfg.Qos, "qos", "q", 0, "qos of mqtt message")

	c.MarkFlagRequired("broker")

	return c
}

type Config struct {
	SubTopic string
	PubTopic string
	Qos      int
	Broker   string
	Username string
	Password string
}

type PowerStatus struct {
	PowerDissipation string `json:"powerDissipation"`
	ElectricQuantity string `json:"electricQuantity"`
}

type Status struct {
	Switch     string      `json:"switch"`
	Brightness int         `json:"brightness"`
	Power      PowerStatus `json:"power"`
	Attr       []int       `json:"attr"`
}

type RoomLight struct {
	cfg    Config
	Status Status
	Client MQTT.Client
}

func (r *RoomLight) Run() {
	tick := time.NewTicker(time.Second * 5)
	var stop = ctrl.SetupSignalHandler()
	if err := r.Init(); err != nil {
		fmt.Println("room light init error:", err.Error())
		return
	}
	if err := r.Subscribe(); err != nil {
		fmt.Println("room light subscribe error:", err.Error())
		return
	}
	for {
		select {
		case <-tick.C:
			if err := r.Report(); err != nil {
				fmt.Println("report error:", err.Error())
			}
		case <-stop:
			return
		}
	}
}

func (r *RoomLight) Init() error {

	r.Status.Switch = "off"
	r.Status.Brightness = 1
	r.Status.Power.PowerDissipation = "10KWh"
	r.Status.Power.ElectricQuantity = "10%"
	r.Status.Attr = append(r.Status.Attr, 13)
	r.Status.Attr = append(r.Status.Attr, 15)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(r.cfg.Broker)
	opts.SetClientID("testRoomLight")
	opts.SetUsername(r.cfg.Username)
	opts.SetPassword(r.cfg.Password)
	opts.SetOrderMatters(true)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(false)

	r.Client = MQTT.NewClient(opts)
	if token := r.Client.Connect(); token.WaitTimeout(waitTimeout) && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (r *RoomLight) Report() error {
	payload, err := json.Marshal(&r.Status)
	if err != nil {
		return err
	}
	token := r.Client.Publish(r.cfg.PubTopic, byte(r.cfg.Qos), true, payload)

	if token.WaitTimeout(waitTimeout) && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("room light report topic: %s, status:%s \n", r.cfg.PubTopic, string(payload))

	return nil
}

func (r *RoomLight) Subscribe() error {
	callback := func(client MQTT.Client, msg MQTT.Message) {
		fmt.Println("Subscribe payload:", string(msg.Payload()))
		if err := json.Unmarshal(msg.Payload(), &r.Status); err != nil {
			fmt.Println("subscribe callback error:", err.Error())
		}
		fmt.Printf("Subscribe update status:%+v\n", r.Status)
	}
	token := r.Client.Subscribe(r.cfg.SubTopic, byte(r.cfg.Qos), callback)
	if token.WaitTimeout(waitTimeout) && token.Error() != nil {
		return token.Error()
	}

	return nil
}
