package whatsapp

import (
	"context"
	"fmt"
	"image"
	"io/fs"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	gozqrcode "github.com/makiuchi-d/gozxing/qrcode"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func (wm *whatsMeow) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if !v.Info.IsFromMe {
			fmt.Println("PESAN DITERIMA!", v.Message.GetConversation())
			userJID := v.Info.Sender.ToNonAD().String()

			fmt.Printf("USER JID: %s \n", userJID)
			invitationCompleteData, err := wm.invitationStore.FindOneCompleteDataByWANumber(context.Background(), "+"+strings.Split(userJID, "@")[0])
			if err != nil {
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(fmt.Sprintf("Maaf, nomor anda belum terdaftar. Silahkan registrasi melalui undangan yang telah dikirimkan")),
				})
				return
			}

			userMessage := v.Message.GetConversation()
			switch userMessage {
			case "1":
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(fmt.Sprintf("Nama yang terdaftar saat ini: %s; Judul video anda di server: %s", invitationCompleteData.User.Name, invitationCompleteData.User.QRImage)),
				})
				return
			case "23":
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String("Silakan kirimkan video anda"),
				})
				return
			// default:
			// 	wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
			// 		Conversation: proto.String("Nama anda telah berhasil disimpan. Ketik 1 untuk mengganti nama, 2 untuk mengirim video, dan 3 untuk melihat nama yang terdaftar saat ini"),
			// 	})
			// 	return
			}

			if v.Message.GetVideoMessage() != nil {
				fmt.Println("VIDEO EXISTS")
				videoMessage := v.Message.GetVideoMessage()
				videoData, err := wm.Client.Download(videoMessage)
				if err != nil {
					fmt.Println("ERROR DOWNLOAD VIDEO")
					return
				}
				videoMIMEType := videoMessage.GetMimetype()
				fmt.Printf("VIDEO MIMETYPE: %s \n", videoMIMEType)

				videoName := uuid.New().String()
				err = os.WriteFile(fmt.Sprintf("./static/videos/%s-video.%s", videoName, strings.Split(videoMIMEType, "/")[1]), videoData, fs.ModePerm)
				if err != nil {
					fmt.Println("ERROR WRITE FILE")
					fmt.Println(err.Error())
				}

				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String("Video anda telah berhasil disimpan."),
				})
				return
			}

			if v.Message.GetImageMessage() != nil {
				fmt.Println("IMAGE EXISTS")
				imageMessage := v.Message.GetImageMessage()
				imageData, err := wm.Client.Download(imageMessage)
				if err != nil {
					fmt.Println("ERROR DOWNLOAD IMAGE")
					return
				}
				imageMIMEType := imageMessage.GetMimetype()
				fmt.Printf("IMAGE MIMETYPE: %s \n", imageMIMEType)

				// permissions := 0644 // or whatever you need
				// byteArray := []byte("to be written to a file\n")
				imageName := uuid.New().String()
				err = os.WriteFile(fmt.Sprintf("./static/qr-codes/%s-image.%s", imageName, strings.Split(imageMIMEType, "/")[1]), imageData, fs.ModePerm)
				if err != nil {
					fmt.Println("ERROR WRITE FILE")
					fmt.Println(err.Error())
				}

				file, err := os.Open(fmt.Sprintf("./static/qr-codes/%s-image.%s", imageName, strings.Split(imageMIMEType, "/")[1]))
				if err != nil {
					fmt.Println("ERROR OPEN IMAGE")
					fmt.Println(err)
				}

				img, _, err := image.Decode(file)
				if err != nil {
					fmt.Println("ERROR DECODE IMAGE")
					fmt.Println(err)
				}

				// prepare BinaryBitmap
				bmp, err := gozxing.NewBinaryBitmapFromImage(img)
				if err != nil {
					fmt.Println("ERROR gozxing bitmap IMAGE")
					fmt.Println(err)
				}

				// decode image
				qrReader := gozqrcode.NewQRCodeReader()
				qrDecodeResult, err := qrReader.Decode(bmp, nil)
				if err != nil {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("QR Code tidak valid."),
					})
					return
				}
				fmt.Println("QR CODE DECODE RESULT")

				// TODO: update attendance, link video
				invitationCompleteData, err := wm.invitationStore.FindOneCompleteDataByUserID(context.Background(), qrDecodeResult.GetText())
				if err != nil {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("QR Code tidak valid. Info pengguna tidak ditemukan."),
					})
					return
				}

				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(fmt.Sprintf("Selamat datang, %s", invitationCompleteData.User.Name)),
				})
				return
			}
		}
	}
}
