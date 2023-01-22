import React, { useState, useEffect } from "react";
import { Endpoints } from "../api/endpoints";
import {
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
} from "@reach/combobox";
import "@reach/combobox/styles.css";
import Counter from "./Counter";

function useExerciseSearch(exerciseInput) {
    console.log(exerciseInput);
    const [exercises, setExercises] = React.useState([]);

    useEffect(() => {
        if (exerciseInput.trim() !== "") {
            let isFresh = true;
            async function fetchData() {
                const fetchedExercises = await fetchExercises(
                    exerciseInput.trim().replace(/ /g, "_")
                );
                console.log(fetchedExercises);
                if (isFresh) setExercises(fetchedExercises.exercises);
            }
            fetchData();
            return () => (isFresh = false);
        }
    }, [exerciseInput]);

    return exercises;
}

const cache = {};
async function fetchExercises(input) {
    console.log(input);
    if (cache[input]) {
        return cache[input];
    }
    const res = await fetch(Endpoints.getExercises, {
        method: "POST",
        body: JSON.stringify({ input }),
        headers: {
            "Content-Type": "application/json",
        },
    });
    const result = await res.json();
    console.log(result);
    cache[input] = result;
    return result;
}

const Searchbar = ({ userID }) => {
    const [exerciseInput, setExerciseInput] = useState("");
    const exercises = useExerciseSearch(exerciseInput);
    const [repCount, setRepCount] = useState(0);
    const [weight, setWeight] = useState(0);
    // const [weightSystem, setWeightSystem] = useState("lb");

    const recordRepData = (i) => {
        console.log(i);
        setRepCount(i);
    };
    const recordWeightData = (i) => {
        console.log(i);
        setWeight(i);
    };

    const handleSearchTermChange = (event) => {
        setExerciseInput(event.target.value);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        console.log(exerciseInput);
        try {
            const res = await fetch(Endpoints.addWorkout, {
                method: "POST",
                body: JSON.stringify({
                    UserID: parseInt(userID),
                    Exercise: exerciseInput,
                    Reps: repCount,
                    Weightlbs: weight,
                    Weightkg: Math.round((weight * 100.0) / 2.205) / 100,
                }),
                headers: {
                    "Content-Type": "application/json",
                },
            });

            const { success } = await res.json();
            if (success) {
                console.log("hi");
            }
        } catch (e) {
            console.log(e.toString());
        }
    };

    return (
        <div>
            <Combobox
                onSelect={(exercise) => {
                    console.log(exercise);
                    setExerciseInput(exercise);
                }}
                aria-label="Exercises"
            >
                <ComboboxInput
                    className="exercise-search-input"
                    onChange={handleSearchTermChange}
                    autocomplete={true}
                />
                {exercises && (
                    <ComboboxPopover className="shadow-popup">
                        {exercises.length > 0 ? (
                            <ComboboxList>
                                {exercises.map((exercise) => {
                                    const str = `${exercise.name}`;
                                    return (
                                        <ComboboxOption key={str} value={str} />
                                    );
                                })}
                            </ComboboxList>
                        ) : (
                            <span style={{ display: "block", margin: 8 }}>
                                No results found
                            </span>
                        )}
                    </ComboboxPopover>
                )}
            </Combobox>
            <form onSubmit={handleSubmit}>
                <Counter
                    recordRepData={recordRepData}
                    recordWeightData={recordWeightData}
                />
                <button type="submit">Add Workout</button>
            </form>
        </div>
    );
};

export default Searchbar;
