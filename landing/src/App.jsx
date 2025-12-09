import Navbar from './components/Navbar'
import Hero from './components/Hero'
import SimulationShowcase from './components/SimulationShowcase'
import ProductSection from './components/ProductSection'
import HowItWorks from './components/HowItWorks'
import ForFounders from './components/ForFounders'
import PricingSection from './components/PricingSection'
import TrustSection from './components/TrustSection'
import Footer from './components/Footer'

function App() {
  return (
    <div className="min-h-screen bg-white">
      <Navbar />
      <Hero />
      <SimulationShowcase />
      <ProductSection />
      <HowItWorks />
      <ForFounders />
      <PricingSection />
      <TrustSection />
      <Footer />
    </div>
  )
}

export default App

