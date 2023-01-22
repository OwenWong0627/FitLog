package api

import (
	"database/sql"
	"fmt"
	"goapp/packages/db"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Workout struct {
	UserID    int     `json:"userid"`
	Exercise  string  `json:"exercise"`
	Reps      int     `json:"reps"`
	Weightlbs float32 `json:"weightlbs"`
	Weightkg  float32 `json:"weightkg"`
}

type UserID struct {
	UserID int `json:"userid"`
}

func (a *App) AddWorkout(c *fiber.Ctx, dbConn *sql.DB) error {
	w := new(Workout)

	if err := c.BodyParser(w); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	_, err := dbConn.Query(db.AddWorkout, w.UserID, w.Exercise, w.Reps, w.Weightlbs, w.Weightkg)
	if err != nil {
		// handle error
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Add Workout failed"}})
	}

	return c.JSON(&fiber.Map{"success": true})
}

func (a *App) GetWorkouts(c *fiber.Ctx, dbConn *sql.DB) error {
	w := new(UserID)

	if err := c.BodyParser(w); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	fmt.Println(w)

	// Need to retrieve an array of workouts acquired from multiple rows from the table
	workout := &[]Workout{}
	// if workout, err := dbConn.Query(db.GetUserByUsernameQuery, w.UserID).
	// 	Scan(&workout.ID, &workout.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
	// 	if err == sql.ErrNoRows {
	// 		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"success": false, "errors": []string{"Retrieve Workout Failed"}})
	// 	}
	// }

	return c.JSON(&fiber.Map{"success": true, "workouts": workout})
}
