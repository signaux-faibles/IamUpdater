package main

import (
	"context"
	"fmt"

	"github.com/signaux-faibles/libwekan"

	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

type Pipeline []PipelineStage

type PipelineStage struct {
	run func(libwekan.Wekan, Users) error
	id  string
}

func (pipeline Pipeline) Run(wekan libwekan.Wekan, fromConfig Users) error {
	for _, stage := range pipeline {
		err := stage.run(wekan, fromConfig)
		if err != nil {
			return PipelineRunError{
				err:   err,
				stage: stage,
			}
		}
	}
	return nil
}

func (pipeline Pipeline) StopAfter(wekan libwekan.Wekan, fromConfig Users, lastStage PipelineStage) error {
	fields := logger.ContextForMethod(pipeline.StopAfter)
	for _, stage := range pipeline {
		fields.AddAny("stage", stage.id)
		logger.Debug("applique le pipeline", fields)
		err := stage.run(wekan, fromConfig)
		if err != nil || stage.id == lastStage.id {
			return err
		}
	}
	return nil
}

var stageCheckBoardSlugs = PipelineStage{checkBoardSlugs, "checkBoardSlugs"}
var stageManageUsers = PipelineStage{manageUsers, "manageUsers"}
var stageManageBoardsMembers = PipelineStage{manageBoardsMembers, "manageBoardsMembers"}
var stageAddMissingRulesAndCardMembership = PipelineStage{addMissingRulesAndCardMembership, "addMissingRulesAndCardMembership"}
var stageRemoveExtraRulesAndCardMembership = PipelineStage{removeExtraRulesAndCardsMembership, "RemoveExtraRulesAndCardMembership"}
var stageCheckNativeUsers = PipelineStage{checkNativeUsers, "checkNativeUsers"}

var pipeline = Pipeline{
	stageCheckBoardSlugs,
	stageCheckNativeUsers,
	stageManageUsers,
	stageManageBoardsMembers,
	stageAddMissingRulesAndCardMembership,
	stageRemoveExtraRulesAndCardMembership,
}

func WekanUpdate(url, database, admin string, users Users, slugDomainRegexp string) error {
	wekan, err := initWekan(url, database, admin, slugDomainRegexp)

	if err != nil {
		return err
	}
	return pipeline.Run(wekan, users.selectScopeWekan())
}

func initWekan(url string, database string, admin string, slugDomainRegexp string) (libwekan.Wekan, error) {
	wekan, err := libwekan.Init(context.Background(), url, database, libwekan.Username(admin), slugDomainRegexp)
	if err != nil {
		return libwekan.Wekan{}, err
	}

	if err := wekan.Ping(context.Background()); err != nil {
		return libwekan.Wekan{}, err
	}
	err = wekan.AssertPrivileged(context.Background())
	if err != nil {
		return libwekan.Wekan{}, err
	}
	return wekan, nil
}

func checkBoardSlugs(wekan libwekan.Wekan, users Users) error {
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	domainActiveBoards := selectSlice(domainBoards, func(board libwekan.Board) bool { return !board.Archived })
	boardToSlug := func(board libwekan.Board) libwekan.BoardSlug { return board.Slug }
	domainActiveSlugs := mapSlice(domainActiveBoards, boardToSlug)

	configBoardsMembers := users.inferBoardsMember()
	var configSlugs []libwekan.BoardSlug
	for slug := range configBoardsMembers {
		configSlugs = append(configSlugs, slug)
	}
	_, _, onlyConfig := intersect(domainActiveSlugs, configSlugs)

	if len(onlyConfig) > 0 {
		return InvalidExcelFileError{msg: fmt.Sprintf("le fichier contient des références de boards inexistantes : %s", onlyConfig)}
	}
	return nil
}
