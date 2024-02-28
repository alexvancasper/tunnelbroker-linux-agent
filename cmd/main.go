package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alexvancasper/TunnelBroker/agent/internal/broker"
	"github.com/alexvancasper/TunnelBroker/agent/internal/doer"
	formatter "github.com/fabienm/go-logrus-formatters"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

func main() {

	//Initialize Logging connections
	var MyLogger = logrus.New()

	gelfFmt := formatter.NewGelf("agent")
	MyLogger.SetFormatter(gelfFmt)
	MyLogger.SetOutput(os.Stdout)
	loglevel, err := logrus.ParseLevel("debug")
	if err != nil {
		MyLogger.WithField("function", "main").Fatalf("error %v", err)
	}
	MyLogger.SetLevel(loglevel)

	var wg sync.WaitGroup

	m, err := broker.MsgBrokerInit(os.Getenv("BROKER_CONN"), os.Getenv("QUEUENAME"))
	if err != nil {
		MyLogger.Fatalf("Message broker error init: %s", err)
	}
	defer m.Close()

	MyLogger.Debug("Message broker connected")

	msgs, err := m.RegisterConsumer()
	if err != nil {
		MyLogger.Fatalf("RegisterConsumer error: %s", err)
	}

	MyLogger.Info(" [*] Waiting for messages. To exit press CTRL+C")

	ctx, ctxCancel := context.WithCancel(context.Background())
	wg.Add(1)
	closeChan := make(chan struct{})
	go Listener(ctx, &wg, MyLogger, msgs, closeChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		MyLogger.Error("Interrupt signal recevied")
	case <-closeChan:
		MyLogger.Error("Shutdown due to communication issue")
		return
	}
	ctxCancel()
	wg.Wait()
	MyLogger.Info("Graceful shutdown")
}

func Listener(ctx context.Context, wg *sync.WaitGroup, log *logrus.Logger, msgs <-chan amqp091.Delivery, closeChan chan<- struct{}) {
	defer wg.Done()
	defer close(closeChan)
	l := log.WithField("function", "Listener")
	h := doer.Handler{
		Log: log,
	}
	for {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			l.Debug("Context closed")
			return
		case msg, ok := <-msgs:
			if !ok {
				closeChan <- struct{}{}
				l.Error("Listener closed due to recv channel is closed")
				return
			}
			l.WithField("message type", msg.Type).WithField("body", msg.Body).Info("Received message from queue")
			switch msg.Type {
			case string(broker.ADD):
				wg.Add(1)
				go h.AddTunnel(wg, msg.Body)
			case string(broker.DELETE):
				wg.Add(1)
				go h.DeleteTunnel(wg, msg.Body)
			}
		}
	}
}
