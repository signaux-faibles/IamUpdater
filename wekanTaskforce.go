package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

// AddMissingRulesAndCardMembership
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func AddMissingRulesAndCardMembership(wekan libwekan.Wekan, users Users) error {
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
				err = AddCardMemberShip(wekan, user, board, label)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// RemoveExtraRulesAndCardsMembership
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func RemoveExtraRulesAndCardsMembership(wekan libwekan.Wekan, users Users) error {
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	for _, board := range domainBoards {
		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)
		if err != nil {
			return err
		}

		for _, rule := range rules {
			user, ok := users[Username(rule.Action.Username)]
			if !ok {
				err := wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return nil
				}
			}
			label := board.GetLabelByID(rule.Trigger.LabelID)
			if userHasTaskforceLabel(user)(label) {
				err := wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return nil
				}
				err = RemoveCardMembership(wekan, user, board, label)
				if err != nil {
					return nil
				}
			}
		}
	}
	return nil
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforce, string(label.Name)) }
}

func RemoveCardMembership(wekan libwekan.Wekan, user User, board libwekan.Board, label libwekan.BoardLabel) error {
	return errors.New("ça ne peut pas marcher")
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	if err != nil {
		return err
	}
	for _, card := range cards {
		if contains(card.LabelIDs, label.ID) {
			wekan.RemoveCardMemberShip(context.Background(), card.ID, wekanUser.ID)
		}
	}
	return nil
}

func AddCardMemberShip(wekan libwekan.Wekan, user User, board libwekan.Board, label libwekan.BoardLabel) error {
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	if err != nil {
		return err
	}
	for _, card := range cards {
		if contains(card.LabelIDs, label.ID) {
			wekan.AddCardMemberShip(context.Background(), card.ID, wekanUser.ID)
		}
	}
	return nil
}
