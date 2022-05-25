package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SNSTopics struct {
	NewImage string
}

type IncomingMessage struct {
	Type      string
	MessageId string
	TopicArn  string
	Message   string
	Timestamp string
}

type StatusMessage struct {
	ImageID string
	Status  string
}

type ConfirmationMessage struct {
	SubscribeURL string
}

func main() {
	log.Println("Initializing API server")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	var snsTopics SNSTopics
	err = json.Unmarshal([]byte(os.Getenv("COPILOT_SNS_TOPIC_ARNS")), &snsTopics)
	if err != nil {
		panic(err)
	}

	bucket := os.Getenv("RECEIPTUPLOADS_NAME")

	r := gin.Default()

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			panic(err)
		}
		id := GenerateID()
		log.Printf("New image received: %s", id)
		tmp := os.TempDir()
		uploadedFile := filepath.Join(tmp, fmt.Sprintf("%s.png", id))
		log.Printf("Saving image to %s", uploadedFile)
		c.SaveUploadedFile(file, uploadedFile)

		uploader := manager.NewUploader(s3.NewFromConfig(cfg))
		openedFile, err := os.Open(uploadedFile)
		if err != nil {
			log.Println("Failed to open the image", uploadedFile, err)
			return
		}
		newKey := fmt.Sprintf("uploads/%s/image.png", id)
		log.Printf("Uploading to S3: %s", newKey)
		s3Result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: &bucket,
			Key:    aws.String(newKey),
			Body:   openedFile,
		})
		if err != nil {
			log.Fatalln("Failed to upload", newKey, err)
		}

		log.Println("Successful upload")
		c.JSON(200, gin.H{
			"imageID": id,
		})

		statusChannels[id] = make(chan string, 3)

		log.Println("Sending SNS message to worker service")
		client := sns.NewFromConfig(cfg)
		input := &sns.PublishInput{
			Message:  &s3Result.Location,
			TopicArn: &snsTopics.NewImage,
		}

		snsResult, err := client.Publish(context.TODO(), input)
		if err != nil {
			fmt.Printf("Error publishing message: %v", err)
		}
		fmt.Printf("Message ID: %s", *snsResult.MessageId)

		log.Println("Cleaning up")
		err = os.Remove(uploadedFile)
		if err != nil {
			panic(err)
		}
	})

	r.GET("/imageStatus/:id/ws", func(c *gin.Context) {
		log.Printf("Websocket connection request for %s", c.Param("id"))
		// websocket request
		wsHandler(c.Writer, c.Request, c.Param("id"))
	})

	r.POST("/imageStatus", func(c *gin.Context) {
		// only SNS should be posting here
		switch c.Request.Header.Get("x-amz-sns-message-type") {
		case "SubscriptionConfirmation":
			var message ConfirmationMessage
			bytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Println(err)
			}
			err = json.Unmarshal(bytes, &message)
			if err != nil {
				log.Println(err)
			}
			log.Println(message.SubscribeURL)

		case "Notification":
			var incomingMessage IncomingMessage
			bytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Println(err)
			}
			err = json.Unmarshal(bytes, &incomingMessage)
			if err != nil {
				log.Println(err)
			}
			var statusMessage StatusMessage
			err = json.Unmarshal([]byte(incomingMessage.Message), &statusMessage)
			if err != nil {
				log.Println(err)
			}
			log.Printf("ImageID: %s, Status: %s", statusMessage.ImageID, statusMessage.Status)
			statusChannels[statusMessage.ImageID] <- statusMessage.Status
			if statusMessage.Status == "Accepted" {
				close(statusChannels[statusMessage.ImageID])
				delete(statusChannels, statusMessage.ImageID)
			}
		default:
			return
		}

	})

	if stage, _ := os.LookupEnv("STAGE"); stage == "prod" {
		r.Static("/static", "/var/www/html")
	}

	r.Run()
}

func GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	id := hex.EncodeToString(b)
	return id
}

var statusChannels = make(map[string](chan string))

func wsHandler(w http.ResponseWriter, r *http.Request, id string) {
	// use the request and response writer to create the websocket connection
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	// a separate thread can communicate with the browser via this connection
	go func(id string, ch chan string) {
		// wait for status messages on the right channel
		log.Printf("Waiting for status message for %s", id)
		for {
			status, more := <-ch
			log.Printf("Message received in thread for %s: %s", id, status)

			if !more {
				return
			}
			conn.WriteJSON(gin.H{
				"status": status,
			})
		}
	}(id, statusChannels[id])
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(req *http.Request) bool {
		return true
	},
}
