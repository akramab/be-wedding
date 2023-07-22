package whatsapp

import (
	"be-wedding/internal/store"
	"be-wedding/pkg/redis"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Config struct {
	EnableNotification bool `toml:"enable_notification"`
	BroadcastMode      bool `toml:"broadcast_mode"`
	DebugMode          bool `toml:"debug_mode"`
	AutoReply          bool `toml:"auto_reply"`
	RSVPCheck          bool `toml:"rsvp_check"`
}

type Client interface {
	SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error
	SendImageMessage(ctx context.Context, recipientNumber string, imageFileName string, captionImageMessage string) error
	SendVideoMessage(ctx context.Context, recipientNumber string, videoFileName string, captionVideoMessage string) error
}

func NewWhatsMeowClient(waCfg Config, userStore store.User, invitationStore store.Invitation, redisCache redis.Client) (Client, error) {
	if waCfg.EnableNotification {
		dbLog := waLog.Stdout("Database", "DEBUG", true)
		// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
		container, err := sqlstore.New("sqlite3", "file:whatsappcred.db?_foreign_keys=on", dbLog)
		if err != nil {
			log.Println(err.Error())
		}
		// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
		deviceStore, err := container.GetFirstDevice()
		if err != nil {
			log.Println(err.Error())
		}
		clientLog := waLog.Stdout("Client", "DEBUG", true)
		client := whatsmeow.NewClient(deviceStore, clientLog)

		if client.Store.ID == nil {
			// No ID stored, new login
			qrChan, _ := client.GetQRChannel(context.Background())
			err = client.Connect()
			if err != nil {
				panic(err)
			}
			for evt := range qrChan {
				if evt.Event == "code" {
					// Render the QR code here
					// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
					fmt.Println("QR code:", evt.Code)
				} else {
					fmt.Println("Login event:", evt.Event)
				}
			}
		} else {
			// Already logged in, just connect
			err = client.Connect()
			if err != nil {
				log.Printf(err.Error())
			}
		}

		wm := &whatsMeow{
			Client:          client,
			userStore:       userStore,
			invitationStore: invitationStore,
			redisCache:      redisCache,
			Config:          waCfg,
		}

		if wm.Config.AutoReply {
			wm.Client.AddEventHandler(wm.eventHandler)
		}

		return wm, nil
		// return &whatsMeow{
		// 	Client: client,
		// }, nil
	}

	return &waClientMock{}, nil
}

type waClientMock struct{}

func (mock *waClientMock) SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error {
	return nil
}

func (mock *waClientMock) SendImageMessage(ctx context.Context, recipientNumber string, imageFileName string, captionImageMessage string) error {
	return nil
}

func (mock *waClientMock) SendVideoMessage(ctx context.Context, recipientNumber string, videoFileName string, captionVideoMessage string) error {
	return nil
}

type whatsMeow struct {
	Client          *whatsmeow.Client
	userStore       store.User
	invitationStore store.Invitation
	redisCache      redis.Client
	Config          Config
}

type whatsMeowEventHandler struct {
	Client          *whatsmeow.Client
	userStore       store.User
	invitationStore store.Invitation
}

func (wm *whatsMeow) SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error {

	// Remove '+' sign from the target recipient number
	recipientNumberCleaned := recipientNumber[1:]

	recipient := types.NewJID(recipientNumberCleaned, "s.whatsapp.net")
	_, err := wm.Client.SendMessage(context.Background(), recipient, message)

	return err
}

func (wm *whatsMeow) SendImageMessage(ctx context.Context, recipientNumber string, imageFileName string, captionImageMessage string) error {
	imageBytes, err := os.ReadFile(fmt.Sprintf("./static/qr-codes/%s", imageFileName)) // still need to change the mimetype
	if err != nil {
		fmt.Println("ERROR read image file")
		fmt.Println(err.Error())
		return err
	}

	resp, err := wm.Client.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
	if err != nil {
		fmt.Println("ERROR UPLOAD IMAGE")
		return err
	}

	var imageMsg *waProto.ImageMessage
	if !wm.Config.BroadcastMode {
		imageMsg = &waProto.ImageMessage{
			Caption:  proto.String(captionImageMessage),
			Mimetype: proto.String("image/png"), // replace this with the actual mime type
			// you can also optionally add other fields like ContextInfo and JpegThumbnail here

			Url:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSha256: resp.FileEncSHA256,
			FileSha256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
		}
	} else {
		thumbnailBytes, err := os.ReadFile(fmt.Sprintf("./static/qr-codes/%s", strings.Split(imageFileName, ".")[0]+"-thumbnail.png"))
		respThumbnail, err := wm.Client.Upload(context.Background(), thumbnailBytes, whatsmeow.MediaImage)
		if err != nil {
			fmt.Println("ERROR UPLOAD IMAGE")
			return err
		}

		imageMsg = &waProto.ImageMessage{
			Caption:  proto.String(captionImageMessage),
			Mimetype: proto.String("image/png"), // replace this with the actual mime type
			// you can also optionally add other fields like ContextInfo and JpegThumbnail here
			ThumbnailDirectPath: &respThumbnail.DirectPath,
			ThumbnailSha256:     respThumbnail.FileSHA256,
			ThumbnailEncSha256:  respThumbnail.FileEncSHA256,
			JpegThumbnail:       thumbnailBytes,

			Url:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSha256: resp.FileEncSHA256,
			FileSha256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
		}
	}

	// Remove '+' sign from the target recipient number
	recipientNumberCleaned := recipientNumber[1:]

	recipient := types.NewJID(recipientNumberCleaned, "s.whatsapp.net")

	_, err = wm.Client.SendMessage(context.Background(), recipient, &waProto.Message{
		ImageMessage: imageMsg,
	})

	if err != nil {
		fmt.Println("SEND IMAGE ERROR")
		fmt.Println(err.Error())
		return err
	}

	return err
}

func (wm *whatsMeow) SendVideoMessage(ctx context.Context, recipientNumber string, videoFileName string, captionVideoMessage string) error {
	imageBytes, err := os.ReadFile(fmt.Sprintf("./static/qr-codes/%s", videoFileName)) // still need to change the mimetype
	if err != nil {
		fmt.Println("ERROR read video file")
		fmt.Println(err.Error())
		return err
	}

	resp, err := wm.Client.Upload(context.Background(), imageBytes, whatsmeow.MediaVideo)
	if err != nil {
		fmt.Println("ERROR UPLOAD VIDEO")
		return err
	}

	videoMsg := &waProto.VideoMessage{
		Caption:  proto.String(captionVideoMessage),
		Mimetype: proto.String("video/mp4"), // replace this with the actual mime type
		// you can also optionally add other fields like ContextInfo and JpegThumbnail here

		Url:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSha256: resp.FileEncSHA256,
		FileSha256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
	}

	// Remove '+' sign from the target recipient number
	recipientNumberCleaned := recipientNumber[1:]

	recipient := types.NewJID(recipientNumberCleaned, "s.whatsapp.net")

	_, err = wm.Client.SendMessage(context.Background(), recipient, &waProto.Message{
		VideoMessage: videoMsg,
	})

	if err != nil {
		fmt.Println("SEND Video ERROR")
		fmt.Println(err.Error())
		return err
	}

	return err
}
