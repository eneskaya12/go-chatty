package main

import (
        "fmt"
        "strconv"
        "log"
        "net"
        "os"
        "strings"
        "go-server-online-chat/models"
        "go-server-online-chat/initializers"
)
const (
        SERVER_HOST = "localhost"
        SERVER_PORT = "9988"
        SERVER_TYPE = "tcp"
)
var clients = make(map[string]net.Conn)
var userListBytes []byte

func init() {
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
        fmt.Println("Server Running...")
        server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
        if err != nil {
                fmt.Println("Error listening:", err.Error())
                os.Exit(1)
        }

        defer server.Close()
        
        fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
        fmt.Println("Waiting for client...")
        
        for {
                connection, err := server.Accept()
                if err != nil {
                        fmt.Println("Error accepting: ", err.Error())
                        os.Exit(1)
                }

                fmt.Println("client connected")
                go processClient(connection)
        }
}

func processClient(connection net.Conn) {
	defer connection.Close()

        username := ""
        users := getAllUsers()

	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}
                
                if strings.HasPrefix(strings.TrimSpace(string(buffer[:mLen])), "/username") {
                        msgprt := strings.SplitN(strings.TrimSpace(string(buffer[:mLen])), " ", 2)
                        username = msgprt[1]
                        continue 
                }
                
                if strings.HasPrefix(strings.TrimSpace(string(buffer[:mLen])), "/user") {
                        connection.Write([]byte(users))
                        continue 
                }

                clients[username] = connection

                fmt.Println("Reveived from "+ username +":", string(buffer[:mLen]))
                message := strings.TrimSpace(string(buffer[:mLen]))

                if strings.HasPrefix(message, "/register"){
                        messageParts := strings.SplitN(message, " ", 3)
                        if len(messageParts) == 3 {
                                register(messageParts[1], messageParts[2], connection)
                                continue
                        }
                } else if strings.HasPrefix(message, "/login"){
                        messageParts := strings.SplitN(message, " ", 3)
                        if len(messageParts) == 3 {
                                login(messageParts[1], messageParts[2], connection)
                                continue
                        }
                } else if strings.HasPrefix(message, "/private") {
			messageParts := strings.SplitN(message, " ", 3)
			recipient := messageParts[1]
			content := strings.Join(messageParts[2:], " ")
                        message := username + ": " + content
                        broadcastMessagePrivate(recipient, username, message, connection)			
		} else if  strings.HasPrefix(message, "/public") {
                        messageParts := strings.SplitN(message, " ", 2)
			broadcastMessagePublic(username, username + ": " + messageParts[1], connection)
		} else{
                        fmt.Println("nabıyon la")
                }
	}
}

func getAllUsers() string {
        var users []models.User
	result := initializers.DB.Find(&users)
        if result.Error != nil {
		log.Fatal(result.Error)
	}

        var userModel []string
	for _, user := range users {
                temp := strconv.FormatUint(uint64(user.ID),10) + " " + user.Name + " " + user.Password
		userModel = append(userModel, temp)
	}
        
        names := strings.Join(userModel, ",")
        names = "/user " + names
        return names
}

func register(username, password string, con net.Conn){
        user := models.User{Name: username, Password: password}
        result := initializers.DB.Create(&user)

        if result.Error != nil {
                fmt.Println("Register Failed")
                con.Write([]byte("/registerFailed"))
        } else {
                fmt.Println("Register Successful")
                con.Write([]byte("/registerSuccessful"))
        }
}

func login(loginUsername, password string, con net.Conn) {
        var user models.User
        initializers.DB.First(&user, "name = ?", loginUsername)

        if user.ID == 0 {
                fmt.Println("Login Failed")
                con.Write([]byte("/loginFailed"))
        } else if user.Password != password {
                fmt.Println("Invalid name or password")
                con.Write([]byte("/invalidCredentials"))
        } else {
                con.Write([]byte("/loginUsername " + loginUsername))
        }
}

func broadcastMessagePublic(sender, message string, con net.Conn) {
	for _, conn := range clients {
                        _, err := conn.Write([]byte("/Public " + message))
                        if err != nil {
                                fmt.Println("Error writing message:", err.Error())
                        }
	}
}

func broadcastMessagePrivate(recipient, username, message string, connection net.Conn) {
        recipientConn, ok := clients[recipient]
        senderConn := clients[username]

        if ok {        
                recipientConn.Write([]byte("/Private " + message))
                senderConn.Write([]byte("/Private " + message))
        } else {        
                fmt.Println("Recipient not found:", recipient)
        }
}