package main

import (
	"log"
	"os"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func handleAnimationCommon(b *tb.Bot, m *tb.Message) (*tb.Message, string, string, error) {
	progressMessage, err := sendMessageWithRepeater(b, m.Chat, "Downloading...")
	if err != nil {
		log.Println(err)
		return nil, "", "", err
	}
	filename, err := justGetTheFile(b, m)
	if err != nil {
		return nil, "", "", err
	}
	animationOutput := filename + ".mp4"
	progressChan := make(chan string, 3)
	go distortVideo(filename, animationOutput, progressChan)
	for report := range progressChan {
		progressMessage, _ = b.Edit(progressMessage, report, &tb.SendOptions{ParseMode: tb.ModeHTML})
	}
	_, err = os.Stat(animationOutput)
	return progressMessage, filename, animationOutput, err
}

func handleVideoCommon(b *tb.Bot, m *tb.Message) (string, error) {
	progressMessage, filename, animationOutput, err := handleAnimationCommon(b, m)
	if err != nil {
		if progressMessage != nil && progressMessage.Text != TooLong {
			doneMessageWithRepeater(b, progressMessage, true)
		}
		return "", err
	}
	defer os.Remove(filename)
	defer os.Remove(animationOutput)
	soundOutput := filename + ".ogg"
	err = distortSound(filename, soundOutput)
	if err != nil {
		soundOutput = ""
	} else {
		defer os.Remove(soundOutput)
	}
	output := filename + "Final.mp4"
	b.Edit(progressMessage, "Muxing frames with sound back together...")
	err = collectAnimationAndSound(animationOutput, soundOutput, output)
	doneMessageWithRepeater(b, progressMessage, err != nil)
	return output, err
}

func doneMessageWithRepeater(b *tb.Bot, m *tb.Message, failed bool) {
	done := "Done!"
	if failed {
		done = Failed
	}
	_, err := b.Edit(m, done)
	for err != nil {
		var timeout int
		timeout, err = extractPossibleTimeout(err)
		if err != nil {
			return
		}
		log.Printf("sleeping for %d before sending Done\n", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		_, err = b.Edit(m, done)
	}
}

func sendMessageWithRepeater(b *tb.Bot, chat *tb.Chat, toSend interface{}) (*tb.Message, error) {
	m, err := b.Send(chat, toSend)
	for err != nil {
		var timeout int
		timeout, err = extractPossibleTimeout(err)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Printf("sleeping for %d\n", timeout)
		time.Sleep(time.Duration(timeout) * time.Second)
		m, err = b.Send(chat, toSend)
		if err != nil {
			log.Println(err)
		}
	}

	return m, nil
}
