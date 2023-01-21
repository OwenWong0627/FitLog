import React, { useState } from 'react'
import { Endpoints } from '../api/endpoints'
import Errors from '../components/Errors'

export default ({ history }) => {
  const [user, setUser] = useState({
    username: '',
    otp: '',
  })

  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errors, setErrors] = useState([])
  const { username, otp } = user

  const handleChange = (e) =>
    setUser({ ...user, [e.target.name]: e.target.value })

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      setIsSubmitting(true)
      const res = await fetch(Endpoints.otp, {
        method: 'POST',
        body: JSON.stringify({
          username,
          otp,
        }),
        headers: {
          'Content-Type': 'application/json',
        },
      })
      const { success, errors = [] } = await res.json()

      if (success) history.push('/login')

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
        <h1>Please Check Your Email and Verify your Account to gain access</h1>
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
          type="otp"
          placeholder="OTP"
          value={otp}
          name="otp"
          onChange={handleChange}
          required
        />

        <button disabled={isSubmitting} onClick={handleSubmit}>
          {isSubmitting ? '.....' : 'Verify'}
        </button>
        <br />
        <a href="/login">{'login'}</a>
        <br />
        <Errors errors={errors} />
      </div>
    </form>
  )
}
