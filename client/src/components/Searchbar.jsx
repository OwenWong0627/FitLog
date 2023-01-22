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

const Searchbar = () => {
    const [exerciseInput, setExerciseInput] = useState("");
    const exercises = useExerciseSearch(exerciseInput);

    const handleSearchTermChange = (event) => {
        setExerciseInput(event.target.value);
    };

    return (
        <div>
            <Combobox aria-label="Exercises">
                <ComboboxInput
                    className="exercise-search-input"
                    onChange={handleSearchTermChange}
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
        </div>
    );
};

export default Searchbar;
