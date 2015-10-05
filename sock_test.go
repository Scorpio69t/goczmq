package goczmq

import (
	"encoding/gob"
	"fmt"
	"io"
	"testing"
)

func TestSendFrame(t *testing.T) {
	pushSock := NewSock(Push)
	defer pushSock.Destroy()

	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind("inproc://test-send-frame")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.Connect("inproc://test-send-frame")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.SendFrame([]byte("Hello"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	frame, flag, err := pullSock.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello", string(frame); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	frame, flag, err = pullSock.RecvFrameNoWait()
	if err == nil {
		t.Error(err)
	}

	if want, got := true, flag == 0; want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	err = pushSock.SendFrame([]byte("World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	frame, flag, err = pullSock.RecvFrameNoWait()
	if err != nil {
		t.Error(err)
	}

	if want, got := true, flag == 0; want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	err = pushSock.SendFrame([]byte("World"), FlagNone)
	if err != nil {
		t.Error(err)
	}
}

func TestSendEmptyFrame(t *testing.T) {
	pushSock := NewSock(Push)
	defer pushSock.Destroy()

	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind("inproc://test-send-empty")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.Connect("inproc://test-send-empty")
	if err != nil {
		t.Error(err)
	}

	var empty []byte
	err = pushSock.SendFrame(empty, FlagNone)
	if err != nil {
		t.Error(err)
	}

	frame, _, err := pullSock.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := 0, len(frame); want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}
}

func TestSendMessage(t *testing.T) {
	pushSock := NewSock(Push)
	defer pushSock.Destroy()

	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind("inproc://test-send-msg")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.Connect("inproc://test-send-msg")
	if err != nil {
		t.Error(err)
	}

	pushSock.SendMessage([][]byte{[]byte("Hello")})
	msg, err := pullSock.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello", string(msg[0]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	msg, err = pullSock.RecvMessageNoWait()
	if err == nil {
		t.Error(err)
	}

	pushSock.SendMessage([][]byte{[]byte("World")})
	msg, err = pullSock.RecvMessageNoWait()
	if err != nil {
		t.Error(err)
	}

	if want, got := "World", string(msg[0]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestPubSub(t *testing.T) {
	bogusPub, err := NewPub("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusPub.Destroy()

	bogusSub, err := NewSub("bogus://bogus", "")
	if err == nil {
		t.Error(err)
	}
	defer bogusSub.Destroy()

	pub, err := NewPub("inproc://pub1,inproc://pub2")
	if err != nil {
		t.Error(err)
	}
	defer pub.Destroy()

	sub, err := NewSub("inproc://pub1,inproc://pub2", "")
	if err != nil {
		t.Error(err)
	}
	defer sub.Destroy()

	err = pub.SendFrame([]byte("test pub sub"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	frame, _, err := sub.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "test pub sub", string(frame); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestReqRep(t *testing.T) {
	bogusReq, err := NewReq("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusReq.Destroy()

	bogusRep, err := NewRep("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusRep.Destroy()

	rep, err := NewRep("inproc://rep1,inproc://rep2")
	if err != nil {
		t.Error(err)
	}
	defer rep.Destroy()

	req, err := NewReq("inproc://rep1,inproc://rep2")
	if err != nil {
		t.Error(err)
	}
	defer req.Destroy()

	err = req.SendFrame([]byte("Hello"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	reqframe, _, err := rep.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello", string(reqframe); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	err = rep.SendFrame([]byte("World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	repframe, _, err := req.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "World", string(repframe); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestPushPull(t *testing.T) {
	bogusPush, err := NewPush("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusPush.Destroy()

	bogusPull, err := NewPull("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusPull.Destroy()

	push, err := NewPush("inproc://push1,inproc://push2")
	if err != nil {
		t.Error(err)
	}
	defer push.Destroy()

	pull, err := NewPull("inproc://push1,inproc://push2")
	if err != nil {
		t.Error(err)
	}
	defer pull.Destroy()

	err = push.SendFrame([]byte("Hello"), FlagMore)
	if err != nil {
		t.Error(err)
	}

	err = push.SendFrame([]byte("World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	msg, err := pull.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello", string(msg[0]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := "World", string(msg[1]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestRouterDealer(t *testing.T) {
	bogusDealer, err := NewDealer("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusDealer.Destroy()

	bogusRouter, err := NewRouter("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusRouter.Destroy()

	dealer, err := NewDealer("inproc://router1,inproc://router2")
	if err != nil {
		t.Error(err)
	}
	defer dealer.Destroy()

	router, err := NewRouter("inproc://router1,inproc://router2")
	if err != nil {
		t.Error(err)
	}
	defer router.Destroy()

	err = dealer.SendFrame([]byte("Hello"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	msg, err := router.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, got := 2, len(msg); want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := "Hello", string(msg[1]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	msg[1] = []byte("World")

	err = router.SendMessage(msg)
	if err != nil {
		t.Error(err)
	}

	msg, err = dealer.RecvMessage()
	if err != nil {
		t.Error(err)
	}

	if want, got := 1, len(msg); want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := "World", string(msg[0]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestXSubXPub(t *testing.T) {
	bogusXPub, err := NewXPub("bogus://bogus")
	if err == nil {
		t.Error("NewXPub should have returned error and did not")
	}
	defer bogusXPub.Destroy()

	bogusXSub, err := NewXSub("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusXSub.Destroy()

	xpub, err := NewXPub("inproc://xpub1,inproc://xpub2")
	if err != nil {
		t.Error(err)
	}
	defer xpub.Destroy()

	xsub, err := NewXSub("inproc://xpub1,inproc://xpub2")
	if err != nil {
		t.Error(err)
	}
	defer xsub.Destroy()
}

func TestPair(t *testing.T) {
	bogusPair, err := NewPair("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusPair.Destroy()

	pair1, err := NewPair(">inproc://pair")
	if err != nil {
		t.Error(err)
	}
	defer pair1.Destroy()

	pair2, err := NewPair("@inproc://pair")
	if err != nil {
		t.Error(err)
	}
	defer pair2.Destroy()
}

func TestStream(t *testing.T) {
	bogusStream, err := NewStream("bogus://bogus")
	if err == nil {
		t.Error(err)
	}
	defer bogusStream.Destroy()

	stream1, err := NewStream(">inproc://stream")
	if err != nil {
		t.Error(err)
	}
	defer stream1.Destroy()

	stream2, err := NewStream("@inproc://stream")
	if err != nil {
		t.Error(err)
	}
	defer stream2.Destroy()

}

func TestPollin(t *testing.T) {
	push, err := NewPush("inproc://pollin")
	if err != nil {
		t.Error(err)
	}
	defer push.Destroy()

	pull, err := NewPull("inproc://pollin")
	if err != nil {
		t.Error(err)
	}
	defer pull.Destroy()

	if want, got := false, pull.Pollin(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	err = push.SendFrame([]byte("Hello World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	if want, got := true, pull.Pollin(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}
}

func TestPollout(t *testing.T) {
	push := NewSock(Push)
	_, err := push.Bind("inproc://pollout")
	if err != nil {
		t.Error(err)
	}
	defer push.Destroy()

	if want, got := false, push.Pollout(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}

	pull := NewSock(Pull)
	defer pull.Destroy()

	err = pull.Connect("inproc://pollout")
	if err != nil {
		t.Error(err)
	}

	if want, got := true, push.Pollout(); want != got {
		t.Errorf("want '%v', got '%v'", want, got)
	}
}

func TestReader(t *testing.T) {
	pushSock := NewSock(Push)
	defer pushSock.Destroy()

	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind("inproc://test-read")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.Connect("inproc://test-read")
	if err != nil {
		t.Error(err)
	}

	err = pushSock.SendFrame([]byte("Hello"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	b := make([]byte, 5)

	n, err := pullSock.Read(b)

	if want, got := 5, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello", string(b); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	err = pushSock.SendFrame([]byte("Hello"), FlagMore)
	if err != nil {
		t.Error(err)
	}

	err = pushSock.SendFrame([]byte(" World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	b = make([]byte, 8)
	n, err = pullSock.Read(b)
	if want, got := ErrSliceFull, err; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := "Hello Wo", string(b); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestReaderWithRouterDealer(t *testing.T) {
	dealerSock := NewSock(Dealer)
	defer dealerSock.Destroy()

	routerSock := NewSock(Router)
	defer routerSock.Destroy()

	_, err := routerSock.Bind("inproc://test-read-router")
	if err != nil {
		t.Error(err)
	}

	err = dealerSock.Connect("inproc://test-read-router")
	if err != nil {
		t.Error(err)
	}

	err = dealerSock.SendFrame([]byte("Hello"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	b := make([]byte, 5)

	n, err := routerSock.Read(b)

	if want, got := 5, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := "Hello", string(b); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	err = dealerSock.SendFrame([]byte("Hello"), FlagMore)
	if err != nil {
		t.Error(err)
	}

	err = dealerSock.SendFrame([]byte(" World"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	b = make([]byte, 8)
	n, err = routerSock.Read(b)

	if want, got := ErrSliceFull, err; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := "Hello Wo", string(b); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	n, err = routerSock.Write([]byte("World"))
	if err != nil {
		t.Error(err)
	}

	if want, got := 5, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	frame, _, err := dealerSock.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "World", string(frame); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestReaderWithRouterDealerAsync(t *testing.T) {
	dealerSock1 := NewSock(Dealer)
	defer dealerSock1.Destroy()

	dealerSock2 := NewSock(Dealer)
	defer dealerSock2.Destroy()

	routerSock := NewSock(Router)
	defer routerSock.Destroy()

	_, err := routerSock.Bind("inproc://test-read-router-async")
	if err != nil {
		t.Error(err)
	}

	err = dealerSock1.Connect("inproc://test-read-router-async")
	if err != nil {
		t.Error(err)
	}

	err = dealerSock1.SendFrame([]byte("Hello From Client 1!"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	err = dealerSock2.Connect("inproc://test-read-router-async")
	if err != nil {
		t.Error(err)
	}

	err = dealerSock2.SendFrame([]byte("Hello From Client 2!"), FlagNone)
	if err != nil {
		t.Error(err)
	}

	msg := make([]byte, 255)

	n, err := routerSock.Read(msg)
	if want, got := 20, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	client1ID := routerSock.GetLastClientID()

	if want, got := 20, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	if want, got := "Hello From Client 1!", string(msg[:n]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	n, err = routerSock.Read(msg)
	if want, got := 20, n; want != got {
		t.Errorf("want '%d', got '%d'", want, got)
	}

	client2ID := routerSock.GetLastClientID()

	if want, got := "Hello From Client 2!", string(msg[:n]); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	routerSock.SetLastClientID(client1ID)
	n, err = routerSock.Write([]byte("Hello Client 1!"))

	if err != nil {
		t.Error(err)
	}

	frame, _, err := dealerSock1.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello Client 1!", string(frame); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	routerSock.SetLastClientID(client2ID)
	n, err = routerSock.Write([]byte("Hello Client 2!"))

	if err != nil {
		t.Error(err)
	}

	frame, _, err = dealerSock2.RecvFrame()
	if err != nil {
		t.Error(err)
	}

	if want, got := "Hello Client 2!", string(frame); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

type encodeMessage struct {
	Foo string
	Bar []byte
	Bat int
}

func TestPushPullEncodeDecode(t *testing.T) {
	push, err := NewPush("inproc://pushpullencode")
	if err != nil {
		t.Error(err)
	}
	defer push.Destroy()

	pull, err := NewPull("inproc://pushpullencode")
	if err != nil {
		t.Error(err)
	}
	defer pull.Destroy()

	encoder := gob.NewEncoder(push)
	decoder := gob.NewDecoder(pull)

	sent := encodeMessage{
		Foo: "the answer",
		Bar: []byte("is"),
		Bat: 42,
	}

	err = encoder.Encode(sent)
	if err != nil {
		t.Error(err)
	}

	var received encodeMessage
	err = decoder.Decode(&received)
	if err != nil {
		t.Error(err)
	}

	if want, got := received.Foo, sent.Foo; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := string(received.Bar), string(sent.Bar); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := received.Bat, sent.Bat; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if received.Bat != sent.Bat {
		t.Errorf("expected '%d', got '%d'", sent.Bat, received.Bat)
	}
}

func TestDealerRouterEncodeDecode(t *testing.T) {
	router, err := NewRouter("inproc://dealerrouterencode")
	if err != nil {
		t.Error(err)
	}
	defer router.Destroy()

	dealer, err := NewDealer("inproc://dealerrouterencode")
	if err != nil {
		t.Error(err)
	}
	defer dealer.Destroy()

	rencoder := gob.NewEncoder(router)
	rdecoder := gob.NewDecoder(router)

	dencoder := gob.NewEncoder(dealer)
	ddecoder := gob.NewDecoder(dealer)

	question := encodeMessage{
		Foo: "what is",
		Bar: []byte("the answer"),
		Bat: 0,
	}

	err = dencoder.Encode(question)
	if err != nil {
		t.Error(err)
	}

	var received encodeMessage
	err = rdecoder.Decode(&received)
	if err != nil {
		t.Error(err)
	}

	if want, got := received.Foo, question.Foo; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := string(received.Bar), string(question.Bar); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := received.Bat, question.Bat; want != got {
		t.Errorf("expected '%d', got '%d'", want, got)
	}

	sent := encodeMessage{
		Foo: "the answer",
		Bar: []byte("is"),
		Bat: 42,
	}

	err = rencoder.Encode(sent)
	if err != nil {
		t.Error(err)
	}

	var answer encodeMessage
	err = ddecoder.Decode(&answer)
	if err != nil {
		t.Error(err)
	}

	if want, got := answer.Foo, sent.Foo; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := string(answer.Bar), string(sent.Bar); want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}

	if want, got := answer.Bat, sent.Bat; want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func ExampleSock_output() {
	// create dealer socket
	dealer, err := NewDealer("inproc://example")
	if err != nil {
		panic(err)
	}
	defer dealer.Destroy()

	// create router socket
	router, err := NewRouter("inproc://example")
	if err != nil {
		panic(err)
	}
	defer router.Destroy()

	// send hello message
	dealer.SendFrame([]byte("Hello"), FlagNone)

	// receive hello message
	request, err := router.RecvMessage()
	if err != nil {
		panic(err)
	}

	// first frame is identify of client - let's append 'World'
	// to the message and route it back.
	request = append(request, []byte("World"))

	// send reply
	err = router.SendMessage(request)
	if err != nil {
		panic(err)
	}

	// receive reply
	reply, err := dealer.RecvMessage()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s %s", string(reply[0]), string(reply[1]))
	// Output: Hello World
}

func benchmarkSockSendFrame(size int, b *testing.B) {
	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind(fmt.Sprintf("inproc://benchSockSendFrame%d", size))
	if err != nil {
		panic(err)
	}

	go func() {
		pushSock := NewSock(Push)
		defer pushSock.Destroy()
		err := pushSock.Connect(fmt.Sprintf("inproc://benchSockSendFrame%d", size))
		if err != nil {
			panic(err)
		}

		payload := make([]byte, size)
		for i := 0; i < b.N; i++ {
			err = pushSock.SendFrame(payload, FlagNone)
			if err != nil {
				panic(err)
			}
		}
	}()

	for i := 0; i < b.N; i++ {
		msg, _, err := pullSock.RecvFrame()
		if err != nil {
			panic(err)
		}
		if len(msg) != size {
			panic("msg too small")
		}

		b.SetBytes(int64(size))
	}
}

func BenchmarkSockSendFrame1k(b *testing.B)  { benchmarkSockSendFrame(1024, b) }
func BenchmarkSockSendFrame4k(b *testing.B)  { benchmarkSockSendFrame(4096, b) }
func BenchmarkSockSendFrame16k(b *testing.B) { benchmarkSockSendFrame(16384, b) }

func benchmarkSockReadWriter(size int, b *testing.B) {
	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	_, err := pullSock.Bind(fmt.Sprintf("inproc://benchSockReadWriter%d", size))
	if err != nil {
		panic(err)
	}

	go func() {
		pushSock := NewSock(Push)
		defer pushSock.Destroy()
		err := pushSock.Connect(fmt.Sprintf("inproc://benchSockReadWriter%d", size))
		if err != nil {
			panic(err)
		}

		payload := make([]byte, size)
		for i := 0; i < b.N; i++ {
			_, err = pushSock.Write(payload)
			if err != nil {
				panic(err)
			}
		}
	}()

	payload := make([]byte, size)
	for i := 0; i < b.N; i++ {
		n, err := pullSock.Read(payload)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n != size {
			panic("msg too small")
		}
		b.SetBytes(int64(size))
	}
}

func BenchmarkSockReadWriter1k(b *testing.B)  { benchmarkSockReadWriter(1024, b) }
func BenchmarkSockReadWriter4k(b *testing.B)  { benchmarkSockReadWriter(4096, b) }
func BenchmarkSockReadWriter16k(b *testing.B) { benchmarkSockReadWriter(16384, b) }

func BenchmarkEncodeDecode(b *testing.B) {
	pullSock := NewSock(Pull)
	defer pullSock.Destroy()

	decoder := gob.NewDecoder(pullSock)

	_, err := pullSock.Bind("inproc://benchSockEncodeDecode")
	if err != nil {
		panic(err)
	}

	go func() {
		pushSock := NewSock(Push)
		defer pushSock.Destroy()
		err := pushSock.Connect("inproc://benchSockEncodeDecode")
		if err != nil {
			panic(err)
		}

		encoder := gob.NewEncoder(pushSock)

		sent := encodeMessage{
			Foo: "the answer",
			Bar: make([]byte, 1024),
			Bat: 42,
		}

		for i := 0; i < b.N; i++ {
			err := encoder.Encode(sent)
			if err != nil {
				panic(err)
			}
		}
	}()

	var received encodeMessage
	for i := 0; i < b.N; i++ {
		err := decoder.Decode(&received)
		if err != nil {
			panic(err)
		}
	}
}
