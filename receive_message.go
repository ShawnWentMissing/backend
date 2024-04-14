package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Message struct {
	ID      string `json:"id"`
	Bounced bool   `json:"bounced"`
	Area    Area   `json:"area"`
}

type TimeArea struct {
	Time float64
	Area Area
}

func pollMessagesSQS(svc *sqs.SQS, queueURL string, msgsCh chan<- *sqs.Message) {
	for {
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: aws.Int64(240),
			WaitTimeSeconds:     aws.Int64(1),
		})
		if err != nil {
			fmt.Println("Error receiving messages:", err)
			continue
		}

		for _, msg := range result.Messages {
			msgsCh <- msg
		}

		time.Sleep(1 * time.Second)
	}
}

func processMessagesSQS(svc *sqs.SQS, queueURL string, msgsCh <-chan *sqs.Message, storage *GameStorage) {
	for msg := range msgsCh {
		var decodedMsg Message
		err := json.Unmarshal([]byte(*msg.Body), &decodedMsg)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})

		if decodedMsg.Bounced {
			endRally, ok := storage.BallBounce(decodedMsg.ID, decodedMsg.Area)
			if !ok {
				fmt.Println("Error updating game")
				return
			}

			if endRally {
				announceMessage()
			}
		}
	}
}

func pollMessagesSQS(svc *sqs.SQS, queueURL string, msgsCh chan<- *sqs.Message) {
	for {
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: aws.Int64(240),
			WaitTimeSeconds:     aws.Int64(1),
		})
		if err != nil {
			fmt.Println("Error receiving messages:", err)
			continue
		}

		for _, msg := range result.Messages {
			msgsCh <- msg
		}

		time.Sleep(1 * time.Second)
	}
}

func processMessages(msgsCh <-chan *Message, storage *GameStorage) {
	for msg := range msgsCh {
		if msg.Bounced {
			endRally, ok := storage.BallBounce(msg.ID, msg.Area)
			if !ok {
				fmt.Println("Error updating game")
				return
			}

			if endRally {
				announceMessage()
			}
		}
	}
}

func pollMessages(timeAreas []TimeArea, startTime time.Time, msgsCh chan<- TimeArea) {
	for _, ta := range timeAreas {
		duration := time.Duration(ta.Time * float64(time.Second))
		time.Sleep(duration)
		msgsCh <- TimeArea{duration.Seconds(), ta.Area}
	}
}

func announceMessage(player1score, player2score int, change bool) {

}
