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
				if err := AddCardMemberShip(wekan, user.email, board, label); err != nil {
					return err
				}

				logger.Debug("s'assure de la règle d'ajout taskforce", fields)
				if modified, err := wekan.EnsureRuleAddTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
					return err
				} else if modified {
					logger.Info("création de la règle d'ajout taskforce", fields)
				}

				logger.Debug("s'assure de la règle de retrait taskforce", fields)
				if modified, err := wekan.EnsureRuleRemoveTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
					return err
				} else if modified {
					logger.Info("création de la règle de retrait taskforce", fields)
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
	logger.Info("examen des règles à supprimer", fields)
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	deleted := 0
	for _, board := range domainBoards {
		fields.Remove("rule")
		fields.AddAny("board", board.Slug)
		logger.Debug("examen des règles du tableau", fields)

		fields.AddAny("board", board.Slug)
		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)

		if err != nil {
			return err
		}

		taskforceRules := append(rules.SelectAddMemberToTaskforceRule(), rules.SelectRemoveMemberFromTaskforceRule()...)
		for _, rule := range taskforceRules {
			fields.AddAny("rule", rule.Title)
			logger.Debug("examen de la règle", fields)

			label := board.GetLabelByID(rule.Trigger.LabelID)
			user := wekanUsers[Username(rule.Action.Username)]
			// l'utilisateur est absent de la config, du scope wekan ou de la board
			if !userHasTaskforceLabel(user)(label) || !contains(user.boards, string(board.Slug)) {
				err := RemoveCardMembership(wekan, Username(rule.Action.Username), board, label)
				if err != nil {
					return err
				}
				logger.Info("suppression de la règle", fields)
				err = wekan.RemoveRuleWithID(context.Background(), rule.ID)
				if err != nil {
					return err
				}
				deleted += 1
			}
		}
	}
	if deleted == 0 {
		fields.Remove("board")
		fields.Remove("rule")
		logger.Info("aucune règle à supprimer", fields)
	}
	return nil
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforces, string(label.Name)) }
}

func RemoveCardMembership(wekan libwekan.Wekan, username Username, board libwekan.Board, label libwekan.BoardLabel) error {
	fields := logger.DataForMethod("RemoveCardMembership")
	fields.AddAny("username", username)
	fields.AddAny("label", label.Name)
	fields.AddAny("board", board.Slug)
	logger.Debug("examen des cartes", fields)
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(username))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })
	if err != nil {
		return err
	}
	var occurences int
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberOutOfCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
			if modified {
				occurences += 1
			}
		}
	}
	if occurences > 0 {
		fields.AddAny("occurences", occurences)
		logger.Info("désinscription des cartes", fields)
	}
	return nil
}

func AddCardMemberShip(wekan libwekan.Wekan, username Username, board libwekan.Board, label libwekan.BoardLabel) error {
	fields := logger.DataForMethod("AddCardMembership")
	fields.AddAny("username", username)
	fields.AddAny("label", label.Name)
	fields.AddAny("board", board.Slug)
	logger.Debug("examen des cartes", fields)
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(username))
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromBoardID(context.Background(), board.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })

	if err != nil {
		return err
	}
	var occurences int
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberInCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
			if modified {
				occurences += 1
			}
		}
	}
	if occurences > 0 {
		fields.AddAny("occurences", occurences)
		logger.Info("inscription sur les cartes", fields)
	}
	return nil
}
