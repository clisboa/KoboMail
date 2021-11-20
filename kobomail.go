package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

const defaultPath = "/mnt/onboard/.adds/kobomail/"
const defaultRootFSPath = "/usr/local/kobomail/"
const defaultLogFile = "kobomail.log"
const defaultConfigFile = "kobomail_cfg.toml"
const defaultLibraryPath = "/mnt/onboard/KoboMailLibrary/"

const defaultNickelMenuPath = "/mnt/onboard/.adds/nm"
const defaultNickelMenuConfigTmpl = defaultRootFSPath + "kobomail_nm.tmpl"
const defaultNickelMenuConfig = defaultNickelMenuPath + "/kobomail"

const defaultUDEVFileTmpl = defaultRootFSPath + "97-kobomail.rules.tmpl"
const defaultUDEVFile = "/etc/udev/rules.d/97-kobomail.rules"

const defaultNickelDbusPath = "/mnt/onboard/.adds/nickeldbus"

const defaultnickelHWstatusPipe = "/tmp/nickel-hardware-status"

const binCP = "/bin/cp"
const binRM = "/bin/rm"
const binQndb = "/usr/bin/qndb"

//config struct
type KoboMailConfig struct {
	IMAP_Config       imap_config
	Execution_Type    execution_type
	Processing_Config processing_config
}

type imap_config struct {
	IMAP_Host       string `toml:"imap_host"`
	IMAP_Port       string `toml:"imap_port"`
	IMAP_User       string `toml:"imap_user"`
	IMAP_Pwd        string `toml:"imap_pwd"`
	Email_Flag_Type string `toml:"email_flag_type"`
	Email_Flag      string `toml:"email_flag"`
	Email_Unseen    string `toml:"email_unseen"`
}

type execution_type struct {
	Type string `toml:"type"`
}

type processing_config struct {
	Filetypes []string `toml:"filetypes"`
	Kepubify  string   `toml:"kepubify"`
}

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAddRemove(action string) {
	nickelPipe, _ := os.OpenFile(defaultnickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug " + action)
	nickelPipe.Close()
}

//let's check if the NickelDbus version is correct
func checkNickelDbusVersion() (ok bool) {
	arg1 := "-m"
	arg2 := "ndbVersion"
	cmd := exec.Command(binQndb, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return false
	}

	currentNickelDbusversion := "0.2.0"
	if strings.TrimSpace(string(stdout)) != currentNickelDbusversion {
		log.Fatal("NickelDbus wrong version, should be ", currentNickelDbusversion, " but we got ", string(stdout))
		return false
	}
	log.Println("dBus ndbVersion called. ", string(stdout))
	return true
}

// send a request ro rescan the library with a timeout
func dBusLibraryRescanFull(timeout string) {
	arg1 := "-t"
	arg2 := timeout
	arg3 := "-s"
	arg4 := "pfmDoneProcessing"
	arg5 := "-m"
	arg6 := "pfmRescanBooksFull"

	cmd := exec.Command(binQndb, arg1, arg2, arg3, arg4, arg5, arg6)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return
	}

	log.Println("dBus Dialog called ", string(stdout))
}

// create a Dbus dialog to show percentage of execution
func dBusDialog(APItype string, args ...string) {
	arg1 := "-m"
	arg2 := APItype
	call_args := append([]string{arg1, arg2}, args...)
	cmd := exec.Command(binQndb, call_args...)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return
	}

	log.Println("dBus Toast called. ", string(stdout))
}

func dbusDialogCreate(initial_msg string) {
	dBusDialog("dlgConfirmCreate")
	dBusDialog("dlgConfirmSetTitle", "KoboMail")
	dBusDialog("dlgConfirmSetBody", initial_msg)
	dBusDialog("dlgConfirmSetModal", "false")
	dBusDialog("dlgConfirmShowClose", "true")
	dBusDialog("dlgConfirmShow")
}

func dbusDialogUpdate(body string) {
	dBusDialog("dlgConfirmSetBody", body)
}

func dbusDialogAddOKButton() {
	dBusDialog("dlgConfirmSetAccept", "OK")
}

// copy UDEV rulesfile to the correct place so we can run KoboMail automatically everytime WIfi is activated
func copyUDEVRulesfile() (ok bool) {
	arg1 := "/usr/local/kobomail/97-kobomail.rules.tmpl"
	arg2 := "/etc/udev/rules.d/97-kobomail.rules"
	cmd := exec.Command(binCP, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return false
	}

	log.Println("UDEV Rules config file put in place", string(stdout))
	return true
}

// copy NickelMenu config file to the correct place so we can run KoboMail mannually
func copyNickelMenufile() (ok bool) {
	arg1 := "/usr/local/kobomail/kobomail_nm.tmpl"
	arg2 := "/mnt/onboard/.adds/nm/kobomail"
	cmd := exec.Command(binCP, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return false
	}

	log.Println("NickelMenu config file put in place", string(stdout))
	return true
}

// delete NickelMenu config file
func deleteNickelMenuConfigMFile() (ok bool) {
	arg1 := "-f"
	arg2 := "/mnt/onboard/.adds/nm/kobomail"
	cmd := exec.Command(binRM, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return false
	}

	log.Println("Removed NickelMenu config file, ", string(stdout))
	return true
}

// delete UDEV rules file
func deleteUDEVRulesFile() (ok bool) {
	arg1 := "-f"
	arg2 := "/etc/udev/rules.d/97-kobomail.rules"
	cmd := exec.Command(binRM, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Something went wrong", err.Error())
		return false
	}

	log.Println("NickelMenu config file put in place", string(stdout))
	return true
}

func checkCurrentExecutionType(exec_type string) (ok bool) {
	if exec_type == "manual" {
		if _, err := os.Stat("/mnt/onboard/.adds/nm"); err == nil {
			log.Println("Found NickelMenu")
			if _, err := os.Stat("/mnt/onboard/.adds/nm/kobomail"); err == nil {
				log.Println("Found KoboMail NickelMenu config file")
				if _, err := os.Stat("/etc/udev/rules.d/97-kobomail.rules"); err == nil {
					log.Println("But also found KoboMail udev rules file in place, we'll need to take care of that")
					return false
				} else {
					log.Println("Everything is corretly in place for manual execution")
					return true
				}
			} else {
				log.Println("But NickelMenu config file is missing, we'll need to take care of that")
				return false
			}
		} else {
			log.Fatal("Did not found NickelMenu, can't change method to manual")
			return false
		}
	} else if exec_type == "auto" {
		//let's check if the udev rules file is in place, if not we'll copy the template
		if _, err := os.Stat("/etc/udev/rules.d/97-kobomail.rules"); err == nil {
			log.Println("Found udev rules file in place")
			if _, err := os.Stat("/mnt/onboard/.adds/nm/kobomail"); err == nil {
				log.Println("But also found NickelMenu config file, we'll have to clean that up")
				return false
			} else {
				log.Println("Everything is corretly in place for auto execution")
				return true
			}
		} else {
			log.Println("Did not found udev rules in place, we'll have to take care of that")
			return false
		}
	}
	log.Fatal("Incorrect execution type option exists in the config file")
	return false
}

func KoboMailExecutionType(exec_type string) (ok bool) {
	log.Println("Execution type is " + exec_type)

	if chkExecType := checkCurrentExecutionType(exec_type); chkExecType {
		return true
	} //else we'll sanitaze whatever needs to be taken care off

	if exec_type == "manual" {
		if _, err := os.Stat("/mnt/onboard/.adds/nm"); err == nil {
			log.Println("Found NickelMenu")
			if copyNMfile := copyNickelMenufile(); !copyNMfile {
				log.Println("Failed to put NickelMenu config file in place")
				return false
			}
			if delUdevRules := deleteUDEVRulesFile(); !delUdevRules {
				log.Println("Failed to delete UDEV rules")
				return false
			}
			log.Println("Found NickelMenu, added the KoboMail NickelMenu file and removed UDEV rules successfully")
			return true
		} else {
			log.Println("Did not found NickelMenu, can't change method to manual")
			return false
		}
	} else if exec_type == "auto" {
		//let's check if the udev rules file is in place, if not we'll copy the template
		if _, err := os.Stat("/etc/udev/rules.d/97-kobomail.rules"); err != nil {
			log.Println("Did not find udev rules file, let's copy it from the template")
			if copyUDEVfile := copyUDEVRulesfile(); !copyUDEVfile {
				log.Println("Failed to delete UDEV rules")
				return false
			}
		}
		if deleteNMFile := deleteNickelMenuConfigMFile(); !deleteNMFile {
			log.Println("Failed to delete NickelMenu config file")
			return false
		}
		return true
	}
	log.Fatal("Incorrect execution type option exists in the config file")
	return false
}

func containsFiletype(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func main() {

	//let's block the log file
	logFile, err := os.OpenFile(defaultPath+defaultLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//let's output to both stoud and log file
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// Check if NickelDbus is installed, if so then for interacting with Nickel
	// for library rescan and user notification will be handled with that
	// if not then let's use the bruteforce method of simulating the usb cable connect
	KM_UseNickelDbus := false
	if _, err := os.Stat(defaultNickelDbusPath); err == nil {
		KM_UseNickelDbus = true
		log.Println("Found NickelDbus")
		if checkDbusVer := checkNickelDbusVersion(); !checkDbusVer {
			log.Println("NickelDbus version mismatch")
			KM_UseNickelDbus = false
		}
	} else {
		log.Println("Did not found NickelDbus")
	}

	if KM_UseNickelDbus {
		//let's start by showing the user we are running opening a dialog
		dbusDialogCreate("Starting up, please wait.")
	}

	//let's read the TOML config file and parse it
	var KM_Config KoboMailConfig
	if _, err := toml.DecodeFile(defaultPath+defaultConfigFile, &KM_Config); err != nil {
		log.Fatal("Couldn't read config. Aborting! ", err)
	}

	//let's check if the user configured Execution Type for KoboMail to be auto or manual
	if exec := KoboMailExecutionType(KM_Config.Execution_Type.Type); !exec {
		log.Fatal("Couldn't change execution method", err)
	}

	tlsn := ""
	connStr := fmt.Sprintf("%s:%s", KM_Config.IMAP_Config.IMAP_Host, KM_Config.IMAP_Config.IMAP_Port)
	tlsc := &tls.Config{}
	if tlsn != "" {
		tlsc.ServerName = tlsn
	}

	//we'll try to login, if not we most likely have a faulty internet connection or an invalid imap host
	num_retries := 3
	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		for num_retries > 0 {
			log.Println("Could not login to server, trying again in 1 second...")
			time.Sleep(1 * time.Second)
			c, err = client.DialTLS(connStr, tlsc)
			if err != nil {
				log.Println(err)
				num_retries -= 1
			} else {
				break
			}
		}
		if err != nil {
			if KM_UseNickelDbus {
				log.Println(err)
				dbusDialogAddOKButton()
				dbusDialogUpdate("Tried 3 times to login to " + KM_Config.IMAP_Config.IMAP_Host + " but failed, please check internet connection")
				os.Exit(0)
			}
			log.Println("Tried 3 times to login " + KM_Config.IMAP_Config.IMAP_Host + " but failed, please check internet connection")
			log.Fatal(err)
		}
	}
	log.Println("Connected")
	defer c.Logout()

	//we connected to the imap host, let's try to login
	if err := c.Login(KM_Config.IMAP_Config.IMAP_User, KM_Config.IMAP_Config.IMAP_Pwd); err != nil {
		if KM_UseNickelDbus {
			log.Println(err)
			dbusDialogAddOKButton()
			dbusDialogUpdate("Failed to authenticate: " + err.Error())
			os.Exit(0)
		}
		log.Fatal(err)
	}
	log.Println("Authenticated")

	//we'll select inbox so we can search on it
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		if KM_UseNickelDbus {
			log.Println(err)
			dbusDialogAddOKButton()
			dbusDialogUpdate("Failed to select INBOX: " + err.Error())
			os.Exit(0)
		}
		log.Fatal(err)
	}
	log.Println("Inbox selected: ", mbox.Name)

	//deciding on the search criteria given the configuration we have
	criteria := imap.NewSearchCriteria()
	if KM_Config.IMAP_Config.Email_Unseen == "true" {
		criteria.WithoutFlags = []string{"\\Seen"}
	}
	if KM_Config.IMAP_Config.Email_Flag_Type == "plus" {
		log.Println("IMAP_Config.Email_Flag_Type == plus")
		KM_Config_To_Kobo := strings.Replace(KM_Config.IMAP_Config.IMAP_User, "@", "+"+KM_Config.IMAP_Config.Email_Flag+"@", 1)
		criteria.Header.Add("TO", KM_Config_To_Kobo)
	} else if KM_Config.IMAP_Config.Email_Flag_Type == "subject" {
		log.Println("IMAP_Config.Email_Flag_Type == subject")
		KM_Config_To_Kobo := KM_Config.IMAP_Config.Email_Flag
		criteria.Header.Add("SUBJECT", KM_Config_To_Kobo)
	}

	//let's apply the search criteria and check if there's any emails with that criteria
	uids, err := c.Search(criteria)
	if err != nil {
		log.Println(err)
	}
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)
	number_emails_found := len(uids)
	log.Printf("Search complete, found %d messages", len(uids))

	if KM_UseNickelDbus {
		if number_emails_found > 0 {
			dbusDialogUpdate("Found " + strconv.Itoa(number_emails_found) + " emails to process. Please wait...")
		} else {
			dbusDialogAddOKButton()
			dbusDialogUpdate("No emails found, nothing to be done.")
			os.Exit(0)
		}
	}

	//let's fetch the emails list we got
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
		log.Println("Fetch complete")
	}()

	//for each email fetched we'll look for the attachments
	number_ebooks_processed := 0
	for msg := range messages {
		if msg != nil {
			log.Printf("got message with address %p\n", msg)

			r := msg.GetBody(section)
			if r == nil {
				if KM_UseNickelDbus {
					dbusDialogAddOKButton()
					dbusDialogUpdate("Exiting, server didn't returned message body: " + err.Error())
					os.Exit(0)
				}
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

			if subject, err := header.Subject(); err == nil {
				log.Println("Subject:", subject)
			}

			// Process each message's part, there might be multiple attachments
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

					fileExtension := strings.Trim(filepath.Ext(filename), ".")
					log.Println("of type: ", fileExtension)

					//we'll only save the attachment if the filetype is allowed
					if containsFiletype(KM_Config.Processing_Config.Filetypes, fileExtension) {
						log.Println("file is in the allowed filetypes, let's save it")
						//but first let's check if the file is a kepub, we need to rename it to .kepub.epub so kobo can properly handle it
						if fileExtension == "kepub" {
							log.Println("file is a kepub, lets rename it to .kepub.epub")
							filename += ".epub"
						}

						b, _ := ioutil.ReadAll(p.Body)
						// write the whole body at once
						err = ioutil.WriteFile(defaultLibraryPath+filename, b, 0644)
						if err != nil {
							panic(err)
						}
						number_ebooks_processed += 1
					} else {
						log.Println("file is not in the allowed filetypes, let's ignore it")
					}

				}
			}
		} else {
			log.Println("no messages matched criteria")
		}
	}

	if KM_UseNickelDbus {
		if number_ebooks_processed > 0 {
			//lets rescan the library for the new ebooks
			dBusLibraryRescanFull("30000")
			dbusDialogCreate("Processed " + strconv.Itoa(number_ebooks_processed) + " new ebooks.")
			dbusDialogAddOKButton()
		} else {
			dbusDialogCreate("No emails found, nothing to be done.")
			dbusDialogAddOKButton()
		}
	} else {
		//now that we finished loading all messages we'll simulate the USB cable connect
		//but only if there were any messages processed, no need to bug the user if there was nothing new
		if number_ebooks_processed > 0 {
			log.Println("Simulating PLugging USB cable and wait 10s for the user to click on the connect button")
			nickelUSBplugAddRemove("add")
			time.Sleep(10 * time.Second)
			log.Println("Simulating unplugging USB cable")
			nickelUSBplugAddRemove("remove")
			//after this Nickel will do the job to import the new files loaded into the KoboMailLibrary folder
		}
	}
}
