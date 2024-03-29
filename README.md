# FitLog

FitLog is a workout tracker that offers a straightforward way to record workouts. It is a web app designed for people who have a general sense of what they want to do in the gym but just want an easy way to track their progress and lifetime Personal Records. It also comes with other features that help show and record different metrics for your body.

- **Front-end**: JavaScript, React
- **Back-end**: Go, Fiber
- **Database**: PostgreSQL
- **Tools**: AWS, Docker 

Inspired by [**this**](https://github.com/aesrael/go-postgres-react-starter) go-postgres-react starter code.

Exercise List retrieved from API Ninja's [**Exercise API**](https://api-ninjas.com/api/exercises)

### To this project using docker
Ensure you have `docker` installed

```bash
make docker-build
make docker-run
```
Server is live on `:8081` and UI is on `:3000`
