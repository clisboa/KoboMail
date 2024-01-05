package main

import (
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
)

type zeit_config struct {
	Zeit_User string `toml:"zeit_user"`
	Zeit_Pwd  string `toml:"zeit_pwd"`
}

const ZEIT_SENDER_ADDRESS = "noreply@digitalabo.mailing.zeit.de"

func processZeitDownloadNotification(p *mail.Part, config zeit_config) bool {

	log.Println("Found email sent by Zeit")

	// TODO: find a more robust way to get to link
	body, err := io.ReadAll(p.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	mailContent := string(body)
	needleStart := strings.Index(mailContent, "automatisch heruntergeladen")
	if needleStart == -1 {
		log.Println("Could not find link start search marker")
		return false
	}
	mailContent = mailContent[needleStart:]
	indexStart := strings.Index(mailContent, "https://t.mailing.zeit.de/lnk/")
	if indexStart == -1 {
		log.Println("Could not find link marker")
		return false
	}
	mailContent = mailContent[indexStart:]
	indexEnd := strings.Index(mailContent, "\"")
	downloadUrl := mailContent[0:indexEnd]

	log.Println("found download link, starting download")
	return downloadZeitEpub(defaultLibraryPath+"zeit_"+time.Now().Format("02-01-2006")+".epub", downloadUrl, config)
}

func downloadZeitEpub(filepath string, address string, config zeit_config) bool {

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Println("Error:", err)
		return false
	}

	client := &http.Client{
		Jar: jar,
	}

	log.Println("Getting login page for CSRF")

	csrf_resp, err := client.Get("https://meine.zeit.de/anmelden")
	if err != nil {
		log.Println("Error:", err)
		return false
	}

	csrf_token := ""
	for _, c := range csrf_resp.Cookies() {
		if strings.Compare(c.Name, "csrf_token") == 0 {
			csrf_token = c.Value
		}
	}
	log.Println("Got csrf token " + csrf_token)

	values := url.Values{}
	values.Set("csrf_token", csrf_token)
	values.Set("pass", config.Zeit_Pwd)
	values.Set("email", config.Zeit_User)

	requestDir, err := http.NewRequest("POST", "https://meine.zeit.de/anmelden", strings.NewReader(values.Encode()))

	requestDir.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// Backend NEEDS to have Origin set, otherwise error
	requestDir.Header.Add("Origin", "https://meine.zeit.de")

	log.Println("Logging in")

	_, err = client.Do(requestDir)
	if err != nil {
		log.Println("Error:", err)
		return false
	}

	log.Println("Retrieving epub")

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Println("Could not create file", err)
		return false
	}
	defer out.Close()

	// Get the data
	resp, err := client.Get(address)
	if err != nil {
		log.Println("Could not retrieve zeit epub file", err)
		return false
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Println("bad status: " + resp.Status)
		return false
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("Could not write to file")
		return false
	}

	return true
}
