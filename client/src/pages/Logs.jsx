import React, { useState, useEffect } from "react";
import { Endpoints } from "../api/endpoints";
import Searchbar from "../components/Searchbar";
import { deleteCookie } from "../utils";
// import Checkbox from "../components/Checkbox";
// import Errors from "../components/Errors"

const Logs = ({ history }) => {
    const [user, setUser] = useState(null);
    const [isFetching, setIsFetching] = useState(false);
    const [workouts, setWorkouts] = useState([]);
    const [errors, setErrors] = useState([]);

    const headers = {
        Accept: "application/json",
        Authorization: document.cookie.split("token=")[1],
    };

    const getUserInfo = async () => {
        try {
            setIsFetching(true);
            const res = await fetch(Endpoints.workouts, {
                method: "GET",
                credentials: "include",
                headers,
            });

            if (!res.ok) logout();

            const { success, errors = [], user } = await res.json();
            setErrors(errors);
            if (!success) history.push("/login");
            setUser(user);
            const fetchWorkouts = await fetch(Endpoints.getWorkouts, {
                method: "POST",
                body: JSON.stringify({
                    UserID: user.id,
                }),
                headers: {
                    "Content-Type": "application/json",
                },
            });
        } catch (e) {
            setErrors([e.toString()]);
        } finally {
            setIsFetching(false);
        }
    };

    const logout = async () => {
        const res = await fetch(Endpoints.logout, {
            method: "GET",
            credentials: "include",
            headers,
        });

        if (res.ok) {
            deleteCookie("token");
            history.push("/login");
        }
    };

    useEffect(() => {
        console.log("verifying user info");
        getUserInfo();
    }, []);

    return (
        <div className="wrapper">
            <div>
                {isFetching ? (
                    <div>fetching details...</div>
                ) : (
                    <div>
                        {user && (
                            <div>
                                <h1>Welcome, {user && user.username}</h1>
                                <h1>Your Id is {user && user.id}</h1>
                                <p>{user && user.email}</p>
                                <br />
                                <Searchbar userID={user.id} />
                                <br />
                                <button onClick={logout}>logout</button>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

export default Logs;