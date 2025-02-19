package main

import (
	"alac-bot/wrapper"
	"fmt"
	"github.com/kamva/mgm/v3"
	"os"
	"regexp"
	"time"

	tg "gopkg.in/telebot.v4"
)

func DownloadSong(ctx tg.Context) error {
	b := Bot()

	msg, err := b.Send(ctx.Chat(), "Getting information...", &tg.SendOptions{ReplyTo: ctx.Message()})
	if err != nil {
		return err
	}

	downloadFolder := "downloads"
	meta, file, err := wrapper.App(ctx.Args()[0], downloadFolder, b, ctx, msg)
	if err != nil {
		fmt.Println("Error in wrapper", err)
		_ = ctx.Reply(fmt.Sprintf("%v", err))
		_ = b.Delete(msg)
		return err
	}

	msg, err = b.Edit(msg, "Uploading "+meta.Attributes.Name)
	if err != nil {
		return err
	}

	song := &tg.Audio{
		File:      tg.FromDisk(file.Name()),
		Duration:  meta.Attributes.DurationInMillis / 1000,
		Title:     meta.Attributes.Name,
		Performer: meta.Attributes.ArtistName,
		FileName:  meta.Attributes.Name,
	}
	err = ctx.Reply(song)
	if err != nil {
		fmt.Println("Failed to upload song.", err)
		_, _ = b.Send(ctx.Sender(), "Failed to upload song")
		return err
	}
	_ = b.Delete(msg)

	newSong := CreateFile([]string{song.FileID}, "song", meta.ID)

	err = mgm.Coll(newSong).Create(newSong)
	if err != nil {
		fmt.Println("Failed to upload song.", err)
	}

	err = os.Remove(file.Name())
	if err != nil {
		fmt.Println("Error removing file", err)
	}

	return err
}

func validateSongUrl(url string) bool {
	// Regular expression for album URLs with an optional 'i' query parameter and other query params
	albumURLRegex := regexp.MustCompile(`^https://music\.apple\.com/([a-z]{2})/album/[a-zA-Z0-9\-]+/([0-9]+)(\?i=([0-9]+).*)?$`)
	// Regular expression for song URLs with optional query params
	songURLRegex := regexp.MustCompile(`^https://music\.apple\.com/([a-z]{2})/song/[a-zA-Z0-9\-]+/([0-9]+)(\?.*)?$`)

	// Check if the URL matches either the album or song pattern
	if albumURLRegex.MatchString(url) || songURLRegex.MatchString(url) {
		return true
	}

	// If no match, return false
	return false
}

// func validateAlbumUrl(url string) bool {
// 	albumURLRegex := regexp.MustCompile(`^https://music\.apple\.com/([a-z]{2})/album/[a-zA-Z0-9\-]+/([0-9]+)(\?.*)?$`)

func sendTempMsg(ctx tg.Context, txt string, t time.Duration) error {
	b := Bot()
	msg, err := b.Send(ctx.Sender(), txt, &tg.SendOptions{ReplyTo: ctx.Message()})
	if err != nil {
		fmt.Println("Failed to send message", err)
		return err
	}
	if t > 0 {
		go func() error {
			time.Sleep(t * time.Second)
			if err := b.Delete(msg); err != nil {
				fmt.Println("Failed to delete message:", err)
			}
			return err
		}()
	}
	return err
}
