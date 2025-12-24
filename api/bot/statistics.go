package bot

import (
	"context"
	"time"

	"haruki-database/api"
	"haruki-database/database/schema/bot"
	"haruki-database/database/schema/bot/dailyrequests"
	"haruki-database/database/schema/bot/hourlyrequests"
	"haruki-database/database/schema/bot/requestsranking"

	"github.com/gofiber/fiber/v3"
)

func (h *StatisticsHandler) RecordStatistics(c fiber.Ctx) error {
	botID := fiber.Params[int](c, "botID", 0)
	if botID <= 0 {
		return api.JSONResponse(c, fiber.StatusBadRequest, "botID required")
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return api.InternalError(c)
	}
	now := time.Now().In(loc)
	ctx := context.Background()
	if err := h.updateRequestsRanking(ctx, botID); err != nil {
		return api.InternalError(c)
	}
	if err := h.updateHourlyRequests(ctx, now); err != nil {
		return api.InternalError(c)
	}
	if err := h.updateDailyRequests(ctx, now, loc); err != nil {
		return api.InternalError(c)
	}
	return api.JSONResponse(c, fiber.StatusOK, "Statistics recorded")
}

func (h *StatisticsHandler) updateRequestsRanking(ctx context.Context, botID int) error {
	rank, err := h.svc.client.RequestsRanking.
		Query().
		Where(requestsranking.BotIDEQ(botID)).
		Only(ctx)
	if err == nil && rank != nil {
		_, err = h.svc.client.RequestsRanking.
			UpdateOne(rank).
			SetCounts(rank.Counts + 1).
			Save(ctx)
	} else {
		_, err = h.svc.client.RequestsRanking.
			Create().
			SetBotID(botID).
			SetCounts(1).
			Save(ctx)
	}
	return err
}

func (h *StatisticsHandler) updateHourlyRequests(ctx context.Context, now time.Time) error {
	hourKey := now.Truncate(time.Hour)
	hourly, err := h.svc.client.HourlyRequests.
		Query().
		Where(hourlyrequests.HourKeyEQ(hourKey)).
		Only(ctx)
	if err == nil && hourly != nil {
		_, err = h.svc.client.HourlyRequests.
			UpdateOne(hourly).
			SetCount(hourly.Count + 1).
			Save(ctx)
	} else {
		_, err = h.svc.client.HourlyRequests.
			Create().
			SetHourKey(hourKey).
			SetCount(1).
			Save(ctx)
	}
	return err
}

func (h *StatisticsHandler) updateDailyRequests(ctx context.Context, now time.Time, loc *time.Location) error {
	dateKey := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	daily, err := h.svc.client.DailyRequests.
		Query().
		Where(dailyrequests.DateKeyEQ(dateKey)).
		Only(ctx)
	if err == nil && daily != nil {
		_, err = h.svc.client.DailyRequests.
			UpdateOne(daily).
			SetCount(daily.Count + 1).
			Save(ctx)
	} else {
		_, err = h.svc.client.DailyRequests.
			Create().
			SetDateKey(dateKey).
			SetCount(1).
			Save(ctx)
	}
	return err
}

func registerStatisticsRoutes(app *fiber.App, client *bot.Client) {
	svc := NewStatisticsService(client)
	h := NewStatisticsHandler(svc)

	app.Post("/bot/statistics/record/:botID", api.VerifyAPIAuthorization(), h.RecordStatistics)
}
