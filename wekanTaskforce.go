package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

// AddMissingRulesAndCardMembership
// Calcule et insère les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func AddMissingRulesAndCardMembership(wekan libwekan.Wekan, users Users) error {
	wekanUsers := users.selectScopeWekan()
	fields := logger.DataForMethod("AddMissingRulesAndCardMembership")
	for _, user := range wekanUsers {
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
				logger.Info("ajoute l'utilisateur sur les cartes portant l'étiquette correspondante", fields)
				err = AddCardMemberShip(wekan, user, board, label)
				if err != nil {
					return err
				}
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

// RemoveExtraRulesAndCardsMembership
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func RemoveExtraRulesAndCardsMembership(wekan libwekan.Wekan, users Users) error {
	wekanUsers := users.selectScopeWekan()
	fields := logger.DataForMethod("AddMissingRulesAndCardMembership")
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	for _, board := range domainBoards {
		fields.AddAny("board", board.Slug)

		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)
		if err != nil {
			return err
		}

		for _, rule := range rules {
			fields.AddAny("rule", rule)

			user, found := wekanUsers[Username(rule.Action.Username)]
			if !found {
				err := wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return nil
				}
			}
			label := board.GetLabelByID(rule.Trigger.LabelID)
			if !userHasTaskforceLabel(user)(label) {
				logger.Info("retrait de l'utilisateur sur les cartes ayant l'étiquette concernée", fields)
				err = RemoveCardMembership(wekan, user, board, label)
				if err != nil {
					return nil
				}
				logger.Info("suppression de la règle", fields)
				err := wekan.RemoveRuleWithID(context.Background(), rule.ID)
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
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })
	if err != nil {
		return err
	}
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			err = wekan.EnsureMemberOutOfCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AddCardMemberShip(wekan libwekan.Wekan, user User, board libwekan.Board, label libwekan.BoardLabel) error {
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromBoardID(context.Background(), board.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })

	if err != nil {
		return err
	}
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			err := wekan.EnsureMemberInCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
