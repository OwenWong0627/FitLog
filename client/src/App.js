import React from 'react'
import { BrowserRouter as Router, Route, } from 'react-router-dom'
import Register from './pages/Register'
import Login from './pages/Login'
import './App.css'
import Workouts from './pages/Workouts'

function App() {
  return (
    <Router>
      <Route exact path="/" component={Workouts} />
      <Route path="/register" component={Register} />
      <Route path="/login" component={Login} />
      <Route path="/workouts" component={Workouts} />
    </Router>
  )
}

export default App
