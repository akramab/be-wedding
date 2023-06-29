package whatsapp

import (
	"context"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Config struct {
	EnableNotification bool `toml:"enable_notification"`
}

type Client interface {
	SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error
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

type whatsMeow struct {
	Client *whatsmeow.Client
}

func (wm *whatsMeow) SendMessage(ctx context.Context, recipientNumber string, message *waProto.Message) error {
	const initialMessage = `*[OTOMATISASI PENILAIAN SPBE]*
` + "```" + `Terima kasih telah menggunakan Aplikasi Otomatisasi Penilaian SPBE. Hasil penilaian anda akan keluar dalam beberapa saat lagi.` + "```"
	// initialTemplateMessage := &waProto.Message{
	// 	TemplateMessage: &waProto.TemplateMessage{
	// 		HydratedTemplate: &waProto.TemplateMessage_HydratedFourRowTemplate{
	// 			Title: &waProto.TemplateMessage_HydratedFourRowTemplate_HydratedTitleText{
	// 				HydratedTitleText: "The Title",
	// 			},
	// 			TemplateId:          proto.String("template-id"),
	// 			HydratedContentText: proto.String("The Content"),
	// 			HydratedFooterText:  proto.String("The Footer"),
	// 			HydratedButtons: []*waProto.HydratedTemplateButton{

	// 				// This for URL button
	// 				{
	// 					Index: proto.Uint32(1),
	// 					HydratedButton: &waProto.HydratedTemplateButton_UrlButton{
	// 						UrlButton: &waProto.HydratedTemplateButton_HydratedURLButton{
	// 							DisplayText: proto.String("The Link"),
	// 							Url:         proto.String("https://fb.me/this"),
	// 						},
	// 					},
	// 				},

	// 				// This for call button
	// 				{
	// 					Index: proto.Uint32(2),
	// 					HydratedButton: &waProto.HydratedTemplateButton_CallButton{
	// 						CallButton: &waProto.HydratedTemplateButton_HydratedCallButton{
	// 							DisplayText: proto.String("Call us"),
	// 							PhoneNumber: proto.String("1234567890"),
	// 						},
	// 					},
	// 				},

	// 				// This is just a quick reply
	// 				{
	// 					Index: proto.Uint32(3),
	// 					HydratedButton: &waProto.HydratedTemplateButton_QuickReplyButton{
	// 						QuickReplyButton: &waProto.HydratedTemplateButton_HydratedQuickReplyButton{
	// 							DisplayText: proto.String("Quick reply"),
	// 							Id:          proto.String("quick-id"),
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	recipient := types.NewJID(recipientNumber, "s.whatsapp.net")
	_, err := wm.Client.SendMessage(context.Background(), recipient, message)

	return err
}
