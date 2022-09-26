package main

import (
	"context"

	"github.com/signaux-faibles/libwekan"
)

type WekanBoardsUsers map[BoardSlug][]User

func WekanUpdate(kc *KeycloakContext, url, database, admin string) error {

	wekan, err := libwekan.Connect(context.TODO(), url, database, libwekan.Username(admin))
	if err != nil {
		return err
	}
	_, err = wekan.AdminUser(context.Background())
	return err
}

//func listUsers(board libwekan.Board) []libwekan.Username{
//return []{}
//}

func (users Users) listBoards() WekanBoardsUsers {
	wekanBoardsUsers := make(WekanBoardsUsers)
	for _, user := range users {
		for _, board := range user.boards {
			if board != "" {
				boardSlug := BoardSlug(board)
				wekanBoardsUsers[boardSlug] = append(wekanBoardsUsers[boardSlug], user)
			}
		}
	}
	return wekanBoardsUsers
}
