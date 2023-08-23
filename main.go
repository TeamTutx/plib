package main

import (
	"fmt"

	"github.com/go-pg/pg/v10"
	"gitlab.com/g-harshit/plib/database/postgresql"
	"gitlab.com/g-harshit/plib/email"
)

func main() {
	//DBTest()
	EmailTest()
}

func DBTest() {
	var (
		conn   *pg.DB
		err    error
		userID []int
	)
	if conn, err = postgresql.Conn(false); err == nil {
		postgresql.StartLogging = true
		query := "SELECT user_id FROM pra_use ;"
		if _, err = conn.Query(&userID, query); err == nil {
			fmt.Println("SUCCESS \n", userID)
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}
}

func EmailTest() {
	email.NewBrevoEmailService()
	email.GetBrevoEmailService().Send(email.EmailModel{
		FromMail:   "gharshit1237@gmail.com",
		ToMail:     []string{"gharshit12371@gmail.com"},
		Subject:    "Hello Mail",
		Body:       "Hi EveryOne",
		ReplyTo:    []string{"gharshit1237@gmail.com"},
		SenderName: "TutxWorld",
	})
}
