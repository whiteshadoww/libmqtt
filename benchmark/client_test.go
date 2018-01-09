/*
 * Copyright GoIIoT (https://github.com/goiiot)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package benchmark

import (
	"net/url"
	"testing"

	pah "github.com/eclipse/paho.mqtt.golang"
	lib "github.com/goiiot/libmqtt"
	gmq "github.com/yosssi/gmq/mqtt/client"
)

//smqM "github.com/surgemq/message"
//smq "github.com/surgemq/surgemq/service"

const (
	testKeepalive = 3600             // prevent keepalive packet disturb
	testServer    = "localhost:1883" // emqttd server address
	testTopic     = "/foo"
	testQos       = 0
	testBufSize   = 1024 // same with gmq default

	testPubCount = 1
)

var (
	// 256 bytes
	testTopicMsg = []byte(
		"1234567890" + "1234567890" + "1234567890" + "1234567890" + "1234567890" +
			"1234567890" + "1234567890" + "1234567890" + "1234567890" + "1234567890" +
			"1234567890" + "1234567890" + "1234567890" + "1234567890" + "1234567890" +
			"1234567890" + "1234567890" + "1234567890" + "1234567890" + "1234567890" +
			"1234567890" + "1234567890" + "1234567890" + "1234567890" + "1234567890",
	)
)

func BenchmarkLibmqttClient(b *testing.B) {
	b.N = testPubCount
	b.ReportAllocs()
	var client lib.Client
	var err error

	if client, err = lib.NewClient(
		lib.WithLog(lib.Error),
		lib.WithServer(testServer),
		lib.WithKeepalive(testKeepalive, 1.2),
		lib.WithRecvBuf(testBufSize),
		lib.WithSendBuf(testBufSize),
		lib.WithCleanSession(true)); err != nil {
		b.Error(err)
	}

	client.HandleUnSub(func(topic []string, err error) {
		if err != nil {
			b.Error(err)
		}
		client.Destroy(true)
	})

	client.Connect(func(server string, code lib.ConnAckCode, err error) {
		if err != nil {
			b.Error(err)
		} else if code != lib.ConnAccepted {
			b.Error(code)
		}

		b.ResetTimer()
		println("connected")
		for i := 0; i < b.N; i++ {
			client.Publish(&lib.PublishPacket{
				TopicName: testTopic,
				Qos:       testQos,
				Payload:   testTopicMsg,
			})
		}
		client.UnSubscribe(testTopic)
	})
	client.Wait()
}

func BenchmarkPahoClient(b *testing.B) {
	b.N = testPubCount
	b.ReportAllocs()

	serverURL, err := url.Parse("tcp://" + testServer)
	if err != nil {
		b.Error(err)
	}

	client := pah.NewClient(&pah.ClientOptions{
		Servers:             []*url.URL{serverURL},
		KeepAlive:           testKeepalive,
		CleanSession:        true,
		ProtocolVersion:     4,
		MessageChannelDepth: testBufSize,
		Store:               pah.NewMemoryStore(),
	})

	t := client.Connect()
	if !t.Wait() {
		b.Fail()
	}

	if err := t.Error(); err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Publish(testTopic, 0, false, testTopicMsg)
	}
	t = client.Unsubscribe(testTopic)
	if !t.Wait() {
		b.Fail()
	}
	if err := t.Error(); err != nil {
		b.Error(err)
	}

	client.Disconnect(0)
}

func BenchmarkGmqClient(b *testing.B) {
	b.N = testPubCount
	b.ReportAllocs()

	client := gmq.New(&gmq.Options{ErrorHandler: func(e error) {
		if e != nil {
			b.Error(e)
		}
	}})
	if err := client.Connect(&gmq.ConnectOptions{
		Network:      "tcp",
		Address:      testServer,
		KeepAlive:    testKeepalive,
		CleanSession: true,
	}); err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := client.Publish(&gmq.PublishOptions{
			QoS:       testQos,
			TopicName: []byte(testTopic),
			Message:   testTopicMsg,
		}); err != nil {
			b.Error(err)
		}
	}
	if err := client.Unsubscribe(&gmq.UnsubscribeOptions{
		TopicFilters: [][]byte{[]byte(testTopic)},
	}); err != nil {
		b.Error(err)
	}

	client.Terminate()
}

//func BenchmarkSurgeClient(b *testing.B) {
//	client := &smq.Client{KeepAlive: 10}
//	err := client.Connect(testServer, &smqM.ConnectMessage{})
//	if err != nil {
//		b.Log(err)
//		b.FailNow()
//	}
//}
