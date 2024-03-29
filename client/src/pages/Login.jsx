import React, { useState } from "react";

import { Endpoints } from "../api/endpoints";
import Errors from "../components/Errors";
import { createCookie } from "../utils";

export default ({ history }) => {
    const [login, setLogin] = useState({
        username: "",
        password: "",
    });

    const [isSubmitting, setIsSubmitting] = useState(false);
    const [errors, setErrors] = useState([]);

    const { username, password } = login;

    const handleChange = (e) =>
        setLogin({ ...login, [e.target.name]: e.target.value });

    const handleSubmit = async (e) => {
        e.preventDefault();
        const { username, password } = login;
        try {
            setIsSubmitting(true);
            const res = await fetch(Endpoints.login, {
                method: "POST",
                body: JSON.stringify({
                    username,
                    password,
                }),
                headers: {
                    "Content-Type": "application/json",
                },
            });

            const { token, success, errors = [], user } = await res.json();
            if (success) {
                // creating a cookie expire in 30 minutes(same time as the token is invalidated on the backend)
                // ordinarily the setcookie from the server should suffice, however it has to be created here manually to bypass browsers
                // restriction on cross-site/non secure cookies on localhost.
                createCookie("token", token, 0.5);
                history.push({ pathname: "/logs", state: user });
            }
            setErrors(errors);
            if (errors[0].startsWith("UserNotConfirmedException"))
                history.push("/otp");
        } catch (e) {
            setErrors([e.toString()]);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <div className="wrapper">
                <h1>Login</h1>

                <input
                    className="input"
                    type="username"
                    placeholder="username"
                    value={username}
                    name="username"
                    onChange={handleChange}
                    required
                />

                <input
                    className="input"
                    type="password"
                    placeholder="password"
                    value={password}
                    name="password"
                    onChange={handleChange}
                    required
                />

                <button disabled={isSubmitting} type="submit">
                    {isSubmitting ? "....." : "login"}
                </button>
                <br />
                <a href="/register">{"create account"}</a>
                <br />
                <Errors errors={errors} />
            </div>
        </form>
    );
};
