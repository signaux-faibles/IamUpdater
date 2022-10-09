package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

// addMissingRulesAndCardMembership
// Calcule et insère les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func addMissingRulesAndCardMembership(wekan libwekan.Wekan, users Users) error {
	wekanUsers := users.selectScopeWekan()
	fields := logger.DataForMethod("addMissingRulesAndCardMembership")
	logger.Info("> ajoute les nouvelles règles", fields)
	occurence := 0
	for _, user := range wekanUsers {
		fields.AddAny("username", user.email)
		logger.Debug(">> examine l'utilisateur", fields)
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
				if err := addCardMemberShip(wekan, user.email, board, label); err != nil {
					return err
				}

				logger.Debug(">>> s'assure de l'ajout à la taskforce", fields)
				if modified, err := wekan.EnsureRuleAddTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
					return err
				} else if modified {
					occurence++
					logger.Info(">>> crée de la règle d'ajout à la taskforce", fields)
				}

				logger.Debug(">>> s'assure de la retrait de la taskforce", fields)
				if modified, err := wekan.EnsureRuleRemoveTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
					return err
				} else if modified {
					occurence++
					logger.Info(">>> crée la règle de retrait de la taskforce", fields)
				}
			}
		}
	}
	if occurence == 0 {
		fields = logger.DataForMethod("addMissingRulesAndCardMembership")
		logger.Info(">> aucune règle à ajouter", fields)
	}
	return nil
}

// removeExtraRulesAndCardsMembership
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func removeExtraRulesAndCardsMembership(wekan libwekan.Wekan, users Users) error {
	wekanUsers := users.selectScopeWekan()
	fields := logger.DataForMethod("RemoveExtraRulesAndCardMembership")
	logger.Info("> supprime les règles obsolètes", fields)
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	deleted := 0
	for _, board := range domainBoards {
		fields.Remove("rule")
		fields.AddAny("board", board.Slug)
		logger.Debug(">> examine les règles du tableau", fields)

		fields.AddAny("board", board.Slug)
		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)

		if err != nil {
			return err
		}

		taskforceRules := append(rules.SelectAddMemberToTaskforceRule(), rules.SelectRemoveMemberFromTaskforceRule()...)
		for _, rule := range taskforceRules {
			fields.AddAny("rule", rule.Title)
			logger.Debug(">>> examine la règle", fields)

			label := board.GetLabelByID(rule.Trigger.LabelID)
			user := wekanUsers[Username(rule.Action.Username)]
			// l'utilisateur est absent de la config, du scope wekan ou de la board
			if !userHasTaskforceLabel(user)(label) || !contains(user.boards, string(board.Slug)) {
				err := removeCardMembership(wekan, Username(rule.Action.Username), board, label)
				if err != nil {
					return err
				}
				logger.Info(">>> supprime la règle", fields)
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
		logger.Info("> aucune règle à supprimer", fields)
	}
	return nil
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforces, string(label.Name)) }
}

func removeCardMembership(wekan libwekan.Wekan, username Username, board libwekan.Board, label libwekan.BoardLabel) error {
	fields := logger.DataForMethod("removeCardMembership")
	fields.AddAny("username", username)
	fields.AddAny("label", label.Name)
	fields.AddAny("board", board.Slug)
	logger.Debug(">>> examine les cartes", fields)
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
		logger.Info(">>> radie l'utilisateur des cartes", fields)
	}
	return nil
}

func addCardMemberShip(wekan libwekan.Wekan, username Username, board libwekan.Board, label libwekan.BoardLabel) error {
	fields := logger.DataForMethod("AddCardMembership")
	fields.AddAny("username", username)
	fields.AddAny("label", label.Name)
	fields.AddAny("board", board.Slug)
	logger.Debug(">>> examen des cartes", fields)
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
				occurences++
			}
		}
	}
	if occurences > 0 {
		fields.AddAny("occurences", occurences)
		logger.Info(">>> inscrit l'utilisateur sur les cartes", fields)
	}
	return nil
}
