package main

import (
	"backendSenior/domain/model"
	"bytes"
	"common/rmq"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"log"
	"proxySenior/domain/encryption"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type Timer struct {
	start time.Time
	last  time.Time
}

func (t *Timer) Reset() {
	t.start = time.Now()
	t.last = t.start
}

func (t *Timer) Lap(msg string) {
	now := time.Now()
	fmt.Printf("%s: %d ns (%d ns since start)\n", msg, now.Sub(t.last).Nanoseconds(), now.Sub(t.start).Nanoseconds())
}

func main() {

	rabbit := rmq.New("amqp://guest:guest@localhost:5672/")
	if err := rabbit.Connect(); err != nil {
		log.Fatalf("can't connect to rabbitmq %s\n", err)
	}
	for _, q := range []string{"upload_task", "upload_result"} {
		if err := rabbit.EnsureQueue(q); err != nil {
			log.Fatalf("error ensuring queue %s: %s\n", q, err)
		}
	}

	// this just block forever
	forever := make(chan struct{})

	msgs, err := rabbit.Consume("upload_task")
	if err != nil {
		log.Fatalf("error consuming queue %s: %s\n", "upload_task", err)
	}

	fmt.Println("successfully connected to rabbitmq")
	clnt := resty.New()
	t := Timer{}

	// TODO: multithread
	go func() {
		for d := range msgs {
			fmt.Printf("recv message %s\n", d.Body)
			var task model.UploadFileTask
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				log.Printf("error parsing task: %s\nmessage was: %s\n", err, d.Body)
				continue
			}

			t.Reset()
			fileData, err := ioutil.ReadFile(task.FilePath)
			if err != nil {
				log.Printf("error reading file: %s\n", err)
				d.Nack(false, true)
				continue
			}
			t.Lap("read file")

			if task.EncryptKey == nil {
				log.Printf("error encrypting file: key is nil\n")
				d.Nack(false, false)
				continue
			}
			fileData, err = encryption.AESEncrypt(fileData, task.EncryptKey)
			if err != nil {
				log.Printf("error encrypting file: %s\n", err)
				d.Nack(false, true)
				continue
			}
			t.Lap("encrypt")

			fmt.Println("posting to url", task.URL)
			res, err := clnt.R().
				SetFormData(task.UploadPostForm).
				SetFileReader("file", "foo.txt", bytes.NewReader(fileData)).
				SetHeader("Content-Type", "multipart/form-data").
				Post(task.URL)
			t.Lap("post url")

			if err != nil {
				fmt.Println("resty req error:", err)
				d.Nack(false, true)
				continue
			}
			if !res.IsSuccess() {
				fmt.Printf("resty status: %d\nbody: %s\n", res.StatusCode(), res.String())
				d.Nack(false, true)
				continue
			}
			data, _ := json.Marshal(map[string]interface{}{
				"taskId": task.TaskID,
			})
			t.Lap("marshal message to send")

			if err := rabbit.Publish("upload_result", data); err != nil {
				fmt.Printf("publish result error %s\n", err)
				continue
			}
			fmt.Println("task", task.TaskID, "success!")
			d.Ack(false)

		}
	}()

	<-forever

}
