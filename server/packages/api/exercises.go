package api

import (
	"encoding/json"
	"goapp/packages/config"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ExerciseInput struct {
	Input string `json:"input"`
}

type Exercise struct {
	Name string `json:"name"`
	// Type         string `json:"type"`
	// Muscle       string `json:"muscle"`
	// Difficulty   string `json:"difficulty"`
	// Instrcutions string `json:"instructions"`
}

func (a *App) Exercises(c *fiber.Ctx) error {
	client := &http.Client{}
	i := new(ExerciseInput)

	if err := c.BodyParser(i); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	req, err := http.NewRequest("GET", "https://api.api-ninjas.com/v1/exercises?name="+i.Input, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header["X-Api-Key"] = []string{config.Config[config.API_NINJA_API_KEY]}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	var exercises []Exercise
	err = json.Unmarshal([]byte(body), &exercises)
	if err != nil {
		return err
	}

	return c.JSON(&fiber.Map{"success": true, "exercises": exercises})
}
