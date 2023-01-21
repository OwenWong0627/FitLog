import React, { useState } from 'react'
import { Endpoints } from '../api/endpoints'
import Errors from '../components/Errors'

export default ({ history }) => {
  const [user, setUser] = useState({
    email: '',
    password: '',
    username: '',
  })

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errors, setErrors] = useState([])
  const { email, password, username } = user

  const handleChange = (e) =>
    setUser({ ...user, [e.target.name]: e.target.value })

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      setIsSubmitting(true)
      const res = await fetch(Endpoints.register, {
        method: 'POST',
        body: JSON.stringify({
          email,
          password,
          username,
        }),
        headers: {
          'Content-Type': 'application/json',
        },
      })
      const { success, errors = [] } = await res.json()

      if (success) history.push('/otp')

      setErrors(errors)
    } catch (e) {
      setErrors([e.toString()])
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <div className="wrapper">
        <h1>Register</h1>
        <input
          className="input"
          type="username"
          placeholder="Username"
          value={username}
          name="username"
          onChange={handleChange}
          required
        />
        <input
          className="input"
          type="email"
          placeholder="Email"
          value={email}
          name="email"
          onChange={handleChange}
          required
        />
        <input
          className="input"
          type="password"
          placeholder="Password"
          value={password}
          name="password"
          onChange={handleChange}
          required
        />

        <button disabled={isSubmitting} onClick={handleSubmit}>
          {isSubmitting ? '.....' : 'Sign Up'}
        </button>
        <br />
        <a href="/login">{'login'}</a>
        <br />
        <Errors errors={errors} />
      </div>
    </form>
  )
}
