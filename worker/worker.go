package main

import (
	"backendSenior/domain/model"
	"bytes"
	"common/rmq"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
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

type Worker struct {
	clnt   *resty.Client
	rabbit *rmq.RMQClient
}

func (w *Worker) handleFileTask(d amqp.Delivery, task model.UploadFileTask, t *Timer) { // timer is for bug only
	t.Reset()
	fileData, err := ioutil.ReadFile(task.FilePath)
	if err != nil {
		log.Printf("error reading file: %s\n", err)
		d.Nack(false, true)
		return
	}
	t.Lap("read file")

	if task.EncryptKey == nil {
		log.Printf("error encrypting file: key is nil\n")
		d.Nack(false, false)
		return
	}
	fileData, err = encryption.AESEncrypt(fileData, task.EncryptKey)
	if err != nil {
		log.Printf("error encrypting file: %s\n", err)
		d.Nack(false, true)
		return
	}
	t.Lap("encrypt")

	fmt.Println("posting to url", task.URL)
	res, err := w.clnt.R().
		SetFormData(task.UploadPostForm).
		SetFileReader("file", "dontcare", bytes.NewReader(fileData)).
		SetHeader("Content-Type", "multipart/form-data").
		Post(task.URL)
	t.Lap("post url")

	if err != nil {
		fmt.Println("resty req error:", err)
		d.Nack(false, true)
		return
	}
	if !res.IsSuccess() {
		fmt.Printf("resty status: %d\nbody: %s\n", res.StatusCode(), res.String())
		d.Nack(false, true)
		return
	}
	data, _ := json.Marshal(map[string]interface{}{
		"taskId": task.TaskID,
	})
	t.Lap("marshal message to send")

	if err := w.rabbit.Publish("upload_result", data); err != nil {
		fmt.Printf("publish result error %s\n", err)
		return
	}
	fmt.Println("task", task.TaskID, "success!")
	d.Ack(false)
}

var mapFormat = map[string]imaging.Format{
	"jpg":  imaging.JPEG,
	"jpeg": imaging.JPEG,
	"png":  imaging.PNG,
	"bmp":  imaging.BMP,
	"gif":  imaging.GIF,
	"tif":  imaging.TIFF,
	"tiff": imaging.TIFF,
}

// handleImageTask: read image, generate thumbnail, then encrypt both of them and upload
func (w *Worker) handleImageTask(d amqp.Delivery, task model.UploadFileTask, t *Timer) {

	// read file
	fileData, err := ioutil.ReadFile(task.FilePath)
	if err != nil {
		log.Printf("error opening image: %s\n", err)
		d.Nack(false, false)
		return
	}
	r := bytes.NewReader(fileData)

	// determine type
	_, format, err := image.DecodeConfig(r)
	if err != nil {
		log.Printf("image: error determining image type: %s, is it corrupt?", err)
		d.Nack(false, false)
		return
	}

	// decode & process
	r.Seek(0, io.SeekStart) // reset seek
	src, err := imaging.Decode(r)
	if err != nil {
		log.Printf("imaging: error decoing image: %s, is it corrupt?", err)
		d.Nack(false, false)
		return
	}

	size := src.Bounds().Size()
	width := size.X
	height := size.Y
	var img *image.NRGBA
	if width > height {
		img = imaging.Resize(src, 400, 0, imaging.Lanczos)
	} else {
		img = imaging.Resize(src, 0, 400, imaging.Lanczos)
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, mapFormat[format])
	if err != nil {
		log.Printf("error encoding to %s:%s\n", format, err)
		d.Nack(false, false)
		return
	}

	// upload orig
	r.Seek(0, io.SeekStart)
	orig, _ := ioutil.ReadAll(r)
	orig, err = encryption.AESEncrypt(orig, task.EncryptKey)
	if err != nil {
		log.Printf("error encrypting to %s\nkey was: %s\n", err, task.EncryptKey)
		d.Nack(false, false)
		return
	}

	res, err := w.clnt.R().
		SetFormData(task.UploadPostForm).
		SetFileReader("file", "dontcare", bytes.NewReader(orig)).
		SetHeader("Content-Type", "multipart/form-data").
		Post(task.URL)
	t.Lap("post url")

	if err != nil {
		fmt.Println("resty req error:", err)
		d.Nack(false, true)
		return
	}
	if !res.IsSuccess() {
		fmt.Printf("resty status: %d\nbody: %s\n", res.StatusCode(), res.String())
		d.Nack(false, true)
		return
	}

	// upload thumb
	thumb := buf.Bytes()
	thumb, err = encryption.AESEncrypt(thumb, task.EncryptKey)
	if err != nil {
		log.Printf("error encrypting to %s\nkey was: %s\n", err, task.EncryptKey)
		d.Nack(false, false)
		return
	}

	res, err = w.clnt.R().
		SetFormData(task.UploadPostForm2).
		SetFileReader("file", "dontcare", bytes.NewReader(thumb)).
		SetHeader("Content-Type", "multipart/form-data").
		Post(task.URL)

	if err != nil {
		fmt.Println("resty req error:", err)
		d.Nack(false, true)
		return
	}
	if !res.IsSuccess() {
		fmt.Printf("resty status: %d\nbody: %s\n", res.StatusCode(), res.String())
		d.Nack(false, true)
		return
	}

	// report result
	data, _ := json.Marshal(map[string]interface{}{
		"taskId": task.TaskID,
	})
	t.Lap("marshal message to send")

	if err := w.rabbit.Publish("upload_result", data); err != nil {
		fmt.Printf("publish result error %s\n", err)
		return
	}
	fmt.Println("task image", task.TaskID, "success!")
	d.Ack(false)

	//src, err := imaging.Open(task.FilePath)
	//if err != nil {
	//	log.Printf("error reading file: %s\n", err)
	//	d.Nack(false, true)
	//	return
	//}
	//
	//
	//
	//t.Lap("read file")
	//
	//if task.EncryptKey == nil {
	//	log.Printf("error encrypting file: key is nil\n")
	//	d.Nack(false, false)
	//	return
	//}
	//fileData, err = encryption.AESEncrypt(fileData, task.EncryptKey)
	//if err != nil {
	//	log.Printf("error encrypting file: %s\n", err)
	//	d.Nack(false, true)
	//	return
	//}
	//t.Lap("encrypt")
	//
	//fmt.Println("posting to url", task.URL)
	//res, err := w.clnt.R().
	//	SetFormData(task.UploadPostForm).
	//	SetFileReader("file", "dontcare", bytes.NewReader(fileData)).
	//	SetHeader("Content-Type", "multipart/form-data").
	//	Post(task.URL)
	//t.Lap("post url")
	//
	//if err != nil {
	//	fmt.Println("resty req error:", err)
	//	d.Nack(false, true)
	//	return
	//}
	//if !res.IsSuccess() {
	//	fmt.Printf("resty status: %d\nbody: %s\n", res.StatusCode(), res.String())
	//	d.Nack(false, true)
	//	return
	//}
	//data, _ := json.Marshal(map[string]interface{}{
	//	"taskId": task.TaskID,
	//})
	//t.Lap("marshal message to send")
	//
	//if err := w.rabbit.Publish("upload_result", data); err != nil {
	//	fmt.Printf("publish result error %s\n", err)
	//	return
	//}
	//fmt.Println("task", task.TaskID, "success!")
	//d.Ack(false)
}

func defaultEnv(key string, defaultVal string) string {
	val, ok := os.LookupEnv(key)
	if ok {
		return val
	}
	return defaultVal
}

func requiredEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("ERROR: environment variable %s is not set", key))
	}
	return val
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("load env error: %s\n", err)
	}

	rabbit := rmq.New(requiredEnv("RABBITMQ_CONN_STRING"))
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

	w := Worker{
		clnt:   clnt,
		rabbit: rabbit,
	}

	t := new(Timer)

	// TODO: multithread
	go func() {
		for d := range msgs {
			var task model.UploadFileTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				d.Ack(false) // ack anyway, to discard malform message
				continue
			} else {
				if task.Type == model.Image {
					log.Println("type image")
					w.handleImageTask(d, task, t)
				} else {
					log.Println("type file")
					w.handleFileTask(d, task, t)
				}
			}
		}
	}()

	<-forever

}
