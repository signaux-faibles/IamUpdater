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
	logger.Info("application des nouvelles règles", fields)
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
				logger.Debug("vérification des carte étiquetées", fields)
				modified, err := AddCardMemberShip(wekan, user, board, label)
				if err != nil {
					return err
				}
				if modified {
					logger.Info("inscription sur les cartes étiquetées", fields)
				}
				logger.Debug("s'assure de la présence de la règle", fields)
				modified, err = wekan.EnsureRuleExists(context.Background(), wekanUser, board, label)
				if err != nil {
					return err
				}
				if modified {
					logger.Info("création de la règle", fields)
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
	fields := logger.DataForMethod("RemoveExtraRulesAndCardMembership")
	logger.Info("effacement des règles obsolètes", fields)
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	for _, board := range domainBoards {
		fields.AddAny("board", board.Slug)
		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)
		taskforceRules := selectSlice(rules, IsTaskforceRule)
		if err != nil {
			return err
		}

		for _, rule := range taskforceRules {
			label := board.GetLabelByID(rule.Trigger.LabelID)
			fields.AddAny("label", label.Name)
			fields.AddAny("username", rule.Action.Username)

			user, found := wekanUsers[Username(rule.Action.Username)]
			if !found {
				err := wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return err
				}
				logger.Info("effacement de la règle", fields)
			}
			if !userHasTaskforceLabel(user)(label) {
				logger.Debug("examen des cartes étiquetées", fields)
				modified, err := RemoveCardMembership(wekan, user, board, label)
				if err != nil {
					return err
				}
				if modified {
					logger.Info("retrait de l'utilisateur des cartes", fields)
				}
				logger.Info("suppression de la règle", fields)
				err = wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforces, string(label.Name)) }
}

func RemoveCardMembership(wekan libwekan.Wekan, user User, board libwekan.Board, label libwekan.BoardLabel) (bool, error) {
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return false, err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })
	if err != nil {
		return false, err
	}
	var someModified bool
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberOutOfCard(context.Background(), card.ID, wekanUser.ID)
			someModified = modified || someModified
			if err != nil {
				return false, err
			}
		}
	}
	return someModified, nil
}

func AddCardMemberShip(wekan libwekan.Wekan, user User, board libwekan.Board, label libwekan.BoardLabel) (bool, error) {
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
	if err != nil {
		return false, err
	}
	cards, err := wekan.SelectCardsFromBoardID(context.Background(), board.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })

	if err != nil {
		return false, err
	}
	var modifiedSome bool
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberInCard(context.Background(), card.ID, wekanUser.ID)
			modifiedSome = modified || modifiedSome
			if err != nil {
				return false, err
			}
		}
	}
	return modifiedSome, nil
}

func IsTaskforceRule(rule libwekan.Rule) bool {
	return rule.Action.Username != "" &&
		rule.Trigger.LabelID != "" &&
		rule.Action.ActionType == "addMember" &&
		rule.Trigger.ActivityType == "addedLabel" &&
		rule.Trigger.UserID == "*"
}
