package main

import (
	"context"
	"github.com/signaux-faibles/libwekan"
)

func WekanUpdate(url, database, admin string) error {

	wekan, err := libwekan.Connect(context.TODO(), url, database, libwekan.Username(admin))
	if err != nil {
		return err
	}

	//wekan.
	return wekan.Ping()

}

//func listUsers(board libwekan.Board) []libwekan.Username{
//return []{}
//}

func (users Users) listBoards() map[string][]User {
	return nil
}
