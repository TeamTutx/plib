package main

import (
	"fmt"

	"github.com/TeamTutx/plib/database/postgresql"
	"github.com/TeamTutx/plib/email"
	"github.com/go-pg/pg/v10"
)

func main() {
	DBTest()
	//EmailTest()
}

func DBTest() {
	var (
		conn   *pg.DB
		err    error
		userID []int
	)
	if conn, err = postgresql.Conn(false); err == nil {
		postgresql.StartLogging = true
		query := "SELECT user_id FROM pra_user ;"
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
