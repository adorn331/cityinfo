package kafkautil

import (
	"context"
	"github.com/segmentio/kafka-go"
)

//const (
//	kafkaBroker = "localhost:9092"
//	topic = "test4"
//	//topic = "city-province"
//)

//var w *kafka.Writer
//var r *kafka.Reader
//
//func init() {
//	w = kafka.NewWriter(kafka.WriterConfig{
//		Brokers: []string{kafkaBroker},
//		Topic:   topic,
//		Balancer: &kafka.LeastBytes{},
//	})
//
//	r = kafka.NewReader(kafka.ReaderConfig{
//		Brokers:   []string{kafkaBroker},
//		Topic:     topic,
//		Partition: 0,
//		MinBytes:  10e3, // 10KB
//		MaxBytes:  10e6, // 10MB
//	})
//}

func ReadMsg(r *kafka.Reader, offset int64) (string, string, error){
	r.SetOffset(offset)

	m, err := r.ReadMessage(context.Background())
	if err != nil {
		return "", "", err
	}

	return string(m.Key), string(m.Value), nil
}

func SendMsg(w *kafka.Writer, key string, val string) error{
	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(val),
		},
	)

	return err
}

func BatchSendMsg(w *kafka.Writer, m map[string]string) error{
	var kafkaMsgs []kafka.Message

	for k, v := range m {
		kafkaMsgs = append(kafkaMsgs, kafka.Message{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	err := w.WriteMessages(context.Background(),
		kafkaMsgs...
	)

	return err
}