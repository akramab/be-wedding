package whatsapp

import (
	"be-wedding/internal/store"
	"context"
	"fmt"
	"image"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	gozqrcode "github.com/makiuchi-d/gozxing/qrcode"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

const (
	StateUploadPhotoVideo = "UPLOAD_PHOTO_VIDEO"
	StateSendQRCode       = "SEND_QR_CODE"
	StateChangeRSPV       = "CHANGE_RSVP"

	DefaultCacheTime      = time.Duration(1) * time.Minute
	DefaultCacheTimeVideo = time.Duration(10) * time.Minute

	GetCurrentVideoList = "CURRENT_VIDE_LIST"
	GetCurrentIndex     = "CURRENT_INDEX"
	StringSeparator     = ","
)

func (wm *whatsMeow) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if !v.Info.IsFromMe {
			fmt.Println("MESSAGE RECEIVED!", v.Message.GetConversation())
			userJID := v.Info.Sender.ToNonAD().String()

			fmt.Printf("USER JID: %s \n", userJID)
			invitationCompleteData, err := wm.invitationStore.FindOneCompleteDataByWANumber(context.Background(), "+"+strings.Split(userJID, "@")[0])
			if err != nil {
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(fmt.Sprintf("Maaf, nomor anda belum terdaftar. Silahkan registrasi melalui undangan yang telah dikirimkan")),
				})
				return
			}

			userState, err := wm.redisCache.Get(context.Background(), invitationCompleteData.User.ID).Result()
			userMessage := strings.TrimSpace(v.Message.GetConversation())
			if userMessage == "" && userState != StateUploadPhotoVideo && userState != StateSendQRCode {
				if v.RawMessage != nil && v.RawMessage.ExtendedTextMessage != nil && v.RawMessage.ExtendedTextMessage.Text != nil {
					userMessage = strings.TrimSpace(*v.RawMessage.ExtendedTextMessage.Text)
				} else {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String(fmt.Sprintf("Terjadi kesalahan teknis pada sistem. Pesan dari nomor yang anda gunakan saat ini tidak bisa diproses oleh sistem. Silakan coba kembali menggunakan nomor lain")),
					})
					return
				}
			}

			switch userState {
			case StateUploadPhotoVideo:
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
						Conversation: proto.String("Terima kasih. Video Ucapan anda telah berhasil disimpan"),
					})
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)
					return
				} else if v.Message.GetImageMessage() != nil {
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
					err = os.WriteFile(fmt.Sprintf("./static/images/%s-image.%s", imageName, strings.Split(imageMIMEType, "/")[1]), imageData, fs.ModePerm)
					if err != nil {
						fmt.Println("ERROR WRITE FILE")
						fmt.Println(err.Error())
					}

					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Terima kasih. Foto Ucapan anda telah berhasil disimpan"),
					})
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)
					return
				} else if userMessage == "0" {
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Pengiriman ucapan dibatalkan"),
					})
					return
				} else {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Silakan kirimkan foto atau video ucapan anda. Tekan 0 jika anda ingin membatalkan"),
					})
					return
				}
			case StateSendQRCode:
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

					// update state to default
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)

					// decode image
					qrReader := gozqrcode.NewQRCodeReader()
					qrDecodeResult, err := qrReader.Decode(bmp, nil)
					if err != nil {
						wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
							Conversation: proto.String("QR Code tidak valid"),
						})
						return
					}
					fmt.Println("QR CODE DECODE RESULT")
					fmt.Println(qrDecodeResult.GetText())

					// TEST
					testUserId := strings.TrimSpace(qrDecodeResult.GetText())

					var currentIdx int
					currentIdxString, _ := wm.redisCache.Get(context.Background(), GetCurrentIndex).Result()
					if currentIdxString == "" {
						currentIdx = 0
					} else {
						currentIdx, _ = strconv.Atoi(currentIdxString)
						currentIdx++
					}

					videoListString, _ := wm.redisCache.Get(context.Background(), GetCurrentVideoList).Result()
					if videoListString == "" {
						videoListString = strings.Join([]string{"https://api.kramili.site/static/3.mp4", "https://api.kramili.site/static/5.mp4"}, StringSeparator)
					}
					videoList := strings.Split(videoListString, ",")

					if testUserId == "a180b098-8568-4ec3-9822-48a313b83047" {
						videoList = append(videoList[:currentIdx+1], videoList[currentIdx:]...)
						videoList[currentIdx] = "https://api.kramili.site/static/1.mp4"
						// videoList = append(videoList, "https://api.kramili.site/static/1.mp4")
					}
					if testUserId == "484d20a6-8097-447a-bf4f-fbdef7db6eca" {
						videoList = append(videoList[:currentIdx+1], videoList[currentIdx:]...)
						videoList[currentIdx] = "https://api.kramili.site/static/2.mp4"
						// videoList = append(videoList, "https://api.kramili.site/static/2.mp4")
					}
					if testUserId == "7e4a645b-341e-420a-81b3-f38f85629ff8" {
						videoList = append(videoList[:currentIdx+1], videoList[currentIdx:]...)
						videoList[currentIdx] = "https://api.kramili.site/static/4.mp4"
						// videoList = append(videoList, "https://api.kramili.site/static/4.mp4")
					}
					if testUserId == "0c467423-e324-4786-a9d9-0c77eb267407" {
						videoList = append(videoList[:currentIdx+1], videoList[currentIdx:]...)
						videoList[currentIdx] = "https://api.kramili.site/static/6.mp4"
						// videoList = append(videoList, "https://api.kramili.site/static/6.mp4")
					}
					if testUserId == "69fee15d-2c45-48ea-982e-4ce6327298fc" {
						videoList = append(videoList[:currentIdx+1], videoList[currentIdx:]...)
						videoList[currentIdx] = "https://api.kramili.site/static/7.mp4"
						// videoList = append(videoList, "https://api.kramili.site/static/7.mp4")
					}
					wm.redisCache.Set(context.Background(), GetCurrentVideoList, strings.Join(videoList, ","), DefaultCacheTimeVideo)

					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String(fmt.Sprintf("Selamat datang, %s", qrDecodeResult.GetText())),
					})

					return

					// TODO: update attendance, link video
					// invitationCompleteData, err := wm.invitationStore.FindOneCompleteDataByUserID(context.Background(), qrDecodeResult.GetText())
					// if err != nil {
					// 	wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					// 		Conversation: proto.String("QR Code tidak valid. Info pengguna tidak ditemukan"),
					// 	})
					// 	return
					// }

					// wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					// 	Conversation: proto.String(fmt.Sprintf("Selamat datang, %s", invitationCompleteData.User.Name)),
					// })
					// return
				} else if v.Message.GetConversation() == "0" {
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Pengiriman code QR dibatalkan"),
					})
					return
				} else {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Mohon kirimkan code QR anda. Tekan 0 jika anda ingin membatalkan"),
					})
					return
				}
			case StateChangeRSPV:
				var changeRSVPValid bool
				var newPeopleCount int
				if newPeopleCount, err = strconv.Atoi(userMessage); err == nil && userMessage != "0" {
					changeRSVPValid = true
				}
				if changeRSVPValid {
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)

					wm.userStore.UpdateRSVPByUserID(context.Background(), &store.UserRSVPData{
						UserID:      invitationCompleteData.User.ID,
						PeopleCount: int64(newPeopleCount),
					})
					replyMessage := fmt.Sprintf(`Data konfirmasi kehadiran anda telah diperbarui

Berikut ini rekap rencana kehadiran yang tercatat:
					
*Nama*			: %s
*Jumlah Orang*	: %d
					
Ketik angka 1 jika anda ingin kembali mengubah jumlah kehadiran`, invitationCompleteData.User.Name, newPeopleCount)
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String(replyMessage),
					})
					return
				} else if userMessage == "0" {
					wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, "", DefaultCacheTime)
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Pengubahan jumlah kehadiran dibatalkan"),
					})
					return
				} else {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Maaf, jumlah kehadiran baru anda tidak dapat diproses. Silakan ketik kembali jumlah kehadiran baru anda dalam *angka*. Ketik 0 jika anda ingin membatalkan"),
					})
					return
				}
			}

			switch userMessage {
			case "1":
				wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, StateChangeRSPV, DefaultCacheTime)
				replyMessage := `Anda akan mengubah jumlah kehadiran

Ketik jumlah kehadiran baru anda (cukup tuliskan dalam *angka*)`
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(replyMessage),
				})
				return
			case "2":
				replyMessage := fmt.Sprintf(`Berikut ini rekap rencana kehadiran yang tercatat:

*Nama*			: %s
*Jumlah Orang*	: %d

*Ketik angka 1 jika anda ingin mengubah jumlah kehadiran*`, invitationCompleteData.User.Name, invitationCompleteData.User.PeopleCount)
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(replyMessage),
				})
				return
			case "3":
				replyMessage := `Berikut ini code QR anda`
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String(replyMessage),
				})

				captionImageMessage := `Tunjukkan code QR saat hendak memasuki venue pada hari H.`
				err = wm.SendImageMessage(context.Background(), invitationCompleteData.User.WhatsAppNumber, invitationCompleteData.User.QRImage, captionImageMessage)
				if err != nil {
					wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
						Conversation: proto.String("Maaf terjadi kesalahan saat mengirimkan code QR. Silakan coba kembali"),
					})
					return
				}
				return
			case "23":
				wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, StateUploadPhotoVideo, DefaultCacheTime)
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String("Silakan kirimkan foto atau video ucapan anda"),
				})
				return
			case "1819":
				wm.redisCache.Set(context.Background(), invitationCompleteData.User.ID, StateSendQRCode, DefaultCacheTime)
				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String("Silakan kirimkan QR code anda"),
				})
				return
			case "2306":
				wm.redisCache.Set(context.Background(), GetCurrentIndex, "", DefaultCacheTimeVideo)
				wm.redisCache.Set(context.Background(), GetCurrentVideoList, "", DefaultCacheTimeVideo)

				wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
					Conversation: proto.String("Daftar video telah berhasil dihapus"),
				})
				return
			case "Broadcast Reminder Ucapan wpVZD":
				if wm.Config.BroadcastMode {
					waNumberList, err := wm.userStore.FindAllWhatsAppNumber(context.Background())
					if err != nil {
						wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
							Conversation: proto.String("Broadcast error. Can't get WhatsApp Number List."),
						})
						return
					}
					// waNumber := "+6282214225921"
					// waNumber := "+628121552492"
					for _, waNumber := range waNumberList {
						firstMessage := `Assalamu'alaikum warahmatullahi wabarakatuh 

Perkenalkan kami Afra Izzati Kamili dan  Muhammad Akram Al Bari. Semoga Bapak/ Ibu telah menerima undangan pernikahan kami.
						
*Tanpa mengurangi rasa hormat, kami tidak menerima karangan bunga secara fisik*. Namun, kami sangat menantikan ucapan selamat dan do'a, berupa foto atau video yang insya Allah, akan ditampilkan pada hari pernikahan. 
						
Berikut kami lampirkan contoh foto dan video yang dimaksud`
						wm.SendMessage(context.Background(), waNumber, &waProto.Message{
							Conversation: proto.String(firstMessage),
						})
						wm.SendImageMessage(context.Background(), waNumber, "contoh-foto-1.png", "")
						wm.SendImageMessage(context.Background(), waNumber, "contoh-foto-2.png", "")
						// VIDEO MESSAGE
						wm.SendVideoMessage(context.Background(), waNumber, "contoh-video-1.mp4", "")
						wm.SendVideoMessage(context.Background(), waNumber, "contoh-video-2.mp4", "")
						secondMessage := `*Pengiriman foto dan/atau video dapat melalui nomor WhatsApp ini* dengan format jpg/png/pdf/mkv/mp4/mov
						
Terima kasih atas perhatian, pengertian, dan do'anya.
Jazaakumullahu khairan katsiraa.
						
Wassalamu'alaikum warahmatullahi wabarakatuh
						
		Afra - Akram 🌹`
						wm.SendMessage(context.Background(), waNumber, &waProto.Message{
							Conversation: proto.String(secondMessage),
						})
					}

				}

				return
			}

			replyMessage := `Pesan anda tidak dikenali

Anda dapat berinteraksi dengan akun WhatsApp ini dengan mengetikkan daftar pesan di bawah ini:
			
- Tekan *1* untuk *mengubah data jumlah konfirmasi kehadiran*
- Tekan *2* untuk *melihat data konfirmasi kehadiran anda*
- Tekan *3* untuk *mendapatkan kembali code QR anda*
- Tekan *23* untuk *mengirim foto atau video ucapan*
			
Terima kasih`
			wm.Client.SendMessage(context.Background(), v.Info.Sender.ToNonAD(), &waProto.Message{
				Conversation: proto.String(replyMessage),
			})
			return
		}
	}
}
