package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"

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
}

type Client interface {
	SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error
	SendImageMessage(ctx context.Context, recipientNumber string, imageFileName string, captionImageMessage string) error
}

func NewWhatsMeowClient(waCfg Config) (Client, error) {
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

		return &whatsMeow{
			Client: client,
		}, nil
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

type whatsMeow struct {
	Client *whatsmeow.Client
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

	imageMsg := &waProto.ImageMessage{
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
