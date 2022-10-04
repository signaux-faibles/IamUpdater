package main

import (
	"context"
	"errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

// AddMissingRules
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
func AddMissingRules(wekan libwekan.Wekan, users Users) error {
	fields := logger.DataForMethod("AddMissingRules")
	for _, user := range users {
		fields.AddAny("username", user.email)
		wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
		if err != nil {
			return err
		}

		for _, boardSlug := range user.boards {
			board, err := wekan.GetBoardFromSlug(context.Background(), libwekan.BoardSlug(boardSlug))
			fields.AddAny("board", boardSlug)
			if err != nil {
				return err
			}
			labels := selectSlice(board.Labels, userHasTaskforceLabel(user))

			for _, label := range labels {
				fields.AddAny("label", label.Name)
				logger.Info("s'assure de la présence de la règle", fields)
				err := wekan.EnsureRuleExists(context.Background(), wekanUser, board, label)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func RemoveExtraRules(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}

func AddMissingCardsMembers(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}

func RemoveExtraCardsMembers(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforce, string(label.Name)) }
}
