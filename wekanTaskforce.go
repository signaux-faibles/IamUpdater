package main

import (
	"context"

	"github.com/signaux-faibles/libwekan"

	"keycloakUpdater/v2/pkg/logger"
)

// addMissingRulesAndCardMembership
// Calcule et insère les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func addMissingRulesAndCardMembership(wekan libwekan.Wekan, users Users) error {
	logContext := logger.ContextForMethod(addMissingRulesAndCardMembership)
	logger.Info("> ajoute les nouvelles règles", logContext)
	occurence := 0
	for _, user := range users {
		logContext.AddAny("username", user.email)
		logger.Debug(">> examine l'utilisateur", logContext)
		wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
		if err != nil {
			return err
		}

		for _, boardSlug := range user.boards {
			logContext.AddString("board", boardSlug)
			logger.Debug(">>> examine le tableau", logContext)
			board, err := wekan.GetBoardFromSlug(context.Background(), libwekan.BoardSlug(boardSlug))
			if err != nil {
				return err
			}
			labels := selectSlice(board.Labels, userHasTaskforceLabel(user))

			for _, label := range labels {
				if err := addCardMemberShip(wekan, wekanUser, board, label); err != nil {
					return err
				}

				if modified, err := EnsureRuleAddTaskforceMemberExists(wekan, wekanUser, board, label); err != nil {
					return err
				} else {
					occurence += modified
				}

				if modified, err := EnsureRuleRemoveTaskforceMemberExists(wekan, wekanUser, board, label); err != nil {
					return err
				} else {
					occurence += modified
				}
			}
		}
	}
	if occurence == 0 {
		logContext = logger.ContextForMethod(addMissingRulesAndCardMembership)
		logger.Info("> aucune règle à ajouter", logContext)
	}
	return nil
}

func EnsureRuleAddTaskforceMemberExists(wekan libwekan.Wekan, wekanUser libwekan.User, board libwekan.Board, label libwekan.BoardLabel) (int, error) {
	logContext := logger.ContextForMethod(EnsureRuleAddTaskforceMemberExists)
	logContext.AddAny("username", wekanUser.Username)
	logContext.AddAny("board", board.Slug)
	logContext.AddAny("label", label.Name)
	logger.Debug(">>> s'assure de l'ajout à la taskforce", logContext)
	if modified, err := wekan.EnsureRuleAddTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
		return 0, err
	} else if modified {
		logger.Notice(">>> crée de la règle d'ajout à la taskforce", logContext)
		return 1, nil
	}
	return 0, nil
}

func EnsureRuleRemoveTaskforceMemberExists(
	wekan libwekan.Wekan,
	wekanUser libwekan.User,
	board libwekan.Board,
	label libwekan.BoardLabel,
) (int, error) {
	logContext := logger.ContextForMethod(EnsureRuleRemoveTaskforceMemberExists).
		AddAny("username", wekanUser.Username).
		AddAny("board", board.Slug).
		AddAny("label", label.Name)
	logger.Debug(">>> s'assure du retrait de la taskforce", logContext)
	if modified, err := wekan.EnsureRuleRemoveTaskforceMemberExists(context.Background(), wekanUser, board, label); err != nil {
		return 0, err
	} else if modified {
		logger.Notice(">>> crée la règle de retrait de la taskforce", logContext)
		return 1, nil
	}
	return 0, nil
}

// removeExtraRulesAndCardsMembership
// Calcule et insert les règles manquantes pour correspondre à la configuration Users
// Ajuste la participation des utilisateurs aux cartes concernées par les labels en cas de changement
func removeExtraRulesAndCardsMembership(wekan libwekan.Wekan, users Users) error {
	logContext := logger.ContextForMethod(removeExtraRulesAndCardsMembership)
	logger.Info("> supprime les règles obsolètes", logContext)
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	deleted := 0
	for _, board := range domainBoards {
		logContext.Remove("rule")
		logContext.AddAny("board", board.Slug)
		logger.Debug(">> examine les règles du tableau", logContext)
		rules, err := wekan.SelectRulesFromBoardID(context.Background(), board.ID)
		if err != nil {
			return err
		}

		taskforceRules := append(rules.SelectAddMemberToTaskforceRule(), rules.SelectRemoveMemberFromTaskforceRule()...)
		for _, rule := range taskforceRules {
			logContext.AddString("rule", rule.Title)
			logger.Debug(">>> examine la règle", logContext)

			label := board.GetLabelByID(rule.Trigger.LabelID)
			logContext.AddAny("label", label)
			username := Username(rule.Action.Username)
			user := users[Username(username)]
			logContext.AddAny("username", username)
			// l'utilisateur est absent de la config, du scope wekan ou de la board
			if !userHasTaskforceLabel(user)(label) || !contains(user.boards, string(board.Slug)) {
				if err := removeCardMembership(wekan, rule.Action.Username, board, label); err != nil {
					return err
				}
				logger.Notice(">>> supprime la règle", logContext)
				if err := wekan.RemoveRuleWithID(context.Background(), rule.ID); err != nil {
					return err
				}
				deleted += 1
			}
		}
	}
	if deleted == 0 {
		logContext.Remove("board")
		logContext.Remove("rule")
		logger.Notice("> aucune règle à supprimer", logContext)
	}
	return nil
}

func userHasTaskforceLabel(user User) func(label libwekan.BoardLabel) bool {
	return func(label libwekan.BoardLabel) bool { return contains(user.taskforces, string(label.Name)) }
}

func removeCardMembership(wekan libwekan.Wekan, wekanUsername libwekan.Username, board libwekan.Board, label libwekan.BoardLabel) error {
	logContext := logger.ContextForMethod(removeCardMembership).
		AddAny("username", wekanUsername).
		AddAny("label", label.Name).
		AddAny("board", board.Slug)
	logger.Debug(">>> examine les cartes", logContext)
	wekanUser, err := wekan.GetUserFromUsername(context.Background(), wekanUsername)
	if err != nil {
		return err
	}
	cards, err := wekan.SelectCardsFromMemberID(context.Background(), wekanUser.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })
	if err != nil {
		return err
	}
	var occurences int
	cardsToBeRemoved := make([]string, 0)
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberOutOfCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
			if modified {
				cardsToBeRemoved = append(cardsToBeRemoved, card.Title)
				occurences += 1
			}
		}
	}
	if occurences > 0 {
		logContext.AddAny("occurences", occurences)
		logger.Info(
			">>> radie l'utilisateur des cartes",
			logContext.AddInt("occurences", occurences).AddArray("cards", cardsToBeRemoved),
		)
	}
	return nil
}

func addCardMemberShip(wekan libwekan.Wekan, wekanUser libwekan.User, board libwekan.Board, label libwekan.BoardLabel) error {
	logContext := logger.ContextForMethod(addCardMemberShip).
		AddAny("username", wekanUser.Username).
		AddAny("label", label.Name).
		AddAny("board", board.Slug)
	logger.Debug(">>> examen des cartes", logContext)

	cards, err := wekan.SelectCardsFromBoardID(context.Background(), board.ID)
	boardCards := selectSlice(cards, func(card libwekan.Card) bool { return card.BoardID == board.ID })

	if err != nil {
		return err
	}
	var occurences int
	cardsToBeAdded := make([]string, 0)
	for _, card := range boardCards {
		if contains(card.LabelIDs, label.ID) {
			modified, err := wekan.EnsureMemberInCard(context.Background(), card.ID, wekanUser.ID)
			if err != nil {
				return err
			}
			if modified {
				cardsToBeAdded = append(cardsToBeAdded, card.Title)
				occurences++
			}
		}
	}
	if occurences > 0 {
		logger.Notice(
			">>> inscrit l'utilisateur sur les cartes",
			logContext.AddInt("occurences", occurences).AddArray("cards", cardsToBeAdded),
		)
	}
	return nil
}
