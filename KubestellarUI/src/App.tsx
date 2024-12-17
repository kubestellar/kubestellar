import './App.css'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Clusters from './components/Clusters'
import Navbar from './components/Navbar'
import ITS from './pages/ITS'
import WDS from './pages/WDS'
import BindingPolicies from './pages/BindingPolicies'

function App() {
  return (
    <Router>
      <div className="w-full min-h-screen bg-base-100">
        <Navbar />
        <Routes>
          <Route path="/" element={<Clusters />} />
          <Route path="/its" element={<ITS />} />
          <Route path="/wds" element={<WDS />} />
          <Route path="/binding-policies" element={<BindingPolicies />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App
