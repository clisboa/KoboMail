package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

//config struct
type KoboMailConfig struct {
	IMAP_Host    string `toml:"imap_host"`
	IMAP_Port    string `toml:"imap_port"`
	IMAP_User    string `toml:"imap_user"`
	IMAP_Pwd     string `toml:"imap_pwd"`
	Email_Flag   string `toml:"email_flag"`
	Email_Unseen string `toml:"email_unseen"`
}

// chkErrFatal prints a message to the Kobo screen, then exits the program
func chkErrFatal(err error, usrMsg string, msgDuration int) {
	if err != nil {
		if usrMsg != "" {
			// fbPrint(usrMsg)
			time.Sleep(time.Duration(msgDuration) * time.Second)
		}
		log.Fatal(err)
	}
}

// // logErrPrint is a convenience function for logging errors
// func logErrPrint(err error) {
// 	if err != nil {
// 		log.Print(err)
// 	}
// }

// nickelUSBplug simulates pugging in a USB cable
func nickelUSBplug() {
	nickelHWstatusPipe := "/tmp/nickel-hardware-status"
	nickelPipe, _ := os.OpenFile(nickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug add")
	nickelPipe.Close()
}

// nickelUSBunplug simulates unplugging a USB cable
func nickelUSBunplug() {
	nickelHWstatusPipe := "/tmp/nickel-hardware-status"
	nickelPipe, _ := os.OpenFile(nickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug remove")
	nickelPipe.Close()
}

func main() {
	// If the file doesn't exist, create it or append to the
	KM_Log_Path := ""
	if _, err := os.Stat("/mnt/onboard/.add/kobomail/logs.txt"); err == nil {
		KM_Log_Path = "/mnt/onboard/.add/kobomail/logs.txt"
	} else if os.IsNotExist(err) {
		KM_Log_Path = "logs.txt"
	}
	logFile, err := os.OpenFile(KM_Log_Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//let's output to both stoud and log file
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	KM_Library_Path := ""
	if _, err := os.Stat("/mnt/onboard/KoboMailLibrary/"); err == nil {
		KM_Library_Path = "/mnt/onboard/KoboMailLibrary/"
	} else if os.IsNotExist(err) {
		KM_Library_Path = "Library/"
	}

	// Read Config file
	KM_Config_Path := ""
	if _, err := os.Stat("/mnt/onboard/.add/kobomail/kobomail_cfg.toml"); err == nil {
		KM_Config_Path = "/mnt/onboard/.add/kobomail/kobomail_cfg.toml"
	} else if os.IsNotExist(err) {
		KM_Config_Path = "kobomail_cfg.toml"
	}

	var KM_Config KoboMailConfig
	if _, err := toml.DecodeFile(KM_Config_Path, &KM_Config); err != nil {
		chkErrFatal(err, "Couldn't read config. Aborting!", 5)
	}

	host := KM_Config.IMAP_Host
	port := KM_Config.IMAP_Port
	user := KM_Config.IMAP_User
	pass := KM_Config.IMAP_Pwd
	tlsn := ""
	if port == "" {
		port = "993"
	}

	connStr := fmt.Sprintf("%s:%s", host, port)

	tlsc := &tls.Config{}
	if tlsn != "" {
		tlsc.ServerName = tlsn
	}

	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	defer c.Logout()

	if err := c.Login(user, pass); err != nil {
		log.Fatal(err)
	}
	log.Println("Authenticated")

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inbox selected: ", mbox.Name)

	criteria := imap.NewSearchCriteria()
	if KM_Config.Email_Unseen == "true" {
		criteria.WithoutFlags = []string{"\\Seen"}
	}
	KM_Config_To_Kobo := strings.Replace(KM_Config.IMAP_User, "@", "+"+KM_Config.Email_Flag+"@", 1)
	criteria.Header.Add("TO", KM_Config_To_Kobo)

	uids, err := c.Search(criteria)
	if err != nil {
		log.Println(err)
	}
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)
	log.Printf("Search complete, found %d messages", len(uids))

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
		log.Println("Fetch complete")
	}()

	for msg := range messages {
		if msg != nil {
			log.Printf("got message with address %p\n", msg)

			r := msg.GetBody(section)
			if r == nil {
				log.Fatal("Server didn't returned message body")
			}
			// Create a new mail reader
			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Fatal(err)
			}

			// Print some info about the message
			header := mr.Header
			if date, err := header.Date(); err == nil {
				log.Println("Date:", date)
			}
			// if from, err := header.AddressList("From"); err == nil {
			// 	log.Println("From:", from)
			// }
			// if to, err := header.AddressList("To"); err == nil {
			// 	log.Println("To:", to)
			// }
			if subject, err := header.Subject(); err == nil {
				log.Println("Subject:", subject)
			}

			// Process each message's part
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}

				switch h := p.Header.(type) {
				case *mail.AttachmentHeader:
					// This is an attachment
					filename, _ := h.Filename()
					log.Println("Got attachment:", filename)
					contenttype, _, _ := h.ContentType()
					log.Println("of type: ", contenttype)
					b, _ := ioutil.ReadAll(p.Body)
					// write the whole body at once
					err = ioutil.WriteFile(KM_Library_Path+filename, b, 0644)
					if err != nil {
						panic(err)
					}

				}
			}
		} else {
			log.Println("no messages matched criteria")
		}
	}
	// if err := <-done; err != nil {
	// 	log.Fatal(err)
	// }

	//now that we finished loading all messages we'll simulate the USB cable connect
	//but only if there were any messages processed, no need to bug the user if there was nothing new
	if len(uids) > 0 {
		log.Println("Simulating PLugging USB cable and wait 10s for the user to click on the connect button")
		nickelUSBplug()

		time.Sleep(10 * time.Second)

		log.Println("Simulating unplugging USB cable")
		nickelUSBunplug()
		//after this Nickel will do the job to import the new files loaded into the KoboMailLibrary folder
	}
}
