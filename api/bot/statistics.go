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

func recordStatistics(client *bot.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		botID := fiber.Params[int](c, "botID", 0)
		if botID <= 0 {
			return api.JSONResponse(c, fiber.StatusBadRequest, "botID required", nil)
		}

		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return api.JSONResponse(c, fiber.StatusInternalServerError, "Failed to load timezone", nil)
		}
		now := time.Now().In(loc)
		ctx := context.Background()
		rank, err := client.RequestsRanking.
			Query().
			Where(requestsranking.BotIDEQ(botID)).
			Only(ctx)
		if err == nil && rank != nil {
			_, err = client.RequestsRanking.
				UpdateOne(rank).
				SetCounts(rank.Counts + 1).
				Save(ctx)
		} else {
			_, err = client.RequestsRanking.
				Create().
				SetBotID(botID).
				SetCounts(1).
				Save(ctx)
		}
		if err != nil {
			return api.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update requests ranking", nil)
		}

		hourly, err := client.HourlyRequests.
			Query().
			Where(hourlyrequests.HourKeyEQ(now.Truncate(time.Hour))).
			Only(ctx)
		if err == nil && hourly != nil {
			_, err = client.HourlyRequests.
				UpdateOne(hourly).
				SetCount(hourly.Count + 1).
				Save(ctx)
		} else {
			_, err = client.HourlyRequests.
				Create().
				SetHourKey(now.Truncate(time.Hour)).
				SetCount(1).
				Save(ctx)
		}
		if err != nil {
			return api.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update hourly requests", nil)
		}

		dateKey := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		daily, err := client.DailyRequests.
			Query().
			Where(dailyrequests.DateKeyEQ(dateKey)).
			Only(ctx)
		if err == nil && daily != nil {
			_, err = client.DailyRequests.
				UpdateOne(daily).
				SetCount(daily.Count + 1).
				Save(ctx)
		} else {
			_, err = client.DailyRequests.
				Create().
				SetDateKey(dateKey).
				SetCount(1).
				Save(ctx)
		}
		if err != nil {
			return api.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update daily requests", nil)
		}

		return api.JSONResponse(c, fiber.StatusOK, "Statistics recorded", nil)
	}
}

func registerStatisticsRoutes(app *fiber.App, client *bot.Client) {
	app.Post("/bot/statistics/record/:botID", api.VerifyAPIAuthorization(), recordStatistics(client))
}
