package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hypebeast/go-osc/osc"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/pkg/errors"
)

type oscClient struct {
	conn *net.UDPConn
}

func newOscClient(addr string) (*oscClient, error) {
	c := &oscClient{}

	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	c.conn = conn

	return c, nil
}

func (c *oscClient) getChannel(ch int) (float32, error) {
	msg := osc.NewMessage(fmt.Sprintf("/ch/%02d/mix/fader", ch))
	data, err := msg.MarshalBinary()
	if err != nil {
		return 0, err
	}

	if _, err = c.conn.Write(data); err != nil {
		return 0, err
	}

	data = make([]byte, 65535)
	_, err = c.conn.Read(data)
	if err != nil {
		return 0, err
	}

	packet, err := osc.ParsePacket(string(data))
	if err != nil {
		return 0, err
	}

	switch packetMsg := packet.(type) {
	case *osc.Message:
		return packetMsg.Arguments[0].(float32), nil
	default:
		return 0, errors.Errorf("Unknown packet/message type %T", packetMsg)
	}
}

func (c *oscClient) close() {
	c.conn.Close()
}

type artnetController struct {
	conn          *net.UDPConn
	remoteAddr    *net.UDPAddr
	channels      [512]byte
	universe      uint8
	channelOffset int
}

func newArtnetController(addr string, universe uint8, channelOffset int) (*artnetController, error) {
	c := &artnetController{}

	remoteAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	c.universe = universe
	c.channelOffset = channelOffset

	return c, nil
}

func (c *artnetController) setChannel(ch int, value byte) {
	c.channels[ch+c.channelOffset] = value
}

func (c *artnetController) sendChannels() error {
	p := &packet.ArtDMXPacket{
		Sequence: 0x00,
		SubUni:   0,
		Net:      0,
		Data:     c.channels,
	}

	data, err := p.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *artnetController) close() {
	c.conn.Close()
}

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: osc-to-artnet OSC_ADDR ARTNET_ADDR ARTNET_UNIVERSE ARTNET_CHANNEL_OFFSET\nExample: osc-to-artnet 192.168.1.65:10023 2.0.0.1:6454 0 0")
	}

	universe, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	channelOffset, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	client, err := newOscClient(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer client.close()

	controller, err := newArtnetController(os.Args[1], uint8(universe), channelOffset)
	if err != nil {
		log.Fatal(err)
	}
	defer controller.close()

loop:
	for {
		for i := 0; i < 32; i++ {
			value, err := client.getChannel(i + 1)
			if err != nil {
				log.Fatal(err)
			}

			controller.setChannel(i, byte(value*255))
		}

		err = controller.sendChannels()
		if err != nil {
			log.Fatal(err)
		}

		select {
		case sig := <-c:
			log.Printf("Got signal %v, exiting.", sig)
			break loop
		default:
			time.Sleep(16 * time.Millisecond)
		}
	}
}
