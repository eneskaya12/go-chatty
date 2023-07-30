package main

import (
	"fmt"
	"net"
	"log"
	"strconv"
	"image/color"
	"strings"
	"go-client-online-chat/models"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
)

var privateChatWindows = make(map[string]fyne.Window)
var reply = make([]byte, 1024)
var messageHistory = make(map[string]bool)

type Client struct {
    socket net.Conn
    data   chan []byte
}

func main() {
	con, err := net.Dial("tcp", "localhost:9988")
	checkErr(err)
    defer con.Close()
	Interface(con)
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func(client *Client) receive(chatHistory *fyne.Container, chatScroll *container.Scroll, privateChatHistory *fyne.Container, privateChatScroll *container.Scroll) {
    for {
        message := make([]byte, 1024)
        _, err := client.socket.Read(message)
        if err != nil {
            client.socket.Close()
            break
        }
        chatMessage := string(message)
		messageParts := strings.SplitN(strings.TrimSpace(chatMessage), " ", 2)

		switch messageParts[0] {
		case "/Public":
			messageArray := strings.Split(messageParts[1], "/")

			if !messageHistory[messageArray[0]]{				
				messageHistory[messageArray[0]] = true
				messageLabel := widget.NewLabel(messageArray[0])
				chatHistory.Add(messageLabel)
				chatScroll.ScrollToBottom()
			}
		case "/Private":
			messageLabel := widget.NewLabel(messageParts[1])
			privateChatHistory.Add(messageLabel)
			privateChatScroll.ScrollToBottom()
		default:
			fmt.Println("something's wrong")
		}
    }
}

func Interface(con net.Conn) {
	//color purple
	colorPurple := color.NRGBA{R:160, G:118, B:249, A:255}

	client := &Client{socket: con}

	a := app.New()
	w := a.NewWindow("Online Chat App")
	w.Resize(fyne.NewSize(1000, 600))

	userListWindow := a.NewWindow("Direct")
	userListWindow.Resize(fyne.NewSize(200, 400))
	
	title := canvas.NewText("", colorPurple)
	title.Alignment = fyne.TextAlignCenter
	
	chatHistory := container.NewVBox()
	chatScroll := container.NewVScroll(chatHistory)
	chatScroll.SetMinSize(fyne.NewSize(500, 500))
	
	directConnection := container.NewHBox(
		canvas.NewText("Direct Connections", colorPurple),
	)

	var usernames []string
	var userModel []string

	con.Write([]byte("/user"))
	text, _ := con.Read(reply)
	chatMessage := string(reply[:text])
	
	if strings.HasPrefix(strings.TrimSpace(chatMessage), "/user") {
		messageParts := strings.SplitN(strings.TrimSpace(chatMessage), " ", 2)
		userModel = strings.Split(messageParts[1], ",")
	}
	
	var users []models.User

	for _, user := range userModel {
		messageParts := strings.SplitN(strings.TrimSpace(user), " ", 3)
		usernames = append(usernames, messageParts[1])
		u64, err := strconv.ParseUint(messageParts[0], 10, 32)
		checkErr(err)

		id := uint(u64)
		userObj := models.User{
			ID:       id,
			Name:     messageParts[1],
			Password: messageParts[2],
		}
		users = append(users, userObj)
	}
	
	userList := widget.NewList(
	    func() int {
	        return len(usernames)
	    },
	    func() fyne.CanvasObject {
	        return widget.NewLabel("")
	    },
	    func(index int, item fyne.CanvasObject) {
	        label := item.(*widget.Label)
	        username := usernames[index]
	        label.SetText(username)
	    },
	)	

	userListContainer := container.NewVScroll(userList)
	userListContainer.SetMinSize(fyne.NewSize(200, 400))

	publicMessageArea := container.NewHBox(
		canvas.NewText("Public Message Area", colorPurple),
	)
	
	input := widget.NewEntry()
	input.SetPlaceHolder("Enter text...")

	privateChatHistory := container.NewVBox()
	privateChatScroll := container.NewVScroll(privateChatHistory)
	
	sendButton := widget.NewButton("Send", func() {
		messageText := input.Text
		if messageText != "" {
			con.Write([]byte("/public " + messageText))
			input.SetText("")
			go client.receive(chatHistory, chatScroll, privateChatHistory, privateChatScroll)
		}
	})

	userList.OnSelected = func(index int) {
		if index >= 0 && index < len(userModel) {
			selectedUser := users[index]
			fmt.Println(users[index])
			if selectedUser.Name != "" {
				if window, ok := privateChatWindows[selectedUser.Name]; ok {
					window.RequestFocus()
				} else {
					privateChatWindow := a.NewWindow("Private Chat - " + selectedUser.Name)
					privateChatWindow.Resize(fyne.NewSize(500, 300))
					privateChatScroll.SetMinSize(fyne.NewSize(500, 250))

					privateInput := widget.NewEntry()
					privateInput.SetPlaceHolder("Enter private message...")
					privateSendButton := widget.NewButton("Send", func() {
						privateMessageText := privateInput.Text
						if privateMessageText != "" {
							con.Write([]byte("/private " + selectedUser.Name + " " + privateMessageText))
							privateInput.SetText("")
							go client.receive(chatHistory, chatScroll, privateChatHistory, privateChatScroll)
						}
					})

					privateChatLayout := container.NewVBox(
						privateChatScroll,
						privateInput,
						privateSendButton,
					)

                	privateChatWindow.SetOnClosed(func() {
                	    delete(privateChatWindows, selectedUser.Name)
                	})

					privateChatWindow.SetContent(privateChatLayout)
					privateChatWindows[selectedUser.Name] = privateChatWindow
					privateChatWindow.Show()
				}
			}
		}
	}
	
	messageArea := container.NewVBox(
		publicMessageArea,
		chatScroll,
		input,
		sendButton,
	)
	
	channelArea := container.NewVBox(
		directConnection,
		userListContainer,
	)

	split := container.NewHSplit(
		channelArea,
		messageArea,
	)
	split.Offset = 0.3
	
	//Login & Register Interface
	label := widget.NewLabel("")

	//register declaration
	registerForm := container.NewVBox()
	registerButton := widget.NewButton("button", func(){})
	registerLoginButton := widget.NewButton("button", func(){})

	//login
	loginName := widget.NewEntry()
	loginName.SetPlaceHolder("Enter your name...")
	loginPassword := widget.NewPasswordEntry()
	loginPassword.SetPlaceHolder("Enter your password...")

	loginForm := container.NewVBox(
		loginName,
		loginPassword,
	)

	loginButton := widget.NewButton("Login", func() {
		loginNameText := loginName.Text
		loginPasswordText := loginPassword.Text

		if loginNameText == "" || loginPasswordText == "" {
			label.Text="You need to enter your name and password"
			label.Refresh()
		} else {
			con.Write([]byte("/login " + loginNameText + " " + loginPasswordText))
			reply := make([]byte, 1024)
			
    		text, _ := con.Read(reply)
			loginMessage := string(reply[:text])

			if loginMessage == "/loginFailed" {
				label.Text = "Login Failed"
			    label.Refresh()

				loginName.SetText("")
				loginPassword.SetText("")

			} else if loginMessage == "/invalidCredentials" {
                label.Text = "Invalid username or password"
			    label.Refresh()

				loginName.SetText("")
				loginPassword.SetText("")

			} else if strings.HasPrefix(strings.TrimSpace(loginMessage), "/loginUsername") {
				messageParts := strings.SplitN(strings.TrimSpace(loginMessage), " ", 2)
				title.Text = messageParts[1]
				con.Write([]byte("/username"+" "+messageParts[1]))
				w.SetContent(container.NewVBox(
					title,
					split,
				))
			}else{
				label.Text="something's wrong"
				label.Refresh()
			}
		}
	})

	loginRegisterButton := widget.NewButton("New to Chat App? Create an account.", func() {
		w.SetContent(
			container.NewVBox(
				registerForm,
				registerButton,
				registerLoginButton,
				label,
		))
	})
	
	//register
	registerName := widget.NewEntry()
	registerName.SetPlaceHolder("Enter your name...")
	registerPassword := widget.NewPasswordEntry()
	registerPassword.SetPlaceHolder("Enter your password...")

	registerForm = container.NewVBox(
		registerName,
		registerPassword,
	)
	
	registerButton = widget.NewButton("Register", func() {
		registerNameText := registerName.Text
		registerPasswordText := registerPassword.Text

		if registerNameText == "" || registerPasswordText == "" {
			label.Text="You need to enter your name and password"
			label.Refresh()
			} else {
				con.Write([]byte("/register " + registerNameText + " " + registerPasswordText))
				reply := make([]byte, 1024)
				text, _ := con.Read(reply)
				registerMessage := string(reply[:text])

				if registerMessage == "/registerFailed" {
					label.Text="Register Failed"
					label.Refresh()

					registerName.SetText("")
					registerPassword.SetText("")
				} else if registerMessage == "/registerSuccessful"{
					w.SetContent(
						container.NewVBox(
							loginForm,
							loginButton,
							loginRegisterButton,
							label,
						),
					)
				} else{
					label.Text="something's wrong"
					label.Refresh()
				}			
		}
	})

	registerLoginButton = widget.NewButton("Already have an account? Login.", func() {
		w.SetContent(
			container.NewVBox(
				loginForm,
				loginButton,
				loginRegisterButton,
				label,
		))
	})
		
	w.SetContent(
		container.NewVBox(
			registerForm,
			registerButton,
			registerLoginButton,
			label,
	))
	
	w.ShowAndRun()
}