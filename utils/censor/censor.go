package censor

import (
	"context"
	"fmt"
	"time"

	ent "haruki-database/database/schema/censor"
	"haruki-database/database/schema/censor/namelog"
	"haruki-database/database/schema/censor/result"
	"haruki-database/database/schema/censor/shortbio"
	"haruki-database/utils"
	"haruki-database/utils/logger"
)

type ResultStatus string

const (
	ResultCompliant    ResultStatus = "合规"
	ResultNonCompliant ResultStatus = "不合规"
)

type Service struct {
	Client    *ent.Client
	CensorAPI *BaiduTextCensorClient
	Logger    *logger.Logger
}

func (s *Service) CensorName(ctx context.Context, imUserID string, userID string, name string, server string) bool {
	serverEnum, _ := utils.ParseBindingServer(server)
	if name == "" || serverEnum == utils.BindingServerCN {
		return true
	}

	existing, err := s.Client.Result.
		Query().
		Where(result.NameEQ(name)).
		Only(ctx)
	if err == nil && existing != nil {
		if existing.Result != nil {
			return *existing.Result == 1
		}
		return false
	}

	data, err := s.CensorAPI.TextCensor(name)
	if err != nil {
		s.Logger.Errorf("审核名字失败1: %v", err)
		return false
	}

	censorResult := 0
	if conclusion, ok := data["conclusion"].(string); ok && conclusion == string(ResultCompliant) {
		censorResult = 1
	} else {
		s.Logger.Debugf("名字审核不通过: imID: %s", imUserID)
	}

	_, err = s.Client.Result.
		Create().
		SetName(name).
		SetResult(censorResult).
		Save(ctx)
	if err != nil {
		s.Logger.Errorf("插入 censor_result 失败: %v", err)
	}

	exists, _ := s.Client.NameLog.
		Query().
		Where(
			namelog.UserIDEQ(fmt.Sprint(userID)),
			namelog.NameEQ(name),
			namelog.ImUserIDEQ(imUserID),
		).
		Exist(ctx)
	if !exists {
		text := string(ResultCompliant)
		if censorResult == 0 {
			text = string(ResultNonCompliant)
		}
		_, err := s.Client.NameLog.
			Create().
			SetUserID(fmt.Sprint(userID)).
			SetName(name).
			SetImUserID(imUserID).
			SetResult(text).
			SetTime(time.Now()).
			Save(ctx)
		if err != nil {
			s.Logger.Errorf("插入 name_log 失败: %v", err)
		}
	}

	return censorResult == 1
}

func (s *Service) CensorShortBio(ctx context.Context, imUserID string, userID string, content string, server string) bool {
	serverEnum, _ := utils.ParseBindingServer(server)
	if content == "" || serverEnum == utils.BindingServerCN {
		return true
	}

	existing, err := s.Client.ShortBio.
		Query().
		Where(shortbio.ContentEQ(content)).
		Only(ctx)
	if err == nil && existing != nil {
		if existing.Result != nil {
			return *existing.Result == string(ResultCompliant)
		}
		return false
	}

	data, err := s.CensorAPI.TextCensor(content)
	if err != nil {
		s.Logger.Errorf("审核短句失败1: %v", err)
		return false
	}

	censorResult := ResultNonCompliant
	if conclusion, ok := data["conclusion"].(string); ok && conclusion == string(ResultCompliant) {
		censorResult = ResultCompliant
	}

	_, err = s.Client.ShortBio.
		Create().
		SetUserID(fmt.Sprint(userID)).
		SetContent(content).
		SetImUserID(imUserID).
		SetResult(string(censorResult)).
		Save(ctx)
	if err != nil {
		s.Logger.Errorf("插入 short_bio 失败: %v", err)
	}

	return censorResult == ResultCompliant
}

func NewService(apiKey, secretKey string, client *ent.Client) *Service {
	censorAPI := NewBaiduTextCensorClient(apiKey, secretKey)
	return &Service{
		Client:    client,
		CensorAPI: censorAPI,
		Logger:    logger.NewLogger("HarukiContentCensorService", "INFO", nil),
	}
}
