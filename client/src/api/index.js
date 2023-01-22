import axios from "axios"

export async function getExercises(exerciseName) {
    try {
      const response = await axios.get(`https://api.api-ninjas.com/v1/exercises?muscle=${exerciseName}&`);
      console.log(response);
    } catch (error) {
      console.error(error);
    }
}