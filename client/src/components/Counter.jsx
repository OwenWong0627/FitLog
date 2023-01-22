import React, { useState } from "react";

const Counter = ({ recordRepData, recordWeightData }) => {
    const [repCount, setRepCount] = useState(0);
    const [weight, setWeight] = useState(0);
    const handleRepDecrement = () => {
        setRepCount(repCount - 1);
        recordRepData(repCount - 1);
    };
    const handleRepIncrement = () => {
        setRepCount(repCount + 1);
        recordRepData(repCount + 1);
    };
    const handleRepChange = (e) => {
        setRepCount(e.target.value);
        recordRepData(e.target.value);
    };
    const handleWeightDecrement = () => {
        setWeight(weight - 2.5);
        recordWeightData(weight - 2.5);
    };
    const handleWeightIncrement = () => {
        setWeight(weight + 2.5);
        recordWeightData(weight + 2.5);
    };
    const handleWeightChange = (e) => {
        setWeight(e.target.value);
        recordWeightData(e.target.value);
    };

    return (
        <div>
            <div display="flex">
                <button
                    disabled={repCount === 0}
                    onClick={handleRepDecrement}
                    type="button"
                >
                    -
                </button>
                <input onChange={handleRepChange} value={repCount} />
                <button onClick={handleRepIncrement} type="button">
                    +
                </button>
            </div>
            <div display="flex">
                <button
                    disabled={weight === 0}
                    onClick={handleWeightDecrement}
                    type="button"
                >
                    -
                </button>
                <input onChange={handleWeightChange} value={weight} />
                <button onClick={handleWeightIncrement} type="button">
                    +
                </button>
            </div>
        </div>
    );
};

export default Counter;
